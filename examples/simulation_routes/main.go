// Offline: go run ./examples/simulation_routes saved-evidence.json [case-name]
package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/0xfnzero/sol-parser-sdk-golang/solparser"
)

type evidence struct {
	Name     *string         `json:"name"`
	Wire     string          `json:"wire"`
	Response json.RawMessage `json:"response"`
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("provide saved simulation evidence JSON")
	}
	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		return err
	}
	var corpus struct {
		Cases []evidence `json:"cases"`
	}
	if err = json.Unmarshal(raw, &corpus); err != nil {
		return err
	}
	if corpus.Cases == nil {
		var c evidence
		if err = json.Unmarshal(raw, &c); err != nil {
			return err
		}
		corpus.Cases = []evidence{c}
	}
	count := 0
	for _, c := range corpus.Cases {
		if len(os.Args) > 2 && (c.Name == nil || *c.Name != os.Args[2]) {
			continue
		}
		wire, err := base64.StdEncoding.DecodeString(c.Wire)
		if err != nil {
			return err
		}
		route, err := solparser.AnalyzeSimulationRoutes(wire, c.Response, nil)
		if err != nil {
			return err
		}
		if err = json.NewEncoder(os.Stdout).Encode(map[string]any{"name": c.Name, "route": route}); err != nil {
			return err
		}
		count++
	}
	if count == 0 {
		return fmt.Errorf("simulation case not found")
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
