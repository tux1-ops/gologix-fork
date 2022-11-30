package gologix

// based on code from https://github.com/loki-os/go-ethernet-ip

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

// The path is formatted like this.
// byte 0: number of 16 bit words
// byte 1: 000. .... path segment type (port segment = 0)
// byte 1: ...0 .... extended link address (0 = false)
// byte 1: .... 0001 port (backplane = 1)
// byte 2: n/a
// byte 3: 001. .... path segment type (logical segment = 1)
// byte 3: ...0 00.. logical segment type class ID (0)
// byte 3: .... ..00 logical segment format: 8-bit (0)
// byte 4: path segment 0x20
// byte 5: 001. .... path segment type (logical segment = 1)
// byte 5: ...0 01.. logical segment type: Instance ID = 1
// byte 5: .... ..00 logical segment format: 8-bit (0)
// byte 6: path segment instance 0x01
// so on...
//msg.Path = [6]byte{0x01, 0x00, 0x20, 0x02, 0x24, 0x01}

// bits 5,6,7 (counting from 0) are the segment type
type SegmentType byte

const (
	SegmentTypePort      SegmentType = 0b0000_0000
	SegmentTypeLogical   SegmentType = 0b0010_0000
	SegmentTypeNetwork   SegmentType = 0b0100_0000
	SegmentTypeSymbolic  SegmentType = 0b0110_0000
	SegmentTypeData      SegmentType = 0b1000_0000
	SegmentTypeDataType1 SegmentType = 0b1010_0000
	SegmentTypeDataType2 SegmentType = 0b1100_0000

	SegmentTypeElement8Bit      SegmentType = 0x28
	SegmentTypeElement16Bit     SegmentType = 0x29
	SegmentTypeElement32Bit     SegmentType = 0x2A
	SegmentTypeClassID8Bit      SegmentType = 0x20
	SegmentTypeClassID16Bit     SegmentType = 0x21
	SegmentTypeInstanceID8Bit   SegmentType = 0x24
	SegmentTypeInstanceID16Bit  SegmentType = 0x25
	SegmentTypeAttributeID8Bit  SegmentType = 0x30
	SegmentTypeAttributeID16Bit SegmentType = 0x31
	SegmentTypeExtendedSymbolic SegmentType = 0x91
)

func Paths(arg ...[]byte) []byte {
	io := bytes.Buffer{}
	for i := 0; i < len(arg); i++ {
		io.Write(arg[i])
	}
	return io.Bytes()
}

// bits 0 and 4 (counting from 0) are the data type bits
type DataTypes byte

const (
	DataTypeSimple DataTypes = 0b0000_0000
	DataTypeANSI   DataTypes = 0b0001_0001 //0x11
)

// bits 2,3, and 4 (counting from 0) are the LogicalType
type LogicalType byte

const (
	LogicalTypeClassID     LogicalType = 0b0000_0000 //0 << 2
	LogicalTypeInstanceID  LogicalType = 0b0000_0100 //1 << 2
	LogicalTypeMemberID    LogicalType = 0b0000_1000 //2 << 2
	LogicalTypeConnPoint   LogicalType = 0b0000_1100 //3 << 2
	LogicalTypeAttributeID LogicalType = 0b0001_0000 //4 << 2
	LogicalTypeSpecial     LogicalType = 0b0001_0100 //5 << 2
	LogicalTypeServiceID   LogicalType = 0b0001_1000 //6 << 2
)

func MarshalPathData(tp DataTypes, data []byte, padded bool) []byte {
	//io := bytes.Buffer{}
	io := make([]byte, 0, 16)

	firstByte := byte(SegmentTypeData) | byte(tp)
	io = append(io, firstByte)
	//io.Write(firstByte)

	length := byte(len(data))
	io = append(io, length)
	//io.Write(length)

	io = append(io, data...)
	//io.Write(data)

	if padded && len(io)%2 == 1 {
		//io.Write(uint8(0))
		io = append(io, 0)
	}

	return io
}

func MarshalPathLogical(tp LogicalType, address uint32, padded bool) []byte {
	format := uint8(0)

	if address <= 255 {
		format = 0
	} else if address > 255 && address <= 65535 {
		format = 1
	} else {
		format = 2
	}

	io := make([]byte, 0, 16)
	firstByte := byte(SegmentTypeLogical) | byte(tp) | format
	io = append(io, firstByte)
	//io.Write(firstByte)

	if address > 255 && address <= 65535 && padded {
		io = append(io, 0)
	}

	if address <= 255 {
		io = append(io, byte(address))
		//io.Write(uint8(address))
	} else if address > 255 && address <= 65535 {
		addr_dat := make([]byte, 2)
		binary.LittleEndian.PutUint16(addr_dat, uint16(address))
		io = append(io, addr_dat...)
		//io.Write(uint16(address))
	} else {
		addr_dat := make([]byte, 4)
		binary.LittleEndian.PutUint32(addr_dat, address)
		io = append(io, addr_dat...)
		//io.Write(address)
	}

	return io
}

func MarshalPathPort(link []byte, portID uint16, padded bool) []byte {
	extendedLinkTag := len(link) > 1
	extendedPortTag := !(portID < 15)

	//io := bytes.Buffer{}
	io := make([]byte, 0, 16)

	firstByte := byte(SegmentTypePort)
	if extendedLinkTag {
		firstByte = firstByte | 0x10
	}

	if !extendedPortTag {
		firstByte = firstByte | byte(portID)
	} else {
		firstByte = firstByte | 0xf
	}

	//io.Write(firstByte)
	io = append(io, firstByte)

	if extendedLinkTag {
		io = append(io, byte(len(link)))
		//io.Write(uint8(len(link)))
	}

	if extendedPortTag {
		port_dat := make([]byte, 2)
		binary.LittleEndian.PutUint16(port_dat, portID)
		io = append(io, port_dat...)
		//io.Write(portID)
	}

	//io.Write(link)
	io = append(io, link...)

	if padded && len(io)%2 == 1 {
		io = append(io, 0)
		//io.Write(uint8(0))
	}

	return io
}

// this function takes a CIP path in the format of 0,1,192.168.2.1,0,1 and converts it into the proper equivalent byte slice.
func ParsePath(path string) ([]byte, error) {
	// get rid of any spaces and square brackets
	path = strings.ReplaceAll(path, " ", "")
	path = strings.ReplaceAll(path, "[", "")
	path = strings.ReplaceAll(path, "]", "")
	// split on commas
	parts := strings.Split(path, ",")

	byte_path := make([]byte, 0, len(parts))

	for _, part := range parts {
		// first see if this looks like an IP address.
		is_ip := strings.Contains(part, ".")
		if is_ip {
			// for some god forsaken reason the path doesn't use the ip address as actual bytes but as an ascii string.
			// we first have to set bit 5 in the previous byte to say we're using an extended address for this part.
			last_pos := len(byte_path) - 1
			last_byte := byte_path[last_pos]
			byte_path[last_pos] = last_byte | 1<<4
			l := len(part)
			byte_path = append(byte_path, byte(l))
			string_bytes := []byte(part)
			byte_path = append(byte_path, string_bytes...)
			continue
		}
		// not an IP address
		val, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("problem converting %v to number. %w", part, err)
		}
		if val < 0 || val > 255 {
			return nil, fmt.Errorf("number out of range. %v", part)
		}
		byte_path = append(byte_path, byte(val))
	}

	return byte_path, nil
}

type Byteable interface {
	Bytes() []byte
}

func BuildPath(ps ...Byteable) (*bytes.Buffer, error) {
	b := new(bytes.Buffer)
	for _, p := range ps {
		_, err := b.Write(p.Bytes())
		if err != nil {
			return nil, err
		}
	}
	return b, nil
}
