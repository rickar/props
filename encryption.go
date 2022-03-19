// (c) 2022 Rick Arnold. Licensed under the BSD license (see LICENSE).

package props

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

// Decrypt returns the plaintext value of a property encrypted with the Encrypt
// function. If the property does not exist, then the default value will be
// returned with a nil error. If the property value could not be decrypted,
// then an error and the default value will be returned.
func Decrypt(password string, val string) (string, error) {
	if strings.Index(val, "]") < 0 {
		return "", fmt.Errorf("missing algorithm")
	}
	alg := val[0 : strings.Index(val, "]")+1]
	if alg == EncryptNone {
		return val[len(alg):], nil
	}

	enc, err := base64.URLEncoding.DecodeString(val[len(alg):])
	if err != nil {
		return "", err
	}
	switch alg {
	case EncryptAESGCM:
		block, err := aes.NewCipher([]byte(password))
		if err != nil {
			return "", err
		}
		// AES guarantees correct block size
		gcm, _ := cipher.NewGCM(block)
		nonceSize := gcm.NonceSize()
		if len(enc) < nonceSize+1 {
			return "", fmt.Errorf("encrypted value too small")
		}
		nonce, enc2 := enc[:nonceSize], enc[nonceSize:]
		dec, err := gcm.Open(nil, nonce, enc2, nil)
		if err != nil {
			return "", err
		}
		return string(dec), nil
	default:
		return "", fmt.Errorf("unknown algorithm")
	}
}

// Encrypt returns the value encrypted with the provided algorithm in base64
// format. If encryption fails, an empty string and error are returned.
func Encrypt(alg, password, value string) (string, error) {
	switch alg {
	case EncryptNone:
		return EncryptNone + value, nil
	case EncryptAESGCM:
		block, err := aes.NewCipher([]byte(password))
		if err != nil {
			return "", fmt.Errorf("unable to init aes encryption [%w]", err)
		}
		// AES guarantees correct block size
		gcm, _ := cipher.NewGCM(block)
		nonce := make([]byte, gcm.NonceSize())
		_, err = rand.Read(nonce)
		if err != nil {
			return "", fmt.Errorf("uanble to create nonce: [%w]", err)
		}

		enc := gcm.Seal(nonce, nonce, []byte(value), nil)
		result := base64.URLEncoding.EncodeToString(enc)
		return EncryptAESGCM + result, nil
	default:
		return "", fmt.Errorf("unknown algorithm %s", alg)
	}
}
