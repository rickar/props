// (c) 2022 Rick Arnold. Licensed under the BSD license (see LICENSE).

package props

import (
	"strings"
	"testing"
)

func TestEncrypt(t *testing.T) {
	val, err := Encrypt("xyz", "password", "plaintext")
	if val != "" || err == nil {
		t.Errorf("want: '', err != nil; got: %s, %v", val, err)
	}

	val, err = Encrypt(EncryptNone, "password", "plaintext")
	if val != "[enc:0]plaintext" || err != nil {
		t.Errorf("want: '[enc:0]plaintext', err == nil; got: %s, %v", val, err)
	}

	val, err = Encrypt(EncryptAESGCM, "small", "plaintext")
	if val != "" || err == nil {
		t.Errorf("want: '', err != nil; got: %s, %v", val, err)
	}

	val, err = Encrypt(EncryptAESGCM, "1234567890123456", "plaintext")
	if len(val) < 58 || !strings.HasPrefix(val, "[enc:1]") || err != nil {
		t.Errorf("want: '[enc:1]...', err == nil; got: %s, %v", val, err)
	}
}

func TestDecrypt(t *testing.T) {
	badpass := "123"
	pass := "1234567890123456"
	pass256 := "12345678901234567890123456789012"

	val, err := Decrypt(badpass, "[enc:0]")
	if val != "" || err != nil {
		t.Errorf("want: '', err == nil; got: %s, %v", val, err)
	}

	val, err = Decrypt(badpass, "[enc:0]plaintext")
	if val != "plaintext" || err != nil {
		t.Errorf("want: 'plaintext', err == nil; got: %s, %v", val, err)
	}

	val, err = Decrypt(badpass, "[enc:1]qzp-F5Cw1G5n9c7D5LFw4NzanvZzdQzONRPlM3JpLLCO6swpBQ==")
	if val != "" || err == nil {
		t.Errorf("want: '', err != nil; got: %s, %v", val, err)
	}

	val, err = Decrypt(pass, "[enc:1]")
	if val != "" || err == nil {
		t.Errorf("want: '', err != nil; got: %s, %v", val, err)
	}

	val, err = Decrypt(pass, "[enc:1]$$$$")
	if val != "" || err == nil {
		t.Errorf("want: '', err != nil; got: %s, %v", val, err)
	}

	val, err = Decrypt(pass, "[enc:1]qzp-F5Cw1G5n9c7D5LFw4NzanvZzdQzONRPlM3JpLLCO6swpBQ==")
	if val != "plaintext" || err != nil {
		t.Errorf("want: 'plaintext', err == nil; got: %s, %v", val, err)
	}

	val, err = Decrypt(pass, "[enc:1]x-f72AdaTiZfRvX_6kekHpd4xUj1JmMFwbwiADbMZXgJjGpJNg==")
	if val != "" || err == nil {
		t.Errorf("want: '', err != nil; got: %s, %v", val, err)
	}

	val, err = Decrypt(pass256, "[enc:1]x-f72AdaTiZfRvX_6kekHpd4xUj1JmMFwbwiADbMZXgJjGpJNg==")
	if val != "plaintext" || err != nil {
		t.Errorf("want: 'plaintext', err == nil; got: %s, %v", val, err)
	}

	val, err = Decrypt(pass, "[enc:x]abcd")
	if val != "" || err == nil {
		t.Errorf("want: '', err != nil; got: %s, %v", val, err)
	}

	val, err = Decrypt(pass, "abcd")
	if val != "" || err == nil {
		t.Errorf("want: '', err != nil; got: %s, %v", val, err)
	}
}
