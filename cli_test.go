package main

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	GetNotFound = "mdb_get: MDB_NOTFOUND: No matching key/data pair found"
	DelNotFound = "mdb_del: MDB_NOTFOUND: No matching key/data pair found"
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
	assert.Equal(suite.T(), "", ExecCmdInCli(""))

	assert.Equal(suite.T(), "unknown command 'non-exist'", ExecCmdInCli("non-exist cmd"))
	// Ignore the case of command name
	assert.Equal(suite.T(), "OK", ExecCmdInCli("SET key value"))
}

func (suite *CmdSuite) TestGet() {
	assert.Equal(suite.T(), "wrong number of arguments for 'get' command", ExecCmdInCli("get"))
	assert.Equal(suite.T(), GetNotFound, ExecCmdInCli("get key"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set key value"))
	assert.Equal(suite.T(), `"value"`, ExecCmdInCli("get key"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set db sub_db key value2"))
	assert.Equal(suite.T(), GetNotFound, ExecCmdInCli("get db non-exist"))
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

	assert.NotEqual(suite.T(), GetNotFound, ExecCmdInCli("stat db"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("set db key value"))
	assert.Regexp(suite.T(), regexp.MustCompile("branch pages"), ExecCmdInCli("stat db"))
}

func (suite *CmdSuite) TestHelp() {
	assert.Equal(suite.T(), len(CmdMap), len(strings.Split(ExecCmdInCli("help"), "\n")))
}

func (suite *CmdSuite) TestExists() {
	assert.Equal(suite.T(), "false", ExecCmdInCli("exists key"))
	assert.Equal(suite.T(), "wrong number of arguments for 'exists' command", ExecCmdInCli("exists"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set key value"))
	assert.Equal(suite.T(), "true", ExecCmdInCli("exists key"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set db key value"))
	assert.Equal(suite.T(), "true", ExecCmdInCli("exists db key"))
	assert.Equal(suite.T(), "true", ExecCmdInCli("exists db"))
}

func (suite *CmdSuite) TestDel() {
	assert.Equal(suite.T(), DelNotFound, ExecCmdInCli("del key"))
	assert.Equal(suite.T(), "wrong number of arguments for 'del' command", ExecCmdInCli("del"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set key value"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("del key"))
	assert.Equal(suite.T(), GetNotFound, ExecCmdInCli("get key"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set db key value"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("del db key"))
	assert.Equal(suite.T(), GetNotFound, ExecCmdInCli("get db key"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("del db"))
	assert.Equal(suite.T(), GetNotFound, ExecCmdInCli("get db"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set db sub_db key value"))
	// There is a limitation to delete middle database
	//assert.Equal(suite.T(), "OK", ExecCmdInCli("del db sub_db"))
	assert.Equal(suite.T(), GetNotFound, ExecCmdInCli("get db sub_db"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("del db"))
	assert.Equal(suite.T(), GetNotFound, ExecCmdInCli("get db"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set db sub_db key value"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("del db"))
	assert.Equal(suite.T(), GetNotFound, ExecCmdInCli("get db"))
}

func (suite *CmdSuite) TestKeys() {
	assert.Equal(suite.T(), "", ExecCmdInCli("keys key"))
	assert.Equal(suite.T(), "wrong number of arguments for 'keys' command", ExecCmdInCli("keys"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set key value"))
	assert.Equal(suite.T(), `1) "key"`, ExecCmdInCli("keys k*"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set key2 value"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("set k value"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("set y value"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("set ky value"))
	assert.Equal(suite.T(), `1) "key"`, ExecCmdInCli("keys key"))
	assert.Equal(suite.T(), `1) "ky"`, ExecCmdInCli("keys k[ey]"))
	assert.Equal(suite.T(), `1) "key"`, ExecCmdInCli("keys k?y"))
	assert.Equal(suite.T(), `1) "key"`+"\n"+`2) "key2"`, ExecCmdInCli("keys ke*"))
	assert.Equal(suite.T(), `1) "key"`+"\n"+`2) "ky"`+"\n"+`3) "y"`,
		ExecCmdInCli("keys *y"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set db key value"))
	assert.Equal(suite.T(), `1) "key"`, ExecCmdInCli("keys db k*"))
	assert.Equal(suite.T(), ``, ExecCmdInCli("keys db y*"))
}

func (suite *CmdSuite) TestCount() {
	assert.Equal(suite.T(), "0", ExecCmdInCli("count key"))
	assert.Equal(suite.T(), "wrong number of arguments for 'count' command", ExecCmdInCli("count"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set key value"))
	assert.Equal(suite.T(), "1", ExecCmdInCli("count k*"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set key2 value"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("set k value"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("set y value"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("set ky value"))
	assert.Equal(suite.T(), "1", ExecCmdInCli("count key"))
	assert.Equal(suite.T(), "2", ExecCmdInCli("count ke*"))
	assert.Equal(suite.T(), "3", ExecCmdInCli("count *y"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set db key value"))
	assert.Equal(suite.T(), "1", ExecCmdInCli("count db k*"))
	assert.Equal(suite.T(), "0", ExecCmdInCli("count db y*"))
}

func (suite *CmdSuite) TestValues() {
	assert.Equal(suite.T(), "", ExecCmdInCli("values key"))
	assert.Equal(suite.T(), "wrong number of arguments for 'values' command", ExecCmdInCli("values"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set key value"))
	assert.Equal(suite.T(), `1) "key" "value"`, ExecCmdInCli("values k*"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set key2 key2"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("set k k"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("set y y"))
	assert.Equal(suite.T(), "OK", ExecCmdInCli("set ky ky"))
	assert.Equal(suite.T(), `1) "key" "value"`, ExecCmdInCli("values key"))
	assert.Equal(suite.T(), `1) "ky" "ky"`, ExecCmdInCli("values k[ey]"))
	assert.Equal(suite.T(), `1) "key" "value"`, ExecCmdInCli("values k?y"))
	assert.Equal(suite.T(), `1) "key" "value"`+"\n"+`2) "key2" "key2"`, ExecCmdInCli("values ke*"))
	assert.Equal(suite.T(), `1) "key" "value"`+"\n"+`2) "ky" "ky"`+"\n"+`3) "y" "y"`,
		ExecCmdInCli("values *y"))

	assert.Equal(suite.T(), "OK", ExecCmdInCli("set db key value"))
	assert.Equal(suite.T(), `1) "key" "value"`, ExecCmdInCli("values db k*"))
	assert.Equal(suite.T(), ``, ExecCmdInCli("values db y*"))
}
