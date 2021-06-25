package claims

import (
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/iden3/go-iden3-core/testgen"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree"
	"github.com/stretchr/testify/assert"
)

func testClaimKeyBabyJub(t *testing.T, i, testKey string) {
	// Create new claim
	var k babyjub.PrivateKey
	hexK := testgen.GetTestValue(i + testKey).(string)
	if _, err := hex.Decode(k[:], []byte(hexK)); err != nil {
		panic(err)
	}
	pk := k.Public()
	c0 := NewClaimKeyBabyJub(pk, BabyJubKeyTypeGeneric)
	c0.Metadata().RevNonce = 5678
	assert.True(t, merkletree.CheckEntryInField(*c0.Entry()))
	e := c0.Entry()
	// Check claim against test vector
	hi, hv, err := e.HiHv()
	assert.Nil(t, err)
	testgen.CheckTestValue(t, "ClaimKeyBabyJub"+i+"_HIndex", hi.Hex())
	testgen.CheckTestValue(t, "ClaimKeyBabyJub"+i+"_HValue", hv.Hex())
	testgen.CheckTestValue(t, "ClaimKeyBabyJub"+i+"_dataString", e.Data.String())
	dataTestOutput(&e.Data)
	c1 := NewClaimKeyBabyJubFromEntry(e)
	c2, err := NewClaimFromEntry(e)
	assert.Nil(t, err)
	assert.Equal(t, c0, c1)
	assert.Equal(t, c0.Metadata(), c1.Metadata())
	assert.Equal(t, c0, c2)
	assert.True(t, merkletree.CheckEntryInField(*e))
	assert.Equal(t, c0.KeyType, c1.KeyType)

	// check Type
	ct1 := ClaimType{}
	// TODO: Update to LittleEndian once claimtype is little endian
	binary.BigEndian.PutUint64(ct1[:], 1)
	assert.Equal(t, ct1, c0.Metadata().header.Type)

	// check KeyType
	c3 := NewClaimKeyBabyJub(pk, BabyJubKeyTypeGeneric)
	c4 := NewClaimKeyBabyJub(pk, BabyJubKeyTypeAuthorizeKSign)
	assert.NotEqual(t, c3.KeyType, c4.KeyType)
	assert.Equal(t, c0.KeyType, c3.KeyType)
	assert.Equal(t, BabyJubKeyTypeGeneric, c3.KeyType)
	assert.Equal(t, BabyJubKeyTypeAuthorizeKSign, c4.KeyType)
}

func TestClaimKeyBabyJub(t *testing.T) {
	testClaimKeyBabyJub(t, "0", "_privateKey")
	testClaimKeyBabyJub(t, "1", "_privateKey")
}

func TestRandomClaimKeyBabyJub(t *testing.T) {
	for i := 0; i < 100; i++ {
		k := babyjub.NewRandPrivKey()
		pk := k.Public()

		c0 := NewClaimKeyBabyJub(pk, BabyJubKeyTypeGeneric)
		assert.True(t, merkletree.CheckEntryInField(*c0.Entry()))
		e := c0.Entry()
		c1 := NewClaimKeyBabyJubFromEntry(e)
		c2, err := NewClaimFromEntry(e)
		assert.Nil(t, err)
		assert.Equal(t, c0, c1)
		assert.Equal(t, c0, c2)
		assert.True(t, merkletree.CheckEntryInField(*e))
	}
}
