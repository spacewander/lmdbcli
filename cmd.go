package main

import (
	"fmt"
	"strings"

	"github.com/bmatsuo/lmdb-go/lmdb"
)

type command struct {
	MinArgc    int
	Call       func(txn *lmdb.Txn, dbi *lmdb.DBI, args []string) (res interface{}, err error)
	needUpdate bool
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

var cmdMap = map[string]command{
	"get":  {1, get, false},
	"set":  {2, put, true},
	"put":  {2, put, true},
	"stat": {0, stat, false},
}

// Exec run given cmd with args
func Exec(cmdName string, args ...string) (res interface{}, err error) {
	cmd, ok := cmdMap[strings.ToLower(cmdName)]
	if !ok {
		return nil, nil
	}

	if cmd.MinArgc == 0 && len(args) == 0 {
		return cmd.Call(nil, nil, args)
	}

	if len(args) < cmd.MinArgc {
		return nil, fmt.Errorf(
			"wrong number of arguments for '%s' command", cmdName)
	}

	keyStartPos := len(args) - cmd.MinArgc

	var run_txn func(op lmdb.TxnOp) (err error)
	if cmd.needUpdate {
		run_txn = Env.Update
	} else {
		run_txn = Env.View
	}

	err = run_txn(func(txn *lmdb.Txn) (err error) {
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

		res, err = cmd.Call(txn, &dbi, args[keyStartPos:])
		return err
	})

	return res, err
}
