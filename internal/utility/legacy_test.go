package utility

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"testing"
	"time"
)

func TestLegacyDecryptAndHash(t *testing.T) {
	salt := []byte("12345678")
	key, iv, e := GenerateKeyAndIV(32, 16, 1, salt, []byte("secret"))
	if e != nil {
		t.Fatal(e)
	}
	plain := []byte("basket")
	for len(plain) < 16 {
		plain = append(plain, 10)
	}
	block, _ := aes.NewCipher(key)
	encrypted := make([]byte, 16)
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(encrypted, plain)
	data := append(append([]byte("Salted__"), salt...), encrypted...)
	s, e := Decrypt(base64.StdEncoding.EncodeToString(data), "secret")
	if e != nil || s != "basket" {
		t.Fatal(s, e)
	}
	if SHA256("abc") != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatal("SHA256 mismatch")
	}
	if _, e = Decrypt("invalid", "secret"); e == nil {
		t.Fatal("invalid ciphertext accepted")
	}
}
func TestValidationAndExpiry(t *testing.T) {
	now := time.Date(2026, 9, 10, 23, 59, 0, 0, time.UTC)
	if ExpiryInSeconds(now) != 60 {
		t.Fatal("expiry mismatch")
	}
	if ValidateExpiryDate("10-09-2026") == nil || ValidateExpiryDate("2026-09-10") != nil {
		t.Fatal("date validation mismatch")
	}
	if ReplaceChar("a?b@c") != "abc" {
		t.Fatal("character replacement")
	}
}
