package encrypt_test

import (
	"encoding/json"
	"github.com/sgerogia/hello-stablecoin/tpp-client/encrypt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type TestData struct {
	Field1 string
	Field2 string
	Field3 string
}

func TestKeyPair_Decrypt(t *testing.T) {
	// arrange
	// ...payload
	data := `{"version":"x25519-xsalsa20-poly1305",
				"nonce":"bD15FwAp5qKHkSXVYIOWdQSTRND/I1xD",
				"ephemPublicKey":"sd3anNO1KATzFua1N8v670lkuM+q6oFDT36fMPyvrzc=",
				"ciphertext":"hdU6KW3SxMbddrCNlz6LXMwt7fqF0oY2f2ZdFd6YNRmwnQSd2BWbPs3zi6eMsoBqCPzVuPjgBWamU4Q="}`
	var eth encrypt.EthSigUtilBox
	require.NoError(t, json.Unmarshal([]byte(data), &eth))
	// ...private key
	key := "13b7285298fd5115015d69f190146a7b9347801127841d4de9bb93b358a6b620"
	keyPair, err := encrypt.NewKeyPairFromHex(key)
	require.NoError(t, err)

	// act
	res, err := keyPair.Decrypt(&eth)

	// assert
	require.NoError(t, err)
	assert.Equal(t, `{"url":"https://bank.com/some-url/consent"}`, string(res))
}

func TestCrypto_Encrypt(t *testing.T) {
	// arrange
	theirKey := "55635c799aebb4ef0ce8776e82d6fc26db2b7f936a7a521777f465466236f598"
	theirKeyPair, err := encrypt.NewKeyPairFromHex(theirKey)
	require.NoError(t, err)
	ourKey := "13b7285298fd5115015d69f190146a7b9347801127841d4de9bb93b358a6b620"
	ourKeyPair, err := encrypt.NewKeyPairFromHex(ourKey)
	require.NoError(t, err)
	// ...data to encrypt
	data := TestData{
		Field1: "Hello world",
		Field2: "Some more",
		Field3: "Hey there!",
	}
	dataStr, _ := json.Marshal(data)

	// act
	encr, err := ourKeyPair.Encrypt(dataStr, theirKeyPair.PublicEncrKeyBytes())

	// assert
	require.NoError(t, err)
	require.NotEmpty(t, encr.Version)
	assert.Equal(t, encrypt.VERSION, encr.Version)
	require.NotEmpty(t, encr.Nonce)
	require.NotEmpty(t, encr.EphemPublicKey)
	require.NotEmpty(t, encr.Ciphertext)

	// FIXME: The following assertion fails. Why?
	// ...and verify decryption compatibility
	decr, err := theirKeyPair.Decrypt(encr)
	require.NoError(t, err)
	assert.Equal(t, string(dataStr), string(decr))
}
