package chain_utils

import (
	"encoding/hex"
	"fmt"

	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
	// "go.uber.org/zap"

	// primitives "github.com/bsv-blockchain/go-sdk/primitives/ec"

	libs "github.com/spycat55/KeymasterMultisigPool/pkg/libs"
	// multisig "github.com/spycat55/KeymasterMultisigPool/pkg/libs"

	tx "github.com/bsv-blockchain/go-sdk/transaction"
	sighash "github.com/bsv-blockchain/go-sdk/transaction/sighash"
	"github.com/bsv-blockchain/go-sdk/transaction/template/p2pkh"
)

// DualEndpoint Pool 双端费用池 dual_endpoint_pool

type BuildStep1Response struct {
	Tx     *tx.Transaction
	Amount uint64
	Index  int
}

// p2pkh to 2t2多签, 不找零
func BuildDualFeePoolBaseTx(
	clientUtxo *[]libs.UTXO, // 发起者 utxos, 我提供的金额就是这个 utxo 的全额
	// serverValue uint64, // 服务器提供金额
	clientPrivateKey *ec.PrivateKey,
	serverPublicKey *ec.PublicKey,
	isMain bool,
	feeRate float64,
) (*BuildStep1Response, error) {
	clientAddress, err := libs.GetAddressFromPubKey(clientPrivateKey.PubKey(), isMain)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	// 获取发起者的地址
	clientPublicKey := clientPrivateKey.PubKey()

	// aAddress, err := script.NewAddressFromPublicKey(clientPublicKey, isMain)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to generate address from public key: %w", err)
	// }

	// 创建交易对象
	transactionData := tx.NewTransaction()

	// 创建解锁脚本模板
	sigHash := sighash.Flag(sighash.ForkID | sighash.All)
	aUnlockingScriptTemplate, err := p2pkh.Unlock(clientPrivateKey, &sigHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create one unlocking script template: %w", err)
	}

	// 前序交易锁定脚本
	prevScript, err := p2pkh.Lock(clientAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create locking script: %w", err)
	}
	prevTxLockingScript := hex.EncodeToString(prevScript.Bytes())

	// 添加我的输入
	var totalValue uint64 = 0
	for _, cUtxo := range *clientUtxo {
		err = transactionData.AddInputFrom(
			cUtxo.TxID,
			cUtxo.Vout,
			prevTxLockingScript,
			cUtxo.Value,
			aUnlockingScriptTemplate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to add input: %w", err)
		}
		totalValue += cUtxo.Value
	}

	// if totalValue < serverValue {
	// 	return nil, fmt.Errorf("not enough balance, need %d, have %d", serverValue, totalValue)
	// }

	// 创建初始交易的锁定脚本
	outputMultisigScript, err := libs.Lock([]*ec.PublicKey{serverPublicKey, clientPublicKey}, 2)
	if err != nil {
		return nil, fmt.Errorf("failed to create server locking script: %w", err)
	}

	// 添加服务器输出
	transactionData.AddOutput(&tx.TransactionOutput{
		Satoshis:      totalValue,
		LockingScript: outputMultisigScript,
	})

	// 找零脚本
	// changeScript, err := p2pkh.Lock(clientAddress)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create change locking script: %w", err)
	// }

	// // 预先加入
	// transactionData.AddOutput(&tx.TransactionOutput{
	// 	Satoshis:      totalValue - serverValue,
	// 	LockingScript: changeScript,
	// })

	// 为每个输入签名，以便正确估计交易大小
	for i := range transactionData.Inputs {
		unlockingScript, err := aUnlockingScriptTemplate.Sign(transactionData, uint32(i))
		if err != nil {
			return nil, fmt.Errorf("failed to sign input %d: %w", i, err)
		}
		transactionData.Inputs[i].UnlockingScript = unlockingScript
	}

	// 计算交易大小（字节）
	txSize := transactionData.Size()
	fee := uint64(float64(txSize) / 1000.0 * float64(feeRate))
	if fee == 0 {
		fee = 1
	}

	// fmt.Printf("fee: %d, txSize: %d, feeRate: %f\n", fee, txSize, c.feeRate)
	// c.Logger.Debug("fee", zap.Int64("fee", int64(fee)), zap.Int64("txSize", int64(txSize)), zap.Float64("feeRate", c.feeRate))

	// if totalValue < serverValue+fee {
	// 	return nil, fmt.Errorf("not enough balance, need %d, have %d", serverValue+fee, totalValue)
	// }

	// 检查是否有足够的余额支付手续费
	if totalValue < fee {
		return nil, fmt.Errorf("not enough balance for fee: need %d, have %d", fee, totalValue)
	}

	// 更新找零输出
	transactionData.Outputs[0].Satoshis = totalValue - fee

	// 重新签名
	for i := range transactionData.Inputs {
		unlockingScript, err := aUnlockingScriptTemplate.Sign(transactionData, uint32(i))
		if err != nil {
			return nil, fmt.Errorf("failed to sign input %d: %w", i, err)
		}
		transactionData.Inputs[i].UnlockingScript = unlockingScript
	}

	finalAmount := totalValue - fee

	return &BuildStep1Response{
		Tx:     transactionData,
		Amount: finalAmount,
		Index:  0,
	}, nil
}
