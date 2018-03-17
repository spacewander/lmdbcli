![screenshot](./screenshot.png)

[![Travis](https://travis-ci.org/spacewander/lmdbcli.svg?branch=master)](https://travis-ci.org/spacewander/lmdbcli)
[![GoReportCard](http://goreportcard.com/badge/spacewander/lmdbcli)](http://goreportcard.com/report/spacewander/lmdbcli)
[![codecov.io](https://codecov.io/github/spacewander/lmdbcli/coverage.svg?branch=master)](https://codecov.io/github/spacewander/lmdbcli?branch=master)
[![license](https://img.shields.io/badge/License-GPLv3-green.svg)](https://github.com/spacewander/lmdbcli/blob/master/LICENSE)

## Feature

* Support CRUD commands in repl-like command line. You can consider it as `redis-cli` for lmdb.
* You can eval Lua script with given database. It makes maintaining lmdb more easily.

Feel free to create an issue or pull request if you think this tool could be improved to satisfy
your requirements.

## Installation

`go get -u github.com/spacewander/lmdbcli`

Note that the lmdb binding is writtern via `cgo`.
If you get `GLIBC_XX symbol not found` error when running the binary,
you need to rebuild it in correspondent environment to fix the dynamic symbol.

## Usage

`lmdbcli  [-e script] /path/to/db`

## Commands

```
$ ./lmdbcli /tmp/node
node> help
stat) STAT get mdb_stat with 'stat' or 'stat db'
exists) EXISTS check if a key exists with 'exists [db...] key'
del) DEL remove a key with 'del [db...] key'
keys) KEYS lists all keys matched given glob pattern with 'keys [db...] pattern'
get) GET a value with 'get [db...] key'
set) SET a value with 'set [db...] key'
put) PUT is an alias of SET
node>
```

## Lua support

You could run a lua script on specific database like this: `lmdbcli -e your.lua db_path`.
`lmdbcli` provides a couple of API within the global variable `lmdb`. For example:
```lua
lmdb.get("key") -- return the value of `key` as a lua string
-- is equal to `> get key` in the command line
```

See [test.lua](./test.lua) as a concrete example.
