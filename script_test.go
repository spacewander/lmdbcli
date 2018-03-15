package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvalLuaScript(t *testing.T) {
	tmpdir, _ := ioutil.TempDir("", "lmdbcli")
	initDB(tmpdir)

	err := StartScript("test.lua")
	assert.Nil(t, err)

	Env.Close()
	os.Remove(tmpdir)
}
