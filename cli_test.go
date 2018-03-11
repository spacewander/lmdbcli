package main

import (
	"io/ioutil"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CmdSuite struct {
	suite.Suite
	dbPath string
}

func (suite *CmdSuite) SetupTest() {
	tmpdir, _ := ioutil.TempDir("", "lmdbcli")
	suite.dbPath = tmpdir
	initDB(suite.dbPath)
}

func TestCmdTestSuite(t *testing.T) {
	suite.Run(t, new(CmdSuite))
}

func (suite *CmdSuite) TearDownTest() {
	Env.Close()
	os.Remove(suite.dbPath)
}

func (suite *CmdSuite) TestExecCmdInCli() {
	assert.Equal(suite.T(), "unknown command 'non-exist'", ExecCmdInCli("non-exist cmd"))
	// Ignore the case of command name
	assert.Equal(suite.T(), "OK", ExecCmdInCli("SET key value"))
}

func (suite *CmdSuite) TestGet() {
	assert.Equal(suite.T(), "wrong number of arguments for 'get' command", ExecCmdInCli("get"))
	assert.Equal(suite.T(), "mdb_get: MDB_NOTFOUND: No matching key/data pair found", ExecCmdInCli("get key"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set key value"))
	assert.Equal(suite.T(), `"value"`, ExecCmdInCli("get key"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set db sub_db key value2"))
	assert.Equal(suite.T(), "mdb_get: MDB_NOTFOUND: No matching key/data pair found",
		ExecCmdInCli("get db non-exist"))
	assert.Equal(suite.T(), `"value2"`, ExecCmdInCli("get db sub_db key"))
}

func (suite *CmdSuite) TestSet() {
	assert.Equal(suite.T(), "wrong number of arguments for 'set' command",
		ExecCmdInCli("set db"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("set key value"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("set db key value"))
	assert.Equal(suite.T(), `"value"`, ExecCmdInCli("get db key"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("set db sub_db key value"))
	assert.Equal(suite.T(), `"value"`, ExecCmdInCli("get db sub_db key"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("set db1 db2 key value"))
	assert.Equal(suite.T(), `"value"`, ExecCmdInCli("get db1 db2 key"))
}

func (suite *CmdSuite) TestStat() {
	assert.Regexp(suite.T(), regexp.MustCompile("branch pages"), ExecCmdInCli("stat"))

	assert.NotEqual(suite.T(), "mdb_get: MDB_NOTFOUND: No matching key/data pair found",
		ExecCmdInCli("stat db"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("set db key value"))
	assert.Regexp(suite.T(), regexp.MustCompile("branch pages"), ExecCmdInCli("stat db"))
}
