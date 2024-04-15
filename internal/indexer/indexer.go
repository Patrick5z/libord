package indexer

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"libord/config"
	"libord/internal/models"
	"libord/pkg/conv"
	"libord/pkg/math"
	"libord/pkg/orm"
	"libord/pkg/rpc"
	"log"
	"strings"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/status-im/keycard-go/hexutils"
)

var (
	OP_FALSE byte = 0
	OP_1     byte = 0x51
	OP_IF    byte = 0x63

	protocols = map[string][]byte{
		"btc":  {OP_FALSE, OP_IF, 3, 'o', 'r', 'd'},
		"ltc":  {OP_FALSE, OP_IF, 3, 'o', 'r', 'd'},
		"doge": {3, 'o', 'r', 'd', OP_1},
	}
)

type Indexer struct {
	Chain string
	Db    *sql.DB
	Rpc   *rpc.Btc
}

func (s *Indexer) Run(startBlock, endBlock int64, minConfirmation int) (err error) {
	if s.Db == nil || s.Rpc == nil {
		err = errors.Errorf("db or rpc is nil, please check")
		return
	}
	_orm := &orm.Orm{Db: s.Db}
	_m := &orm.Model{TablePrefix: strings.ToLower(s.Chain) + "_"}

	dictKey := strings.ToLower(s.Chain) + ".ord.indexer.block"
	if startBlock == 0 || endBlock == 0 {
		// No specified block, query the last completed block from the database.
		obj := &models.Dict{}
		block, _err := _orm.One(_m.Bind(obj).Where("Key", dictKey), "value")
		if _err != nil {
			err = _err
			return
		}
		startBlock = conv.Int64(block)

		if startBlock <= 0 {
			startBlock = config.Instance().OrdGenesisBlock[strings.ToLower(s.Chain)]
			obj := &models.Dict{Key: dictKey, Value: conv.String(startBlock)}
			if _, _, err = _orm.Save(_m.Bind(obj).BatchData(obj)); err != nil {
				return
			}
		}
		endBlock, err = s.Rpc.GetBlockNumber()
		if err != nil {
			return
		}
		endBlock = endBlock - int64(minConfirmation)
	} else {
		// Indexed at a specified block height, no need to update dict table.
		dictKey = ""
	}

	log.Printf("ord index block start from %d to %d", startBlock, endBlock)
	for block := startBlock + 1; block <= endBlock; block++ {
		if err = s.indexBlock(block); err != nil {
			return
		}
		if dictKey != "" {
			obj := &models.Dict{}
			if _, err = _orm.Update(_m.Bind(obj).Update("Value", block).Where("Key", dictKey)); err != nil {
				return
			}
		}
	}
	return
}

func (s *Indexer) indexBlock(block int64) (err error) {
	log.Printf("indexing block:%d", block)
	var info map[string]any
	if info, err = s.Rpc.GetBlockByNumber(block, true); err != nil {
		return
	}

	for txIdx, tx := range info["tx"].([]any) {
		if err = s.indexTx(block, txIdx, tx, conv.Int64(info["time"])); err != nil {
			return
		}
	}
	return
}

func (s *Indexer) indexTx(block int64, txIdx int, tx any, blockTime int64) (err error) {
	_orm := &orm.Orm{Db: s.Db}
	_m := &orm.Model{TablePrefix: strings.ToLower(s.Chain) + "_"}
	txMap := tx.(map[string]any)
	txid := conv.String(txMap["txid"])
	vins := txMap["vin"].([]any)
	vouts := txMap["vout"].([]any)
	for inputIdx, _vin := range vins {
		vin := _vin.(map[string]any)
		var hexList []string
		if strings.EqualFold(s.Chain, "doge") {
			if vin["scriptSig"] != nil {
				hexList = append(hexList, conv.String(vin["scriptSig"].(map[string]any)["hex"]))
			}
		} else {
			witness := vin["txinwitness"]
			if witness != nil && len(witness.([]any)) >= 2 {
				size := len(witness.([]any))
				hexList = append(hexList, conv.String(witness.([]any)[size-2])) // for other taproot
				hexList = append(hexList, conv.String(witness.([]any)[size-1])) // for taproot annex， TODO: maybe it's better to check the prefix of the annex here.
			}
		}
		for _, hex := range hexList {
			meta, contents := s.parseOrd(txid, hex)
			m := conv.Map(contents)
			// To avoid issues with non-standard inscriptions in some wallets, we will reconfirm and record the problems.
			_bytes := hexutils.HexToBytes(hex)
			if m == nil && strings.Contains(string(_bytes), "\"op\"") && strings.Contains(string(_bytes), "\"tick\"") {
				log.Printf("[ERROR] parse ordinal result is nothing,but has ordinal segment string at tx:%s", txid)
			} else if m != nil && strings.EqualFold(config.Instance().OrdProtocolName[strings.ToLower(s.Chain)], conv.String(m["p"])) {
				op := strings.ToLower(conv.String(m["op"]))
				if op == "transfer" {
					op = "inscribe-transfer"
				}
				tx := &models.Tx{
					TxId:          txid,
					InscriptionId: txid + "i0",
					Operation:     op,
					Tick:          conv.String(m["tick"]),
					Amount:        conv.String(m["amt"]),
					To:            s.getOutputAddress(vouts[0].(map[string]any)),
					SatOffset:     "0," + conv.Decimal(vouts[0].(map[string]any)["value"]).Shift(8).StringFixed(0),
					BlockHeight:   block,
					BlockTime:     blockTime,
					Position:      txIdx,
					InputIndex:    inputIdx,
					OutputIndex:   0,
					Content:       string(contents),
					Meta:          meta,
				}
				if op == "deploy" {
					tick := &models.Tick{
						Name:           tx.Tick,
						Dec:            conv.Int(m["dec"], 18),
						Supply:         conv.String(m["max"]),
						MintLimit:      conv.String(m["lim"]),
						DeployTx:       txid,
						DeployAddress:  tx.To,
						DeployTime:     tx.BlockTime,
						DeployPosition: txIdx,
					}
					// if duplication, db will ignore insert
					if _, _, err = _orm.Save(_m.Bind(tick).BatchData(tick)); err != nil {
						return
					}
				}

				if _, _, err = _orm.Save(_m.Bind(tx).BatchData(tx)); err != nil {
					return
				}
				// Found the inscription record, return it, and only one inscription per transaction is allowed.
				return
			}
		}
	}

	// find transfer tx
	inputIdx2ValueMap := make(map[int]string)
	for idx, _vin := range vins {
		vin := _vin.(map[string]any)
		prevTxID := conv.String(vin["txid"])
		if conv.Int(vin["vout"]) != 0 { // inscribe tx output index must be 0
			continue
		}
		// Determine if the previous transaction is a inscribe-transfer transaction.
		obj := &models.Tx{}
		var item any
		if item, err = _orm.One(_m.Bind(obj).Where("TxId", prevTxID).Where("Operation", "inscribe-transfer"), ""); err != nil {
			return
		} else if item != nil {
			obj = item.(*models.Tx)
			toAddress, outputIdx, satOffset, _err := s.calReceiveAddress(obj.SatOffset, idx, inputIdx2ValueMap, vins, vouts)
			if _err != nil {
				err = _err
				return
			}
			tx := &models.Tx{
				TxId:          txid,
				InscriptionId: obj.InscriptionId,
				Operation:     "transfer",
				Tick:          obj.Tick,
				Amount:        obj.Amount,
				From:          obj.To,
				To:            toAddress,
				SatOffset:     satOffset,
				BlockHeight:   block,
				BlockTime:     blockTime,
				Position:      txIdx,
				InputIndex:    idx,
				OutputIndex:   outputIdx,
			}
			if _, _, err = _orm.Save(_m.Bind(tx).BatchData(tx)); err != nil {
				return
			}
		}
	}
	return
}

func (s *Indexer) parseOrd(txid, hex string) (meta string, contents []byte) {
	protocol := protocols[strings.ToLower(s.Chain)]
	_bytes := hexutils.HexToBytes(hex)
	_len := len(_bytes) - len(protocol)
	dataBeginIdx := 0
	for i := 0; i < len(_bytes); i++ {
		if i < _len {
			ok := true
			for j := 0; j < len(protocol); j++ {
				if _bytes[i+j] != protocol[j] {
					ok = false
					break
				}
			}
			if ok {
				dataBeginIdx = i + len(protocol)
				break
			}
		}
	}
	if dataBeginIdx == 0 {
		return
	}
	var tag byte
	for i := dataBeginIdx; i < len(_bytes); {
		token, idx := s.nextOp(_bytes, i)
		i = idx
		if len(token) == 1 {
			tag = token[0]
		}
		if len(token) < len("text/plain") { // Some are unnecessary strings; for safety reasons, ignore them.
			continue
		}
		if tag == 0x1 && meta == "" {
			meta = strings.ToLower(strings.TrimSpace(string(token)))
		} else if meta != "" {
			contents = token // {x}rc20 direct break, but nft need continue append, contents = append(contents, token...)
			break
		}
	}
	return
}

func (s *Indexer) nextOp(_bytes []byte, idx int) (result []byte, start int) {
	op := _bytes[idx]
	start = idx + 1
	total := len(_bytes)
	if start >= total {
		return
	}
	if op > 0 && op < 0x4c { // The next opcode bytes is data to be pushed onto the stack
		if start+int(op) <= total {
			result = _bytes[start : start+int(op)]
		}
		start += int(op)
		return
	} else if op == 0x4c { // OP_PUSHDATA1
		size := int(_bytes[start])
		start += 1
		if start+size <= total {
			result = _bytes[start : start+size]
		}
		start += size
		return
	} else if op == 0x4d && start+1 < total { // OP_PUSHDATA2
		size := int(binary.LittleEndian.Uint16([]byte{_bytes[start], _bytes[start+1]}))
		start += 2
		if start+size <= total {
			result = _bytes[start : start+size]
		}
		start += size
		return
	} else if op == 0x4e && start+3 < total { // OP_PUSHDATA4
		size := int(binary.LittleEndian.Uint32([]byte{_bytes[start], _bytes[start+1], _bytes[start+2], _bytes[start+3]}))
		start += 4
		if start+size <= total {
			result = _bytes[start : start+size]
		}
		start += size
		return
	}
	return
}

func (s *Indexer) getOutputAddress(vout map[string]any) string {
	scriptPubKey := vout["scriptPubKey"].(map[string]any)
	address := conv.String(scriptPubKey["address"])
	if address != "" {
		return address
	}
	if addresses := scriptPubKey["addresses"]; addresses != nil && len(addresses.([]any)) > 0 {
		return conv.String(addresses.([]any)[0])
	}
	return ""
}

func (s *Indexer) getInputValue(vin map[string]any) (value string, err error) {
	prevOut := vin["prevout"]
	if prevOut != nil {
		value = conv.String(prevOut.(map[string]any)["value"])
		return
	}
	var tx map[string]any
	if tx, err = s.Rpc.GetTransactionByHash(conv.String(vin["txid"])); err != nil {
		return
	}
	if tx != nil {
		vouts := tx["vout"].([]any)
		vout := conv.Int(vin["vout"])
		if len(vouts) > vout {
			value = conv.String(vouts[vout].(map[string]any)["value"])
		}
	}
	return
}

func (s *Indexer) getInputAddress(vin map[string]any) (address string, err error) {
	prevOut := vin["prevout"]
	if prevOut != nil && prevOut.(map[string]any)["scriptPubKey"] != nil {
		address = conv.String(prevOut.(map[string]any)["scriptPubKey"].(map[string]any)["address"])
		return
	}
	var tx map[string]any
	if tx, err = s.Rpc.GetTransactionByHash(conv.String(vin["txid"])); err != nil {
		return
	}
	if tx != nil {
		vouts := tx["vout"].([]any)
		vout := conv.Int(vin["vout"])
		if len(vouts) > vout {
			address = s.getOutputAddress(vouts[vout].(map[string]any))
		}
	}
	return
}

func (s *Indexer) calReceiveAddress(prevSatOffset string, currentInputIdx int, inputIdx2ValueMap map[int]string, vins, vouts []any) (to string, outputIdx int, satOffset string, err error) {
	if len(vins) == 0 || len(vouts) == 0 {
		return
	}
	split := strings.Split(prevSatOffset, ",")
	prevOutputOffset := conv.Int64(split[0])
	prevOutputEnd := conv.Int64(split[1])
	input2OffsetMap := make(map[int]int64)
	output2OffsetMap := make(map[int]int64)

	inputTotalAmount := decimal.Zero
	outputTotalAmount := decimal.Zero

	for idx, vin := range vins {
		if _, ok := inputIdx2ValueMap[idx]; !ok {
			var _value string
			if _value, err = s.getInputValue(vin.(map[string]any)); err != nil {
				return
			}
			inputIdx2ValueMap[idx] = _value
		}
		inputTotalAmount = inputTotalAmount.Add(conv.Decimal(inputIdx2ValueMap[idx]).Shift(8))
	}
	for _, vout := range vouts {
		outputTotalAmount = outputTotalAmount.Add(conv.Decimal(vout.(map[string]any)["value"]).Shift(8))
	}
	lastAddress, _err := s.getInputAddress(vins[len(vins)-1].(map[string]any))
	if _err != nil {
		err = _err
		return
	}

	// fill fee into outputs for calculate output address
	fee := inputTotalAmount.Sub(outputTotalAmount)
	if fee.GreaterThan(decimal.Zero) {
		feeF, _ := fee.Shift(-8).Float64()
		vouts = append(vouts, map[string]any{
			"scriptPubKey": map[string]any{
				"address": lastAddress,
			},
			"value": feeF,
		})
	}

	for i := 0; i <= currentInputIdx; i++ {
		if i > 0 {
			input2OffsetMap[i] = input2OffsetMap[i-1] + conv.Decimal(inputIdx2ValueMap[i-1]).Shift(8).IntPart()
		}
	}
	for idx := range vouts {
		if idx > 0 {
			output2OffsetMap[idx] = output2OffsetMap[idx-1] + conv.Decimal(vouts[idx-1].(map[string]any)["value"]).Shift(8).IntPart()
		}
	}
	inputOffset := input2OffsetMap[currentInputIdx] + prevOutputOffset
	inputEnd := input2OffsetMap[currentInputIdx] + prevOutputEnd
	for i := len(vouts) - 1; i >= 0; i-- {
		if inputOffset >= output2OffsetMap[i] {
			outputValue := conv.Decimal(vouts[i].(map[string]any)["value"]).Shift(8).IntPart()
			satOffset = fmt.Sprintf("%d,%d", inputOffset-output2OffsetMap[i], math.MinInt64(outputValue, inputEnd-output2OffsetMap[i]))
			outputIdx = i
			to = s.getOutputAddress(vouts[i].(map[string]any))
			break
		}
	}
	return
}
