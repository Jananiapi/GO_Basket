// Package utility preserves non-framework helpers from CommonUtils, CodifiUtil,
// StringUtil and ValidateUtil. Legacy crypto is for interoperability only.
package utility

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
)

func SHA256(s string) string { d := sha256.Sum256([]byte(s)); return hex.EncodeToString(d[:]) }
func GenerateSHAKey(user, authCode, secret string) string {
	return strings.ToUpper(SHA256(user + authCode + secret))
}
func HashData1000(s string) string {
	d := sha256.Sum256([]byte(s))
	for i := 1; i < 1000; i++ {
		d = sha256.Sum256(d[:])
	}
	return hex.EncodeToString(d[:])
}
func RandomAlphaNumeric(n int) (string, error) {
	return randomString(n, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
}
func Random256Key() (string, error) {
	return randomString(256, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvxyz")
}
func randomString(n int, alphabet string) (string, error) {
	if n < 0 {
		return "", errors.New("negative length")
	}
	b := make([]byte, n)
	for i := range b {
		x, e := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if e != nil {
			return "", e
		}
		b[i] = alphabet[x.Int64()]
	}
	return string(b), nil
}
func NotifyID() (string, error) {
	n, e := rand.Int(rand.Reader, big.NewInt(999))
	if e != nil {
		return "", e
	}
	return fmt.Sprintf("%03d", n.Int64()), nil
}
func ExpiryInSeconds(now time.Time) int {
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return int(midnight.Sub(now) / time.Second)
}

// GenerateKeyAndIV implements the source's OpenSSL EVP_BytesToKey MD5 derivation.
func GenerateKeyAndIV(keyLen, ivLen, iterations int, salt, password []byte) ([]byte, []byte, error) {
	if keyLen < 0 || ivLen < 0 || iterations < 1 || keyLen+ivLen > 4096 {
		return nil, nil, errors.New("invalid key derivation parameters")
	}
	if salt != nil && len(salt) < 8 {
		return nil, nil, errors.New("salt must contain 8 bytes")
	}
	var all, previous []byte
	for len(all) < keyLen+ivLen {
		h := md5.New()
		h.Write(previous)
		h.Write(password)
		if salt != nil {
			h.Write(salt[:8])
		}
		previous = h.Sum(nil)
		for i := 1; i < iterations; i++ {
			d := md5.Sum(previous)
			previous = d[:]
		}
		all = append(all, previous...)
	}
	return all[:keyLen], all[keyLen : keyLen+ivLen], nil
}
func Decrypt(cipherText, secret string) (string, error) {
	b, e := base64.StdEncoding.DecodeString(cipherText)
	if e != nil {
		return "", e
	}
	if len(b) < 32 || string(b[:8]) != "Salted__" || (len(b)-16)%aes.BlockSize != 0 {
		return "", errors.New("invalid OpenSSL ciphertext")
	}
	key, iv, e := GenerateKeyAndIV(32, 16, 1, b[8:16], []byte(secret))
	if e != nil {
		return "", e
	}
	block, e := aes.NewCipher(key)
	if e != nil {
		return "", e
	}
	plain := make([]byte, len(b)-16)
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, b[16:])
	pad := int(plain[len(plain)-1])
	if pad < 1 || pad > aes.BlockSize || pad > len(plain) {
		return "", errors.New("invalid padding")
	}
	for _, p := range plain[len(plain)-pad:] {
		if int(p) != pad {
			return "", errors.New("invalid padding")
		}
	}
	return string(plain[:len(plain)-pad]), nil
}
func NullOrEmpty(s *string) bool { return s == nil || strings.TrimSpace(*s) == "" }
func Equal(a, b *string) bool    { return a != nil && b != nil && strings.EqualFold(*a, *b) }
func RemoveEmpty(xs []string) []string {
	out := []string{}
	for _, s := range xs {
		if strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}
func ReplaceChar(s string) string {
	for _, ch := range []string{"?", "*", "@", "(", ")", "'", "!", "\""} {
		s = strings.ReplaceAll(s, ch, "")
	}
	return s
}
func RestrictCharacter(s string) bool {
	const key = "sender_message_reference="
	if i := strings.LastIndex(s, key); i >= 0 {
		value := strings.SplitN(s[i+len(key):], "&", 2)[0]
		return len(value) <= 50
	}
	return true
}
func ConvertTimeByZone(s, zone string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	if len(s) == 17 {
		s = strings.ReplaceAll(s, "Z", ":00Z")
	}
	t, e := time.Parse(time.RFC3339, s)
	if e != nil {
		return s
	}
	if zone == "" {
		zone = "Asia/Dubai"
	}
	loc, e := time.LoadLocation(zone)
	if e != nil {
		loc = time.UTC
	}
	return t.In(loc).Format("2006-01-02T15:04:05")
}

var datePattern = regexp.MustCompile(`^(19|20)\d{2}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$`)

func ValidateExpiryDate(s string) error {
	if strings.TrimSpace(s) != "" && !datePattern.MatchString(s) {
		return errors.New("Invalid Date format.Date should be in the format:YYYY-MM-DD")
	}
	return nil
}
func ValidateBasketID(id *int64) error {
	if id == nil || *id == 0 {
		return errors.New("Basket id is required")
	}
	if *id < 0 {
		return errors.New("Basket id should be positive numeric ")
	}
	return nil
}

// LegacyOTP returns a caller-supplied compatibility value. The Java helper used
// the fixed value 123456; it was not called by any Basket API. No authentication
// flow uses this function.
func LegacyOTP(configuredValue string) string { return configuredValue }
func Split(s, separator string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	r, e := regexp.Compile(separator)
	if e != nil {
		return nil
	}
	parts := r.Split(s, -1)
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}
