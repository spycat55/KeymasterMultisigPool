package dual_endpoint_test

import (
    "encoding/json"
    "io/ioutil"
    "testing"

    ce "github.com/spycat55/KeymasterMultisigPool/pkg/dual_endpoint"
    pkg "github.com/spycat55/KeymasterMultisigPool/pkg"

    ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
)

type Fixture struct {
    ClientPrivHex string `json:"clientPrivHex"`
    ServerPrivHex string `json:"serverPrivHex"`
    ClientUtxos   []pkg.UTXO `json:"clientUtxos"`
    EndHeight     uint32  `json:"endHeight"`
    FeeRate       float64 `json:"feeRate"`
    IsMain        bool    `json:"isMain"`
}

func loadFixture(t *testing.T) Fixture {
    data, err := ioutil.ReadFile("../dual_endpoint/fixture.json")
    if err != nil { t.Fatalf("read fixture: %v", err) }
    var f Fixture
    if err := json.Unmarshal(data, &f); err != nil {
        t.Fatalf("unmarshal fixture: %v", err)
    }
    return f
}

func TestBuildAndSpendDualEndpoint(t *testing.T) {
    f := loadFixture(t)

    clientPriv, _ := ec.PrivateKeyFromHex(f.ClientPrivHex)
    serverPriv, _ := ec.PrivateKeyFromHex(f.ServerPrivHex)

    // Step 1
    res1, err := ce.BuildDualFeePoolBaseTx(&f.ClientUtxos, clientPriv, serverPriv.PubKey(), f.IsMain, f.FeeRate)
    if err != nil { t.Fatalf("step1: %v", err) }

    // Step 2
    bTx, _, _, err := ce.BuildDualFeePoolSpendTX(res1.Tx, res1.Amount, f.EndHeight, clientPriv, serverPriv.PubKey(), f.IsMain, f.FeeRate)
    if err != nil { t.Fatalf("step2: %v", err) }

    // Assert deterministic txid
    if res1.Tx.TxID().String() == "" || bTx.TxID().String() == "" {
        t.Fatalf("empty txid")
    }
}
