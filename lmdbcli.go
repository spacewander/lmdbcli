package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bmatsuo/lmdb-go/lmdb"
)

var (
	// Env is the global environment inited at the beginning
	Env *lmdb.Env

	shouldPrintVersion = flag.Bool("version", false, "Output version and exit.")
	version            = "1.0.0"
)

func initDB(dbPath string) {
	env, err := lmdb.NewEnv()
	if err != nil {
		log.Fatal("Could not create lmdb environment")
	}

	var maxDbs = 256
	err = env.SetMaxDBs(maxDbs)
	if err != nil {
		log.Fatalf("Could not set max dbs to %v: %v", maxDbs, err)
	}

	var mapSize int64 = 1073741824 // 1GB
	err = env.SetMapSize(mapSize)
	if err != nil {
		log.Fatalf("Could not set map size to %v: %v", mapSize, err)
	}

	err = env.Open(dbPath, 0, 0644)
	if err != nil {
		log.Fatalf("Could not open %s: %v", dbPath, err)
	}
	Env = env
}

func printVersion() {
	fmt.Printf("lmdbcli %s\n", version)
}

func main() {
	flag.Parse()
	if *shouldPrintVersion {
		printVersion()
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		log.Fatalf("database filename is required.")
	}

	initDB(flag.Arg(0))
	defer Env.Close()
	StartCli()
}
