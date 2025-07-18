package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
)

// matches both "Step1Hex:" and "Step1 - Hex:"
var stepRegexp = regexp.MustCompile(`Step[12](?:\s*-)?\s*Hex[:\s]*([0-9a-fA-F]+)`) // matches Step1Hex or Step2Hex

func capture(cmd *exec.Cmd) ([]string, error) {
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%v: %s", err, out.String())
	}

	matches := stepRegexp.FindAllSubmatch(out.Bytes(), -1)
	if len(matches) != 2 {
		return nil, fmt.Errorf("expected 2 StepHex outputs, got %d. output=%s", len(matches), out.String())
	}

	return []string{string(matches[0][1]), string(matches[1][1])}, nil
}

func main() {
	goHex, err := capture(exec.Command("go", "run", "examples/txtest/go_runner/main.go"))
	if err != nil {
		fmt.Printf("Run go_runner failed: %v\n", err)
		return
	}
	fmt.Println("Go Step1Hex", goHex[0])
	fmt.Println("Go Step2Hex", goHex[1])

	tsHex, err := capture(exec.Command("bun", "run", "examples/txtest/ts_runner.ts"))
	if err != nil {
		fmt.Printf("Run ts_runner failed: %v\n", err)
		return
	}
	fmt.Println("TS Step1Hex", tsHex[0])
	fmt.Println("TS Step2Hex", tsHex[1])

	pass := true
	for i := 0; i < 2; i++ {
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
