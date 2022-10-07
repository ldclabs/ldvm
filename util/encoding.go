// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
)

const (
	checksumLen = 4
)

func CheckSumHex(data []byte) string {
	buf := make([]byte, len(data)*2+2)
	copy(buf[:2], "0x")
	hex.Encode(buf[2:], data[:])

	hash := Sum256(data)
	bf := buf[2:]
	for i := 0; i < len(bf); i++ {
		if bf[i] > '9' && i < 256 && (hash[i/8]>>uint(7-i%8))&1 > 0 {
			bf[i] -= 32
		}
	}
	return string(buf[:])
}

// EncodeToString [bytes] to a string using raw base64url format.
func EncodeToString(bytes []byte) string {
	bytesLen := len(bytes)
	buf := make([]byte, bytesLen+checksumLen)
	copy(buf, bytes)
	copy(buf[bytesLen:], Sum256(bytes)[:checksumLen])
	return base64.RawURLEncoding.EncodeToString(buf)
}

func EncodeToQuoteString(bytes []byte) string {
	bytesLen := len(bytes)
	src := make([]byte, bytesLen+checksumLen)
	copy(src, bytes)
	copy(src[bytesLen:], Sum256(bytes)[:checksumLen])

	buf := make([]byte, base64.RawURLEncoding.EncodedLen(len(src))+2)
	buf[0] = '"'
	base64.RawURLEncoding.Encode(buf[1:], src)
	buf[len(buf)-1] = '"'
	return string(buf)
}

// DecodeString [str] to bytes from raw base64url.
func DecodeString(str string) ([]byte, error) {
	buf, err := base64.RawURLEncoding.DecodeString(str)
	if err != nil {
		return nil, errors.New("util.DecodeString: " + err.Error())
	}

	bytesLen := len(buf)
	if bytesLen < checksumLen {
		return nil, errors.New("util.DecodeString: no checksum bytes")
	}
	// Verify the checksum
	rawBytes := buf[:bytesLen-checksumLen]
	checksum := buf[bytesLen-checksumLen:]
	if !bytes.Equal(checksum, Sum256(rawBytes)[:checksumLen]) {
		return nil, errors.New("util.DecodeString: invalid input checksum")
	}
	return rawBytes, nil
}

func DecodeQuoteString(str string) ([]byte, error) {
	strLen := len(str)
	if strLen < 2 || str[0] != '"' || str[strLen-1] != '"' {
		return nil, errors.New("util.DecodeQuoteString: invalid quote string")
	}
	return DecodeString(str[1 : strLen-1])
}

func DecodeStringWithLen(str string, expectedLen int) ([]byte, error) {
	buf, err := DecodeString(str)
	if err != nil {
		return nil, errors.New("util.DecodeStringWithLen: " + err.Error())
	}
	if bytesLen := len(buf); bytesLen != expectedLen {
		return nil, errors.New("util.DecodeStringWithLen: invalid bytes length, expected " +
			strconv.Itoa(expectedLen) + ", got " + strconv.Itoa(bytesLen))
	}
	return buf, nil
}

func DecodeQuoteStringWithLen(str string, expectedLen int) ([]byte, error) {
	strLen := len(str)
	if strLen < 2 || str[0] != '"' || str[strLen-1] != '"' {
		return nil, errors.New("util.DecodeQuoteStringWithLen: invalid quote string")
	}

	return DecodeStringWithLen(str[1:strLen-1], expectedLen)
}
