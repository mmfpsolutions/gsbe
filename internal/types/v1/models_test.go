/*
 * Copyright 2026 Scott Walter, MMFP Solutions LLC
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published by the Free
 * Software Foundation; either version 3 of the License, or (at your option)
 * any later version.  See LICENSE for more details.
 */

package v1types

import (
	"encoding/json"
	"testing"
)

// These tests cover the decode boundary against the node REST API. GSBE talks
// to more than one coin and more than one Core vintage, and the fields that
// DIFFER between them are typed `interface{}` for exactly that reason. The
// point here is that every shape a real node returns decodes without error —
// a decode failure takes out the whole page, not one field.

// ---------------------------------------------------------------------------
// ChainInfo
// ---------------------------------------------------------------------------

// Bitcoin-style: one proof-of-work algorithm, so `difficulty` is a bare number
// and `difficulties` is absent.
func TestChainInfoSingleAlgoDifficulty(t *testing.T) {
	const raw = `{
		"chain": "main",
		"blocks": 875432,
		"headers": 875432,
		"bestblockhash": "0000000000000000000149c9a2c1b0f7cbd0f8b7bfbb90b53bd1f4b9e0a1c2d3",
		"difficulty": 108522647629298.4,
		"time": 1735689600,
		"mediantime": 1735689000,
		"verificationprogress": 0.9999987,
		"pruned": false,
		"size_on_disk": 682341234567,
		"warnings": ""
	}`

	var info ChainInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		t.Fatalf("decoding single-algo chaininfo: %v", err)
	}

	if info.Chain != "main" || info.Blocks != 875432 {
		t.Errorf("chain/blocks = %q/%d, want main/875432", info.Chain, info.Blocks)
	}
	// A bare JSON number lands in the interface as float64. Any consumer type
	// asserting to anything else panics, so this is worth pinning explicitly.
	diff, ok := info.Difficulty.(float64)
	if !ok {
		t.Fatalf("Difficulty is %T, want float64", info.Difficulty)
	}
	if diff != 108522647629298.4 {
		t.Errorf("Difficulty = %v, want 108522647629298.4", diff)
	}
	if info.Difficulties != nil {
		t.Errorf("Difficulties = %v, want nil for a single-algo chain", info.Difficulties)
	}
	if info.SizeOnDisk != 682341234567 {
		t.Errorf("SizeOnDisk = %d, want 682341234567 — must not overflow int32", info.SizeOnDisk)
	}
}

// DigiByte: five algorithms, so `difficulties` is a map and `difficulty`
// carries the value for the current algo. Both must survive the same decode.
func TestChainInfoMultiAlgoDifficulties(t *testing.T) {
	const raw = `{
		"chain": "main",
		"blocks": 21500000,
		"headers": 21500000,
		"bestblockhash": "00000000000000012a3b4c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f8091",
		"difficulty": 1234.5678,
		"difficulties": {
			"sha256d": 1234.5678,
			"scrypt": 87654.321,
			"skein": 4321.8765,
			"qubit": 5678.1234,
			"odo": 91011.1213
		},
		"time": 1735689600,
		"mediantime": 1735689000,
		"verificationprogress": 1,
		"pruned": false,
		"size_on_disk": 45678901234,
		"warnings": ""
	}`

	var info ChainInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		t.Fatalf("decoding multi-algo chaininfo: %v", err)
	}

	if len(info.Difficulties) != 5 {
		t.Fatalf("len(Difficulties) = %d, want 5", len(info.Difficulties))
	}
	for _, algo := range []string{"sha256d", "scrypt", "skein", "qubit", "odo"} {
		if _, ok := info.Difficulties[algo]; !ok {
			t.Errorf("Difficulties missing %q", algo)
		}
	}
	if info.Difficulties["scrypt"] != 87654.321 {
		t.Errorf("Difficulties[scrypt] = %v, want 87654.321", info.Difficulties["scrypt"])
	}
}

// `warnings` changed shape in Core 25: it was a single string, it is now an
// array of strings. GSBE points at whatever the operator is running, so BOTH
// must decode. This is the field most likely to break a node upgrade, and the
// reason it is typed interface{}.
func TestChainInfoWarningsAcceptsBothShapes(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		wantType string
		check    func(*testing.T, interface{})
	}{
		{
			name:     "pre-Core-25 empty string",
			raw:      `{"chain":"main","warnings":""}`,
			wantType: "string",
			check: func(t *testing.T, w interface{}) {
				if w.(string) != "" {
					t.Errorf("warnings = %q, want empty", w)
				}
			},
		},
		{
			name:     "pre-Core-25 populated string",
			raw:      `{"chain":"main","warnings":"Unknown new rules activated"}`,
			wantType: "string",
			check: func(t *testing.T, w interface{}) {
				if w.(string) != "Unknown new rules activated" {
					t.Errorf("warnings = %q, want the warning text", w)
				}
			},
		},
		{
			name:     "Core 25+ empty array",
			raw:      `{"chain":"main","warnings":[]}`,
			wantType: "[]interface {}",
			check: func(t *testing.T, w interface{}) {
				if len(w.([]interface{})) != 0 {
					t.Errorf("warnings = %v, want empty", w)
				}
			},
		},
		{
			name:     "Core 25+ populated array",
			raw:      `{"chain":"main","warnings":["Unknown new rules activated","Large reorg detected"]}`,
			wantType: "[]interface {}",
			check: func(t *testing.T, w interface{}) {
				if len(w.([]interface{})) != 2 {
					t.Errorf("warnings = %v, want 2 entries", w)
				}
			},
		},
		{
			// Some nodes omit it entirely; the zero interface must not be
			// mistaken for an empty string by a consumer doing a type switch.
			name:     "absent",
			raw:      `{"chain":"main"}`,
			wantType: "<nil>",
			check:    func(t *testing.T, w interface{}) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var info ChainInfo
			if err := json.Unmarshal([]byte(tt.raw), &info); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if got := typeName(info.Warnings); got != tt.wantType {
				t.Fatalf("warnings is %s, want %s", got, tt.wantType)
			}
			tt.check(t, info.Warnings)
		})
	}
}

func typeName(v interface{}) string {
	if v == nil {
		return "<nil>"
	}
	switch v.(type) {
	case string:
		return "string"
	case []interface{}:
		return "[]interface {}"
	default:
		return "unknown"
	}
}

// ---------------------------------------------------------------------------
// Block
// ---------------------------------------------------------------------------

// A verbosity-2 block as returned for the chain tip, with a coinbase
// transaction. Exercises the fields the block page reads.
const tipBlockJSON = `{
	"hash": "00000000000000000001a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f7",
	"confirmations": 1,
	"height": 875432,
	"version": 536870912,
	"merkleroot": "4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c",
	"time": 1735689600,
	"mediantime": 1735689000,
	"nonce": 3947285016,
	"bits": "17034219",
	"difficulty": 108522647629298.4,
	"chainwork": "0000000000000000000000000000000000000000a1b2c3d4e5f60718293a4b5c",
	"nTx": 2,
	"previousblockhash": "000000000000000000029384756abcdef0123456789abcdef0123456789abcde",
	"size": 1543210,
	"weight": 3993210,
	"strippedsize": 816667,
	"tx": [
		{
			"txid": "a1b2c3d4e5f60718293a4b5c6d7e8f9012345678901234567890123456789012",
			"hash": "a1b2c3d4e5f60718293a4b5c6d7e8f9012345678901234567890123456789012",
			"version": 2,
			"size": 285,
			"vsize": 258,
			"weight": 1032,
			"locktime": 0,
			"vin": [
				{
					"coinbase": "03a8580d1b4d696e656420627920534d",
					"txinwitness": ["0000000000000000000000000000000000000000000000000000000000000000"],
					"sequence": 4294967295
				}
			],
			"vout": [
				{
					"value": 3.14159265,
					"n": 0,
					"scriptPubKey": {
						"asm": "OP_DUP OP_HASH160 89abcdef0123456789abcdef0123456789abcdef OP_EQUALVERIFY OP_CHECKSIG",
						"hex": "76a91489abcdef0123456789abcdef0123456789abcdef88ac",
						"address": "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2",
						"type": "pubkeyhash"
					}
				}
			]
		},
		{
			"txid": "b2c3d4e5f60718293a4b5c6d7e8f901234567890123456789012345678901234",
			"hash": "c3d4e5f60718293a4b5c6d7e8f90123456789012345678901234567890123456",
			"version": 2,
			"size": 225,
			"vsize": 144,
			"weight": 573,
			"locktime": 875431,
			"vin": [
				{
					"txid": "d4e5f60718293a4b5c6d7e8f9012345678901234567890123456789012345678",
					"vout": 1,
					"sequence": 4294967293,
					"scriptSig": {"asm": "", "hex": ""}
				}
			],
			"vout": [
				{
					"value": 0.00054321,
					"n": 0,
					"scriptPubKey": {
						"hex": "0014abcdef0123456789abcdef0123456789abcdef01",
						"address": "bc1q40x77qy3nzefu6l0hupsn2vhr9m8sdcxvfvphv",
						"type": "witness_v0_keyhash"
					}
				}
			],
			"fee": 0.00001234
		}
	]
}`

func TestBlockDecode(t *testing.T) {
	var b Block
	if err := json.Unmarshal([]byte(tipBlockJSON), &b); err != nil {
		t.Fatalf("decoding block: %v", err)
	}

	if b.Height != 875432 || b.NTx != 2 {
		t.Errorf("height/nTx = %d/%d, want 875432/2", b.Height, b.NTx)
	}
	// nTx is `nTx` on the wire, not `ntx` — Go's decoder is case-insensitive
	// so this passes either way, but the tag documents the real field name.
	if len(b.Tx) != 2 {
		t.Fatalf("len(Tx) = %d, want 2", len(b.Tx))
	}

	// Nonce exceeds int32; it must not be truncated or sign-flipped.
	if b.Nonce != 3947285016 {
		t.Errorf("Nonce = %d, want 3947285016", b.Nonce)
	}
	// Bits is a hex STRING, not a number — a node returns "17034219", and
	// decoding it as an int would fail outright.
	if b.Bits != "17034219" {
		t.Errorf("Bits = %q, want the hex string 17034219", b.Bits)
	}

	// The tip has no nextblockhash. Its absence must read as empty, which is
	// what the template branches on to hide the "next" link.
	if b.NextBlockHash != "" {
		t.Errorf("NextBlockHash = %q, want empty on the tip", b.NextBlockHash)
	}
	if b.PreviousBlockHash == "" {
		t.Error("PreviousBlockHash is empty, want the parent hash")
	}
	// PowAlgo is DigiByte-only and absent here.
	if b.PowAlgo != "" {
		t.Errorf("PowAlgo = %q, want empty for a single-algo chain", b.PowAlgo)
	}
}

func TestBlockCoinbaseVsSpendingInput(t *testing.T) {
	var b Block
	if err := json.Unmarshal([]byte(tipBlockJSON), &b); err != nil {
		t.Fatalf("decoding block: %v", err)
	}

	// The coinbase input has a `coinbase` field and NO txid/scriptSig. This is
	// how the UI decides to render "Newly generated coins" instead of an input
	// address, so the distinction has to survive decoding.
	cb := b.Tx[0].Vin[0]
	if cb.Coinbase == "" {
		t.Error("coinbase input has no Coinbase field")
	}
	if cb.TxID != "" {
		t.Errorf("coinbase input TxID = %q, want empty", cb.TxID)
	}
	if cb.ScriptSig != nil {
		t.Errorf("coinbase input ScriptSig = %+v, want nil", cb.ScriptSig)
	}
	if len(cb.TxInWitness) != 1 {
		t.Errorf("len(TxInWitness) = %d, want 1 — the segwit commitment witness", len(cb.TxInWitness))
	}
	if cb.Sequence != 4294967295 {
		t.Errorf("Sequence = %d, want 4294967295 — must not overflow uint32", cb.Sequence)
	}

	// A spending input is the mirror image: txid + vout + scriptSig, no
	// coinbase. Note vout is 1 here, so a `Vout != 0` presence test would be
	// wrong for the common vout-0 case — the coinbase field is the signal.
	spend := b.Tx[1].Vin[0]
	if spend.Coinbase != "" {
		t.Errorf("spending input Coinbase = %q, want empty", spend.Coinbase)
	}
	if spend.TxID == "" {
		t.Error("spending input has no TxID")
	}
	if spend.Vout != 1 {
		t.Errorf("spending input Vout = %d, want 1", spend.Vout)
	}
	if spend.ScriptSig == nil {
		t.Error("spending input ScriptSig = nil, want the (empty, segwit) script object")
	}
}

func TestTransactionValuesAndFee(t *testing.T) {
	var b Block
	if err := json.Unmarshal([]byte(tipBlockJSON), &b); err != nil {
		t.Fatalf("decoding block: %v", err)
	}

	// Values arrive as decimal coin amounts, not satoshis. 8-dp values must
	// survive float64 exactly enough to render.
	if got := b.Tx[0].Vout[0].Value; got != 3.14159265 {
		t.Errorf("coinbase output value = %v, want 3.14159265", got)
	}
	if got := b.Tx[1].Vout[0].Value; got != 0.00054321 {
		t.Errorf("output value = %v, want 0.00054321", got)
	}

	// Fee is `omitempty` and absent on the coinbase — a coinbase has no fee,
	// and 0 is the correct read.
	if b.Tx[0].Fee != 0 {
		t.Errorf("coinbase Fee = %v, want 0", b.Tx[0].Fee)
	}
	if b.Tx[1].Fee != 0.00001234 {
		t.Errorf("Fee = %v, want 0.00001234", b.Tx[1].Fee)
	}

	// txid != hash for a segwit transaction; conflating them breaks lookups.
	if b.Tx[1].TxID == b.Tx[1].Hash {
		t.Error("segwit tx has TxID == Hash; the wtxid must decode separately")
	}
	if b.Tx[1].LockTime != 875431 {
		t.Errorf("LockTime = %d, want 875431", b.Tx[1].LockTime)
	}
}

// A DigiByte block names its algorithm. The field is omitempty, so it must be
// present when the node sends it and absent when it does not.
func TestBlockPowAlgoRoundTrip(t *testing.T) {
	const raw = `{"hash":"abc","height":21500000,"pow_algo":"scrypt","bits":"1a2b3c4d","tx":[]}`

	var b Block
	if err := json.Unmarshal([]byte(raw), &b); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if b.PowAlgo != "scrypt" {
		t.Fatalf("PowAlgo = %q, want scrypt", b.PowAlgo)
	}

	out, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var reencoded map[string]interface{}
	if err := json.Unmarshal(out, &reencoded); err != nil {
		t.Fatalf("re-decode: %v", err)
	}
	if reencoded["pow_algo"] != "scrypt" {
		t.Errorf("re-encoded pow_algo = %v, want scrypt", reencoded["pow_algo"])
	}

	var noAlgo Block
	if err := json.Unmarshal([]byte(`{"hash":"abc","tx":[]}`), &noAlgo); err != nil {
		t.Fatalf("decode without algo: %v", err)
	}
	out2, err := json.Marshal(noAlgo)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var reencoded2 map[string]interface{}
	if err := json.Unmarshal(out2, &reencoded2); err != nil {
		t.Fatalf("re-decode: %v", err)
	}
	if _, present := reencoded2["pow_algo"]; present {
		t.Error("pow_algo present when unset, want it omitted")
	}
}

// A block whose transactions were not expanded (verbosity 1) returns tx as an
// array of hash strings, which does NOT fit []Transaction. Pinned so the
// failure is understood as a verbosity mismatch rather than a corrupt node.
func TestBlockVerbosityOneIsARejectedShape(t *testing.T) {
	const raw = `{"hash":"abc","height":1,"tx":["a1b2c3","d4e5f6"]}`

	var b Block
	err := json.Unmarshal([]byte(raw), &b)
	if err == nil {
		t.Fatal("decoding a verbosity-1 block succeeded; Tx is []Transaction and must reject bare hashes")
	}
}

// ---------------------------------------------------------------------------
// ScriptPubKey / NodeStatus / MempoolInfo
// ---------------------------------------------------------------------------

// Not every output has an address: OP_RETURN and bare-multisig outputs have a
// type but no address, and the UI must render them rather than skipping them.
func TestScriptPubKeyWithoutAddress(t *testing.T) {
	const raw = `{
		"value": 0,
		"n": 1,
		"scriptPubKey": {
			"asm": "OP_RETURN aa21a9ed1234567890",
			"hex": "6a24aa21a9ed1234567890",
			"type": "nulldata"
		}
	}`

	var v Vout
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if v.ScriptPubKey.Type != "nulldata" {
		t.Errorf("type = %q, want nulldata", v.ScriptPubKey.Type)
	}
	if v.ScriptPubKey.Address != "" {
		t.Errorf("address = %q, want empty for an OP_RETURN output", v.ScriptPubKey.Address)
	}
	if v.Value != 0 {
		t.Errorf("value = %v, want 0", v.Value)
	}
}

func TestNodeStatusOmitsMessageWhenOnline(t *testing.T) {
	online := NodeStatus{
		ID: "abc12345", Name: "DigiByte", Symbol: "DGB",
		Network: "mainnet", Online: true, ChainHeight: 21500000,
	}
	out, err := json.Marshal(online)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, present := m["message"]; present {
		t.Error("message present on a healthy node, want it omitted")
	}
	// online and chain_height are NOT omitempty: a node that is down must
	// serialise `"online": false`, not drop the key and read as undefined.
	if v, present := m["online"]; !present || v != true {
		t.Errorf("online = %v (present=%v), want true", v, present)
	}

	offline := NodeStatus{ID: "abc12345", Online: false, ChainHeight: 0, Message: "connection refused"}
	out2, err := json.Marshal(offline)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m2 map[string]interface{}
	if err := json.Unmarshal(out2, &m2); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if v, present := m2["online"]; !present || v != false {
		t.Errorf("online = %v (present=%v), want an explicit false", v, present)
	}
	if v, present := m2["chain_height"]; !present || v != float64(0) {
		t.Errorf("chain_height = %v (present=%v), want an explicit 0", v, present)
	}
	if m2["message"] != "connection refused" {
		t.Errorf("message = %v, want the failure reason", m2["message"])
	}
}

// MempoolInfo is a bare map alias precisely because coins disagree about the
// field set. Anything the node sends has to survive.
func TestMempoolInfoAcceptsDivergentFields(t *testing.T) {
	const raw = `{
		"loaded": true,
		"size": 142,
		"bytes": 58231,
		"usage": 312480,
		"total_fee": 0.00234,
		"maxmempool": 300000000,
		"mempoolminfee": 0.00001,
		"minrelaytxfee": 0.00001,
		"incrementalrelayfee": 0.00001,
		"unbroadcastcount": 0,
		"fullrbf": true
	}`

	var info MempoolInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if info["size"] != float64(142) {
		t.Errorf("size = %v, want 142", info["size"])
	}
	if info["fullrbf"] != true {
		t.Errorf("fullrbf = %v, want true", info["fullrbf"])
	}
	// An unknown field from a coin GSBE has never seen must pass through
	// rather than being dropped.
	var extended MempoolInfo
	if err := json.Unmarshal([]byte(`{"size":1,"some_new_coin_field":"xyz"}`), &extended); err != nil {
		t.Fatalf("decode with unknown field: %v", err)
	}
	if extended["some_new_coin_field"] != "xyz" {
		t.Error("unknown mempool field was dropped; the map alias exists to keep it")
	}
}
