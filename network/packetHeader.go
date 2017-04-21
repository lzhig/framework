package network

import (
	"io"
	"unsafe"
)

var packetHeader IPacketHeader = &PacketDefaultHeader{}

func SetPacketHeader(header IPacketHeader) {
	packetHeader = header
}

func GetPacketHeader() IPacketHeader {
	return packetHeader
}

var defaultHeaderFlag = [4]byte{0x12, 0x34, 0x45, 0x67}

type IPacketHeader interface {
	BuildHeader(bodyLen int, data []byte) error
	ParsePacketHeader(data []byte) (ok bool, headerLen int32, packetLen int32, err error)
	GetHeaderLen() int
}

type PacketDefaultHeader struct {
}

func (this *PacketDefaultHeader) BuildHeader(bodyLen int, data []byte) error {
	if len(data) < len(defaultHeaderFlag)+int(unsafe.Sizeof(bodyLen)) {
		return io.EOF
	}
	copy(data, defaultHeaderFlag[:])
	ndx := len(defaultHeaderFlag)
	data[ndx] = byte((uint32(bodyLen) & 0xFF000000) >> 24)
	data[ndx+1] = byte((bodyLen & 0xFF0000) >> 16)
	data[ndx+2] = byte((bodyLen & 0xFF00) >> 8)
	data[ndx+3] = byte(bodyLen & 0xFF)
	return nil
}

func (this *PacketDefaultHeader) ParsePacketHeader(data []byte) (ok bool, headerLen int32, packetLen int32, err error) {
	totalLen := len(defaultHeaderFlag) + int(unsafe.Sizeof(packetLen))
	if len(data) < totalLen {
		return false, int32(totalLen), 0, nil
	}

	for k, v := range defaultHeaderFlag {
		if data[k] != v {
			return false, int32(totalLen), 0, &ErrorInvalidPacketHeader{ErrorNetwork{s: "Invalid packet header"}}
		}
	}

	packetLen = (int32(data[4]) << 24) + (int32(data[5]) << 16) + (int32(data[6]) << 8) + int32(data[7])

	return true, int32(totalLen), packetLen, nil
}

func (this *PacketDefaultHeader) GetHeaderLen() int {
	return 8
}
