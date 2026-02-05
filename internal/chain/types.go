package chain

type Block struct {
	Number       string        `json:"number"`
	Hash         string        `json:"hash"`
	ParentHash   string        `json:"parentHash"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	Hash  string `json:"hash"`
	From  string `json:"from"`
	To    string `json:"to"`
	Value string `json:"value"`
	Data  string `json:"input"`
}

type TxPoolContent struct {
	Pending map[string]map[string]Transaction `json:"pending"`
	Queued  map[string]map[string]Transaction `json:"queued"`
}

type CallRequest struct {
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	Gas      string `json:"gas,omitempty"`
	GasPrice string `json:"gasPrice,omitempty"`
	Value    string `json:"value,omitempty"`
	Data     string `json:"data,omitempty"`
	MaxFee   string `json:"maxFeePerGas,omitempty"`
	MaxPrio  string `json:"maxPriorityFeePerGas,omitempty"`
}

type SendTransactionRequest struct {
	From                 string `json:"from"`
	To                   string `json:"to,omitempty"`
	Gas                  string `json:"gas,omitempty"`
	Value                string `json:"value,omitempty"`
	Data                 string `json:"data,omitempty"`
	MaxFeePerGas         string `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas,omitempty"`
	Nonce                string `json:"nonce,omitempty"`
	Type                 string `json:"type,omitempty"`
	ChainID              string `json:"chainId,omitempty"`
}

type SignedTransaction struct {
	Raw  string `json:"raw"`
	Tx   string `json:"tx"`
	Hash string `json:"hash"`
}

type BundleItem struct {
	Txs           []string `json:"txs"`
	BlockNumber   string   `json:"blockNumber"`
	MinTimestamp  int64    `json:"minTimestamp,omitempty"`
	MaxTimestamp  int64    `json:"maxTimestamp,omitempty"`
	RevertingTxs  []string `json:"revertingTxs,omitempty"`
	UUID          string   `json:"uuid,omitempty"`
	ReplacementID string   `json:"replacementUuid,omitempty"`
}

type BundleRequest struct {
	JSONRPC string       `json:"jsonrpc"`
	ID      int          `json:"id"`
	Method  string       `json:"method"`
	Params  []BundleItem `json:"params"`
}
