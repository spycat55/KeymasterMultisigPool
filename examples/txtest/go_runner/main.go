package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	ce "github.com/spycat55/KeymasterMultisigPool/pkg/dual_endpoint"
	libs "github.com/spycat55/KeymasterMultisigPool/pkg/libs"

	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
)

type Fixture struct {
	ClientPrivHex string      `json:"clientPrivHex"`
	ServerPrivHex string      `json:"serverPrivHex"`
	ClientUtxos   []libs.UTXO `json:"clientUtxos"`
	EndHeight     uint32      `json:"endHeight"`
	FeeRate       float64     `json:"feeRate"`
	IsMain        bool        `json:"isMain"`
}

func loadFixture() Fixture {
	// 获取当前源文件所在目录
	_, srcPath, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatalf("cannot determine caller path")
	}

	dir := filepath.Dir(srcPath) // .../tests/txtest/go_runner
	// 向上查找 fixture.json
	var fixturePath string
	for {
		candidate := filepath.Join(dir, "fixture.json")
		if _, err := os.Stat(candidate); err == nil {
			fixturePath = candidate
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // 到根目录
		}
		dir = parent
	}

	// 最后退而求其次：尝试当前工作目录
	if fixturePath == "" {
		cwd, _ := os.Getwd()
		fixturePath = filepath.Join(cwd, "fixture.json")
	}

	data, err := os.ReadFile(fixturePath)
	if err != nil {
		log.Fatalf("read fixture (%s): %v", fixturePath, err)
	}
	var f Fixture
	if err := json.Unmarshal(data, &f); err != nil {
		log.Fatalf("unmarshal fixture: %v", err)
	}
	return f
}

// helper to save produced utxo json for next round
func saveNewUTXO(newUtxo libs.UTXO) {
	out, _ := json.MarshalIndent(newUtxo, "", "  ")
	fmt.Printf("NEW_UTXO: %s\n", string(out))
}

func main() {
	f := loadFixture()

	clientPriv, _ := ec.PrivateKeyFromHex(f.ClientPrivHex)
	serverPriv, _ := ec.PrivateKeyFromHex(f.ServerPrivHex)

	// Step1 Base tx
	res1, err := ce.BuildDualFeePoolBaseTx(&f.ClientUtxos, clientPriv, serverPriv.PubKey(), f.IsMain, f.FeeRate)
	if err != nil {
		log.Fatalf("step1: %v", err)
	}

	// Step2 Spend tx
	const serverAmount uint64 = 100
	bTx, clientSignBytes, amount, err := ce.BuildDualFeePoolSpendTX(res1.Tx, res1.Amount, serverAmount, f.EndHeight, clientPriv, serverPriv.PubKey(), f.IsMain, f.FeeRate)
	if err != nil {
		log.Fatalf("step2: %v", err)
	}

	fmt.Printf("Step1Hex %s\n", res1.Tx.String())
	fmt.Printf("Step2Hex %x\n", *clientSignBytes)

	// Step3 Server sign
	serverSignBytes, err := ce.SpendTXServerSign(bTx, res1.Amount, serverPriv, clientPriv.PubKey())
	if err != nil {
		log.Fatalf("step3: %v", err)
	}

	fmt.Printf("Step3Hex %x\n", *serverSignBytes)

	// Step4 Client update and re-sign
	const newServerAmount uint64 = 150 // 修改服务器金额
	const newSequenceNumber uint32 = 2 // 修改序列号

	updatedTx, err := ce.LoadTx(bTx.String(), nil, newSequenceNumber, newServerAmount, serverPriv.PubKey(), clientPriv.PubKey(), res1.Amount)
	if err != nil {
		log.Fatalf("step4 load tx: %v", err)
	}

	clientUpdateSignBytes, err := ce.ClientDualFeePoolSpendTXUpdateSign(updatedTx, clientPriv, serverPriv.PubKey())
	if err != nil {
		log.Fatalf("step4 client sign: %v", err)
	}

	fmt.Printf("Step4Hex %x\n", *clientUpdateSignBytes)

	// Step5 Server update sign
	serverUpdateSignBytes, err := ce.ServerDualFeePoolSpendTXUpdateSign(updatedTx, serverPriv, clientPriv.PubKey())
	if err != nil {
		log.Fatalf("step5 server sign: %v", err)
	}

	fmt.Printf("Step5Hex %x\n", *serverUpdateSignBytes)
	_ = amount
	// _ = clientSignBytes

	// // new utxo comes from bTx output[1] (client P2PKH)
	// newUtxo := libs.UTXO{
	// 	TxID:  bTx.TxID().String(),
	// 	Vout:  1,
	// 	Value: amount,
	// }
	// saveNewUTXO(newUtxo)
}
