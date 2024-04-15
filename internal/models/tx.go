package models

type TxStatus int

const (
	TxStatusUnknown TxStatus = iota
	TxStatusValid
	TxStatusInvalid
)

type Tx struct {
	meta          string   `table:"ord_tx"`
	Id            int64    `json:"id"`
	TxId          string   `json:"txid"`
	InscriptionId string   `json:"inscription_id"`
	Operation     string   `json:"op"`
	Tick          string   `json:"tick"`
	Amount        string   `json:"amt"`
	ValidAmount   string   `json:"valid_amt"` // valid amount, If the total supply is 100 and 98 has already been mined, then minted 10 will result in a valid amount of 2(100-98), not 10.
	From          string   `json:"from"`
	To            string   `json:"to"`
	SatOffset     string   `json:"sat_offset"` // sat's offset range, for instance: if input is the second position and sat's offset range is [10, 15], then the inscription's offset in the transaction is [input[1].offset + 10, input[1].offset + 15]. This is then compared with the output's offset range to select the corresponding output.
	BlockHeight   int64    `json:"block_height"`
	BlockTime     int64    `json:"block_time"`
	Position      int      `json:"pos"`
	InputIndex    int      `json:"input_idx"`
	OutputIndex   int      `json:"output_idx"`
	Status        TxStatus `json:"status"` // 0:not validated 1:valid 2:invalid
	Reason        string   `json:"reason"`
	Meta          string   `json:"meta"`    // ordinal meta, e.g: text/plain
	Content       string   `json:"content"` // ordinal raw content
}
