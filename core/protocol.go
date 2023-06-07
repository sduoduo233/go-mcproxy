package core

import (
	"errors"
	"io"
)

type (
	VarInt int32
	String string
	UShort uint16
	Long   int64
)

func readByte(r io.Reader) (byte, error) {
	b := [1]byte{0}
	_, err := r.Read(b[:])
	return b[0], err
}

func (v *Long) ReadFrom(r io.Reader) (int64, error) {
	var buf [8]byte
	n, err := io.ReadFull(r, buf[:])
	if err != nil {
		return int64(n), err
	}

	*v = Long(
		int64(buf[0])<<56 |
			int64(buf[1])<<48 |
			int64(buf[2])<<40 |
			int64(buf[3])<<32 |
			int64(buf[4])<<24 |
			int64(buf[5])<<16 |
			int64(buf[6])<<8 |
			int64(buf[7]),
	)
	return int64(n), nil
}

func (v Long) WriteTo(w io.Writer) (int64, error) {
	n := int64(v)
	var buf [8]byte
	buf[0] = byte(n >> 56)
	buf[1] = byte(n >> 48)
	buf[2] = byte(n >> 40)
	buf[3] = byte(n >> 32)
	buf[4] = byte(n >> 24)
	buf[5] = byte(n >> 16)
	buf[6] = byte(n >> 8)
	buf[7] = byte(n)

	n1, err := w.Write(buf[:])
	return int64(n1), err
}

func (v *VarInt) ReadFrom(r io.Reader) (int64, error) {
	var SEGMENT_BITS byte = 0x7F
	var CONTINUE_BIT byte = 0x80

	var value int32 = 0
	var position int = 0
	var currentByte byte = 0
	var err error
	var n int64 = 0

	for {
		n++
		currentByte, err = readByte(r)
		if err != nil {
			return 0, err
		}

		value |= (int32(currentByte&SEGMENT_BITS) << position)

		if (currentByte & CONTINUE_BIT) == 0 {
			break
		}

		position += 7
		if position >= 32 {
			return 0, errors.New("varint is too large")
		}
	}

	*v = VarInt(value)
	return n, nil
}

func (v VarInt) WriteTo(w io.Writer) (int64, error) {
	len := v.Len()
	num := int32(v)
	buf := make([]byte, len)

	i := 0
	for {
		b := num & 0x7f
		num >>= 7
		if num != 0 {
			b |= 0x80
		}
		buf[i] = byte(b)
		i++
		if num == 0 {
			break
		}
	}

	n, err := w.Write(buf)
	return int64(n), err
}

func (v VarInt) Len() int {
	switch {
	case v < 0:
		return 5
	case v < 1<<(7*1):
		return 1
	case v < 1<<(7*2):
		return 2
	case v < 1<<(7*3):
		return 3
	case v < 1<<(7*4):
		return 4
	default:
		return 5
	}
}

func (s *String) ReadFrom(r io.Reader) (int64, error) {
	// string length
	var len VarInt
	n, err := len.ReadFrom(r)
	if err != nil {
		return n, err
	}

	// read string
	buf := make([]byte, len)
	n2, err := io.ReadFull(r, buf)
	n += int64(n2)
	if err != nil {
		return n, err
	}

	*s = String(string(buf))
	return n, nil
}

func (s String) WriteTo(w io.Writer) (int64, error) {
	bytesStr := []byte(s)

	n, err := VarInt(len(bytesStr)).WriteTo(w)
	if err != nil {
		return n, err
	}

	n2, err := w.Write(bytesStr)
	n += int64(n2)
	return n, err
}

func (s *UShort) ReadFrom(r io.Reader) (int64, error) {
	buf := make([]byte, 2)
	n, err := io.ReadFull(r, buf)
	if err != nil {
		return int64(n), err
	}

	*s = UShort(uint16(buf[0])<<8 + uint16(buf[1]))
	return int64(n), nil
}

func (s UShort) WriteTo(w io.Writer) (int64, error) {
	var buf [2]byte
	buf[0] = byte(s >> 8)
	buf[1] = byte(s)

	n, err := w.Write(buf[:])
	return int64(n), err
}
