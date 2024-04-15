package models

type Address struct {
	meta          string `table:"ord_address"`
	Id            int64  `json:"id"`
	Address       string `json:"address"`
	Tick          string `json:"tick"`
	Available     string `json:"available"`
	Transferable  string `json:"transferable"`
	BlockAtUpdate int64  `json:"block"` // block height at last update balance
}
