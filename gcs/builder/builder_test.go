// Copyright (c) 2017 The btcsuite developers
// Copyright (c) 2017 The Lightning Network Developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package builder_test

import (
	"bytes"
	"encoding/hex"
	"github.com/DFWallet/xec/txscript"
	"testing"
	"time"

	"github.com/DFWallet/xec/chaincfg/chainhash"
	"github.com/DFWallet/xec/wire"
	"github.com/DFWallet/xecutil/gcs"
	"github.com/DFWallet/xecutil/gcs/builder"
)

var (
	// No need to allocate an err variable in every test
	err error

	// List of values for building a filter
	contents = [][]byte{
		[]byte("Alex"),
		[]byte("Bob"),
		[]byte("Charlie"),
		[]byte("Dick"),
		[]byte("Ed"),
		[]byte("Frank"),
		[]byte("George"),
		[]byte("Harry"),
		[]byte("Ilya"),
		[]byte("John"),
		[]byte("Kevin"),
		[]byte("Larry"),
		[]byte("Michael"),
		[]byte("Nate"),
		[]byte("Owen"),
		[]byte("Paul"),
		[]byte("Quentin"),
	}

	testKey = [16]byte{0x4c, 0xb1, 0xab, 0x12, 0x57, 0x62, 0x1e, 0x41,
		0x3b, 0x8b, 0x0e, 0x26, 0x64, 0x8d, 0x4a, 0x15}

	testHash = "000000000000000000496d7ff9bd2c96154a8d64260e8b3b411e625712abb14c"

	testAddrScript = "a914e95dbb25283cc35d4a6aefa76c0382e63ce0fa3687"
)

// TestUseBlockHash tests using a block hash as a filter key.
func TestUseBlockHash(t *testing.T) {
	// Block hash #448710, pretty high difficulty.
	hash, err := chainhash.NewHashFromStr(testHash)
	if err != nil {
		t.Fatalf("Hash from string failed: %s", err.Error())
	}

	// wire.OutPoint
	outPoint := wire.OutPoint{
		Hash:  *hash,
		Index: 4321,
	}

	addrBytes, err := hex.DecodeString(testAddrScript)
	if err != nil {
		t.Fatalf("Address script build failed: %s", err.Error())
	}

	// Create a GCSBuilder with a key hash and check that the key is derived
	// correctly, then test it.
	b := builder.WithKeyHash(hash)
	key, err := b.Key()
	if err != nil {
		t.Fatalf("Builder instantiation with key hash failed: %s",
			err.Error())
	}
	if key != testKey {
		t.Fatalf("Key not derived correctly from key hash:\n%s\n%s",
			hex.EncodeToString(key[:]),
			hex.EncodeToString(testKey[:]))
	}
	BuilderTest(b, hash, builder.DefaultP, outPoint, addrBytes, t)

	// Create a GCSBuilder with a key hash and non-default P and test it.
	b = builder.WithKeyHashPM(hash, 30, 90)
	BuilderTest(b, hash, 30, outPoint, addrBytes, t)

	// Create a GCSBuilder with a random key, set the key from a hash
	// manually, check that the key is correct, and test it.
	b = builder.WithRandomKey()
	b.SetKeyFromHash(hash)
	key, err = b.Key()
	if err != nil {
		t.Fatalf("Builder instantiation with known key failed: %s",
			err.Error())
	}
	if key != testKey {
		t.Fatalf("Key not copied correctly from known key:\n%s\n%s",
			hex.EncodeToString(key[:]),
			hex.EncodeToString(testKey[:]))
	}
	BuilderTest(b, hash, builder.DefaultP, outPoint, addrBytes, t)

	// Create a GCSBuilder with a random key and test it.
	b = builder.WithRandomKey()
	key1, err := b.Key()
	if err != nil {
		t.Fatalf("Builder instantiation with random key failed: %s",
			err.Error())
	}
	t.Logf("Random Key 1: %s", hex.EncodeToString(key1[:]))
	BuilderTest(b, hash, builder.DefaultP, outPoint, addrBytes, t)

	// Create a GCSBuilder with a random key and non-default P and test it.
	b = builder.WithRandomKeyPM(30, 90)
	key2, err := b.Key()
	if err != nil {
		t.Fatalf("Builder instantiation with random key failed: %s",
			err.Error())
	}
	t.Logf("Random Key 2: %s", hex.EncodeToString(key2[:]))
	if key2 == key1 {
		t.Fatalf("Random keys are the same!")
	}
	BuilderTest(b, hash, 30, outPoint, addrBytes, t)

	// Create a GCSBuilder with a known key and test it.
	b = builder.WithKey(testKey)
	key, err = b.Key()
	if err != nil {
		t.Fatalf("Builder instantiation with known key failed: %s",
			err.Error())
	}
	if key != testKey {
		t.Fatalf("Key not copied correctly from known key:\n%s\n%s",
			hex.EncodeToString(key[:]),
			hex.EncodeToString(testKey[:]))
	}
	BuilderTest(b, hash, builder.DefaultP, outPoint, addrBytes, t)

	// Create a GCSBuilder with a known key and non-default P and test it.
	b = builder.WithKeyPM(testKey, 30, 90)
	key, err = b.Key()
	if err != nil {
		t.Fatalf("Builder instantiation with known key failed: %s",
			err.Error())
	}
	if key != testKey {
		t.Fatalf("Key not copied correctly from known key:\n%s\n%s",
			hex.EncodeToString(key[:]),
			hex.EncodeToString(testKey[:]))
	}
	BuilderTest(b, hash, 30, outPoint, addrBytes, t)

	// Create a GCSBuilder with a known key and too-high P and ensure error
	// works throughout all functions that use it.
	b = builder.WithRandomKeyPM(33, 99).SetKeyFromHash(hash).SetKey(testKey)
	b.SetP(30).AddEntry(hash.CloneBytes()).AddEntries(contents).
		AddHash(hash).AddEntry(addrBytes)
	_, err = b.Key()
	if err != gcs.ErrPTooBig {
		t.Fatalf("No error on P too big!")
	}
	_, err = b.Build()
	if err != gcs.ErrPTooBig {
		t.Fatalf("No error on P too big!")
	}
}

func BuilderTest(b *builder.GCSBuilder, hash *chainhash.Hash, p uint8,
	outPoint wire.OutPoint, addrBytes []byte, t *testing.T) {

	key, err := b.Key()
	if err != nil {
		t.Fatalf("Builder instantiation with key hash failed: %s",
			err.Error())
	}

	// Build a filter and test matches.
	b.AddEntries(contents)
	f, err := b.Build()
	if err != nil {
		t.Fatalf("Filter build failed: %s", err.Error())
	}
	if f.P() != p {
		t.Fatalf("Filter built with wrong probability")
	}
	match, err := f.Match(key, []byte("Nate"))
	if err != nil {
		t.Fatalf("Filter match failed: %s", err)
	}
	if !match {
		t.Fatal("Filter didn't match when it should have!")
	}
	match, err = f.Match(key, []byte("weks"))
	if err != nil {
		t.Fatalf("Filter match failed: %s", err)
	}
	if match {
		t.Logf("False positive match, should be 1 in 2**%d!",
			builder.DefaultP)
	}

	// Add a hash, build a filter, and test matches
	b.AddHash(hash)
	f, err = b.Build()
	if err != nil {
		t.Fatalf("Filter build failed: %s", err.Error())
	}
	match, err = f.Match(key, hash.CloneBytes())
	if err != nil {
		t.Fatalf("Filter match failed: %s", err)
	}
	if !match {
		t.Fatal("Filter didn't match when it should have!")
	}

	// Add a script, build a filter, and test matches
	b.AddEntry(addrBytes)
	f, err = b.Build()
	if err != nil {
		t.Fatalf("Filter build failed: %s", err.Error())
	}
	match, err = f.MatchAny(key, [][]byte{addrBytes})
	if err != nil {
		t.Fatalf("Filter match any failed: %s", err)
	}
	if !match {
		t.Fatal("Filter didn't match when it should have!")
	}

	// Check that adding duplicate items does not increase filter size.
	originalSize := f.N()
	b.AddEntry(addrBytes)
	f, err = b.Build()
	if err != nil {
		t.Fatalf("Filter build failed: %s", err.Error())
	}
	if f.N() != originalSize {
		t.Fatal("Filter size increased with duplicate items")
	}
}

// TestBuildBasicFilter tests building the filter against a test block
// and makes sure everything that should match the filter does and things
// that should not match do not.
func TestBuildBasicFilter(t *testing.T) {
	filter, err := builder.BuildBasicFilter(&TestBlock)
	if err != nil {
		t.Fatal(err)
	}

	blockHash := TestBlock.BlockHash()
	key := builder.DeriveKey(&blockHash)

	// Test coinbase
	var buf bytes.Buffer
	TestBlock.Transactions[0].TxIn[0].PreviousOutPoint.Serialize(&buf)
	matches, err := filter.Match(key, buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if matches {
		t.Errorf("Coinbase previous outpoint matches filter when it should not")
	}

	matches, err = filter.Match(key, []byte{0x01, 0x02, 0x03})
	if err != nil {
		t.Fatal(err)
	}
	if matches {
		t.Errorf("Coinbase OP_RETURN output matches filter when it should not")
	}

	matches, err = filter.Match(key, TestBlock.Transactions[0].TxOut[0].PkScript)
	if err != nil {
		t.Fatal(err)
	}
	if !matches {
		t.Errorf("Coinbase output 0 does not match the filter when it should")
	}

	// Test non-coinbase transactions
	for i, tx := range TestBlock.Transactions[1:] {
		for x, in := range tx.TxIn {
			var buf bytes.Buffer
			in.PreviousOutPoint.Serialize(&buf)
			matches, err := filter.Match(key, buf.Bytes())
			if err != nil {
				t.Fatal(err)
			}
			if !matches {
				t.Errorf("Previous outpoint %d for transaction %d not in filter", x, i)
			}
		}
		for x, out := range tx.TxOut {
			if out.PkScript[0] != txscript.OP_RETURN {
				matches, err := filter.Match(key, out.PkScript)
				if err != nil {
					t.Fatal(err)
				}
				if !matches {
					t.Errorf("Output %d for transaction %d not in filter", x, i)
				}
			}
		}
	}
}

var TestBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version: 1,
		PrevBlock: chainhash.Hash([32]byte{ // Make go vet happy.
			0x50, 0x12, 0x01, 0x19, 0x17, 0x2a, 0x61, 0x04,
			0x21, 0xa6, 0xc3, 0x01, 0x1d, 0xd3, 0x30, 0xd9,
			0xdf, 0x07, 0xb6, 0x36, 0x16, 0xc2, 0xcc, 0x1f,
			0x1c, 0xd0, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00,
		}), // 000000000002d01c1fccc21636b607dfd930d31d01c3a62104612a1719011250
		MerkleRoot: chainhash.Hash([32]byte{ // Make go vet happy.
			0x66, 0x57, 0xa9, 0x25, 0x2a, 0xac, 0xd5, 0xc0,
			0xb2, 0x94, 0x09, 0x96, 0xec, 0xff, 0x95, 0x22,
			0x28, 0xc3, 0x06, 0x7c, 0xc3, 0x8d, 0x48, 0x85,
			0xef, 0xb5, 0xa4, 0xac, 0x42, 0x47, 0xe9, 0xf3,
		}), // f3e94742aca4b5ef85488dc37c06c3282295ffec960994b2c0d5ac2a25a95766
		Timestamp: time.Unix(1293623863, 0), // 2010-12-29 11:57:43 +0000 UTC
		Bits:      0x1b04864c,               // 453281356
		Nonce:     0x10572b0f,               // 274148111
	},
	Transactions: []*wire.MsgTx{
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash:  chainhash.Hash{},
						Index: 0xffffffff,
					},
					SignatureScript: []byte{
						0x04, 0x4c, 0x86, 0x04, 0x1b, 0x02, 0x06, 0x02,
					},
					Sequence: 0xffffffff,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 0x12a05f200, // 5000000000
					PkScript: []byte{
						0x41, // OP_DATA_65
						0x04, 0x1b, 0x0e, 0x8c, 0x25, 0x67, 0xc1, 0x25,
						0x36, 0xaa, 0x13, 0x35, 0x7b, 0x79, 0xa0, 0x73,
						0xdc, 0x44, 0x44, 0xac, 0xb8, 0x3c, 0x4e, 0xc7,
						0xa0, 0xe2, 0xf9, 0x9d, 0xd7, 0x45, 0x75, 0x16,
						0xc5, 0x81, 0x72, 0x42, 0xda, 0x79, 0x69, 0x24,
						0xca, 0x4e, 0x99, 0x94, 0x7d, 0x08, 0x7f, 0xed,
						0xf9, 0xce, 0x46, 0x7c, 0xb9, 0xf7, 0xc6, 0x28,
						0x70, 0x78, 0xf8, 0x01, 0xdf, 0x27, 0x6f, 0xdf,
						0x84, // 65-byte signature
						0xac, // OP_CHECKSIG
					},
				},
				{
					Value: 0,
					PkScript: []byte{
						0x6a, // OP_RETURN
						0x03, // OP_DATA_9
						0x01, 0x02, 0x03,
					},
				},
			},
			LockTime: 0,
		},
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash: chainhash.Hash([32]byte{ // Make go vet happy.
							0x03, 0x2e, 0x38, 0xe9, 0xc0, 0xa8, 0x4c, 0x60,
							0x46, 0xd6, 0x87, 0xd1, 0x05, 0x56, 0xdc, 0xac,
							0xc4, 0x1d, 0x27, 0x5e, 0xc5, 0x5f, 0xc0, 0x07,
							0x79, 0xac, 0x88, 0xfd, 0xf3, 0x57, 0xa1, 0x87,
						}), // 87a157f3fd88ac7907c05fc55e271dc4acdc5605d187d646604ca8c0e9382e03
						Index: 0,
					},
					SignatureScript: []byte{
						0x49, // OP_DATA_73
						0x30, 0x46, 0x02, 0x21, 0x00, 0xc3, 0x52, 0xd3,
						0xdd, 0x99, 0x3a, 0x98, 0x1b, 0xeb, 0xa4, 0xa6,
						0x3a, 0xd1, 0x5c, 0x20, 0x92, 0x75, 0xca, 0x94,
						0x70, 0xab, 0xfc, 0xd5, 0x7d, 0xa9, 0x3b, 0x58,
						0xe4, 0xeb, 0x5d, 0xce, 0x82, 0x02, 0x21, 0x00,
						0x84, 0x07, 0x92, 0xbc, 0x1f, 0x45, 0x60, 0x62,
						0x81, 0x9f, 0x15, 0xd3, 0x3e, 0xe7, 0x05, 0x5c,
						0xf7, 0xb5, 0xee, 0x1a, 0xf1, 0xeb, 0xcc, 0x60,
						0x28, 0xd9, 0xcd, 0xb1, 0xc3, 0xaf, 0x77, 0x48,
						0x01, // 73-byte signature
						0x41, // OP_DATA_65
						0x04, 0xf4, 0x6d, 0xb5, 0xe9, 0xd6, 0x1a, 0x9d,
						0xc2, 0x7b, 0x8d, 0x64, 0xad, 0x23, 0xe7, 0x38,
						0x3a, 0x4e, 0x6c, 0xa1, 0x64, 0x59, 0x3c, 0x25,
						0x27, 0xc0, 0x38, 0xc0, 0x85, 0x7e, 0xb6, 0x7e,
						0xe8, 0xe8, 0x25, 0xdc, 0xa6, 0x50, 0x46, 0xb8,
						0x2c, 0x93, 0x31, 0x58, 0x6c, 0x82, 0xe0, 0xfd,
						0x1f, 0x63, 0x3f, 0x25, 0xf8, 0x7c, 0x16, 0x1b,
						0xc6, 0xf8, 0xa6, 0x30, 0x12, 0x1d, 0xf2, 0xb3,
						0xd3, // 65-byte pubkey
					},
					Sequence: 0xffffffff,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 0x2123e300, // 556000000
					PkScript: []byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0xc3, 0x98, 0xef, 0xa9, 0xc3, 0x92, 0xba, 0x60,
						0x13, 0xc5, 0xe0, 0x4e, 0xe7, 0x29, 0x75, 0x5e,
						0xf7, 0xf5, 0x8b, 0x32,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					},
				},
				{
					Value: 0x108e20f00, // 4444000000
					PkScript: []byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0x94, 0x8c, 0x76, 0x5a, 0x69, 0x14, 0xd4, 0x3f,
						0x2a, 0x7a, 0xc1, 0x77, 0xda, 0x2c, 0x2f, 0x6b,
						0x52, 0xde, 0x3d, 0x7c,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					},
				},
				{
					Value: 0x108e20f00, // 4444000000
					PkScript: []byte{
						0x6a, // OP_RETURN
						0x09, // OP_DATA_9
						0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x09, 0x09,
						0x04, // OP_DATA_4
						0xff, 0xff, 0xff, 0xff,
						0x5d, // OP_7
					},
				},
			},
			LockTime: 0,
		},
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash: chainhash.Hash([32]byte{ // Make go vet happy.
							0xc3, 0x3e, 0xbf, 0xf2, 0xa7, 0x09, 0xf1, 0x3d,
							0x9f, 0x9a, 0x75, 0x69, 0xab, 0x16, 0xa3, 0x27,
							0x86, 0xaf, 0x7d, 0x7e, 0x2d, 0xe0, 0x92, 0x65,
							0xe4, 0x1c, 0x61, 0xd0, 0x78, 0x29, 0x4e, 0xcf,
						}), // cf4e2978d0611ce46592e02d7e7daf8627a316ab69759a9f3df109a7f2bf3ec3
						Index: 1,
					},
					SignatureScript: []byte{
						0x47, // OP_DATA_71
						0x30, 0x44, 0x02, 0x20, 0x03, 0x2d, 0x30, 0xdf,
						0x5e, 0xe6, 0xf5, 0x7f, 0xa4, 0x6c, 0xdd, 0xb5,
						0xeb, 0x8d, 0x0d, 0x9f, 0xe8, 0xde, 0x6b, 0x34,
						0x2d, 0x27, 0x94, 0x2a, 0xe9, 0x0a, 0x32, 0x31,
						0xe0, 0xba, 0x33, 0x3e, 0x02, 0x20, 0x3d, 0xee,
						0xe8, 0x06, 0x0f, 0xdc, 0x70, 0x23, 0x0a, 0x7f,
						0x5b, 0x4a, 0xd7, 0xd7, 0xbc, 0x3e, 0x62, 0x8c,
						0xbe, 0x21, 0x9a, 0x88, 0x6b, 0x84, 0x26, 0x9e,
						0xae, 0xb8, 0x1e, 0x26, 0xb4, 0xfe, 0x01,
						0x41, // OP_DATA_65
						0x04, 0xae, 0x31, 0xc3, 0x1b, 0xf9, 0x12, 0x78,
						0xd9, 0x9b, 0x83, 0x77, 0xa3, 0x5b, 0xbc, 0xe5,
						0xb2, 0x7d, 0x9f, 0xff, 0x15, 0x45, 0x68, 0x39,
						0xe9, 0x19, 0x45, 0x3f, 0xc7, 0xb3, 0xf7, 0x21,
						0xf0, 0xba, 0x40, 0x3f, 0xf9, 0x6c, 0x9d, 0xee,
						0xb6, 0x80, 0xe5, 0xfd, 0x34, 0x1c, 0x0f, 0xc3,
						0xa7, 0xb9, 0x0d, 0xa4, 0x63, 0x1e, 0xe3, 0x95,
						0x60, 0x63, 0x9d, 0xb4, 0x62, 0xe9, 0xcb, 0x85,
						0x0f, // 65-byte pubkey
					},
					Sequence: 0xffffffff,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 0xf4240, // 1000000
					PkScript: []byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0xb0, 0xdc, 0xbf, 0x97, 0xea, 0xbf, 0x44, 0x04,
						0xe3, 0x1d, 0x95, 0x24, 0x77, 0xce, 0x82, 0x2d,
						0xad, 0xbe, 0x7e, 0x10,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					},
				},
				{
					Value: 0x11d260c0, // 299000000
					PkScript: []byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0x6b, 0x12, 0x81, 0xee, 0xc2, 0x5a, 0xb4, 0xe1,
						0xe0, 0x79, 0x3f, 0xf4, 0xe0, 0x8a, 0xb1, 0xab,
						0xb3, 0x40, 0x9c, 0xd9,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					},
				},
			},
			LockTime: 0,
		},
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						Hash: chainhash.Hash([32]byte{ // Make go vet happy.
							0x0b, 0x60, 0x72, 0xb3, 0x86, 0xd4, 0xa7, 0x73,
							0x23, 0x52, 0x37, 0xf6, 0x4c, 0x11, 0x26, 0xac,
							0x3b, 0x24, 0x0c, 0x84, 0xb9, 0x17, 0xa3, 0x90,
							0x9b, 0xa1, 0xc4, 0x3d, 0xed, 0x5f, 0x51, 0xf4,
						}), // f4515fed3dc4a19b90a317b9840c243bac26114cf637522373a7d486b372600b
						Index: 0,
					},
					SignatureScript: []byte{
						0x49, // OP_DATA_73
						0x30, 0x46, 0x02, 0x21, 0x00, 0xbb, 0x1a, 0xd2,
						0x6d, 0xf9, 0x30, 0xa5, 0x1c, 0xce, 0x11, 0x0c,
						0xf4, 0x4f, 0x7a, 0x48, 0xc3, 0xc5, 0x61, 0xfd,
						0x97, 0x75, 0x00, 0xb1, 0xae, 0x5d, 0x6b, 0x6f,
						0xd1, 0x3d, 0x0b, 0x3f, 0x4a, 0x02, 0x21, 0x00,
						0xc5, 0xb4, 0x29, 0x51, 0xac, 0xed, 0xff, 0x14,
						0xab, 0xba, 0x27, 0x36, 0xfd, 0x57, 0x4b, 0xdb,
						0x46, 0x5f, 0x3e, 0x6f, 0x8d, 0xa1, 0x2e, 0x2c,
						0x53, 0x03, 0x95, 0x4a, 0xca, 0x7f, 0x78, 0xf3,
						0x01, // 73-byte signature
						0x41, // OP_DATA_65
						0x04, 0xa7, 0x13, 0x5b, 0xfe, 0x82, 0x4c, 0x97,
						0xec, 0xc0, 0x1e, 0xc7, 0xd7, 0xe3, 0x36, 0x18,
						0x5c, 0x81, 0xe2, 0xaa, 0x2c, 0x41, 0xab, 0x17,
						0x54, 0x07, 0xc0, 0x94, 0x84, 0xce, 0x96, 0x94,
						0xb4, 0x49, 0x53, 0xfc, 0xb7, 0x51, 0x20, 0x65,
						0x64, 0xa9, 0xc2, 0x4d, 0xd0, 0x94, 0xd4, 0x2f,
						0xdb, 0xfd, 0xd5, 0xaa, 0xd3, 0xe0, 0x63, 0xce,
						0x6a, 0xf4, 0xcf, 0xaa, 0xea, 0x4e, 0xa1, 0x4f,
						0xbb, // 65-byte pubkey
					},
					Sequence: 0xffffffff,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 0xf4240, // 1000000
					PkScript: []byte{
						0x76, // OP_DUP
						0xa9, // OP_HASH160
						0x14, // OP_DATA_20
						0x39, 0xaa, 0x3d, 0x56, 0x9e, 0x06, 0xa1, 0xd7,
						0x92, 0x6d, 0xc4, 0xbe, 0x11, 0x93, 0xc9, 0x9b,
						0xf2, 0xeb, 0x9e, 0xe0,
						0x88, // OP_EQUALVERIFY
						0xac, // OP_CHECKSIG
					},
				},
			},
			LockTime: 0,
		},
	},
}
