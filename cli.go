package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/bmatsuo/lmdb-go/lmdb"
	"github.com/chzyer/readline"
)

// Format ["o1", "o2"] to string
// 1) "o1"\n
// 2) "o2"
func formatListToStr(list []string) string {
	paddingNum := strconv.Itoa(int(math.Log10(float64(len(list)))) + 1)
	padded := make([]string, len(list))
	for i, data := range list {
		padded[i] = fmt.Sprintf("%"+paddingNum+`d) "%s"`, i+1, data)
	}
	return strings.Join(padded, "\n")
}

func formatMapToStr(strMap map[string]string) string {
	formatted := make([]string, len(strMap))
	i := 0
	for k := range strMap {
		formatted[i] = fmt.Sprintf(`%s) %v`, k, strMap[k])
		i++
	}
	return strings.Join(formatted, "\n")
}

func formatStatToStr(stat *lmdb.Stat) string {
	formatted := make([]string, 6)
	formatted[0] = fmt.Sprintf(`branch pages) %v`, stat.BranchPages)
	formatted[1] = fmt.Sprintf(`depth) %v`, stat.Depth)
	formatted[2] = fmt.Sprintf(`entries) %v`, stat.Entries)
	formatted[3] = fmt.Sprintf(`leaf pages) %v`, stat.LeafPages)
	formatted[4] = fmt.Sprintf(`overflow pages) %v`, stat.OverflowPages)
	formatted[5] = fmt.Sprintf(`psize) %v`, stat.PSize)
	return strings.Join(formatted, "\n")
}

func help() string {
	tips := make(map[string]string, len(CmdMap))
	for name, cmd := range CmdMap {
		tips[name] = cmd.tip
	}
	return formatMapToStr(tips)
}

// ExecCmdInCli executes one command and returns the string result for
// displaying in terminal.
func ExecCmdInCli(line string) string {
	fields := strings.Fields(strings.TrimSpace(line))
	if fields[0] == "help" {
		return help()
	}

	res, err := Exec(fields[0], fields[1:]...)
	if err != nil {
		return err.Error()
	}

	if res == nil {
		return fmt.Sprintf("unknown command '%s'", fields[0])
	}

	switch res := res.(type) {
	case bool:
		if fields[0] == "exists" {
			return strconv.FormatBool(res)
		}
		return "OK"
	case string:
		return fmt.Sprintf("\"%s\"", res)
	case []string:
		return formatListToStr(res)
	case *lmdb.Stat:
		return formatStatToStr(res)
	default:
		panic(fmt.Sprintf(
			"The type of result returns from command '%s' is unsupported",
			fields[0]))
	}
}

func buildCompleter() readline.AutoCompleter {
	cmds := []readline.PrefixCompleterInterface{}
	for k := range CmdMap {
		cmds = append(cmds, readline.PcItem(k))
	}

	cmds = append(cmds, readline.PcItem("help"))
	return readline.NewPrefixCompleter(cmds...)
}

func getHomeDir() string {
	env := "HOME"
	if runtime.GOOS == "windows" {
		env = "USERPROFILE"
	}
	return os.Getenv(env)
}

// StartCli starts the cli mode
func StartCli() {
	historyFileDir := filepath.Join(getHomeDir(), ".cache")
	if _, err := os.Stat(historyFileDir); os.IsNotExist(err) {
		// simply ignore error since the history feature is optional.
		os.Mkdir(historyFileDir, 0644)
	}
	l, err := readline.NewEx(&readline.Config{
		AutoComplete:    buildCompleter(),
		Prompt:          DbName + "> ",
		HistoryFile:     filepath.Join(historyFileDir, "lmdbclihistory"),
		HistoryLimit:    1000,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		println(ExecCmdInCli(line))
	}
}
