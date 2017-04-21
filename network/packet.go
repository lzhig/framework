package network

import (
	"io"
	"math"
)

type Packet struct {
	data     []byte
	len      int
	writePos int
	readPos  int
	cap      int
}

func (this *Packet) reset() {
	this.len = 0
	this.writePos = 0
	this.readPos = 0
	this.cap = 0
}

func (this *Packet) Attach(data []byte) (old []byte) {
	old = this.data
	this.reset()
	this.data = data
	this.cap = len(data)
	this.len = this.cap
	this.writePos = this.len
	return old
}

func (this *Packet) Detach() (data []byte) {
	data = this.data
	this.data = nil
	this.reset()
	return data
}

func (this *Packet) isEnough(l int) (bool, error) {
	if this.cap-this.writePos < l {
		return false, &ErrorPacketBufferSizeTooSmall{ErrorNetwork{s: "packet buffer is too small"}}
	}
	return true, nil
}

func (this *Packet) GetPacketLen() int {
	return this.len
}

func (this *Packet) GetData() []byte {
	return this.data[:this.len]
}

func (this *Packet) MoveReadPos(step int) {
	this.readPos += step
}

func (this *Packet) MoveWritePos(step int) {
	this.writePos += step
}

func (this *Packet) WriteSlice(b []byte) error {
	l := len(b)
	if ok, err := this.isEnough(l); !ok {
		return err
	}

	copy(this.data[this.writePos:], b)
	this.len += l
	this.writePos += l
	return nil
}

func (this *Packet) WriteByte(b byte) error {
	if ok, err := this.isEnough(1); !ok {
		return err
	}

	this.data[this.writePos] = b
	this.len++
	this.writePos++

	return nil
}

func (this *Packet) WriteInt8(b int8) error {
	return this.WriteByte(byte(b))
}

func (this *Packet) WriteUInt8(b uint8) error {
	return this.WriteByte(byte(b))
}

func (this *Packet) WriteInt16(b int16) error {
	if ok, err := this.isEnough(2); !ok {
		return err
	}

	this.data[this.writePos] = byte((int(b) & 0xFF00) >> 8)
	this.data[this.writePos+1] = byte(b & 0xFF)
	this.len += 2
	this.writePos += 2

	return nil
}

func (this *Packet) WriteUInt16(b uint16) error {
	return this.WriteInt16(int16(b))
}

func (this *Packet) WriteInt32(b int32) error {
	if ok, err := this.isEnough(4); !ok {
		return err
	}

	this.data[this.writePos] = byte((uint(b) & 0xFF000000) >> 24)
	this.data[this.writePos+1] = byte((uint(b) & 0xFF0000) >> 16)
	this.data[this.writePos+2] = byte((uint(b) & 0xFF00) >> 8)
	this.data[this.writePos+3] = byte(b & 0xFF)
	this.len += 4
	this.writePos += 4

	return nil
}

func (this *Packet) WriteUInt32(b uint32) error {
	return this.WriteInt32(int32(b))
}

func (this *Packet) WriteInt64(b int64) error {
	if ok, err := this.isEnough(8); !ok {
		return err
	}

	this.data[this.writePos] = byte((uint64(b) & 0xFF00000000000000) >> 56)
	this.data[this.writePos+1] = byte((uint64(b) & 0xFF000000000000) >> 48)
	this.data[this.writePos+2] = byte((uint64(b) & 0xFF0000000000) >> 40)
	this.data[this.writePos+3] = byte((uint64(b) & 0xFF00000000) >> 32)
	this.data[this.writePos+4] = byte((uint64(b) & 0xFF000000) >> 24)
	this.data[this.writePos+5] = byte((uint64(b) & 0xFF0000) >> 16)
	this.data[this.writePos+6] = byte((uint64(b) & 0xFF00) >> 8)
	this.data[this.writePos+7] = byte(b & 0xFF)
	this.len += 8
	this.writePos += 8

	return nil
}

func (this *Packet) WriteUInt64(b uint64) error {
	return this.WriteInt64(int64(b))
}

func (this *Packet) WriteString(b string) error {
	l := len(b)
	if ok, err := this.isEnough(l + 2); !ok || l > math.MaxUint16 {
		return err
	}

	if err := this.WriteUInt16(uint16(l)); err != nil {
		return err
	}

	copy(this.data[this.writePos:], b)
	this.len += l
	this.writePos += l

	return nil
}

func (this *Packet) ReadByte() (byte, error) {
	if this.readPos+1 > this.cap {
		return 0, io.EOF
	}
	c := this.data[this.readPos]
	this.readPos++
	return c, nil
}

func (this *Packet) ReadInt8() (int8, error) {
	if this.readPos+1 > this.cap {
		return 0, io.EOF
	}
	c := int8(this.data[this.readPos])
	this.readPos++
	return c, nil
}

func (this *Packet) ReadUInt8() (uint8, error) {
	if this.readPos+1 > this.cap {
		return 0, io.EOF
	}
	c := uint8(this.data[this.readPos])
	this.readPos++
	return c, nil
}

func (this *Packet) ReadInt16() (int16, error) {
	if this.readPos+2 > this.cap {
		return 0, io.EOF
	}
	c := (int16(this.data[this.readPos]) << 8) + int16(this.data[this.readPos+1])
	this.readPos += 2
	return c, nil
}

func (this *Packet) ReadUInt16() (uint16, error) {
	if this.readPos+2 > this.cap {
		return 0, io.EOF
	}
	c := (uint16(this.data[this.readPos]) << 8) + uint16(this.data[this.readPos+1])
	this.readPos += 2
	return c, nil
}

func (this *Packet) ReadInt32() (int32, error) {
	if this.readPos+4 > this.cap {
		return 0, io.EOF
	}
	c := (int32(this.data[this.readPos]) << 24) + (int32(this.data[this.readPos+1]) << 16) + (int32(this.data[this.readPos+2]) << 8) + int32(this.data[this.readPos+3])
	this.readPos += 4
	return c, nil
}

func (this *Packet) ReadUInt32() (uint32, error) {
	if this.readPos+4 > this.cap {
		return 0, io.EOF
	}
	c := (uint32(this.data[this.readPos]) << 24) + (uint32(this.data[this.readPos+1]) << 16) + (uint32(this.data[this.readPos+2]) << 8) + uint32(this.data[this.readPos+3])
	this.readPos += 4
	return c, nil
}

func (this *Packet) ReadInt64() (int64, error) {
	if this.readPos+8 > this.cap {
		return 0, io.EOF
	}
	c := (int64(this.data[this.readPos]) << 56) +
		(int64(this.data[this.readPos+1]) << 48) +
		(int64(this.data[this.readPos+2]) << 40) +
		(int64(this.data[this.readPos+3]) << 32) +
		(int64(this.data[this.readPos+4]) << 24) +
		(int64(this.data[this.readPos+5]) << 16) +
		(int64(this.data[this.readPos+6]) << 8) +
		int64(this.data[this.readPos+7])
	this.readPos += 8
	return c, nil
}

func (this *Packet) ReadUInt64() (uint64, error) {
	if this.readPos+8 > this.cap {
		return 0, io.EOF
	}
	c := (uint64(this.data[this.readPos]) << 56) +
		(uint64(this.data[this.readPos+1]) << 48) +
		(uint64(this.data[this.readPos+2]) << 40) +
		(uint64(this.data[this.readPos+3]) << 32) +
		(uint64(this.data[this.readPos+4]) << 24) +
		(uint64(this.data[this.readPos+5]) << 16) +
		(uint64(this.data[this.readPos+6]) << 8) +
		uint64(this.data[this.readPos+7])
	this.readPos += 8
	return c, nil
}

func (this *Packet) ReadSlice(n int) (b []byte, err error) {
	if this.readPos+n > this.cap {
		return nil, io.EOF
	}

	b = append(b, this.data[this.readPos:this.readPos+n]...)
	this.readPos += n
	return b, nil
}

func (this *Packet) ReadString() (string, error) {
	if this.readPos+2 > this.cap {
		return "", io.EOF
	}

	l, err := this.ReadUInt16()
	if err != nil {
		return "", err
	}

	if this.readPos+int(l) > this.cap {
		return "", io.EOF
	}

	b, err := this.ReadSlice(int(l))
	return string(b), err
}
