package javaser

import (
	"basket/internal/model"
	"bytes"
	"os"
	"testing"
)

func decodeBytes(b []byte) (any, error) {
	r := bytes.NewReader(b)
	return Decode(func() byte {
		b, e := r.ReadByte()
		if e != nil {
			panic(e)
		}
		return b
	})
}
func TestOriginalJavaContract(t *testing.T) {
	b, e := os.ReadFile("testdata/contract.ser")
	if e != nil {
		t.Fatal(e)
	}
	v, e := decodeBytes(b)
	if e != nil {
		t.Fatal(e)
	}
	m := v.(map[string]any)
	if m["token"] != "47310" || m["lotSize"] != "65" || m["expiry"] != int64(1789410600000) || m["companyName"] != "Test 😀\x00" {
		t.Fatal(m)
	}
}
func TestOriginalJavaCustomer(t *testing.T) {
	b, e := os.ReadFile("testdata/customer.ser")
	if e != nil {
		t.Fatal(e)
	}
	v, e := decodeBytes(b)
	if e != nil {
		t.Fatal(e)
	}
	m := v.(map[string]any)
	if m["stringPkey4"] != "fixture-key" || m["tomcatcount"] != "3" {
		t.Fatal("customer credentials mismatch")
	}
	settings := m["userSettingDto"].(map[string]any)
	if settings["s_prdt_ali"] != "CNC:CNC" {
		t.Fatal("nested login settings mismatch")
	}
}
func TestDeviceEncoding(t *testing.T) {
	user := "USER1"
	b, e := EncodeDevices([]model.DeviceMappingEntity{{CommonEntity: model.CommonEntity{Id: 7, ActiveStatus: 1, CreatedOn: model.Now()}, UserId: &user}})
	if e != nil {
		t.Fatal(e)
	}
	v, e := decodeBytes(b)
	if e != nil {
		t.Fatal(e)
	}
	a := v.([]any)
	if len(a) != 1 || a[0].(map[string]any)["userId"] != "USER1" {
		t.Fatal(v)
	}
}
func TestMalformedStream(t *testing.T) {
	for _, b := range [][]byte{nil, {0xac, 0xed, 0, 5, 0x71, 0, 0, 0, 0}, {0xac, 0xed, 0, 5, 0x75}} {
		if _, e := decodeBytes(b); e == nil {
			t.Fatal("expected decoder error")
		}
	}
}
