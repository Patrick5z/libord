package models

type Tick struct {
	meta           string `table:"ord_tick"`
	Id             int64  `json:"id"`
	Name           string `json:"name"`
	Dec            int    `json:"dec"`
	Supply         string `json:"supply"`
	MintLimit      string `json:"mint_limit"`
	MintedAmount   string `json:"minted"`
	DeployTx       string `json:"deploy_tx"`
	DeployPosition int    `json:"deploy_pos"` // The position of the block where the transaction deploying this tick is located.
	DeployAddress  string `json:"deploy_by"`
	DeployTime     int64  `json:"deploy_time"`
	FinishMintTx   string `json:"finish_mint_tx"`
	FinishMintTime int64  `json:"finish_mint_time"`
	BlockAtUpdate  int64  `json:"block"` // block height at last update
}
