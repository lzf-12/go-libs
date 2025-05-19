package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"sync"
)

func main() {

	mode := flag.String("mode", "rest", "available mode: rest | graphql | all")
	flag.Parse()

	serverMode := strings.ToLower(*mode)

	var wg sync.WaitGroup
	wg.Add(1)

	switch serverMode {
	case "rest":
		ServeRest()
	case "graphql":
		ServeGraphql()
	case "all":
		ServeRest()
		ServeGraphql()
	default:
		log.Printf("%s. is invalid mode. valid mode are: rest, graphql, all", serverMode)
		os.Exit(1)
	}

	wg.Wait()
}
