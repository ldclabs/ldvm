// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
//
// Much love to the original authors for their work.
// **********
// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"strconv"

	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fxamacker/cbor/v2"

	"golang.org/x/crypto/sha3"
)

const (
	vOffset      = 64
	legacySigAdj = 27
)

var SignatureEmpty = Signature{}

type Signature [crypto.SignatureLength]byte

func (id Signature) String() string {
	str, _ := formatting.EncodeWithChecksum(formatting.Hex, id[:])
	return str
}

func (id Signature) GoString() string {
	return id.String()
}

func (id Signature) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *Signature) UnmarshalText(b []byte) error {
	if id == nil {
		return fmt.Errorf("Signature: UnmarshalText on nil pointer")
	}

	str := string(b)
	if str == "" { // If "null", do nothing
		return nil
	}

	sid, err := formatting.Decode(formatting.Hex, str)
	switch {
	case err != nil:
		return err
	case len(sid) == crypto.SignatureLength:
		copy(id[:], sid)
	default:
		return fmt.Errorf("Signature: UnmarshalText on invalid length bytes %d", len(sid))
	}
	return err
}

func (id Signature) MarshalJSON() ([]byte, error) {
	return []byte("\"" + id.String() + "\""), nil
}

func (id *Signature) UnmarshalJSON(b []byte) error {
	if id == nil {
		return fmt.Errorf("Signature: UnmarshalJSON on nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}
	lastIndex := len(str) - 1
	if str[0] != '"' || str[lastIndex] != '"' {
		return fmt.Errorf("Signature: UnmarshalJSON on invalid string %s", strconv.Quote(str))
	}

	return id.UnmarshalText([]byte(str[1:lastIndex]))
}

func (id Signature) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(id[:])
}

func (id *Signature) UnmarshalCBOR(data []byte) error {
	if id == nil {
		return fmt.Errorf("Signature: UnmarshalCBOR on nil pointer")
	}
	var b []byte
	if err := cbor.Unmarshal(data, &b); err != nil {
		return err
	}
	if len(b) != crypto.SignatureLength {
		return fmt.Errorf("Signature: UnmarshalCBOR on invalid length bytes: %d", len(b))
	}
	copy((*id)[:], b)
	return nil
}

func SignaturesFromStrings(ss []string) ([]Signature, error) {
	sigs := make([]Signature, len(ss))
	for i, s := range ss {
		if err := (&sigs[i]).UnmarshalText([]byte(s)); err != nil {
			return nil, err
		}
	}
	return sigs, nil
}

func Sign(data []byte, priv *ecdsa.PrivateKey) (Signature, error) {
	dh := sha3.Sum256(data)
	return signHash(dh[:], priv)
}

func signHash(datahash []byte, priv *ecdsa.PrivateKey) (Signature, error) {
	sig := Signature{}
	data, err := crypto.Sign(datahash, priv)
	if err != nil {
		return sig, err
	}
	if len(data) != crypto.SignatureLength {
		return sig, fmt.Errorf("invalid signature length")
	}
	data[vOffset] = data[vOffset] + legacySigAdj
	copy(sig[:], data)
	return sig, nil
}

func DeriveSigner(data []byte, sig []byte) (EthID, error) {
	if len(data) == 0 || len(sig) != crypto.SignatureLength {
		return EthIDEmpty, fmt.Errorf("no data or signature to derive")
	}
	dh := sha3.Sum256(data)
	pk, err := derivePublicKey(dh[:], sig)
	if err != nil {
		return EthIDEmpty, err
	}
	return EthID(crypto.PubkeyToAddress(*pk)), nil
}

func DeriveSigners(data []byte, sigs []Signature) ([]EthID, error) {
	if len(data) == 0 || len(sigs) == 0 {
		return nil, fmt.Errorf("no data or signatures to derive")
	}
	signers := make([]EthID, len(sigs))
	dh := sha3.Sum256(data)
	for i, sig := range sigs {
		pk, err := derivePublicKey(dh[:], sig[:])
		if err != nil {
			return nil, err
		}
		signers[i] = EthID(crypto.PubkeyToAddress(*pk))
	}
	return signers, nil
}

func derivePublicKey(dh []byte, sig []byte) (*ecdsa.PublicKey, error) {
	if len(sig) != crypto.SignatureLength {
		return nil, errors.New("invalid signature")
	}
	// Avoid modifying the signature in place in case it is used elsewhere
	sigcpy := make([]byte, crypto.SignatureLength)
	copy(sigcpy, sig)

	// Support signers that don't apply offset (ex: ledger)
	if sigcpy[vOffset] >= legacySigAdj {
		sigcpy[vOffset] = sig[vOffset] - legacySigAdj
	}
	return crypto.SigToPub(dh, sigcpy)
}

func SatisfySigning(threshold uint8, keepers, signers []EthID, whenZero bool) bool {
	if threshold == 0 || len(keepers) == 0 {
		return whenZero
	}
	if len(signers) < int(threshold) || len(keepers) < int(threshold) {
		return false
	}

	set := make(map[EthID]struct{}, len(keepers))
	for _, v := range keepers {
		set[v] = struct{}{}
	}
	t := uint8(0)
	for _, id := range signers {
		if _, ok := set[id]; ok {
			t++
		}
	}
	return t >= threshold
}

// SatisfySigningPlus verify for updating keepers.
func SatisfySigningPlus(threshold uint8, keepers, signers []EthID) bool {
	if int(threshold) < len(keepers) {
		threshold += 1
	}
	return SatisfySigning(threshold, keepers, signers, false)
}
