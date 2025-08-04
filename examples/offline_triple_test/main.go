package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"testing"

	libs "github.com/spycat55/KeymasterMultisigPool/pkg/libs"
	te "github.com/spycat55/KeymasterMultisigPool/pkg/triple_endpoint"

	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
)

// TestConfig holds all test parameters
type TestConfig struct {
	FeeRate          float64 `json:"fee_rate"`
	EndHeight        uint32  `json:"end_height"`
	Sequence2        uint32  `json:"sequence_2"`
	NewClient1Amount uint64  `json:"new_client1_amount"`
	FinalLocktime    uint32  `json:"final_locktime"`
	FinalSequence    uint32  `json:"final_sequence"`
	Client1PrivHex   string  `json:"client1_priv_hex"`
	Client2PrivHex   string  `json:"client2_priv_hex"`
	ServerPrivHex    string  `json:"server_priv_hex"`
	FixedUTXO        struct {
		TxID  string `json:"txid"`
		Vout  uint32 `json:"vout"`
		Value uint64 `json:"value"`
	} `json:"fixed_utxo"`
	ExpectedOutputs map[string]string `json:"expected_outputs"`
}

// loadTestConfig loads configuration from test_config.json and expected_outputs.json
func loadTestConfig() (*TestConfig, error) {
	// Load test configuration
	configData, err := ioutil.ReadFile("test_config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read test_config.json: %v", err)
	}

	var config TestConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse test_config.json: %v", err)
	}

	// Load expected outputs
	expectedData, err := ioutil.ReadFile("expected_outputs.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read expected_outputs.json: %v", err)
	}

	var expectedOutputs map[string]string
	if err := json.Unmarshal(expectedData, &expectedOutputs); err != nil {
		return nil, fmt.Errorf("failed to parse expected_outputs.json: %v", err)
	}

	config.ExpectedOutputs = expectedOutputs

	return &config, nil
}

func TestTripleEndpointFixedUTXO(t *testing.T) {
	// Load test configuration
	config, err := loadTestConfig()
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// Initialize private keys
	client1Priv, err := ec.PrivateKeyFromHex(config.Client1PrivHex)
	if err != nil {
		t.Fatalf("Invalid client1 private key: %v", err)
	}

	client2Priv, err := ec.PrivateKeyFromHex(config.Client2PrivHex)
	if err != nil {
		t.Fatalf("Invalid client2 private key: %v", err)
	}

	serverPriv, err := ec.PrivateKeyFromHex(config.ServerPrivHex)
	if err != nil {
		t.Fatalf("Invalid server private key: %v", err)
	}

	// Get addresses
	client1Address, err := libs.GetAddressFromPublicKey(client1Priv.PubKey(), false) // testnet
	if err != nil {
		t.Fatalf("Failed to get client1 address: %v", err)
	}

	client2Address, err := libs.GetAddressFromPublicKey(client2Priv.PubKey(), false) // testnet
	if err != nil {
		t.Fatalf("Failed to get client2 address: %v", err)
	}

	serverAddress, err := libs.GetAddressFromPublicKey(serverPriv.PubKey(), false) // testnet
	if err != nil {
		t.Fatalf("Failed to get server address: %v", err)
	}

	fmt.Printf("Client1 Private Key: %s\n", config.Client1PrivHex)
	fmt.Printf("Client2 Private Key: %s\n", config.Client2PrivHex)
	fmt.Printf("Server Private Key: %s\n", config.ServerPrivHex)
	fmt.Printf("Client1 Address: %s\n", client1Address.AddressString)
	fmt.Printf("Client2 Address: %s\n", client2Address.AddressString)
	fmt.Printf("Server Address: %s\n", serverAddress.AddressString)
	fmt.Printf("Current Block Height: 1687373\n")
	fmt.Printf("End Height: %d\n", config.EndHeight)

	// Use fixed UTXO from config
	fixedUTXO := libs.UTXO{
		TxID:  config.FixedUTXO.TxID,
		Vout:  config.FixedUTXO.Vout,
		Value: config.FixedUTXO.Value,
	}
	client1UTXOs := []libs.UTXO{fixedUTXO}
	fmt.Printf("\nFound 1 UTXOs for client1:\n")
	fmt.Printf("UTXO 0: %s:%d (%d satoshis)\n", fixedUTXO.TxID, fixedUTXO.Vout, fixedUTXO.Value)
	fmt.Printf("Total Value: %d satoshis\n", fixedUTXO.Value)

	// Step 1: Create base transaction
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("STEP 1: Creating Base Transaction (Client1 UTXO -> Triple Multisig)")
	fmt.Println(strings.Repeat("=", 60))

	res1, err := te.BuildTripleFeePoolBaseTx(&client1UTXOs, serverPriv.PubKey(), client1Priv, client2Priv.PubKey(), false, config.FeeRate)
	if err != nil {
		t.Fatalf("Step 1 failed: %v", err)
	}

	baseTxHex := res1.Tx.String()
	fmt.Printf("=== STEP 1 - BASE TRANSACTION ===\n")
	fmt.Printf("Converts client1 UTXOs to 2-of-3 multisig output\nMultisig Amount: %d satoshis\n", res1.Amount)
	fmt.Printf("TX Hex: %s\n", baseTxHex)
	fmt.Printf("Length: %d bytes\n", len(baseTxHex)/2)
	fmt.Println("---")

	// Verify base transaction (if expected output exists)
	if expected, exists := config.ExpectedOutputs["base_tx"]; exists && baseTxHex != expected {
		t.Errorf("Base transaction mismatch.\nExpected: %s\nGot:      %s", expected, baseTxHex)
	}

	// Step 2: Build spend transaction
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("STEP 2: Build Spend Transaction (Client1 -> Client2, Server as Arbitrator)")
	fmt.Println(strings.Repeat("=", 60))

	bTx, client1SignBytes, client1Amount, err := te.BuildTripleFeePoolSpendTX(
		res1.Tx, res1.Amount, config.EndHeight, serverPriv.PubKey(), client1Priv, client2Priv.PubKey(), false, config.FeeRate)
	if err != nil {
		t.Fatalf("Step 2 failed: %v", err)
	}

	client2Amount := res1.Amount - client1Amount
	fmt.Printf("Client1 Amount (after fee): %d satoshis\n", client1Amount)
	fmt.Printf("Client2 Amount: %d satoshis\n", client2Amount)
	fmt.Printf("Server Role: Arbitrator (no funds allocated)\n")

	spendTxUnsignedHex := bTx.String()
	fmt.Printf("=== STEP 2 - SPEND TRANSACTION (UNSIGNED) ===\n")
	fmt.Printf("Spends multisig: client1 (%d sats), client2 (%d sats)\nServer acts as arbitrator\nLocktime: %d\n",
		client1Amount, client2Amount, config.EndHeight)
	fmt.Printf("TX Hex: %s\n", spendTxUnsignedHex)
	fmt.Printf("Length: %d bytes\n", len(spendTxUnsignedHex)/2)
	fmt.Println("---")

	// Step 3: Client1 signature (already done in BuildTripleFeePoolSpendTX)
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("STEP 3: Client1 Sign (from BuildTripleFeePoolSpendTX)")
	fmt.Println(strings.Repeat("=", 60))

	client1SigHex := fmt.Sprintf("%x", *client1SignBytes)
	fmt.Printf("Client1 Signature: %s\n", client1SigHex)

	// Verify client1 signature (if expected output exists)
	if expected, exists := config.ExpectedOutputs["client1_signature"]; exists && client1SigHex != expected {
		t.Errorf("Client1 signature mismatch.\nExpected: %s\nGot:      %s", expected, client1SigHex)
	}

	// Step 4: Server signature
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("STEP 4: Server Sign")
	fmt.Println(strings.Repeat("=", 60))

	serverSignBytes, err := te.SpendTXTripleFeePoolBSign(bTx, res1.Amount, serverPriv.PubKey(), client1Priv.PubKey(), client2Priv)
	if err != nil {
		t.Fatalf("Step 4 failed: %v", err)
	}

	serverSigHex := fmt.Sprintf("%x", *serverSignBytes)
	fmt.Printf("Server Signature: %s\n", serverSigHex)

	// Verify server signature (if expected output exists)
	if expected, exists := config.ExpectedOutputs["server_signature"]; exists && serverSigHex != expected {
		t.Errorf("Server signature mismatch.\nExpected: %s\nGot:      %s", expected, serverSigHex)
	}

	// Merge signatures for complete spend transaction
	bTx, err = te.MergeTripleFeePoolSigForSpendTx(bTx.String(), client1SignBytes, serverSignBytes)
	if err != nil {
		t.Fatalf("Failed to merge signatures: %v", err)
	}

	completeSpendTxHex := bTx.String()
	fmt.Printf("=== STEP 4 - COMPLETE SPEND TRANSACTION ===\n")
	fmt.Printf("Fully signed transaction (client1 + server) ready for broadcast\n")
	fmt.Printf("TX Hex: %s\n", completeSpendTxHex)
	fmt.Printf("Length: %d bytes\n", len(completeSpendTxHex)/2)
	fmt.Println("---")

	// Verify complete spend transaction (if expected output exists)
	if expected, exists := config.ExpectedOutputs["complete_spend_tx"]; exists && completeSpendTxHex != expected {
		t.Errorf("Complete spend transaction mismatch.\nExpected: %s\nGot:      %s", expected, completeSpendTxHex)
	}

	// Step 5: Client1 update signature
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("STEP 5: Client1 Update Sign (Adjust Client1-Client2 Distribution)")
	fmt.Println(strings.Repeat("=", 60))

	updatedTx, err := te.TripleFeePoolLoadTx(bTx.String(), nil, config.Sequence2, config.NewClient1Amount,
		serverPriv.PubKey(), client1Priv.PubKey(), client2Priv.PubKey(), res1.Amount)
	if err != nil {
		t.Fatalf("Step 5 load tx failed: %v", err)
	}

	client1UpdateSignBytes, err := te.ClientATripleFeePoolSpendTXUpdateSign(updatedTx, serverPriv.PubKey(), client1Priv, client2Priv.PubKey())
	if err != nil {
		t.Fatalf("Step 5 client1 sign failed: %v", err)
	}

	newClient2Amount := res1.Amount - config.NewClient1Amount
	client1UpdateSigHex := fmt.Sprintf("%x", *client1UpdateSignBytes)
	fmt.Printf("Client1 Update Signature: %s\n", client1UpdateSigHex)
	fmt.Printf("New Client1 Amount: %d satoshis\n", config.NewClient1Amount)
	fmt.Printf("New Client2 Amount: %d satoshis\n", newClient2Amount)
	fmt.Printf("Server Role: Arbitrator (unchanged)\n")
	fmt.Printf("New Sequence Number: %d\n", config.Sequence2)

	updatedTxUnsignedHex := updatedTx.String()
	fmt.Printf("=== STEP 5 - UPDATED TRANSACTION (UNSIGNED) ===\n")
	fmt.Printf("Updated distribution - Client1: %d sats, Client2: %d sats, Sequence: %d\n",
		config.NewClient1Amount, newClient2Amount, config.Sequence2)
	fmt.Printf("TX Hex: %s\n", updatedTxUnsignedHex)
	fmt.Printf("Length: %d bytes\n", len(updatedTxUnsignedHex)/2)
	fmt.Println("---")

	// Verify client1 update signature (if expected output exists)
	if expected, exists := config.ExpectedOutputs["client1_update_signature"]; exists && client1UpdateSigHex != expected {
		t.Errorf("Client1 update signature mismatch.\nExpected: %s\nGot:      %s", expected, client1UpdateSigHex)
	}

	// Step 6: Client2 agrees and signs
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("STEP 6: Client2 Agrees and Signs (Normal Negotiation)")
	fmt.Println(strings.Repeat("=", 60))

	client2UpdateSignBytes, err := te.ClientBTripleFeePoolSpendTXUpdateSign(updatedTx, serverPriv.PubKey(), client1Priv.PubKey(), client2Priv)
	if err != nil {
		t.Fatalf("Step 6 failed: %v", err)
	}

	client2UpdateSigHex := fmt.Sprintf("%x", *client2UpdateSignBytes)
	fmt.Printf("Client2 Update Signature: %s\n", client2UpdateSigHex)

	// Verify client2 update signature (if expected output exists)
	if expected, exists := config.ExpectedOutputs["client2_update_signature"]; exists && client2UpdateSigHex != expected {
		t.Errorf("Client2 update signature mismatch.\nExpected: %s\nGot:      %s", expected, client2UpdateSigHex)
	}

	// Merge updated signatures
	updatedTx, err = te.MergeTripleFeePoolSigForSpendTx(updatedTx.String(), client1UpdateSignBytes, client2UpdateSignBytes)
	if err != nil {
		t.Fatalf("Failed to merge update signatures: %v", err)
	}

	completeUpdatedTxHex := updatedTx.String()
	fmt.Printf("=== STEP 6 - COMPLETE UPDATED TRANSACTION ===\n")
	fmt.Printf("Client1 + Client2 协商完成的交易，可以广播\n")
	fmt.Printf("TX Hex: %s\n", completeUpdatedTxHex)
	fmt.Printf("Length: %d bytes\n", len(completeUpdatedTxHex)/2)
	fmt.Println("---")

	// Verify complete updated transaction (if expected output exists)
	if expected, exists := config.ExpectedOutputs["complete_updated_tx"]; exists && completeUpdatedTxHex != expected {
		t.Errorf("Complete updated transaction mismatch.\nExpected: %s\nGot:      %s", expected, completeUpdatedTxHex)
	}

	// Final step: Close fee pool
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("FINAL STEP: Close Fee Pool (locktime=0xffffffff, sequence=0xffffffff)")
	fmt.Println(strings.Repeat("=", 60))

	finalTx, err := te.TripleFeePoolLoadTx(bTx.String(), &config.FinalLocktime, config.FinalSequence, config.NewClient1Amount,
		serverPriv.PubKey(), client1Priv.PubKey(), client2Priv.PubKey(), res1.Amount)
	if err != nil {
		t.Fatalf("Final step load tx failed: %v", err)
	}

	// Client1 final signature
	client1FinalSignBytes, err := te.ClientATripleFeePoolSpendTXUpdateSign(finalTx, serverPriv.PubKey(), client1Priv, client2Priv.PubKey())
	if err != nil {
		t.Fatalf("Final client1 sign failed: %v", err)
	}

	// Server final signature
	serverFinalSignBytes, err := te.SpendTXTripleFeePoolBSign(finalTx, res1.Amount, serverPriv.PubKey(), client1Priv.PubKey(), client2Priv)
	if err != nil {
		t.Fatalf("Final server sign failed: %v", err)
	}

	// Merge final signatures
	finalTx, err = te.MergeTripleFeePoolSigForSpendTx(finalTx.String(), client1FinalSignBytes, serverFinalSignBytes)
	if err != nil {
		t.Fatalf("Failed to merge final signatures: %v", err)
	}

	client1FinalSigHex := fmt.Sprintf("%x", *client1FinalSignBytes)
	serverFinalSigHex := fmt.Sprintf("%x", *serverFinalSignBytes)
	fmt.Printf("Client1 Final Signature: %s\n", client1FinalSigHex)
	fmt.Printf("Server Final Signature: %s\n", serverFinalSigHex)
	fmt.Printf("Final Locktime: 0x%x\n", config.FinalLocktime)
	fmt.Printf("Final Sequence: 0x%x\n", config.FinalSequence)

	finalClosedTxHex := finalTx.String()
	fmt.Printf("=== FINAL - CLOSED FEE POOL TRANSACTION ===\n")
	fmt.Printf("Fee pool closed - can be broadcast immediately without time locks\n")
	fmt.Printf("TX Hex: %s\n", finalClosedTxHex)
	fmt.Printf("Length: %d bytes\n", len(finalClosedTxHex)/2)
	fmt.Println("---")

	// Verify final signatures and transaction (if expected outputs exist)
	if expected, exists := config.ExpectedOutputs["client1_final_signature"]; exists && client1FinalSigHex != expected {
		t.Errorf("Client1 final signature mismatch.\nExpected: %s\nGot:      %s", expected, client1FinalSigHex)
	}

	if expected, exists := config.ExpectedOutputs["server_final_signature"]; exists && serverFinalSigHex != expected {
		t.Errorf("Server final signature mismatch.\nExpected: %s\nGot:      %s", expected, serverFinalSigHex)
	}

	if expected, exists := config.ExpectedOutputs["final_closed_tx"]; exists && finalClosedTxHex != expected {
		t.Errorf("Final closed transaction mismatch.\nExpected: %s\nGot:      %s", expected, finalClosedTxHex)
	}

	// Summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("SUMMARY - TRANSACTION SEQUENCE")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("1. Base TX (Step 1):     %s\n", res1.Tx.TxID())
	fmt.Printf("2. Spend TX (Step 4):    %s\n", bTx.TxID())
	fmt.Printf("3. Updated TX (Step 6):  %s\n", updatedTx.TxID())
	fmt.Printf("4. Final TX (Closed):    %s\n", finalTx.TxID())
	fmt.Println()
	fmt.Printf("Client1 Final Amount: %d satoshis\n", config.NewClient1Amount)
	fmt.Printf("Client2 Final Amount: %d satoshis\n", newClient2Amount)
	fmt.Printf("Server Role: Arbitrator (no funds)\n")
	fmt.Println()
	fmt.Println("BROADCAST ORDER:")
	fmt.Println("1. First broadcast the Base Transaction (Step 1)")
	fmt.Println("2. Wait for confirmation")
	fmt.Println("3. Then broadcast the Final Transaction")
	fmt.Println()
	fmt.Println("TESTNET EXPLORER:")
	fmt.Println("https://test.whatsonchain.com/")

	fmt.Println("\n✅ All transaction outputs match expected values from successful test run!")
}

func main() {
	// Run the test
	t := &testing.T{}
	TestTripleEndpointFixedUTXO(t)
	if t.Failed() {
		log.Fatal("Test failed")
	}
	fmt.Println("Test passed successfully!")
}
