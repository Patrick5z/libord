package validator

import (
	"database/sql"
	"fmt"
	"libord/config"
	"libord/internal/models"
	"libord/pkg/conv"
	"libord/pkg/orm"
	"libord/pkg/slice"
	"log"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type Validator struct {
	Chain string
	Db    *sql.DB

	tickMap    map[string]*models.Tick
	addressMap map[string]*models.Address

	validateTicks []string // the ticks which need to be validated
}

func (s *Validator) Run() (err error) {
	if s.Db == nil {
		err = errors.Errorf("db is nil, please check")
		return
	}

	validatorDictKey := strings.ToLower(s.Chain) + ".ord.validator.block"

	// The highest block height already indexed by the indexer.
	indexerBlock := s.getDictValue(strings.ToLower(s.Chain) + ".ord.indexer.block")
	// The highest block height already validated by the validator.
	validatorBlock := s.getDictValue(validatorDictKey)

	log.Printf("found indexed block:%d validated block:%d", indexerBlock, validatorBlock)

	if validatorBlock <= 0 {
		validatorBlock = s.getGenesisBlock()
		if err = s.saveDict(validatorDictKey, validatorBlock); err != nil {
			return
		}
	}
	if validatorBlock < indexerBlock {
		if err = s.loadTickData(); err != nil {
			return
		}
		if err = s.loadAddressData(); err != nil {
			return
		}
	}

	for block := validatorBlock + 1; block <= indexerBlock; block++ {
		if err = s.validateBlock(block); err != nil {
			return
		}
		if err = s.updateDict(validatorDictKey, block); err != nil {
			return
		}
	}
	return
}

func (s *Validator) Revalidate(startBlock, endBlock int64, ticks []string) (err error) {
	log.Printf("revalidating ticks:%s", conv.String(ticks))
	if len(ticks) == 0 {
		err = errors.Errorf("ticks not allowed empty")
		return
	}
	s.validateTicks = slice.Map(ticks, func(item string) string {
		return strings.ToLower(item)
	})

	if err = s.loadTickData(); err != nil {
		return
	}

	if err = s.loadAddressData(); err != nil {
		return
	}

	for _, tick := range s.validateTicks {
		s.tickMap[tick].MintedAmount = "0"
		s.tickMap[tick].BlockAtUpdate = 0
	}

	for _, addr := range s.addressMap {
		if slice.Contains(s.validateTicks, strings.ToLower(addr.Tick)) {
			// Reset the balance of the address, i.e., start calculating the balance from beginning.
			addr.Available = ""
			addr.Transferable = ""
			addr.BlockAtUpdate = 0
		}
	}
	for block := startBlock + 1; block <= endBlock; block++ {
		if err = s.validateBlock(block); err != nil {
			return
		}
	}
	return
}

// loadTickData: load tick data from db for speedy validation
func (s *Validator) loadTickData() (err error) {
	log.Printf("loading all tick table data")
	if s.tickMap == nil {
		s.tickMap = make(map[string]*models.Tick)
	}
	_orm := &orm.Orm{Db: s.Db}
	obj := &models.Tick{}
	startId := int64(0)
	limit := 2000
	for {
		log.Printf("load tick, start id:%d", startId)
		if items, _err := _orm.Find((&orm.Model{TablePrefix: strings.ToLower(s.Chain) + "_"}).Bind(obj).WhereGT("Id", startId).Extra("order by id asc limit ?", limit)); _err != nil {
			err = _err
			return
		} else {
			var mutex sync.Mutex
			if err = s.batchExec(items, func(_info any) (funcErr error) {
				tick := _info.(*models.Tick)
				var tx any
				_m := &orm.Model{TablePrefix: strings.ToLower(s.Chain) + "_"}
				if tick.DeployTx != "" && tick.DeployPosition <= 0 {
					if tx, funcErr = _orm.One(_m.Bind(&models.Tx{}).Where("TxId", tick.DeployTx), ""); funcErr != nil {
						return
					} else { // If tx is nil, an exception will be thrown. This helps us identify situations where a tick lacks a deploy transaction, although such cases are generally rare.
						tick.DeployPosition = tx.(*models.Tx).Position
						if _, err = _orm.Update(_m.Bind(tick).Update("DeployPosition", tick.DeployPosition).Where("Id", tick.Id)); err != nil {
							return
						}
					}
				}
				mutex.Lock()
				defer mutex.Unlock()
				s.tickMap[strings.ToLower(tick.Name)] = tick
				return
			}); err != nil {
				return
			}
			if len(items) < limit {
				break
			}
			startId = items[len(items)-1].(*models.Tick).Id
		}
	}
	return
}

// loadAddressData: load address data from db for speedy validation
func (s *Validator) loadAddressData() (err error) {
	log.Printf("loading all address balance table data")
	if s.addressMap == nil {
		s.addressMap = make(map[string]*models.Address)
	}
	_orm := &orm.Orm{Db: s.Db}
	_m := &orm.Model{TablePrefix: strings.ToLower(s.Chain) + "_"}
	obj := &models.Address{}
	startId := int64(0)
	limit := 2000
	for {
		if items, _err := _orm.Find(_m.Bind(obj).WhereGT("Id", startId).Extra("order by id asc limit ?", limit)); _err != nil {
			err = _err
			return
		} else {
			for _, item := range items {
				address := item.(*models.Address)
				startId = address.Id
				s.addressMap[strings.ToLower(fmt.Sprintf("%s,%s", address.Tick, address.Address))] = address
			}
			if len(items) < limit {
				break
			}
		}
	}
	return
}

func (s *Validator) validateBlock(block int64) (err error) {
	log.Printf("validating block:%d", block)
	dirtyTick := make(map[string]bool)
	dirtyAddress := make(map[string]bool)
	var dirtyTransactions []any

	// Because we batch update all transactions under a block, we need to cache these transactions.
	// This ensures that 'transfer' transactions can obtain the correct 'inscribe-transfer' status before the database is updated.
	txMap := make(map[string][]*models.Tx)

	_orm := &orm.Orm{Db: s.Db}
	_m := &orm.Model{TablePrefix: strings.ToLower(s.Chain) + "_"}

	start := 0
	limit := 2000
	for {
		if items, _err := _orm.Find(_m.Bind(&models.Tx{}).WhereGTE("BlockHeight", block).WhereLT("BlockHeight", block+1).Extra("order by pos asc,input_idx asc limit ?,?", start, limit)); _err != nil {
			err = _err
			return
		} else {
			for _, item := range items {
				tx := item.(*models.Tx)
				if len(s.validateTicks) > 0 && !slice.Contains(s.validateTicks, strings.ToLower(tx.Tick)) {
					continue
				}
				// No need for revalidation if it's a manual patch.
				isPatch := strings.Index(tx.Reason, "patch:") == 0
				if isPatch && tx.Status == models.TxStatusInvalid {
					continue
				}
				tick := s.tickMap[strings.ToLower(tx.Tick)]
				tx.Reason = ""
				if tick == nil {
					tx.Reason = fmt.Sprintf("The tick:%s has not been deployed yet.", tx.Tick)
				} else {
					senderKey := strings.ToLower(fmt.Sprintf("%s,%s", tx.Tick, tx.From))
					recipientKey := strings.ToLower(fmt.Sprintf("%s,%s", tx.Tick, tx.To))
					// ensure address has balance records
					if tx.From != "" && s.addressMap[senderKey] == nil {
						s.addressMap[senderKey] = &models.Address{
							Tick:    tick.Name,
							Address: tx.From,
						}
					}
					if tx.To != "" && s.addressMap[recipientKey] == nil {
						s.addressMap[recipientKey] = &models.Address{
							Tick:    tick.Name,
							Address: tx.To,
						}
					}
					if tx.Reason = s.validateCommon(tx); tx.Reason == "" {
						amount := conv.Decimal(tx.Amount)
						switch strings.ToLower(tx.Operation) {
						case "deploy": // No need to validate name, dec, max, lim; it seems redundant, so ignore them.
							if tick.DeployTx != tx.TxId {
								tx.Reason = fmt.Sprintf("The tick:%s has been deployed at %s.", tx.Tick, tick.DeployTx)
							}
						case "mint":
							if tx.Reason = s.validateMint(tx, tick); isPatch || tx.Reason == "" {
								remainMintAmount := conv.Decimal(tick.Supply).Sub(conv.Decimal(tick.MintedAmount))
								if remainMintAmount.LessThanOrEqual(amount) { // remain mint amount <= tx amount
									if tick.BlockAtUpdate < block {
										tx.ValidAmount = remainMintAmount.String()
										tick.MintedAmount = tick.Supply
										tick.FinishMintTx = tx.TxId
										tick.FinishMintTime = tx.BlockTime
										dirtyTick[strings.ToLower(tick.Name)] = true
									}

									if s.addressMap[recipientKey].BlockAtUpdate < block {
										s.addressMap[recipientKey].Available = conv.Decimal(s.addressMap[recipientKey].Available).Add(remainMintAmount).String()
										dirtyAddress[recipientKey] = true
									}
								} else { // remain mint amount is sufficient
									if tick.BlockAtUpdate < block {
										tick.MintedAmount = conv.Decimal(tick.MintedAmount).Add(amount).String()
										dirtyTick[strings.ToLower(tick.Name)] = true
									}

									if s.addressMap[recipientKey].BlockAtUpdate < block {
										s.addressMap[recipientKey].Available = conv.Decimal(s.addressMap[recipientKey].Available).Add(amount).String()
										dirtyAddress[recipientKey] = true
									}
								}
							}
						case "inscribe-transfer":
							if tx.Reason = s.validateInscribeTransfer(tx, recipientKey); isPatch || tx.Reason == "" {
								if s.addressMap[recipientKey].BlockAtUpdate < block {
									s.addressMap[recipientKey].Available = conv.Decimal(s.addressMap[recipientKey].Available).Sub(amount).String()
									s.addressMap[recipientKey].Transferable = conv.Decimal(s.addressMap[recipientKey].Transferable).Add(amount).String()
									dirtyAddress[recipientKey] = true
								}
							}
						case "transfer":
							if tx.Reason, err = s.validateTransfer(tx, senderKey, txMap); err != nil {
								return
							} else if isPatch || tx.Reason == "" {
								// Do not revalidate addresses that have been verified before to avoid discrepancies caused by duplicate changes in amounts.
								// Validation must occur incrementally for each block; it cannot be done intermittently.
								// Otherwise, transactions that were verified later may be invalid, requiring revalidation.
								if s.addressMap[senderKey].BlockAtUpdate < block {
									// Deduct transferable-amount from the sender.
									s.addressMap[senderKey].Transferable = conv.Decimal(s.addressMap[senderKey].Transferable).Sub(amount).String()
									dirtyAddress[senderKey] = true
								}
								if s.addressMap[senderKey].BlockAtUpdate < block {
									// Credit available-amount to the recipient.
									s.addressMap[recipientKey].Available = conv.Decimal(s.addressMap[recipientKey].Available).Add(amount).String()
									dirtyAddress[recipientKey] = true
								}
							}
						default:
							tx.Reason = fmt.Sprintf("unknown op:%s", tx.Operation)
						}
					}
				}
				if tx.Reason != "" {
					tx.Status = models.TxStatusInvalid
				} else {
					tx.Status = models.TxStatusValid
				}

				if !isPatch {
					dirtyTransactions = append(dirtyTransactions, tx)
				}
				txMap[tx.TxId] = append(txMap[tx.TxId], tx)
			}
			start += limit
			if len(items) < limit {
				break
			}
		}
	}

	log.Printf("updating %d tx", len(dirtyTransactions))
	// use goroutine to boost speedy update
	if err = s.batchExec(dirtyTransactions, func(_info any) (funcErr error) {
		info := _info.(*models.Tx)
		_model := &orm.Model{TablePrefix: strings.ToLower(s.Chain) + "_"}
		_, funcErr = _orm.Update(_model.Bind(&models.Tx{}).Update("Status", info.Status).Update("Reason", info.Reason).Update("ValidAmount", info.ValidAmount).Where("Id", info.Id))
		return
	}); err != nil {
		return
	}

	log.Printf("updating %d tick", len(dirtyTick))
	var ticks []any
	for key := range dirtyTick {
		ticks = append(ticks, s.tickMap[key])
	}
	if err = s.batchExec(ticks, func(_info any) (funcErr error) {
		info := _info.(*models.Tick)
		_model := &orm.Model{TablePrefix: strings.ToLower(s.Chain) + "_"}
		if info.BlockAtUpdate < block {
			info.BlockAtUpdate = block
		}
		_, funcErr = _orm.Update(_model.Bind(&models.Tick{}).Update("MintedAmount", info.MintedAmount).Update("FinishMintTx", info.FinishMintTx).Update("FinishMintTime", info.FinishMintTime).Update("BlockAtUpdate", info.BlockAtUpdate).Where("Id", info.Id))
		return
	}); err != nil {
		return
	}

	log.Printf("updating %d address", len(dirtyAddress))
	var addresses []any
	for key := range dirtyAddress {
		addresses = append(addresses, s.addressMap[key])
	}
	if err = s.batchExec(addresses, func(_info any) (funcErr error) {
		info := _info.(*models.Address)
		_model := &orm.Model{TablePrefix: strings.ToLower(s.Chain) + "_"}
		_obj := &models.Address{}
		if info.BlockAtUpdate < block {
			info.BlockAtUpdate = block
		}
		if info.Id <= 0 {
			var addressId int64
			if _, addressId, funcErr = _orm.Save(_model.Bind(_obj).BatchData(info)); err != nil {
				return
			} else if addressId > 0 {
				info.Id = addressId
			}
		} else {
			_, funcErr = _orm.Update(_model.Bind(_obj).Update("Available", info.Available).Update("Transferable", info.Transferable).Update("BlockAtUpdate", info.BlockAtUpdate).Where("Id", info.Id))
		}
		return
	}); err != nil {
		return
	}
	return
}

func (s *Validator) validateCommon(tx *models.Tx) string {
	if tx.From == "" && tx.To == "" {
		return "'from' and 'to' address are both empty"
	}

	if !strings.EqualFold(tx.Operation, "transfer") {
		protocolName := config.Instance().OrdProtocolName[strings.ToLower(s.Chain)]
		if m := conv.Map(tx.Content); m != nil && !strings.EqualFold(conv.String(m["p"]), protocolName) {
			return "not " + protocolName + " protocol"
		}
		contentType := strings.ToLower(strings.TrimSpace(tx.Meta))
		if strings.Index(contentType, "text/plain") != 0 && strings.Index(contentType, "application/json") != 0 {
			return fmt.Sprintf("content-type:%s is not valid", tx.Meta)
		}
	}

	if !strings.EqualFold(tx.Operation, "deploy") && conv.Decimal(tx.Amount).LessThanOrEqual(decimal.Zero) {
		return fmt.Sprintf("The amount:%s not valid", tx.Amount)
	}
	return ""
}

func (s *Validator) validateMint(tx *models.Tx, tick *models.Tick) string {
	amount := conv.Decimal(tx.Amount)
	if tick.DeployTime > tx.BlockTime || (tick.DeployTime == tx.BlockTime && tick.DeployPosition > tx.Position) {
		return fmt.Sprintf("The tick:%s has not been deployed before %d.", tx.Tick, tx.BlockTime)
	} else if amount.GreaterThan(conv.Decimal(tick.MintLimit)) {
		return fmt.Sprintf("The mint amount:%s has exceeded mint limit:%s", tx.Amount, tick.MintLimit)
	} else {
		remainMintAmount := conv.Decimal(tick.Supply).Sub(conv.Decimal(tick.MintedAmount))
		if remainMintAmount.LessThanOrEqual(decimal.Zero) { // remain mint amount is zero
			return fmt.Sprintf("The tick:%s have already been full minted.", tick.Name)
		}
	}
	return ""
}

func (s *Validator) validateInscribeTransfer(tx *models.Tx, addressKey string) string {
	amount := conv.Decimal(tx.Amount)
	if conv.Decimal(s.addressMap[addressKey].Available).LessThan(amount) {
		return fmt.Sprintf("Insufficient balance for inscription; 'available balance' is only '%s'", s.addressMap[addressKey].Available)
	}
	return ""
}

func (s *Validator) validateTransfer(tx *models.Tx, addressKey string, txMap map[string][]*models.Tx) (reason string, err error) {
	amount := conv.Decimal(tx.Amount)
	_orm := &orm.Orm{Db: s.Db}
	_m := &orm.Model{TablePrefix: strings.ToLower(s.Chain) + "_"}

	if conv.Decimal(s.addressMap[addressKey].Transferable).LessThan(amount) {
		reason = fmt.Sprintf("Insufficient balance for inscription; 'transferable balance' is only '%s'", s.addressMap[addressKey].Transferable)
	} else {
		inscribeTx := tx.InscriptionId[0:64]
		if len(txMap[inscribeTx]) > 0 && txMap[inscribeTx][0].Status == models.TxStatusValid {
			return
		}
		if inscribeTx, _err := _orm.One(_m.Bind(&models.Tx{}).Where("TxId", inscribeTx).Where("Operation", "inscribe-transfer").Where("Tick", tx.Tick).Where("Status", 1), ""); _err != nil {
			err = _err
			return
		} else if inscribeTx == nil {
			reason = fmt.Sprintf("The previous inscribe-transfer tx:%s failed.", tx.InscriptionId)
		}
	}
	return
}

func (s *Validator) getDictValue(key string) int64 {
	_orm := &orm.Orm{Db: s.Db}
	_m := &orm.Model{TablePrefix: strings.ToLower(s.Chain) + "_"}
	obj := &models.Dict{}
	if str, _err := _orm.One(_m.Bind(obj).Where("Key", key), "value"); _err != nil {
		log.Fatalf("read dict db error:%+v", _err)
	} else {
		return conv.Int64(str)
	}
	return 0
}

func (s *Validator) updateDict(key string, value any) error {
	_orm := &orm.Orm{Db: s.Db}
	_m := &orm.Model{TablePrefix: strings.ToLower(s.Chain) + "_"}
	obj := &models.Dict{}
	_, err := _orm.Update(_m.Bind(obj).Update("Value", value).Where("Key", key))
	return err
}

func (s *Validator) saveDict(key string, value any) error {
	_orm := &orm.Orm{Db: s.Db}
	_m := &orm.Model{TablePrefix: strings.ToLower(s.Chain) + "_"}
	obj := &models.Dict{Key: key, Value: conv.String(value)}
	_, _, err := _orm.Save(_m.Bind(obj).BatchData(obj))
	return err
}

func (s *Validator) getGenesisBlock() int64 {
	if tx, _err := (&orm.Orm{Db: s.Db}).One((&orm.Model{TablePrefix: strings.ToLower(s.Chain) + "_"}).Bind(&models.Tx{}).Extra("order by block_height asc limit 1"), ""); _err != nil {
		log.Fatalf("get db error:%+v", _err)
	} else if tx != nil {
		return tx.(*models.Tx).BlockHeight - 1
	}
	return 0
}

func (s *Validator) batchExec(items []any, callback func(_info any) error) (err error) {
	for parti := range slice.Partition(len(items), 20) {
		partiItems := items[parti.Low:parti.High]
		var wg sync.WaitGroup
		errCh := make(chan error, len(partiItems))
		for _, item := range partiItems {
			wg.Add(1)
			go func(info any) {
				defer wg.Done()
				if info == nil {
					return
				}
				if _err := callback(info); _err != nil {
					errCh <- _err
					return
				}
			}(item)
		}
		go func() {
			wg.Wait()
			close(errCh)
		}()
		if err = <-errCh; err != nil {
			return
		}
	}
	return
}
