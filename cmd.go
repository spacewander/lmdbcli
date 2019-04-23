package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/gobwas/glob"
	"github.com/spacewander/lmdb-go/lmdb"
)

var (
	globSpecialChar = regexp.MustCompile(`[*\[{}?\\]`)
)

type command struct {
	minArgc    int
	call       func(txn *lmdb.Txn, dbi *lmdb.DBI, args []string) (res interface{}, err error)
	needUpdate bool
	tip        string
}

func get(txn *lmdb.Txn, dbi *lmdb.DBI, args []string) (res interface{}, err error) {
	v, err := txn.Get(*dbi, []byte(args[0]))
	if err != nil {
		return nil, err
	}
	return string(v), nil
}

func put(txn *lmdb.Txn, dbi *lmdb.DBI, args []string) (res interface{}, err error) {
	err = txn.Put(*dbi, []byte(args[0]), []byte(args[1]), 0)
	if err != nil {
		return nil, err
	}
	return true, nil
}

func stat(txn *lmdb.Txn, dbi *lmdb.DBI, args []string) (res interface{}, err error) {
	if txn == nil {
		return Env.Stat()
	}
	return txn.Stat(*dbi)
}

func exists(txn *lmdb.Txn, dbi *lmdb.DBI, args []string) (res interface{}, err error) {
	_, err = txn.Get(*dbi, []byte(args[0]))
	if err != nil {
		if lmdb.IsNotFound(err) {
			return false, nil
		}
		return nil, err
	}
	return true, nil
}

func del(txn *lmdb.Txn, dbi *lmdb.DBI, args []string) (res interface{}, err error) {
	v, err := txn.Get(*dbi, []byte(args[0]))
	if err != nil {
		if !lmdb.IsNotFound(err) {
			return nil, err
		}
	}

	err = txn.Del(*dbi, []byte(args[0]), v)
	if err != nil {
		return nil, err
	}
	return true, nil
}

func keysHelper(txn *lmdb.Txn, dbi *lmdb.DBI, args []string, countOnly bool) ([]string, int, error) {
	byteKey := []byte(args[0])
	globIdx := globSpecialChar.FindIndex(byteKey)
	var prefix []byte
	if globIdx == nil {
		prefix = byteKey
	} else if globIdx[0] > 0 {
		prefix = byteKey[:globIdx[0]]
	}

	pattern, err := glob.Compile(args[0])
	if err != nil {
		return nil, 0, err
	}

	cur, err := txn.OpenCursor(*dbi)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close()

	var count int
	var res []string

	if !countOnly {
		res = []string{}
	}

	rangeFound := prefix == nil
	for {
		var k []byte
		if !rangeFound {
			k, _, err = cur.Get(prefix, nil, lmdb.SetRange)
			rangeFound = true
		} else {
			k, _, err = cur.Get(nil, nil, lmdb.Next)
		}

		if lmdb.IsNotFound(err) {
			return res, count, nil
		}
		if err != nil {
			return nil, 0, err
		}

		if !bytes.HasPrefix(k, prefix) {
			return res, count, nil
		}

		s := string(k)
		if pattern.Match(s) {
			if countOnly {
				count++
			} else {
				res = append(res, s)
			}
		}
	}
}

func keys(txn *lmdb.Txn, dbi *lmdb.DBI, args []string) (res interface{}, err error) {
	res, _, err = keysHelper(txn, dbi, args, false)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func count(txn *lmdb.Txn, dbi *lmdb.DBI, args []string) (res interface{}, err error) {
	_, res, err = keysHelper(txn, dbi, args, true)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// CmdMap holds commands used in all modes
var CmdMap = map[string]command{
	"get":  {1, get, false, "GET a value with 'get [db...] key'"},
	"set":  {2, put, true, "SET a value with 'set [db...] key'"},
	"put":  {2, put, true, "PUT is an alias of SET"},
	"stat": {0, stat, false, "STAT get mdb_stat with 'stat' or 'stat db'"},
	"exists": {1, exists, false,
		"EXISTS check if a key exists with 'exists [db...] key'"},
	"del": {1, del, true, "DEL remove a key with 'del [db...] key'"},
	"keys": {1, keys, false,
		"KEYS lists all keys matched given glob pattern with 'keys [db...] pattern'"},
	"count": {1, count, false,
		"COUNT works like KEYS but only returns the number of keys"},
}

// Exec run given cmd with args
func Exec(cmdName string, args ...string) (res interface{}, err error) {
	cmdName = strings.ToLower(cmdName)
	cmd, ok := CmdMap[cmdName]
	if !ok {
		return nil, nil
	}

	if cmd.minArgc == 0 && len(args) == 0 {
		return cmd.call(nil, nil, args)
	}

	if len(args) < cmd.minArgc {
		return nil, fmt.Errorf(
			"wrong number of arguments for '%s' command", cmdName)
	}

	keyStartPos := len(args) - cmd.minArgc

	var runTxn func(op lmdb.TxnOp) (err error)
	if cmd.needUpdate {
		runTxn = Env.Update
	} else {
		runTxn = Env.View
	}

	err = runTxn(func(txn *lmdb.Txn) (err error) {
		dbi, err := txn.OpenRoot(0)
		if err != nil {
			return err
		}

		for i := 0; i < keyStartPos; i++ {
			dbi, err = txn.OpenDBI(args[i], 0)
			if err != nil {
				if cmd.needUpdate && lmdb.IsNotFound(err) {
					dbi, err = txn.CreateDBI(args[i])
					if err == nil {
						continue
					}
				}
				return err
			}
		}

		res, err = cmd.call(txn, &dbi, args[keyStartPos:])
		return err
	})

	if cmdName == "del" && lmdb.IsErrno(err, lmdb.Incompatible) {
		// Is a database? Try again with drop in other txn.
		// The former txn may create new DBI.
		err = runTxn(func(txn *lmdb.Txn) (err error) {
			dbi, err := txn.OpenRoot(0)
			if err != nil {
				return err
			}

			for i := 0; i < len(args); i++ {
				dbi, err = txn.OpenDBI(args[i], 0)
				if err != nil {
					return err
				}
			}

			err = txn.Drop(dbi, true)
			if err == nil {
				res = true
			}
			return err
		})
	}

	return res, err
}
