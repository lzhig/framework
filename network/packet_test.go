package network

import (
	"fmt"
	//"globaltedinc/framework/network"
	"reflect"
	"testing"
)

type testType struct {
	v interface{}
	f interface{}
	t string
}

func Test_Packet(t *testing.T) {
	p := NewPacket(100)
	tt := []testType{
		{v: byte(1), f: p.WriteByte, t: "write byte"},
		{v: byte(1), f: p.ReadByte, t: "read byte"},
		{v: int8(2), f: p.WriteInt8, t: "write int8"},
		{v: int8(2), f: p.ReadInt8, t: "read int8"},
		{v: uint8(3), f: p.WriteUInt8, t: "write uint8"},
		{v: uint8(3), f: p.ReadUInt8, t: "read uint8"},
		{v: int16(4), f: p.WriteInt16, t: "write int16"},
		{v: int16(4), f: p.ReadInt16, t: "read int16"},
		{v: uint16(5), f: p.WriteUInt16, t: "write uint16"},
		{v: uint16(5), f: p.ReadUInt16, t: "read uint16"},
		{v: int32(6), f: p.WriteInt32, t: "write int32"},
		{v: int32(6), f: p.ReadInt32, t: "read int32"},
		{v: uint32(7), f: p.WriteUInt32, t: "write uint32"},
		{v: uint32(7), f: p.ReadUInt32, t: "read uint32"},
		{v: int64(6), f: p.WriteInt64, t: "write int64"},
		{v: int64(6), f: p.ReadInt64, t: "read int64"},
		{v: uint64(7), f: p.WriteUInt64, t: "write uint64"},
		{v: uint64(7), f: p.ReadUInt64, t: "read uint64"},
	}

	for k, v := range tt {
		f := reflect.ValueOf(v.f)

		if k%2 == 0 {
			params := make([]reflect.Value, 1)
			params[0] = reflect.ValueOf(v.v)
			if err := f.Call(params); err[0].Interface() != nil {
				fmt.Println("Failed to call ", v.t, ", Error:", err[0])
			}
		} else {
			ret := f.Call(nil)
			if ret[1].Interface() != nil {
				t.Error("Failed to call", v.t)
			}

			if ret[0].Interface() != v.v {
				t.Error("Failed to call", v.t)
			}
		}
	}

	slice := []byte{8, 9, 10, 11, 12, 13}
	if err := p.WriteSlice(slice); err != nil {
		t.Error("Failed to call write slice")
	}

	r_slice, err := p.ReadSlice(len(slice))
	if err != nil || len(slice) != len(r_slice) {
		t.Error("Failed to call read slice")
	}
	for k, v := range slice {
		if v != r_slice[k] {
			t.Error("Failed to call read slice")
		}
	}

	s := "Test write string"
	if err := p.WriteString(s); err != nil {
		t.Error("Failed to call write string")
	}

	if s1, err := p.ReadString(); s1 != s || err != nil {
		t.Error("Failed to call read string")
	}
}
