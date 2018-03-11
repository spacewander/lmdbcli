package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/bmatsuo/lmdb-go/lmdb"
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

func execCmdInCli(line string) string {
	fields := strings.Fields(strings.TrimSpace(line))
	res, err := Exec(fields[0], fields[1:]...)
	if err != nil {
		return err.Error()
	}

	switch res := res.(type) {
	case bool:
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

func StartCli() {
	println(execCmdInCli("stat"))
}
