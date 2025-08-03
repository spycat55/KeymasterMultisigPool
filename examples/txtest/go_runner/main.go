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

	fmt.Printf("Step1 - Hex: %s\n", res1.Tx.String())
	// 在客户端收到 server 私钥的环境下，完成多签输入脚本
	// sigHash := sighash.Flag(sighash.ForkID | sighash.All)
	// ms, err := libs.Unlock([]*ec.PrivateKey{serverPriv, clientPriv}, []*ec.PublicKey{serverPriv.PubKey(), clientPriv.PubKey()}, 2, &sigHash)
	// if err != nil {
	// 	log.Fatalf("unlock template: %v", err)
	// }
	// unlockScript, err := ms.Sign(bTx, 0)
	// if err != nil {
	// 	log.Fatalf("final sign: %v", err)
	// }
	// bTx.Inputs[0].UnlockingScript = unlockScript

	fmt.Printf("Step2 - Hex: %s\n", bTx.String())
	_ = clientSignBytes
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
