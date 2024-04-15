package rpc

import (
	"encoding/base64"
	"libord/pkg/conv"
	"libord/pkg/ghttp"
	"libord/pkg/slice"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type Btc struct {
	Chain    string
	Url      string
	User     string
	Password string
}

func (r *Btc) GetBlockNumber() (result int64, errRet error) {
	if ret, err := r.Request("getblockcount", nil); err != nil {
		errRet = err
	} else if ret != nil {
		result = conv.Int64(ret["result"])
	}
	return
}

func (r *Btc) GetBlockHashByNumber(number int64) (result string, errRet error) {
	if ret, err := r.Request("getblockhash", []any{number}); err != nil {
		errRet = err
	} else if ret != nil {
		result = conv.String(ret["result"])
	}
	return
}

func (r *Btc) GetBlockByNumber(number int64, retrieveTxJson bool) (result map[string]any, errRet error) {
	if hash, err := r.GetBlockHashByNumber(number); err != nil {
		errRet = err
	} else {
		return r.GetBlockByHash(hash, retrieveTxJson)
	}
	return
}

func (r *Btc) GetBlockByHash(hash string, retrieveTxJson bool) (result map[string]any, errRet error) {
	verbosity := 1
	if retrieveTxJson {
		verbosity = 3
	}
	switch r.Chain {
	case "doge":
		if ret, err := r.Request("getblock", []any{hash}); err != nil {
			errRet = err
		} else if ret != nil {
			result = ret["result"].(map[string]any)
			if retrieveTxJson {
				if result["tx"] != nil {
					var txs []any
					var m sync.Mutex
					_txs := result["tx"].([]any)
					txMap := make(map[any]any)
					for part := range slice.Partition(len(_txs), 10) {
						items := _txs[part.Low:part.High]
						var wg sync.WaitGroup
						for _, item := range items {
							wg.Add(1)
							go func(_item any) {
								defer wg.Done()
								if info, err2 := r.GetTransactionByHash(conv.String(_item)); err2 != nil {
									errRet = err2
									return
								} else {
									m.Lock()
									defer m.Unlock()
									txMap[_item] = info
								}
							}(item)
						}
						wg.Wait()
					}
					if errRet != nil {
						return
					}
					for _, txId := range _txs {
						txs = append(txs, txMap[txId])
					}
					result["tx"] = txs
				}
			}
		}
	default:
		if ret, err := r.Request("getblock", []any{hash, verbosity}); err != nil {
			errRet = err
		} else if ret != nil {
			result = ret["result"].(map[string]any)
		}
	}
	return
}

func (r *Btc) GetMemPoolTxs() (result []string, errRet error) {
	if ret, err := r.Request("getrawmempool", []any{}); err != nil {
		errRet = err
	} else if ret != nil {
		result = conv.SliceStr(ret["result"].([]any))
	}
	return
}

func (r *Btc) GetTransactionByHash(hash string) (result map[string]any, errRet error) {
	// (default: 0) A numeric parameter that can take one of the following values: '0' for hex-encoded data, '1' for JSON object and '2' for JSON object with fee and prevout
	if ret, err := r.Request("getrawtransaction", []any{hash, 2}); err != nil {
		errRet = err
	} else if ret != nil {
		result = ret["result"].(map[string]any)
		if result["txid"] != hash {
			errRet = errors.Errorf("tx:%v response:%v not match hash", hash, conv.String(ret))
		}
	}
	return
}

func (r *Btc) Request(method string, params any) (ret map[string]any, errRet error) {
	for i := 0; i < 5; i++ { // retry 5 times
		if result, err := r.doReq(method, params); err == nil {
			return result, nil
		} else {
			errRet = err
			time.Sleep(time.Second)
		}
	}
	return
}

func (r *Btc) doReq(method string, params any) (ret map[string]any, errRet error) {
	if params == nil {
		params = []any{}
	}
	id := "1"
	req := &ghttp.Request{
		Method:      http.MethodPost,
		ReadTimeOut: time.Minute,
		Url:         r.Url,
		Body: conv.String(map[string]any{
			"id":      id,
			"jsonrpc": "2.0",
			"method":  method,
			"params":  params,
		}),
	}
	if r.User != "" && r.Password != "" {
		req.Headers = map[string]string{"Content-Type": "application/json", "Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(r.User+":"+r.Password))}
	}
	if respBytes, httpStatusCode, err := req.DoReq(); err != nil {
		errRet = err
	} else {
		if httpStatusCode != http.StatusOK {
			errRet = errors.Errorf("req:%v body:%v http status code:%v resp:%s", req.Url, req.Body, httpStatusCode, string(respBytes))
		} else {
			ret = conv.Map(respBytes)
			if ret["id"] != id {
				errRet = errors.Errorf("response id not %v", id)
			} else if ret["result"] == nil {
				errRet = errors.Errorf("req:%v body:%v http status code:%v resp:%v", req.Url, req.Body, httpStatusCode, string(respBytes))
			}
		}
	}
	return
}
