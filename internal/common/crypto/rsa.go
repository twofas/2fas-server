package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"

	"github.com/pkg/errors"
)

func PublicKeyToBase64(key *rsa.PublicKey) string {
	keyAsStr := ExportRsaPublicKeyAsPemStr(key)

	b64Key := base64.StdEncoding.EncodeToString([]byte(keyAsStr))

	return b64Key
}

func PrivateKeyToBase64(key *rsa.PrivateKey) string {
	keyAsStr := ExportRsaPrivateKeyAsPemStr(key)

	b64Key := base64.StdEncoding.EncodeToString([]byte(keyAsStr))

	return b64Key
}

func Base64ToPublicKey(b64Key string) (*rsa.PublicKey, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(b64Key)

	if err != nil {
		return nil, err
	}

	publicKey, err := BytesToPublicKey(publicKeyBytes)

	if err != nil {
		return nil, err
	}

	return publicKey, nil
}

func Base64ToPrivateKey(b64Key string) (*rsa.PrivateKey, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(b64Key)

	if err != nil {
		return nil, err
	}

	privateKey, err := BytesToPrivateKey(privateKeyBytes)

	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func BytesToPrivateKey(priv []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(priv)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes

	var err error

	if enc {
		b, err = x509.DecryptPEMBlock(block, nil)

		if err != nil {
			return nil, err
		}
	}

	key, err := x509.ParsePKCS1PrivateKey(b)

	if err != nil {
		return nil, err
	}

	return key, nil
}

func BytesToPublicKey(pub []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pub)
	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes

	var err error

	if enc {
		b, err = x509.DecryptPEMBlock(block, nil)

		if err != nil {
			return nil, err
		}
	}

	ifc, err := x509.ParsePKIXPublicKey(b)

	if err != nil {
		return nil, err
	}

	key, ok := ifc.(*rsa.PublicKey)

	if !ok {
		return nil, errors.New("Something went wrong")
	}

	return key, nil
}

func EncryptWithPublicKey(publicKey *rsa.PublicKey, text []byte) ([]byte, error) {
	hash := sha512.New()

	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, publicKey, text, nil)

	if err != nil {
		return []byte{}, err
	}

	return ciphertext, nil
}

func DecryptWithPrivateKey(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	hash := sha512.New()

	text, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, ciphertext, nil)

	if err != nil {
		return []byte{}, err
	}

	return text, nil
}

type KeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

func GenerateKeyPair(bits int) *KeyPair {
	key, _ := rsa.GenerateKey(rand.Reader, bits)

	return &KeyPair{
		PrivateKey: key,
		PublicKey:  &key.PublicKey,
	}
}

func ExportRsaPrivateKeyAsPemStr(privkey *rsa.PrivateKey) string {
	privkey_bytes := x509.MarshalPKCS1PrivateKey(privkey)

	privkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privkey_bytes,
		},
	)

	return string(privkey_pem)
}

func ParseRsaPrivateKeyFromPemStr(privPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privPEM))

	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	return priv, nil
}

func ExportRsaPublicKeyAsPemStr(pubkey *rsa.PublicKey) string {
	pubkey_bytes, err := x509.MarshalPKIXPublicKey(pubkey)

	if err != nil {
		return ""
	}

	pubkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubkey_bytes,
		},
	)

	return string(pubkey_pem)
}

func ParseRsaPublicKeyFromPemStr(pubPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubPEM))

	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		break
	}
	return nil, errors.New("Key type is not RSA")
}
