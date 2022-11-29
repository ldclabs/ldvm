// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package signer

import (
	"errors"
	"math"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/util/encoding"
)

type Key []byte

type Keys []Key

func (k Key) Kind() Kind {
	switch len(k) {
	case 20:
		return Secp256k1
	case 32:
		return Ed25519
	case 48:
		return BLS12381
	default:
		return Unknown
	}
}

var (
	empty20b [20]byte
	empty32b [32]byte
	empty48b [48]byte

	empty20 = string(empty20b[:])
	empty32 = string(empty32b[:])
	empty48 = string(empty48b[:])
)

func (k Key) Valid() error {
	if len(k) == 0 {
		return errors.New("signer.Key.Valid: empty key")
	}

	s := string(k)
	switch k.Kind() {
	case Secp256k1:
		if s == empty20 {
			return errors.New("signer.Key.Valid: empty Secp256k1 key")
		}

	case Ed25519:
		if s == empty32 {
			return errors.New("signer.Key.Valid: empty Ed25519 key")
		}

	case BLS12381:
		if s == empty48 {
			return errors.New("signer.Key.Valid: empty BLS12-381 key")
		}

	default:
		return errors.New("signer.Key.Valid: invalid key " + encoding.EncodeToString(k))
	}

	return nil
}

func (k Key) ValidOrEmpty() error {
	switch {
	case k == nil:
		return errors.New("signer.Key.Valid: nil key")

	case len(k) == 0:
		return nil

	default:
		return k.Valid()
	}
}

func (k Key) IsAddress(addr ids.Address) bool {
	switch len(k) {
	case 20:
		return string(addr[:]) == string(k)
	case 32, 48:
		return string(addr[:]) == string(encoding.Sum256(k)[:20])
	}

	return false
}

func (k Key) Address() ids.Address {
	var addr [20]byte

	switch len(k) {
	case 20:
		copy(addr[:], k)
	case 32, 48:
		copy(addr[:], encoding.Sum256(k)[:20])
	}

	return addr
}

func (k Key) Bytes() []byte {
	return k
}

func (k Key) Ptr() *Key {
	return &k
}

func (k Key) AsKey() string {
	return string(k)
}

func (k Key) String() string {
	return encoding.EncodeToString(k)
}

func (k Key) GoString() string {
	return k.String()
}

func (k Key) Equal(b Key) bool {
	if k == nil || b == nil {
		return k == nil && b == nil
	}
	return string(k) == string(b)
}

func (k Key) MarshalText() ([]byte, error) {
	return []byte(k.String()), nil
}

func (k *Key) UnmarshalText(b []byte) error {
	if k == nil {
		return errors.New("signer.Key.UnmarshalText: nil pointer")
	}

	b, err := encoding.DecodeString(string(b))
	if err != nil {
		return errors.New("signer.Key.UnmarshalText: " + err.Error())
	}

	*k = append((*k)[0:0], b...)
	return nil
}

func (k Key) MarshalJSON() ([]byte, error) {
	return []byte(encoding.EncodeToQuoteString(k)), nil
}

func (k *Key) UnmarshalJSON(b []byte) error {
	if k == nil {
		return errors.New("signer.Key.UnmarshalJSON: nil pointer")
	}

	b, err := encoding.DecodeQuoteString(string(b))
	if err != nil {
		return errors.New("signer.Key.UnmarshalJSON: " + err.Error())
	}

	*k = append((*k)[0:0], b...)
	return nil
}

func (k Key) MarshalCBOR() ([]byte, error) {
	data, err := encoding.MarshalCBOR(k.Bytes())
	if err != nil {
		return nil, errors.New("signer.Key.MarshalCBOR: " + err.Error())
	}
	return data, nil
}

func (k *Key) UnmarshalCBOR(data []byte) error {
	if k == nil {
		return errors.New("signer.Key.UnmarshalCBOR: nil pointer")
	}

	if len(data) == 1 {
		switch data[0] {
		case 0xf6: // nil
			*k = nil
			return nil
		case 0x40:
			*k = Key{}
			return nil
		}
	}

	var b []byte
	if err := encoding.UnmarshalCBOR(data, &b); err != nil {
		return errors.New("signer.Key.UnmarshalCBOR: " + err.Error())
	}

	*k = append((*k)[0:0], b...)
	return nil
}

func (k Key) Clone() Key {
	if k == nil {
		return nil
	}

	nk := make([]byte, len(k))
	copy(nk, k)
	return nk
}

func (k Key) Verify(digestHash []byte, sigs Sigs) bool {
	for _, sig := range sigs {
		if i := sig.FindKey(digestHash, k); i >= 0 {
			return true
		}
	}
	return false
}

func (ks Keys) Has(key Key) bool {
	s := string(key)
	for _, k := range ks {
		if s == string(k) {
			return true
		}
	}
	return false
}

func (ks Keys) HasKeys(keys Keys, threshold uint16) bool {
	if len(keys) == 0 || threshold == 0 || len(ks) < int(threshold) || len(keys) < int(threshold) {
		return false
	}
	if keys.Valid() != nil {
		return false
	}
	i := uint16(0)
	for _, k := range keys {
		if ks.Has(k) {
			i++
			if i >= threshold {
				return true
			}
		}
	}
	return false
}

func (ks Keys) HasAddress(addr ids.Address) bool {
	s := string(addr[:])
	for _, k := range ks {
		switch len(k) {
		case 20:
			if s == string(k) {
				return true
			}
		case 32:
			if s == string(encoding.Sum256(k)[:20]) {
				return true
			}
		}
	}

	return false
}

func (ks Keys) FindKeyOrAddr(addr ids.Address) Key {
	s := string(addr[:])
	for _, k := range ks {
		switch len(k) {
		case 20:
			if s == string(k) {
				return k
			}
		case 32, 48:
			if s == string(encoding.Sum256(k)[:20]) {
				return k
			}
		}
	}

	return Key(addr[:])
}

func (ks Keys) Valid() error {
	dset := make(map[string]struct{}, len(ks))
	for _, k := range ks {
		keyStr := k.AsKey()
		if _, ok := dset[keyStr]; ok {
			return errors.New("signer.Keys.Valid: duplicate key " + k.String())
		}
		dset[keyStr] = struct{}{}

		if err := k.Valid(); err != nil {
			return errors.New("signer.Keys.Valid: " + err.Error())
		}
	}

	return nil
}

func (ks Keys) Clone() Keys {
	if ks == nil {
		return nil
	}

	nks := make([]Key, len(ks))
	for i, k := range ks {
		nks[i] = k.Clone()
	}
	return nks
}

func (ks Keys) Verify(digestHash []byte, sigs Sigs, threshold uint16) bool {
	ksLen := len(ks)
	if ksLen == 0 || ksLen > math.MaxUint8 || threshold == 0 ||
		ksLen < int(threshold) || len(sigs) < int(threshold) {
		return false
	}

	t := uint16(0)
	remaining := make([]Key, ksLen)
	copy(remaining, ks)

	dset := make(map[string]struct{}, len(sigs))

	for _, sig := range sigs {
		sigStr := sig.AsKey()
		if _, ok := dset[sigStr]; ok {
			return false
		}
		dset[sigStr] = struct{}{}

		if i := sig.FindKey(digestHash, remaining...); i >= 0 {
			t += 1
			if t >= threshold {
				return true
			}

			remaining = append(remaining[:i], remaining[i+1:]...)
		}
	}

	return false
}

func (ks Keys) VerifyPlus(digestHash []byte, sigs Sigs, threshold uint16) bool {
	if int(threshold) < len(ks) {
		threshold += 1
	}
	return ks.Verify(digestHash, sigs, threshold)
}
