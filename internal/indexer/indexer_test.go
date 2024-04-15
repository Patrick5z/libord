package indexer

import (
	"libord/pkg/conv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseBtcOrd(t *testing.T) {
	_indexer := &Indexer{Chain: "btc"}

	txid := "d20f829557ecc07ee55341a95771585854d655d3abda9ab6e990f3115e0cbfa6"
	hex := "20117f692257b2331233b5705ce9c682be8719ff1b2b64cbca290bd6faeb54423eac06756e6973617406281429c686016d0063036f7264010118746578742f706c61696e3b636861727365743d7574662d38004d00017b0d0a2020202020202020202020202020202020202020202020202020202020202270223a20226272632d3230222c20202020200d0a202020202020202020202020202020202020202020202020202020202020202020202020226f70223a20226d696e74222c20202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020227469636b223a20226f726469222c20202020200d0a20202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202020202022616d74223a202231303030220d0a7d68"

	meta, contentBytes := _indexer.parseOrd(txid, hex)
	m := conv.Map(contentBytes)
	assert.Equal(t, meta, "text/plain;charset=utf-8")
	assert.EqualValues(t, m["p"], "brc-20")
	assert.EqualValues(t, m["op"], "mint")
	assert.EqualValues(t, m["tick"], "ordi")
	assert.EqualValues(t, m["amt"], "1000")

	txid = "57d9a8040877f854a8c1bab33d3b5906d7bf2b433b0b81923d05320146da2bdd"
	hex = "20117f692257b2331233b5705ce9c682be8719ff1b2b64cbca290bd6faeb54423eac06a5b604fe8701750063036f7264010118746578742f706c61696e3b636861727365743d7574662d3800397b2270223a226272632d3230222c226f70223a227472616e73666572222c227469636b223a22564d5058222c22616d74223a2232393430227d68"
	meta, contentBytes = _indexer.parseOrd(txid, hex)
	m = conv.Map(contentBytes)
	assert.Equal(t, meta, "text/plain;charset=utf-8")
	assert.EqualValues(t, m["p"], "brc-20")
	assert.EqualValues(t, m["op"], "transfer")
	assert.EqualValues(t, m["tick"], "VMPX")
	assert.EqualValues(t, m["amt"], "2940")

	// cbrc20
	txid = "ab0be4c01c293c92f70fb5d37a8083a055847297a03a9249e7d1cff1b9e366d1"
	hex = "0063036f7264010713636272632d32303a6d696e743a474f494e3d3101010a746578742f706c61696e003a7b2270223a226272632d3230222c226f70223a227472616e73666572222c227469636b223a226d696365222c22616d74223a223430303030227d68"
	meta, contentBytes = _indexer.parseOrd(txid, hex)
	m = conv.Map(contentBytes)
	assert.Equal(t, meta, "text/plain")
	assert.EqualValues(t, m["p"], "brc-20")
	assert.EqualValues(t, m["op"], "transfer")
	assert.EqualValues(t, m["tick"], "mice")
	assert.EqualValues(t, m["amt"], "40000")

	txid = "b61b0172d95e266c18aea0c624db987e971a5d6d4ebc2aaed85da4642d635735i0"
	hex = "209e2849b90a2353691fccedd467215c88eec89a5d0dcf468e6cf37abed344d746ac0063036f7264010118746578742f706c61696e3b636861727365743d7574662d38004c5e7b200a20202270223a20226272632d3230222c0a2020226f70223a20226465706c6f79222c0a2020227469636b223a20226f726469222c0a2020226d6178223a20223231303030303030222c0a2020226c696d223a202231303030220a7d68"
	meta, contentBytes = _indexer.parseOrd(txid, hex)
	m = conv.Map(contentBytes)
	assert.Equal(t, meta, "text/plain;charset=utf-8")
	assert.EqualValues(t, m["p"], "brc-20")
	assert.EqualValues(t, m["op"], "deploy")
	assert.EqualValues(t, m["tick"], "ordi")
	assert.EqualValues(t, m["max"], "21000000")
	assert.EqualValues(t, m["lim"], "1000")
}

func Test_ParseLtcOrd(t *testing.T) {
	_indexer := &Indexer{Chain: "ltc"}

	txid := "d51c20d107a4a01140ec116ad82533d5bcd5ec0e68429cbb80f3588f8190798e"
	hex := "20f0c1f71c0816ee449a66ad5fa5e081a87fae13407001066e1d160bf8c2178a01ac0063036f7264010118746578742f706c61696e3b636861727365743d7574662d3800487b2270223a226c74632d3230222c226f70223a226465706c6f79222c227469636b223a226c697465222c226d6178223a223834303030303030222c226c696d223a2234303030227d68"
	meta, contentBytes := _indexer.parseOrd(txid, hex)
	m := conv.Map(contentBytes)
	assert.Equal(t, meta, "text/plain;charset=utf-8")
	assert.EqualValues(t, m["p"], "ltc-20")
	assert.EqualValues(t, m["op"], "deploy")
	assert.EqualValues(t, m["tick"], "lite")
	assert.EqualValues(t, m["max"], "84000000")
	assert.EqualValues(t, m["lim"], "4000")

	txid = "daf8dc07ace5ac13f76ee69bcbcd3a54e6e03accb9b812e575e9d871476879c6"
	hex = "20bcccf22072b0b4dade40f6f63c46094cefa0cfd1ecfc70befaa9cedcd906fd7aac0063036f7264010118746578742f706c61696e3b636861727365743d7574662d3800357b2270223a226c74632d3230222c226f70223a226d696e74222c227469636b223a22666f6d6f222c22616d74223a2234303030227d68"
	meta, contentBytes = _indexer.parseOrd(txid, hex)
	m = conv.Map(contentBytes)
	assert.Equal(t, meta, "text/plain;charset=utf-8")
	assert.EqualValues(t, m["p"], "ltc-20")
	assert.EqualValues(t, m["op"], "mint")
	assert.EqualValues(t, m["tick"], "fomo")
	assert.EqualValues(t, m["amt"], "4000")

	txid = "9fcd43ad331c33ecda6d7b9788bcb44811b00853a3724b98446ac91a5a6e3ee4"
	hex = "207af4299099b48a49f65e8f327bac8ff0e49c224a3504eae0d91a35412893057dac0063036f7264010118746578742f706c61696e3b636861727365743d7574662d38003b7b2270223a226c74632d3230222c226f70223a227472616e73666572222c227469636b223a22666f6d6f222c22616d74223a22383030303030227d68"
	meta, contentBytes = _indexer.parseOrd(txid, hex)
	m = conv.Map(contentBytes)
	assert.Equal(t, meta, "text/plain;charset=utf-8")
	assert.EqualValues(t, m["p"], "ltc-20")
	assert.EqualValues(t, m["op"], "transfer")
	assert.EqualValues(t, m["tick"], "fomo")
	assert.EqualValues(t, m["amt"], "800000")
}

func Test_ParseDogeOrd(t *testing.T) {
	_indexer := &Indexer{Chain: "doge"}

	txid := "0bd32d69ca2221f3fc34d99aa14bccc2af10eedc7514770ae842ab9a72468743"
	hex := "036f72645119746578742f706c61696e3b20636861727365743d7574662d38004c647b200d0a20202270223a20226472632d3230222c0d0a2020226f70223a20226465706c6f79222c0d0a2020227469636b223a2022646f6769222c0d0a2020226d6178223a20223231303030303030222c0d0a2020226c696d223a202231303030220d0a7d4830450221008f20b47cab433bb680114700b7ec5c140d74522120d9746fd16ce20ec07f3a4502207ea121b03b038cb6838cacb1d6e5ce191f745b1af8940ec0079d3bfda26ed5e80129210321802b1bbff4781a29049a8fa84e71ed1e553ba16c8c196c0b0149c3f283a988ad757575757551"
	meta, contentBytes := _indexer.parseOrd(txid, hex)
	m := conv.Map(contentBytes)
	assert.Equal(t, meta, "text/plain; charset=utf-8")
	assert.EqualValues(t, m["p"], "drc-20")
	assert.EqualValues(t, m["op"], "deploy")
	assert.EqualValues(t, m["tick"], "dogi")
	assert.EqualValues(t, m["max"], "21000000")
	assert.EqualValues(t, m["lim"], "1000")

	txid = "702e1b8ac65c561f66c71172ebe807f775ad6dfb6c51d2e846e1d01dec9e5a1f"
	hex = "036f72645117746578742f706c61696e3b636861727365743d7574663800467b0a20202270223a20226472632d3230222c0a2020226f70223a20226d696e74222c0a2020227469636b223a202262696f70222c0a202022616d74223a202231303030220a7d483045022100836948a8f7362bb8d4e39f27a5620c85451497e27ea72cacbbb8fb0e2f6f517102201ca38ae0d2bda776c198e30e88993289223e736a5ad8f3e8343b1c49eb37dee20129210285230c884117ba81d98b0f3857c27ca676606ae0e32dc817ecfb0d26f6585705ad757575757551"
	meta, contentBytes = _indexer.parseOrd(txid, hex)
	m = conv.Map(contentBytes)
	assert.Equal(t, meta, "text/plain;charset=utf8")
	assert.EqualValues(t, m["p"], "drc-20")
	assert.EqualValues(t, m["op"], "mint")
	assert.EqualValues(t, m["tick"], "biop")
	assert.EqualValues(t, m["amt"], "1000")

	txid = "7d11289ca2ea0d555cb8902f21df942b38ff8b5540064d58c15ade61c01e3ac9"
	hex = "036f72645117746578742f706c61696e3b636861727365743d7574663800377b2270223a226472632d3230222c226f70223a227472616e73666572222c227469636b223a22646f6769222c22616d74223a223530227d47304402206c6f364a39645563e3ab934c4d0f4cf34055ce3cf97349efcc95b55fddb82c60022065f995a4f9149c407e53f6137a5741a5dd3074ba1706947aa0673b14fc62b24701292103a8160e2442ce02cf12de7de63200845cf5418aa0a374b865b3712b3979a2173aad757575757551"
	meta, contentBytes = _indexer.parseOrd(txid, hex)
	m = conv.Map(contentBytes)
	assert.Equal(t, meta, "text/plain;charset=utf8")
	assert.EqualValues(t, m["p"], "drc-20")
	assert.EqualValues(t, m["op"], "transfer")
	assert.EqualValues(t, m["tick"], "dogi")
	assert.EqualValues(t, m["amt"], "50")
}
