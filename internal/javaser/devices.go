package javaser

import (
	"basket/internal/model"
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

// EncodeDevices writes the Java ArrayList<DeviceMappingEntity> value used by
// Hazelcast's deviceMappingDetails map, including CommonEntity inheritance.
func EncodeDevices(devices []model.DeviceMappingEntity) ([]byte, error) {
	w := writer{}
	w.u16(0xaced)
	w.u16(5)
	w.byte(0x73)
	list := &class{name: "java.util.ArrayList", uid: 8683452581122892189, flags: 3, fields: []field{{'I', "size"}}}
	w.desc(list)
	w.u32(uint32(len(devices)))
	w.byte(0x77)
	w.byte(4)
	w.u32(uint32(len(devices)))
	common := &class{name: "in.codifi.basket.entity.primary.CommonEntity", uid: 1, flags: 2, fields: []field{{'I', "activeStatus"}, {'J', "id"}, {'L', "createdBy"}, {'L', "createdOn"}, {'L', "updatedBy"}, {'L', "updatedOn"}}}
	device := &class{name: "in.codifi.basket.entity.primary.DeviceMappingEntity", uid: 1, flags: 2, super: common, fields: []field{{'L', "deviceId"}, {'L', "deviceType"}, {'L', "userId"}, {'L', "userName"}}}
	for _, d := range devices {
		w.byte(0x73)
		w.desc(device)
		w.u32(uint32(d.ActiveStatus))
		w.u64(uint64(d.Id))
		w.stringPtr(d.CreatedBy)
		w.date(d.CreatedOn)
		w.stringPtr(d.UpdatedBy)
		w.date(d.UpdatedOn)
		w.stringPtr(d.DeviceId)
		w.stringPtr(d.DeviceType)
		w.stringPtr(d.UserId)
		w.stringPtr(d.UserName)
	}
	w.byte(0x78)
	return w.b, w.err
}

type writer struct {
	b   []byte
	err error
}

func (w *writer) byte(v byte)  { w.b = append(w.b, v) }
func (w *writer) u16(v uint16) { w.b = binary.BigEndian.AppendUint16(w.b, v) }
func (w *writer) u32(v uint32) { w.b = binary.BigEndian.AppendUint32(w.b, v) }
func (w *writer) u64(v uint64) { w.b = binary.BigEndian.AppendUint64(w.b, v) }
func (w *writer) utf(s string) {
	b := []byte{}
	for _, r := range utf16.Encode([]rune(s)) {
		switch {
		case r > 0 && r < 128:
			b = append(b, byte(r))
		case r < 2048:
			b = append(b, 0xc0|byte(r>>6), 0x80|byte(r&63))
		default:
			b = append(b, 0xe0|byte(r>>12), 0x80|byte(r>>6&63), 0x80|byte(r&63))
		}
	}
	if len(b) > 65535 {
		w.err = fmt.Errorf("Java short UTF exceeds 65535 bytes")
		return
	}
	w.u16(uint16(len(b)))
	w.b = append(w.b, b...)
}
func (w *writer) stringPtr(s *string) {
	if s == nil {
		w.byte(0x70)
		return
	}
	w.byte(0x74)
	w.utf(*s)
}
func (w *writer) desc(c *class) {
	if c == nil {
		w.byte(0x70)
		return
	}
	w.byte(0x72)
	w.utf(c.name)
	w.u64(c.uid)
	w.byte(c.flags)
	w.u16(uint16(len(c.fields)))
	for _, f := range c.fields {
		w.byte(f.kind)
		w.utf(f.name)
		if f.kind == 'L' {
			kind := "Ljava/lang/String;"
			if f.name == "createdOn" || f.name == "updatedOn" {
				kind = "Ljava/util/Date;"
			}
			w.stringPtr(&kind)
		}
	}
	w.byte(0x78)
	w.desc(c.super)
}
func (w *writer) date(t *model.Timestamp) {
	if t == nil {
		w.byte(0x70)
		return
	}
	w.byte(0x73)
	w.desc(&class{name: "java.util.Date", uid: 7523967970034938905, flags: 3})
	w.byte(0x77)
	w.byte(8)
	w.u64(uint64(t.Time().UnixMilli()))
	w.byte(0x78)
}
