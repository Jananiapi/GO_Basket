package utility

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

func TestCryptoHashing(t *testing.T) {
	// SHA256
	h := SHA256("test_string")
	if len(h) != 64 {
		t.Fatalf("expected 64 char hex string, got %d chars: %s", len(h), h)
	}

	// GenerateSHAKey
	key := GenerateSHAKey("USER1", "AUTH123", "SECRET")
	if len(key) != 64 || key != strings.ToUpper(key) {
		t.Fatalf("expected uppercase 64 char hex key, got %s", key)
	}

	// HashData1000
	h1000 := HashData1000("test_data")
	if len(h1000) != 64 {
		t.Fatalf("expected 64 char hex string from HashData1000, got %s", h1000)
	}
}

func TestRandomGeneration(t *testing.T) {
	// RandomAlphaNumeric
	s, err := RandomAlphaNumeric(16)
	if err != nil || len(s) != 16 {
		t.Fatalf("RandomAlphaNumeric(16) failed: %v, got %s", err, s)
	}

	// Random256Key
	k, err := Random256Key()
	if err != nil || len(k) != 256 {
		t.Fatalf("Random256Key failed: %v, got length %d", err, len(k))
	}

	// Negative length
	if _, err := randomString(-5, "ABC"); err == nil {
		t.Fatal("expected error for negative length in randomString, got nil")
	}

	// NotifyID
	nid, err := NotifyID()
	if err != nil || len(nid) != 3 {
		t.Fatalf("NotifyID failed: %v, got %s", err, nid)
	}
}

func TestExpiryInSeconds(t *testing.T) {
	now := time.Date(2026, 9, 21, 23, 59, 50, 0, time.UTC)
	exp := ExpiryInSeconds(now)
	if exp != 10 {
		t.Fatalf("expected 10 seconds to midnight, got %d", exp)
	}
}

func TestGenerateKeyAndIVErrors(t *testing.T) {
	// Invalid parameters
	if _, _, err := GenerateKeyAndIV(-1, 16, 1, nil, []byte("pass")); err == nil {
		t.Fatal("expected error for keyLen < 0")
	}
	if _, _, err := GenerateKeyAndIV(32, -1, 1, nil, []byte("pass")); err == nil {
		t.Fatal("expected error for ivLen < 0")
	}
	if _, _, err := GenerateKeyAndIV(32, 16, 0, nil, []byte("pass")); err == nil {
		t.Fatal("expected error for iterations < 1")
	}
	if _, _, err := GenerateKeyAndIV(3000, 2000, 1, nil, []byte("pass")); err == nil {
		t.Fatal("expected error for keyLen+ivLen > 4096")
	}
	if _, _, err := GenerateKeyAndIV(32, 16, 1, []byte("short"), []byte("pass")); err == nil {
		t.Fatal("expected error for salt < 8 bytes")
	}

	// Valid key & IV with 8-byte salt
	salt := []byte("12345678")
	key, iv, err := GenerateKeyAndIV(32, 16, 2, salt, []byte("password"))
	if err != nil {
		t.Fatalf("GenerateKeyAndIV failed: %v", err)
	}
	if len(key) != 32 || len(iv) != 16 {
		t.Fatalf("expected 32 key and 16 iv, got %d, %d", len(key), len(iv))
	}
}

func TestDecryptBranches(t *testing.T) {
	secret := "my_secret_pass"
	plaintext := "Hello Secure World!"

	// 1. Build a valid OpenSSL formatted ciphertext
	salt := []byte("8byteslt")
	key, iv, err := GenerateKeyAndIV(32, 16, 1, salt, []byte(secret))
	if err != nil {
		t.Fatal(err)
	}

	// PKCS7 padding
	padLen := aes.BlockSize - (len(plaintext) % aes.BlockSize)
	padded := append([]byte(plaintext), make([]byte, padLen)...)
	for i := len(plaintext); i < len(padded); i++ {
		padded[i] = byte(padLen)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	encrypted := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(encrypted, padded)

	// Combine: "Salted__" + salt + encrypted
	fullPayload := append([]byte("Salted__"), salt...)
	fullPayload = append(fullPayload, encrypted...)
	b64Cipher := base64.StdEncoding.EncodeToString(fullPayload)

	// Decrypt valid
	dec, err := Decrypt(b64Cipher, secret)
	if err != nil {
		t.Fatalf("Decrypt valid failed: %v", err)
	}
	if dec != plaintext {
		t.Fatalf("expected %q, got %q", plaintext, dec)
	}

	// Invalid Base64
	if _, err := Decrypt("not-valid-base64!@#$", secret); err == nil {
		t.Fatal("expected error for invalid base64")
	}

	// Ciphertext too short (< 32 bytes)
	shortB64 := base64.StdEncoding.EncodeToString([]byte("short"))
	if _, err := Decrypt(shortB64, secret); err == nil {
		t.Fatal("expected error for ciphertext < 32 bytes")
	}

	// Missing "Salted__" header
	noSaltHeader := make([]byte, 48)
	copy(noSaltHeader, []byte("NotSaltd"))
	if _, err := Decrypt(base64.StdEncoding.EncodeToString(noSaltHeader), secret); err == nil {
		t.Fatal("expected error for missing Salted__ header")
	}

	// Invalid length not multiple of aes.BlockSize
	badLen := make([]byte, 45)
	copy(badLen, []byte("Salted__"))
	if _, err := Decrypt(base64.StdEncoding.EncodeToString(badLen), secret); err == nil {
		t.Fatal("expected error for length not multiple of BlockSize")
	}

	// Invalid padding (tampered last byte)
	tamperedPayload := make([]byte, len(fullPayload))
	copy(tamperedPayload, fullPayload)
	// Make bad padding
	tamperedPayload[len(tamperedPayload)-1] = 0 // invalid pad value
	if _, err := Decrypt(base64.StdEncoding.EncodeToString(tamperedPayload), secret); err == nil {
		t.Fatal("expected error for invalid padding byte 0")
	}
}

func TestStringAndValidationHelpers(t *testing.T) {
	// NullOrEmpty
	var nilStr *string
	emptyStr := ""
	whitespaceStr := "   "
	validStr := "hello"
	if !NullOrEmpty(nilStr) || !NullOrEmpty(&emptyStr) || !NullOrEmpty(&whitespaceStr) {
		t.Fatal("expected true for nil or empty string")
	}
	if NullOrEmpty(&validStr) {
		t.Fatal("expected false for valid string")
	}

	// Equal
	strA := "Hello"
	strB := "hello"
	strC := "world"
	if !Equal(&strA, &strB) {
		t.Fatal("expected case-insensitive Equal to be true")
	}
	if Equal(&strA, &strC) || Equal(&strA, nil) || Equal(nil, &strB) || Equal(nil, nil) {
		t.Fatal("expected false for mismatched or nil Equal calls")
	}

	// RemoveEmpty
	input := []string{"  ", "alpha", "", "beta", "   "}
	cleaned := RemoveEmpty(input)
	if len(cleaned) != 2 || cleaned[0] != "alpha" || cleaned[1] != "beta" {
		t.Fatalf("RemoveEmpty failed, got %v", cleaned)
	}

	// ReplaceChar
	rawPunct := `Hello?*@(World)'!"`
	replaced := ReplaceChar(rawPunct)
	if replaced != "HelloWorld" {
		t.Fatalf("ReplaceChar failed, got %q", replaced)
	}

	// RestrictCharacter
	validRef := "https://api.com?sender_message_reference=SHORT_REF&other=123"
	if !RestrictCharacter(validRef) {
		t.Fatal("expected true for short reference")
	}
	longRef := "https://api.com?sender_message_reference=" + strings.Repeat("A", 55)
	if RestrictCharacter(longRef) {
		t.Fatal("expected false for reference exceeding 50 chars")
	}
	noRef := "https://api.com?param=value"
	if !RestrictCharacter(noRef) {
		t.Fatal("expected true when parameter is absent")
	}
}

func TestConvertTimeByZone(t *testing.T) {
	if ConvertTimeByZone("", "Asia/Kolkata") != "" {
		t.Fatal("expected empty string for empty input")
	}

	// 17-char timestamp (e.g. 2026-09-21T10:00Z)
	shortTime := "2026-09-21T10:00Z"
	converted := ConvertTimeByZone(shortTime, "UTC")
	if !strings.HasPrefix(converted, "2026-09-21T10:00:00") {
		t.Fatalf("unexpected result for 17-char time: %s", converted)
	}

	// Invalid RFC3339 format returns original string
	invalidTime := "not-a-valid-time"
	if ConvertTimeByZone(invalidTime, "UTC") != invalidTime {
		t.Fatalf("expected original string returned on error, got %s", ConvertTimeByZone(invalidTime, "UTC"))
	}

	// Empty zone defaults to Asia/Dubai
	validRFC := "2026-09-21T10:00:00Z"
	dubaiTime := ConvertTimeByZone(validRFC, "")
	if dubaiTime == "" {
		t.Fatal("expected formatted time for empty zone")
	}

	// Invalid timezone location defaults to UTC
	utcTime := ConvertTimeByZone(validRFC, "NonExistent/Timezone")
	if !strings.HasPrefix(utcTime, "2026-09-21T10:00:00") {
		t.Fatalf("expected UTC fallback time, got %s", utcTime)
	}
}

func TestValidateExpiryDate(t *testing.T) {
	if err := ValidateExpiryDate(""); err != nil {
		t.Fatal("expected nil for empty expiry date")
	}
	if err := ValidateExpiryDate("2026-09-21"); err != nil {
		t.Fatalf("expected nil for valid date, got %v", err)
	}
	if err := ValidateExpiryDate("21-09-2026"); err == nil {
		t.Fatal("expected error for invalid date format")
	}
}

func TestValidateBasketID(t *testing.T) {
	var zero int64 = 0
	var neg int64 = -1
	var pos int64 = 42

	if err := ValidateBasketID(nil); err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("expected required error for nil basket ID, got %v", err)
	}
	if err := ValidateBasketID(&zero); err == nil || !strings.Contains(err.Error(), "required") {
		t.Fatalf("expected required error for zero basket ID, got %v", err)
	}
	if err := ValidateBasketID(&neg); err == nil || !strings.Contains(err.Error(), "positive") {
		t.Fatalf("expected positive error for negative basket ID, got %v", err)
	}
	if err := ValidateBasketID(&pos); err != nil {
		t.Fatalf("expected nil for valid positive basket ID, got %v", err)
	}
}

func TestLegacyOTPAndSplit(t *testing.T) {
	if LegacyOTP("custom_123") != "custom_123" {
		t.Fatal("LegacyOTP did not return supplied value")
	}

	if Split("", ",") != nil {
		t.Fatal("expected nil for empty string in Split")
	}
	if Split("a,b,c", "[invalid-regex") != nil {
		t.Fatal("expected nil for invalid regex in Split")
	}
	parts := Split("one,two,three,,,", ",")
	if len(parts) != 3 || parts[0] != "one" || parts[2] != "three" {
		t.Fatalf("expected trimmed trailing empty elements, got %v", parts)
	}
}
