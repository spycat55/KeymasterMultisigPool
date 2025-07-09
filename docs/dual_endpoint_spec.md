# Dual Endpoint Fee Pool – Cross-Language Specification

This document describes the canonical behaviour of the *Dual Endpoint* fee-pool implementation.  Both the Go code under `pkg/dual_endpoint` and the TypeScript code under `src/dual_endpoint` **MUST** follow this spec byte-for-byte so that transactions produced by either language are identical.

---

## 1. Terminology
| Term | Meaning |
|------|---------|
| **client** | The party that initiates the pool (provides the initial UTXOs). |
| **server** | The service that co-signs the multisig and pays the spending fee. |
| **A-Tx** | Base transaction produced by *Step 1* (`1base_tx`). |
| **B-Tx** | Spend transaction produced by *Step 2* (`2client_spend_tx`). |

---

## 2. Multisig Script
A **2-of-2** multisig script is used in all pool outputs.

```
OP_2 <serverPub> <clientPub> OP_2 OP_CHECKMULTISIG
```

*Ordering* of public keys **MUST** be `[serverPub, clientPub]`.

---

## 3. Network Parameter
All functions take
```
isMain: boolean  // true → mainnet; false → testnet
```
When `isMain` is omitted (TS originally), default to `true`.  The two SDKs must encode addresses identically.

---

## 4. Step 1 – buildDualFeePoolBaseTx / BuildDualFeePoolBaseTx
```
inputs:
    clientUtxos     []UTXO      // full value is consumed
    clientPrivKey   PrivateKey
    serverPubKey    PublicKey
    isMain          bool
    feeRate         float64     // sat/byte
returns (struct):
    tx      Transaction
    amount  uint64   // multisig output value (= Σutxo - fee)
    index   int      // always 0
```
Algorithm (identical in both languages):
1. Derive `clientAddress` from `clientPrivKey`.
2. For each UTXO produce P2PKH input with temporary unlocking script.
3. Create multisig output with value `Σ(utxo)`.
4. Apply fake signatures to every input → measure `txSize`.
5. `fee = max(1, floor(txSize/1000 * feeRate))`.
6. `output[0].satoshis = Σ(utxo) - fee`.
7. Re-sign every input.

---

## 5. Step 2 – buildDualFeePoolSpendTX / BuildDualFeePoolSpendTX
```
inputs:
    aTx             Transaction  // output of step 1
    serverValue     uint64       // same as aTx.outputs[0].satoshis
    endHeight       uint32       // lock-up height
    clientPrivKey   PrivateKey
    serverPubKey    PublicKey
    isMain          bool
    feeRate         float64
returns (struct):
    tx              Transaction // B-Tx
    clientSignBytes []byte      // DER + sighash byte
    amount          uint64      // transferred to client (= serverValue - fee)
```
Sub-Steps:
1. **SubBuildDualFeePoolSpendTX** – produces unsigned B-Tx with fake unlock for size calc.
2. **SpendTXDualFeePoolClientSign** – generates client’s signature bytes.
3. **BuildDualFeePoolSpendTX** – wraps the above.

Fee calculation identical to Step 1, paid by `serverValue`.

---

## 6. Placeholder / Fake Signature
For size estimation the following rules apply:
* Unlock script starts with `OP_0` (due to off-by-one bug).
* Each fake signature is **73 bytes** (72 dummy + 1 sighash).
* Script structure is therefore:
  `OP_0 <fakeSig> <fakeSig>`.

Both languages must use the same fake signature routine to avoid size drift.

---

## 7. Return Types
Go and TS should both expose small structs/objects with identical field names (`tx`, `amount`, `index`, `clientSignBytes`).  This simplifies cross-lang testing.

---

## 8. Logging
Production builds must not emit logs by default.  Provide optional debug flag.

---

## 9. Future Steps – Server Sign & Updates
The Go implementation includes files `3server_sign.go`, `4client_spend_tx_update.go`, `5server_sign_update.go` that finish the protocol.  TypeScript must implement the same APIs with byte-level identical behaviour.  Details will be added here once ported.

---

## 10. Test Vectors
A shared set of JSON fixtures will be placed in `tests/dual_endpoint/*.json` covering:
1. Deterministic keys & UTXOs for Step 1.
2. Expected hex of A-Tx.
3. Expected hex of B-Tx after client sign.

Both runtimes must pass these tests.

---

*Last updated*: 2025-07-09
