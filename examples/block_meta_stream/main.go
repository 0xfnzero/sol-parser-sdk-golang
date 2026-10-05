// Read one gRPC BlockMeta using GRPC_URL/GRPC_TOKEN. No RPC.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/0xfnzero/sol-parser-sdk-golang/solparser"
	"os"
	"time"
)

func run() error {
	if os.Getenv("GRPC_URL") == "" {
		return fmt.Errorf("set GRPC_URL")
	}
	client := solparser.NewYellowstoneGrpc(os.Getenv("GRPC_URL"))
	client.SetXToken(os.Getenv("GRPC_TOKEN"))
	defer client.Disconnect()
	sub, err := client.SubscribeDexEvents(nil, nil, solparser.EventTypeFilterIncludeOnly([]solparser.EventType{solparser.EventTypeBlockMeta}))
	if err != nil {
		return err
	}
	defer sub.Cancel()
	select {
	case event, ok := <-sub.Events:
		if !ok {
			return fmt.Errorf("event stream closed")
		}
		if event.Type != solparser.EventTypeBlockMeta {
			return fmt.Errorf("expected BlockMeta")
		}
		data, err := json.Marshal(event.GetMetadata())
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	case err, ok := <-sub.Errors:
		if !ok || err == nil {
			return fmt.Errorf("error stream closed")
		}
		return err
	case <-time.After(30 * time.Second):
		return fmt.Errorf("BlockMeta timeout")
	}
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
