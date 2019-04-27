package main

import (
	"encoding/base32"
	"strings"

	"github.com/Shopify/go-lua"
	"github.com/spacewander/lmdb-go/lmdb"
)

var (
	vm     *lua.State
	curCmd string
)

func init() {
	vm = lua.NewState()
	lua.OpenLibraries(vm)
	injectAPI(vm)
}

func injectAPI(L *lua.State) {
	L.CreateTable(0, 1)

	L.CreateTable(0, 1)
	L.PushGoFunction(dispatchCmd)
	L.SetField(-2, "__index")
	L.SetMetaTable(-2)

	L.CreateTable(0, 1)
	L.PushGoFunction(toHex)
	L.SetField(-2, "to_hex")
	L.SetField(-2, "utils")

	L.Global("package")
	L.Field(-1, "loaded")
	L.PushValue(-3)
	L.SetField(-2, "lmdb")
	L.Pop(2)
	L.SetGlobal("lmdb")
}

// Exported utils functions start ...
// Once you adds a util function, please update the README to document it.

func toHex(L *lua.State) int {
	nargs := L.Top()
	if nargs < 1 {
		L.PushString("at least one argument expected")
		L.Error()
	}

	var s string
	var ok bool
	if s, ok = lua.ToStringMeta(L, 1); !ok {
		L.PushString("arg 1 must be a string")
		L.Error()
	}

	L.PushString(base32.StdEncoding.EncodeToString([]byte(s)))
	return 1
}

// Exported utils functions end

func dispatchCmd(L *lua.State) int {
	if s, ok := lua.ToStringMeta(L, 2); ok {
		s = strings.ToLower(s)
		_, ok = CmdMap[s]
		if ok {
			curCmd = s
			L.PushGoFunction(execCmdInLuaScript)
			return 1
		}
	}
	// it is equal to return nil
	return 0
}

func pushList(L *lua.State, res []string) {
	L.CreateTable(len(res), 0)
	for i, s := range res {
		L.PushString(s)
		L.RawSetInt(-2, i+1)
	}
}

func pushStat(L *lua.State, stat *lmdb.Stat) {
	L.CreateTable(0, 6)
	L.PushString("branch_pages")
	L.PushNumber(float64(stat.BranchPages))
	L.RawSet(-3)

	L.PushString("depth")
	L.PushUnsigned(stat.Depth)
	L.RawSet(-3)

	L.PushString("entries")
	L.PushNumber(float64(stat.Entries))
	L.RawSet(-3)

	L.PushString("leaf_pages")
	L.PushNumber(float64(stat.LeafPages))
	L.RawSet(-3)

	L.PushString("overflow_pages")
	L.PushNumber(float64(stat.OverflowPages))
	L.RawSet(-3)

	L.PushString("psize")
	L.PushUnsigned(stat.PSize)
	L.RawSet(-3)
}

func execCmdInLuaScript(L *lua.State) int {
	args := []string{}
	nargs := L.Top()
	for i := 1; i <= nargs; i++ {
		luaType := L.TypeOf(i)
		switch luaType {
		case lua.TypeNumber:
			fallthrough
		case lua.TypeString:
			if s, ok := lua.ToStringMeta(L, i); ok {
				args = append(args, s)
			}
		default:
			// arg x is one based, like other stuff in Lua land
			L.PushFString("The type of arg %d is incorrect, only number and string are acceptable", i)
			L.Error()
		}
	}

	res, err := Exec(curCmd, args...)
	if err != nil {
		L.PushNil()
		// standardize common error message to make error detection easier
		if lmdb.IsNotFound(err) {
			L.PushString("not found")
		} else if lmdb.IsErrno(err, lmdb.Incompatible) {
			L.PushString("incompatible operation")
		} else {
			L.PushString(err.Error())
		}
		return 2
	}
	switch res := res.(type) {
	case bool:
		L.PushBoolean(res)
	case string:
		L.PushString(res)
	case int:
		L.PushNumber(float64(res))
	case *lmdb.Stat:
		pushStat(L, res)
	case []string:
		pushList(L, res)
	default:
		L.PushFString("The type of result returns from command '%s' is unsupported",
			curCmd)
		L.Error()
	}
	return 1
}

// StartScript evals given script file
func StartScript(script string) error {
	return lua.DoFile(vm, script)
}
