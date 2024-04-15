package models

type Dict struct {
	meta  string `table:"ord_dict"`
	Id    int64  `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}
