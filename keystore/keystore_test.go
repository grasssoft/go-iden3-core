package keystore

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/iden3/go-iden3-core/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	data := []byte("Top secret")
	pass := []byte("my passphrase")
	encData, err := EncryptData(data, pass, LightScryptN, LightScryptP)
	assert.Equal(t, nil, err)
	data1, err := DecryptData(encData, pass)
	assert.Equal(t, nil, err)
	assert.Equal(t, data, data1)
}

func TestKeyStore(t *testing.T) {
	pass := []byte("my passphrase")
	storage := MemStorage([]byte{})
	ks, err := NewKeyStore(&storage, LightKeyStoreParams)
	assert.Equal(t, nil, err)
	pk, err := ks.NewKey(pass)
	assert.Equal(t, nil, err)
	fmt.Println("pk", common.Hex(pk[:]))
	fmt.Printf("encryptedKeys %+v\n", ks.encryptedKeys)
	fmt.Println("storage", string(storage))
	fmt.Println("keys", common.Hex(ks.Keys()[0][:]))

	// Unlock key
	err = ks.UnlockKey(pk, pass)
	assert.Equal(t, nil, err)

	// Make a new key store with the same storage
	ks1, err := NewKeyStore(&storage, LightKeyStoreParams)
	assert.Equal(t, nil, err)
	assert.Equal(t, ks.encryptedKeys, ks1.encryptedKeys)

	// Import a key
	storage2 := MemStorage([]byte{})
	ks2, err := NewKeyStore(&storage2, LightKeyStoreParams)
	assert.Equal(t, nil, err)
	_, err = ks2.ImportKey(*ks.cache[*pk], pass)
	assert.Equal(t, nil, err)
	assert.Equal(t, ks.Keys(), ks2.Keys())
}

func TestSignVerify(t *testing.T) {
	pass := []byte("my passphrase")
	msg := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
	storage := MemStorage([]byte{})
	ks, err := NewKeyStore(&storage, LightKeyStoreParams)
	assert.Equal(t, nil, err)
	pk, err := ks.NewKey(pass)
	assert.Equal(t, nil, err)
	var sk [32]byte
	_, err = hex.Decode(sk[:], []byte("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"))
	require.Nil(t, err)

	_, err = ks.ImportKey(sk, pass)
	assert.Equal(t, nil, err)

	if err := ks.UnlockKey(pk, pass); err != nil {
		panic(err)
	}
	sig, date, err := ks.Sign(pk, PrefixMinorUpdate, msg)
	assert.Equal(t, nil, err)
	ok, err := VerifySignature(pk, sig, PrefixMinorUpdate, date, msg)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, ok)
	// check that with a different date the verification gives error
	ok, err = VerifySignature(pk, sig, PrefixMinorUpdate, date+1, msg)
	assert.Equal(t, nil, err)
	assert.Equal(t, false, ok)
}
