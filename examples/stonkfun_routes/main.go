// go run ./examples/stonkfun_routes <getTransaction-base64.json>
package main

import (
	"encoding/json"
	"fmt"
	"github.com/0xfnzero/sol-parser-sdk-golang/solparser"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		panic("provide a saved compiled/base64 getTransaction response")
	}
	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	registry, err := solparser.NewStonkFunPoolRegistry(nil)
	if err != nil {
		panic(err)
	}
	if _, err = registry.ObserveRPCTransaction(raw); err != nil {
		panic(err)
	}
	route, err := solparser.AnalyzeRPCTransactionRoutes(raw, registry.VerifiedCpmmPools())
	if err != nil {
		panic(err)
	}
	result, err := json.MarshalIndent(route, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(result))
}
