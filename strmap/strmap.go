package strmap

import (
	"bytes"
	"encoding/binary"
	"io"
	"unsafe"
)

func VarInt(x int64) []byte {
	b := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(b, x)
	return b[:n]
}

func Marshal(buf *bytes.Buffer, m map[string]string) {
	buf.Write(VarInt(int64(len(m))))

	for k, v := range m {
		buf.Write(VarInt(int64(len(k))))
		buf.WriteString(k)
		buf.Write(VarInt(int64(len(v))))
		buf.WriteString(v)
	}
}

func Unmarshal(r *bytes.Reader) (map[string]string, error) {
	l, err := binary.ReadVarint(r)
	if err != nil {
		return nil, err
	}

	m := make(map[string]string, l)

	for i := int64(0); i < l; i++ {
		k, err := readOne(r)
		if err != nil {
			return nil, err
		}

		v, err := readOne(r)
		if err != nil {
			return nil, err
		}

		m[k] = v
	}

	return m, nil
}

func readOne(r *bytes.Reader) (string, error) {
	l, err := binary.ReadVarint(r)
	if err != nil {
		return "", err
	}

	b := make([]byte, l)

	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}

	// turn []byte into string without allocation and copy
	// taken from https://cs.opensource.google/go/go/+/refs/tags/go1.17:src/strings/builder.go;l=47
	return *(*string)(unsafe.Pointer(&b)), nil
}
