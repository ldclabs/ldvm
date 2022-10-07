// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package signer

import (
	"bytes"
	"crypto/ed25519"
	"errors"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ldclabs/ldvm/util"
)

type Sig []byte

type Sigs []Sig

func (s Sig) Kind() Kind {
	switch len(s) {
	case 64:
		return Ed25519
	case 65:
		return Secp256k1
	default:
		return Unknown
	}
}

func (s Sig) Valid() error {
	if s.Kind() == Unknown {
		return errors.New("signer.Sig.Valid: unknown sig " + s.String())
	}

	return nil
}

func (s Sig) Bytes() []byte {
	return s
}

func (s Sig) Ptr() *Sig {
	return &s
}

func (s Sig) AsKey() string {
	switch len(s) {
	case 65:
		return string(s[:64])
	default:
		return string(s)
	}
}

func (s Sig) String() string {
	return util.EncodeToString(s)
}

func (s Sig) GoString() string {
	return s.String()
}

func (s Sig) Equal(b Sig) bool {
	if s == nil || b == nil {
		return s == nil && b == nil
	}
	return string(s) == string(b)
}

func (s Sig) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *Sig) UnmarshalText(b []byte) error {
	if s == nil {
		return errors.New("signer.Sig.UnmarshalText: nil pointer")
	}

	b, err := util.DecodeString(string(b))
	if err != nil {
		return errors.New("signer.Sig.UnmarshalText: " + err.Error())
	}

	*s = append((*s)[0:0], b...)
	return nil
}

func (s Sig) MarshalJSON() ([]byte, error) {
	return []byte(util.EncodeToQuoteString(s)), nil
}

func (s *Sig) UnmarshalJSON(b []byte) error {
	if s == nil {
		return errors.New("signer.Sig.UnmarshalJSON: nil pointer")
	}

	b, err := util.DecodeQuoteString(string(b))
	if err != nil {
		return errors.New("signer.Sig.UnmarshalJSON: " + err.Error())
	}

	*s = append((*s)[0:0], b...)
	return nil
}

func (s Sig) MarshalCBOR() ([]byte, error) {
	data, err := util.MarshalCBOR(s.Bytes())
	if err != nil {
		return nil, errors.New("signer.Sig.MarshalCBOR: " + err.Error())
	}
	return data, nil
}

func (s *Sig) UnmarshalCBOR(data []byte) error {
	if s == nil {
		return errors.New("signer.Sig.UnmarshalCBOR: nil pointer")
	}

	if len(data) == 1 {
		switch data[0] {
		case 0xf6: // nil
			*s = nil
			return nil
		case 0x40:
			*s = Sig{}
			return nil
		}
	}

	var b []byte
	if err := util.UnmarshalCBOR(data, &b); err != nil {
		return errors.New("signer.Sig.UnmarshalCBOR: " + err.Error())
	}

	*s = append((*s)[0:0], b...)
	return nil
}

func (s Sig) Clone() Sig {
	if s == nil {
		return nil
	}

	ns := make([]byte, len(s))
	copy(ns, s)
	return ns
}

func (s Sig) FindKey(digestHash []byte, keys ...Key) int {
	if len(digestHash) != 32 || len(keys) == 0 {
		return -1
	}

	switch s.Kind() {
	case Ed25519:
		for i, k := range keys {
			if k.Kind() == Ed25519 && ed25519.Verify(ed25519.PublicKey(k.Bytes()), digestHash, s) {
				return i
			}
		}

	case Secp256k1:
		sigcpy := make([]byte, crypto.SignatureLength)
		copy(sigcpy, s)
		if sigcpy[64] >= 27 {
			sigcpy[64] = sigcpy[64] - 27
		}
		pk, err := crypto.SigToPub(digestHash, s)
		if err != nil {
			return -1
		}

		id := crypto.PubkeyToAddress(*pk)
		for i, k := range keys {
			if k.Kind() == Secp256k1 && bytes.Equal(id[:], k.Bytes()) {
				return i
			}
		}
	}

	return -1
}

func (ss Sigs) Valid() error {
	dset := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		sigStr := s.AsKey()
		if _, ok := dset[sigStr]; ok {
			return errors.New("signer.Sigs.Valid: duplicate sig " + s.String())
		}
		dset[sigStr] = struct{}{}

		if err := s.Valid(); err != nil {
			return errors.New("signer.Sigs.Valid: " + err.Error())
		}
	}

	return nil
}
