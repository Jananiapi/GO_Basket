// Package javaser reads the Java Object Serialization stream format without
// loading classes or executing readObject methods. It is limited to data values
// needed by the supplied Basket module and fails on unsupported stream features.
package javaser

import (
	"encoding/binary"
	"fmt"
	"math"
	"unicode/utf16"
)

type field struct {
	kind byte
	name string
}
type class struct {
	name   string
	uid    uint64
	flags  byte
	fields []field
	super  *class
}
type object struct {
	class  *class
	fields map[string]any
	extra  map[string][]any
}
type block []byte
type reader struct {
	read         func() byte
	handles      []any
	depth, count int
}

func Decode(read func() byte) (v any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Java serialization: %v", r)
		}
	}()
	d := &reader{read: read}
	if d.u16() != 0xaced || d.u16() != 5 {
		return nil, fmt.Errorf("invalid Java serialization header")
	}
	raw := d.value(d.byte())
	return plain(raw, 0), nil
}
func (d *reader) byte() byte {
	d.count++
	if d.count > 16<<20 {
		panic("stream exceeds 16 MiB")
	}
	return d.read()
}
func (d *reader) bytes(n int) []byte {
	if n < 0 || n > 16<<20 {
		panic("invalid length")
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = d.byte()
	}
	return b
}
func (d *reader) u16() uint16 { return binary.BigEndian.Uint16(d.bytes(2)) }
func (d *reader) u32() uint32 { return binary.BigEndian.Uint32(d.bytes(4)) }
func (d *reader) u64() uint64 { return binary.BigEndian.Uint64(d.bytes(8)) }
func (d *reader) add(v any) {
	if len(d.handles) > 1<<20 {
		panic("too many references")
	}
	d.handles = append(d.handles, v)
}
func (d *reader) utf(n int) string {
	b := d.bytes(n)
	units := []uint16{}
	for i := 0; i < len(b); {
		c := b[i]
		switch {
		case c < 0x80:
			units = append(units, uint16(c))
			i++
		case c&0xe0 == 0xc0:
			if i+1 >= len(b) {
				panic("truncated UTF")
			}
			units = append(units, uint16(c&31)<<6|uint16(b[i+1]&63))
			i += 2
		case c&0xf0 == 0xe0:
			if i+2 >= len(b) {
				panic("truncated UTF")
			}
			units = append(units, uint16(c&15)<<12|uint16(b[i+1]&63)<<6|uint16(b[i+2]&63))
			i += 3
		default:
			panic("invalid modified UTF-8")
		}
	}
	return string(utf16.Decode(units))
}
func (d *reader) shortUTF() string { return d.utf(int(d.u16())) }
func (d *reader) annotations() []any {
	a := []any{}
	for {
		tag := d.byte()
		if tag == 0x78 {
			return a
		}
		a = append(a, d.value(tag))
	}
}
func (d *reader) descriptor() *class {
	v := d.value(d.byte())
	if v == nil {
		return nil
	}
	c, ok := v.(*class)
	if !ok {
		panic("expected class descriptor")
	}
	return c
}
func (d *reader) primitive(kind byte) any {
	switch kind {
	case 'B':
		return int8(d.byte())
	case 'Z':
		return d.byte() != 0
	case 'S':
		return int16(d.u16())
	case 'C':
		return string(rune(d.u16()))
	case 'I':
		return int32(d.u32())
	case 'J':
		return int64(d.u64())
	case 'F':
		return math.Float32frombits(d.u32())
	case 'D':
		return math.Float64frombits(d.u64())
	case 'L', '[':
		return d.value(d.byte())
	default:
		panic("unsupported field type")
	}
}
/*
func (d *reader) value(tag byte) any {
	d.depth++
	defer func() { d.depth-- }()
	if d.depth > 128 {
		panic("object nesting limit")
	}
	switch tag {
	case 0x70:
		return nil
	case 0x71:
		i := int(d.u32()) - 0x7e0000
		if i < 0 || i >= len(d.handles) {
			panic("invalid reference")
		}
		return d.handles[i]
	case 0x72:
		c := &class{name: d.shortUTF(), uid: d.u64()}
		d.add(c)
		c.flags = d.byte()
		n := int(d.u16())
		for i := 0; i < n; i++ {
			f := field{kind: d.byte(), name: d.shortUTF()}
			if f.kind == 'L' || f.kind == '[' {
				d.value(d.byte())
			}
			c.fields = append(c.fields, f)
		}
		d.annotations()
		c.super = d.descriptor()
		return c
	case 0x73:
		c := d.descriptor()
		if c == nil {
			panic("object without class")
		}
		o := &object{class: c, fields: map[string]any{}, extra: map[string][]any{}}
		d.add(o)
		chain := []*class{}
		for x := c; x != nil; x = x.super {
			chain = append(chain, x)
			if len(chain) > 128 {
				panic("class hierarchy limit")
			}
		}
		for i := len(chain) - 1; i >= 0; i-- {
			x := chain[i]
			if x.flags&4 != 0 {
				panic("Externalizable values are unsupported")
			}
			for _, f := range x.fields {
				o.fields[f.name] = d.primitive(f.kind)
			}
			if x.flags&1 != 0 {
				o.extra[x.name] = d.annotations()
			}
		}
		return o
	case 0x74:
		s := d.shortUTF()
		d.add(s)
		return s
	case 0x7c:
		n := d.u64()
		if n > 16<<20 {
			panic("string limit")
		}
		s := d.utf(int(n))
		d.add(s)
		return s
	case 0x75:
		c := d.descriptor()
		if c == nil || len(c.name) < 2 {
			panic("invalid array class")
		}
		n := int(d.u32())
		if n < 0 || n > 1<<20 {
			panic("array limit")
		}
		a := make([]any, n)
		d.add(a)
		for i := range a {
			a[i] = d.primitive(c.name[1])
		}
		return a
	case 0x76:
		c := d.descriptor()
		d.add(c)
		return c
	case 0x77:
		return block(d.bytes(int(d.byte())))
	case 0x7a:
		return block(d.bytes(int(d.u32())))
	case 0x7e:
		c := d.descriptor()
		o := &object{class: c, fields: map[string]any{}}
		d.add(o)
		o.fields["name"] = d.value(d.byte())
		return o
	case 0x79:
		d.handles = nil
		return d.value(d.byte())
	default:
		panic(fmt.Sprintf("unsupported tag 0x%x", tag))
	}
}
*/

// added for sonarqube
func (d *reader) decodeReference() any {
	i := int(d.u32()) - 0x7e0000
	if i < 0 || i >= len(d.handles) {
		panic("invalid reference")
	}
	return d.handles[i]
}

// added for sonarqube
func (d *reader) decodeClassDesc() any {
	c := &class{name: d.shortUTF(), uid: d.u64()}
	d.add(c)
	c.flags = d.byte()
	n := int(d.u16())
	for i := 0; i < n; i++ {
		f := field{kind: d.byte(), name: d.shortUTF()}
		if f.kind == 'L' || f.kind == '[' {
			d.value(d.byte())
		}
		c.fields = append(c.fields, f)
	}
	d.annotations()
	c.super = d.descriptor()
	return c
}

// added for sonarqube
func (d *reader) decodeObject() any {
	c := d.descriptor()
	if c == nil {
		panic("object without class")
	}
	o := &object{class: c, fields: map[string]any{}, extra: map[string][]any{}}
	d.add(o)
	chain := []*class{}
	for x := c; x != nil; x = x.super {
		chain = append(chain, x)
		if len(chain) > 128 {
			panic("class hierarchy limit")
		}
	}
	for i := len(chain) - 1; i >= 0; i-- {
		x := chain[i]
		if x.flags&4 != 0 {
			panic("Externalizable values are unsupported")
		}
		for _, f := range x.fields {
			o.fields[f.name] = d.primitive(f.kind)
		}
		if x.flags&1 != 0 {
			o.extra[x.name] = d.annotations()
		}
	}
	return o
}

// added for sonarqube
func (d *reader) decodeArray() any {
	c := d.descriptor()
	if c == nil || len(c.name) < 2 {
		panic("invalid array class")
	}
	n := int(d.u32())
	if n < 0 || n > 1<<20 {
		panic("array limit")
	}
	a := make([]any, n)
	d.add(a)
	for i := range a {
		a[i] = d.primitive(c.name[1])
	}
	return a
}

// added for sonarqube
func (d *reader) decodeBlockData(long bool) any {
	if long {
		return block(d.bytes(int(d.u32())))
	}
	return block(d.bytes(int(d.byte())))
}

// added for sonarqube
func (d *reader) decodeEnum() any {
	c := d.descriptor()
	o := &object{class: c, fields: map[string]any{}}
	d.add(o)
	o.fields["name"] = d.value(d.byte())
	return o
}

// added for sonarqube
func (d *reader) value(tag byte) any {
	d.depth++
	defer func() { d.depth-- }()
	if d.depth > 128 {
		panic("object nesting limit")
	}
	switch tag {
	case 0x70:
		return nil
	case 0x71:
		return d.decodeReference()
	case 0x72:
		return d.decodeClassDesc()
	case 0x73:
		return d.decodeObject()
	case 0x74:
		s := d.shortUTF()
		d.add(s)
		return s
	case 0x7c:
		n := d.u64()
		if n > 16<<20 {
			panic("string limit")
		}
		s := d.utf(int(n))
		d.add(s)
		return s
	case 0x75:
		return d.decodeArray()
	case 0x76:
		c := d.descriptor()
		d.add(c)
		return c
	case 0x77:
		return d.decodeBlockData(false)
	case 0x7a:
		return d.decodeBlockData(true)
	case 0x7e:
		return d.decodeEnum()
	case 0x79:
		d.handles = nil
		return d.value(d.byte())
	default:
		panic(fmt.Sprintf("unsupported tag 0x%x", tag))
	}
}

/*
func plain(v any, depth int) any {
	if depth > 128 {
		panic("cyclic object graph")
	}
	switch x := v.(type) {
	case *object:
		if x.class.flags&16 != 0 {
			return plain(x.fields["name"], depth+1)
		}
		switch x.class.name {
		case "java.lang.Integer", "java.lang.Long", "java.lang.Short", "java.lang.Byte", "java.lang.Float", "java.lang.Double", "java.lang.Boolean", "java.lang.Character":
			return x.fields["value"]
		}
		if extra, ok := x.extra["java.util.Date"]; ok {
			for _, e := range extra {
				if b, ok := e.(block); ok && len(b) >= 8 {
					return int64(binary.BigEndian.Uint64(b[:8]))
				}
			}
		}
		if extra, ok := x.extra["java.util.ArrayList"]; ok {
			a := []any{}
			for _, e := range extra {
				if _, ok := e.(block); !ok {
					a = append(a, plain(e, depth+1))
				}
			}
			return a
		}
		if extra, ok := x.extra["java.util.HashMap"]; ok {
			vals := []any{}
			for _, e := range extra {
				if _, ok := e.(block); !ok {
					vals = append(vals, e)
				}
			}
			m := map[string]any{}
			for i := 0; i+1 < len(vals); i += 2 {
				m[fmt.Sprint(plain(vals[i], depth+1))] = plain(vals[i+1], depth+1)
			}
			return m
		}
		m := map[string]any{}
		for k, v := range x.fields {
			m[k] = plain(v, depth+1)
		}
		return m
	case []any:
		a := make([]any, len(x))
		for i, v := range x {
			a[i] = plain(v, depth+1)
		}
		return a
	case *class:
		return x.name
	default:
		return v
	}
}
*/

// added for sonarqube
func plainBuiltinDate(x *object) (any, bool) {
	if extra, ok := x.extra["java.util.Date"]; ok {
		for _, e := range extra {
			if b, ok := e.(block); ok && len(b) >= 8 {
				return int64(binary.BigEndian.Uint64(b[:8])), true
			}
		}
	}
	return nil, false
}

// added for sonarqube
func plainBuiltinCollections(x *object, depth int) (any, bool) {
	if extra, ok := x.extra["java.util.ArrayList"]; ok {
		a := []any{}
		for _, e := range extra {
			if _, ok := e.(block); !ok {
				a = append(a, plain(e, depth+1))
			}
		}
		return a, true
	}
	if extra, ok := x.extra["java.util.HashMap"]; ok {
		vals := []any{}
		for _, e := range extra {
			if _, ok := e.(block); !ok {
				vals = append(vals, e)
			}
		}
		m := map[string]any{}
		for i := 0; i+1 < len(vals); i += 2 {
			m[fmt.Sprint(plain(vals[i], depth+1))] = plain(vals[i+1], depth+1)
		}
		return m, true
	}
	return nil, false
}

// added for sonarqube
func plainBuiltin(x *object, depth int) (any, bool) {
	switch x.class.name {
	case "java.lang.Integer", "java.lang.Long", "java.lang.Short", "java.lang.Byte", "java.lang.Float", "java.lang.Double", "java.lang.Boolean", "java.lang.Character":
		return x.fields["value"], true
	}
	if val, ok := plainBuiltinDate(x); ok {
		return val, true
	}
	return plainBuiltinCollections(x, depth)
}

// added for sonarqube
func plainObject(x *object, depth int) any {
	if x.class.flags&16 != 0 {
		return plain(x.fields["name"], depth+1)
	}
	if val, ok := plainBuiltin(x, depth); ok {
		return val
	}
	m := map[string]any{}
	for k, v := range x.fields {
		m[k] = plain(v, depth+1)
	}
	return m
}

// added for sonarqube
func plain(v any, depth int) any {
	if depth > 128 {
		panic("cyclic object graph")
	}
	switch x := v.(type) {
	case *object:
		return plainObject(x, depth)
	case []any:
		a := make([]any, len(x))
		for i, val := range x {
			a[i] = plain(val, depth+1)
		}
		return a
	case *class:
		return x.name
	default:
		return v
	}
}
