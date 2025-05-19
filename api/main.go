package main

import (
	"api/graphql"
	"api/grpc"
	"api/rest"
	"flag"
	"log"
	"os"
	"strings"
	"sync"
)

func main() {

	mode := flag.String("mode", "rest", "available mode: rest | graphql | grpc")
	flag.Parse()

	serverMode := strings.ToLower(*mode)

	var wg sync.WaitGroup
	wg.Add(1)

	switch serverMode {
	case "rest":
		rest.ServeRest()
	case "graphql":
		graphql.ServeGraphql()
	case "grpc":
		grpc.ServeGrpc()
	default:
		log.Printf("%s. is invalid mode. valid mode are: rest, graphql, all", serverMode)
		os.Exit(1)
	}

	wg.Wait()
}
