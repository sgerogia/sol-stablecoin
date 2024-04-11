package encrypt_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sgerogia/hello-stablecoin/tpp-client/encrypt"
	"github.com/stretchr/testify/assert"
)

func TestEthSigUtilBox_UnmarshalJSON(t *testing.T) {
	// arrange
	data := `{"version":"x25519-xsalsa20-poly1305",
				"nonce":"bD15FwAp5qKHkSXVYIOWdQSTRND/I1xD",
				"ephemPublicKey":"sd3anNO1KATzFua1N8v670lkuM+q6oFDT36fMPyvrzc=",
				"ciphertext":"hdU6KW3SxMbddrCNlz6LXMwt7fqF0oY2f2ZdFd6YNRmwnQSd2BWbPs3zi6eMsoBqCPzVuPjgBWamU4Q="}`
	eth := encrypt.EthSigUtilBox{}
	// act
	json.Unmarshal([]byte(data), &eth)

	// assert
	assert.Equal(t, "x25519-xsalsa20-poly1305", eth.Version)
	assert.Equal(t, 24, len(*eth.Nonce))
	assert.NotEmpty(t, *eth.Nonce)
	assert.Equal(t, 32, len(*eth.EphemPublicKey))
	assert.NotEmpty(t, *eth.EphemPublicKey)
	assert.NotEmpty(t, *eth.Ciphertext)
}

func TestEthSigUtilBox_MarshalJSON(t *testing.T) {
	// arrange
	nonce := make([]byte, 32)
	rand.Read(nonce)
	eph := make([]byte, 32)
	rand.Read(eph)
	cipher := make([]byte, 128)
	rand.Read(cipher)
	data := encrypt.EthSigUtilBox{
		Version:        "x25519-xsalsa20-poly1305",
		Nonce:          (*[24]byte)(nonce),
		EphemPublicKey: (*[32]byte)(eph),
		Ciphertext:     &cipher}

	//act
	j, err := json.Marshal(data)

	// assert
	require.NoError(t, err)
	var tmp map[string]interface{}
	err = json.Unmarshal([]byte(j), &tmp)
	require.NoError(t, err)
	assert.Equal(t, "x25519-xsalsa20-poly1305", tmp["version"])
	assert.NotEmpty(t, tmp["nonce"])
	assert.NotEmpty(t, tmp["ephemPublicKey"])
	assert.NotEmpty(t, tmp["ciphertext"])
}
