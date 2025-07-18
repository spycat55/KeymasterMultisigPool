# Transaction Cross-Test (`examples/txtest`)

This directory contains a small integration test that guarantees the **Go** and **TypeScript** implementations of the Keymaster multisig pool create **bit-identical transactions** from the same fixture data.

## Contents

| File | Role |
| ---- | ---- |
| `fixture.json` | Input data (UTXOs, keys, amounts) shared by both implementations. |
| `go_runner/main.go` | Builds **Step 1** (fund) and **Step 2** (spend) transactions using the Go SDK and prints their hex. |
| `ts_runner.ts` | Does the same in TypeScript using `@bsv/sdk`. |
| `compare.go` | Runs the two runners, extracts their hex strings and asserts they are identical. Prints **PASS/FAIL**. |

## Prerequisites

* Go 1.20+
* [Bun](https://bun.sh/) ≥ 1.0 (for running TypeScript)
* Node dependencies installed (`bun install` at project root)

## Running the test

From the project root:

```bash
# 1. Generate TS transactions (optional – useful while debugging)
bun run examples/txtest/ts_runner.ts

# 2. Generate Go transactions (optional)
go run examples/txtest/go_runner/main.go

# 3. One-shot comparison (recommended)
go run examples/txtest/compare.go
```

Expected output on success:

```
PASS: Go and TS transactions are identical
```

If you see **FAIL**, the program will show the differing Step hex so you can diff them and locate the root cause (signature hash, script, fee, etc.).

## Debug Tips

* Both runners print the generated signature-hash preimage (`TS sighash:`) and its double-SHA256 digest (`sighash32`). Comment out these `console.log` / `fmt.Printf` lines or guard them with an environment variable when you no longer need verbose output.
* Keep `compare.go` in your CI pipeline to ensure future changes maintain byte-level parity between languages.
