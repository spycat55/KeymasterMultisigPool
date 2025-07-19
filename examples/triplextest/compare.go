package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
)

var stepRegexp = regexp.MustCompile(`Step[123](?:\s*-)?\s*Hex[:\s]*([0-9a-fA-F]+)`) // matches Step1/2/3

func capture(cmd *exec.Cmd) ([]string, error) {
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%v: %s", err, out.String())
	}
	matches := stepRegexp.FindAllSubmatch(out.Bytes(), -1)
	if len(matches) != 3 {
		return nil, fmt.Errorf("expected 3 StepHex outputs, got %d. output=%s", len(matches), out.String())
	}
	hexes := make([]string, 3)
	for i, m := range matches {
		hexes[i] = string(m[1])
	}
	return hexes, nil
}

func main() {
	goHex, err := capture(exec.Command("go", "run", "examples/triplextest/go_runner/main.go"))
	if err != nil {
		fmt.Printf("Run go_runner failed: %v\n", err)
		return
	}

	tsHex, err := capture(exec.Command("bun", "run", "examples/triplextest/ts_runner_refactor.ts"))
	if err != nil {
		fmt.Printf("Run ts_runner failed: %v\n", err)
		return
	}

	// Print captured hexes for reference
	for i := 0; i < 3; i++ {
		fmt.Printf("Go Step%dHex: %s\n", i+1, goHex[i])
		fmt.Printf("TS Step%dHex: %s\n", i+1, tsHex[i])
	}

	pass := true
	for i := 0; i < 3; i++ {
		if goHex[i] != tsHex[i] {
			fmt.Printf("Mismatch at Step%d\n", i+1)
			pass = false
		}
	}
	if pass {
		fmt.Println("PASS: Go and TS transactions are identical")
	} else {
		fmt.Println("FAIL: Transactions differ")
	}
}
