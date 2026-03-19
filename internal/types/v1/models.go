package v1types

// ChainInfo holds blockchain info from the node REST API
type ChainInfo struct {
	Chain                string             `json:"chain"`
	Blocks               int64              `json:"blocks"`
	Headers              int64              `json:"headers"`
	BestBlockHash        string             `json:"bestblockhash"`
	Difficulty           interface{}        `json:"difficulty"`
	Difficulties         map[string]float64 `json:"difficulties,omitempty"`
	Time                 int64              `json:"time"`
	MedianTime           int64              `json:"mediantime"`
	VerificationProgress float64            `json:"verificationprogress"`
	Pruned               bool               `json:"pruned"`
	SizeOnDisk           int64              `json:"size_on_disk"`
	Warnings             interface{}        `json:"warnings"`
}

// Block represents a blockchain block
type Block struct {
	Hash              string        `json:"hash"`
	Confirmations     int64         `json:"confirmations"`
	Height            int64         `json:"height"`
	Version           int           `json:"version"`
	MerkleRoot        string        `json:"merkleroot"`
	Time              int64         `json:"time"`
	MedianTime        int64         `json:"mediantime"`
	Nonce             uint64        `json:"nonce"`
	Bits              string        `json:"bits"`
	Difficulty        interface{}   `json:"difficulty"`
	PowAlgo           string        `json:"pow_algo,omitempty"`
	Chainwork         string        `json:"chainwork"`
	NTx               int           `json:"nTx"`
	PreviousBlockHash string        `json:"previousblockhash,omitempty"`
	NextBlockHash     string        `json:"nextblockhash,omitempty"`
	Size              int           `json:"size"`
	Weight            int           `json:"weight"`
	StrippedSize      int           `json:"strippedsize"`
	Tx                []Transaction `json:"tx"`
}

// Transaction represents a blockchain transaction
type Transaction struct {
	TxID     string  `json:"txid"`
	Hash     string  `json:"hash"`
	Version  int     `json:"version"`
	Size     int     `json:"size"`
	VSize    int     `json:"vsize"`
	Weight   int     `json:"weight"`
	LockTime uint32  `json:"locktime"`
	Vin      []Vin   `json:"vin"`
	Vout     []Vout  `json:"vout"`
	Fee      float64 `json:"fee,omitempty"`
}

// Vin represents a transaction input
type Vin struct {
	TxID        string   `json:"txid,omitempty"`
	Vout        int      `json:"vout,omitempty"`
	Coinbase    string   `json:"coinbase,omitempty"`
	TxInWitness []string `json:"txinwitness,omitempty"`
	Sequence    uint32   `json:"sequence"`
	ScriptSig   *Script  `json:"scriptSig,omitempty"`
}

// Vout represents a transaction output
type Vout struct {
	Value        float64      `json:"value"`
	N            int          `json:"n"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

// Script represents input script data
type Script struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

// ScriptPubKey represents output script data
type ScriptPubKey struct {
	Asm     string `json:"asm,omitempty"`
	Hex     string `json:"hex,omitempty"`
	Address string `json:"address,omitempty"`
	Type    string `json:"type,omitempty"`
	Desc    string `json:"desc,omitempty"`
}

// MempoolInfo uses a flexible map since different coins return different fields
type MempoolInfo = map[string]interface{}

// NodeStatus represents the status of a configured node
type NodeStatus struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Network     string `json:"network"`
	Online      bool   `json:"online"`
	ChainHeight int64  `json:"chain_height"`
	Message     string `json:"message,omitempty"`
}
