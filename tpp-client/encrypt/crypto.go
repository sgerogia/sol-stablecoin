package encrypt

import (
	"crypto/ecdsa"
	"crypto/rand"
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/curve25519"
	"io"
	"strings"

	"golang.org/x/crypto/nacl/box"
)

// ----- WARNING! HERE BE DRAGONS! ------
// The following crypto code is for demo purposes only!
// It has not been security audited and should not be used anywhere near production!
// -----

const VERSION = "x25519-xsalsa20-poly1305"

type KeyPair struct {
	PublicEncrKey []byte
	PrivateKey    *ecdsa.PrivateKey
	PublicKey     *ecdsa.PublicKey
}

func NewKeyPairFromHex(privKey string) (*KeyPair, error) {

	if strings.HasPrefix(privKey, "0x") {
		privKey = privKey[2:]
	}

	raw, err := hex.DecodeString(privKey)
	if err != nil {
		return nil, errors.New("Error decoding hex string")
	}

	pubEncrKey, err := curve25519.X25519(raw[:], curve25519.Basepoint)
	if err != nil {
		return nil, errors.New("Error generating public encryption key")
	}

	priv, err := crypto.HexToECDSA(privKey)
	if err != nil {
		return nil, errors.New("Error decoding private key")
	}
	return &KeyPair{
		PublicEncrKey: pubEncrKey,
		PrivateKey:    priv,
		PublicKey:     &priv.PublicKey}, nil
}

func NewKeyPair() (*KeyPair, error) {
	var random [32]byte
	if _, err := io.ReadFull(rand.Reader, random[:]); err != nil {
		return nil, err
	}

	encrPubKey, err := curve25519.X25519(random[:], curve25519.Basepoint)
	if err != nil {
		return nil, errors.New("Error generating public encryption key")
	}

	signPrivKey, err := crypto.ToECDSA(random[:])
	signPubKey := &signPrivKey.PublicKey
	if err != nil {
		return nil, err
	}
	return &KeyPair{
		PublicEncrKey: encrPubKey,
		PrivateKey:    signPrivKey,
		PublicKey:     signPubKey}, nil
}

// Encrypt a given message using the counter-party's public key.
func (keyPair *KeyPair) Encrypt(msg []byte, theirPublicKey *[32]byte) (*EthSigUtilBox, error) {

	// Generate an ephemeral keypair
	ephemPubKey, ephemPrivKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, errors.New("Error generating ephemeral keypair")
	}

	// create random nonce
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}

	var encr []byte
	encr = box.Seal(encr, msg, &nonce, theirPublicKey, ephemPrivKey)

	return &EthSigUtilBox{
		Version:        VERSION,
		Nonce:          &nonce,
		EphemPublicKey: ephemPubKey,
		Ciphertext:     &encr,
	}, nil
}

// Decrypt Takes an incoming EthSigUtilBox and decrypts the cipher text, using our private key.
func (keyPair *KeyPair) Decrypt(payload *EthSigUtilBox) ([]byte, error) {

	var msg []byte
	msg, ok := box.Open(msg, *payload.Ciphertext, payload.Nonce, payload.EphemPublicKey, keyPair.PrivateKeyBytes())
	if !ok {
		return nil, errors.New("Error decrypting")
	}
	return msg, nil
}

func (keyPair *KeyPair) PublicKeyBytes() *[32]byte {
	return (*[32]byte)(crypto.FromECDSAPub(keyPair.PublicKey))
}

func (keyPair *KeyPair) PrivateKeyBytes() *[32]byte {
	return (*[32]byte)(crypto.FromECDSA(keyPair.PrivateKey))
}

func (keyPair *KeyPair) PublicEncrKeyBytes() *[32]byte {
	return (*[32]byte)(keyPair.PublicEncrKey)
}

// The Go representation of the output of eth-sig-util.encrypt. Used in unmarshalling.
type EthSigUtilBox struct {
	Version        string
	Nonce          *[24]byte
	EphemPublicKey *[32]byte
	Ciphertext     *[]byte
}

// Intermediate representation of the output of eth-sig-util.encrypt.Used in marshalling.
type ethSigUtilBoxBase64 struct {
	Version        string `json:"version"`
	Nonce          string `json:"nonce"`
	EphemPublicKey string `json:"ephemPublicKey"`
	Ciphertext     string `json:"ciphertext"`
}

// UnmarshalJSON Custom JSON unmarshalling, taking care of Base64 to []byte conversions.
func (e *EthSigUtilBox) UnmarshalJSON(data []byte) error {
	var tmp map[string]interface{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	e.Version = tmp["version"].(string)

	nonce, err := b64.StdEncoding.DecodeString(tmp["nonce"].(string))
	if err != nil {
		return err
	}
	if len(nonce) != 24 {
		return errors.New("Nonce is not 24 bytes")
	}
	e.Nonce = (*[24]byte)(nonce)

	pubKey, err := b64.StdEncoding.DecodeString(tmp["ephemPublicKey"].(string))
	if err != nil {
		return err
	}
	if len(pubKey) != 32 {
		return errors.New("Ephem. pub. key is not 32 bytes")
	}
	e.EphemPublicKey = (*[32]byte)(pubKey)

	cipher, err := b64.StdEncoding.DecodeString(tmp["ciphertext"].(string))
	if err != nil {
		return err
	}
	e.Ciphertext = &cipher

	return nil
}

// MarshalJSON Custom JSON marshalling, taking care of []byte to Base64 conversions.
func (e EthSigUtilBox) MarshalJSON() ([]byte, error) {
	tmp := ethSigUtilBoxBase64{
		Version:        e.Version,
		Nonce:          b64.StdEncoding.EncodeToString((*e.Nonce)[:]),
		EphemPublicKey: b64.StdEncoding.EncodeToString((*e.EphemPublicKey)[:]),
		Ciphertext:     b64.StdEncoding.EncodeToString(*e.Ciphertext),
	}
	return json.Marshal(tmp)
}
