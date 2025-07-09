package pkg

type UTXO struct {
	TxID  string `json:"tx_id"`
	Vout  uint32 `json:"vout"`
	Value uint64 `json:"value"`
}
