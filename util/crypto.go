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
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fxamacker/cbor/v2"

	"golang.org/x/crypto/sha3"
)

var SignatureEmpty = Signature{}

type Signature [crypto.SignatureLength]byte

func (id Signature) String() string {
	return hex.EncodeToString(id[:])
}

func (id Signature) GoString() string {
	return id.String()
}

func (id Signature) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

func (id *Signature) UnmarshalText(b []byte) error {
	errp := ErrPrefix("Signature.UnmarshalText error: ")
	if id == nil {
		return errp.Errorf("nil pointer")
	}

	str := string(b)
	if str == "" { // If "null", do nothing
		return nil
	}

	sid, err := hex.DecodeString(str)
	switch {
	case err != nil:
		return err
	case len(sid) == crypto.SignatureLength:
		copy(id[:], sid)
	default:
		return errp.Errorf("invalid bytes length, expected 65, got %d", len(sid))
	}
	return err
}

func (id Signature) MarshalJSON() ([]byte, error) {
	return []byte("\"" + id.String() + "\""), nil
}

func (id *Signature) UnmarshalJSON(b []byte) error {
	errp := ErrPrefix("Signature.UnmarshalJSON error: ")
	if id == nil {
		return errp.Errorf("nil pointer")
	}

	str := string(b)
	if str == "null" || str == `""` { // If "null", do nothing
		return nil
	}
	lastIndex := len(str) - 1
	if str[0] != '"' || str[lastIndex] != '"' {
		return errp.Errorf("invalid string %s", strconv.Quote(str))
	}

	return id.UnmarshalText([]byte(str[1:lastIndex]))
}

func (id Signature) MarshalCBOR() ([]byte, error) {
	data, err := MarshalCBOR(id[:])
	if err != nil {
		return nil, ErrPrefix("Signature.MarshalCBOR error: ").ErrorIf(err)
	}
	return data, nil
}

func (id *Signature) UnmarshalCBOR(data []byte) error {
	errp := ErrPrefix("Signature.UnmarshalCBOR error: ")
	if id == nil {
		return errp.Errorf("nil pointer")
	}
	var b []byte
	if err := cbor.Unmarshal(data, &b); err != nil {
		return err
	}
	if len(b) != crypto.SignatureLength {
		return errp.Errorf("invalid bytes length, expected 65, got %d", len(b))
	}
	copy((*id)[:], b)
	return nil
}

func SignaturesFromStrings(ss []string) ([]Signature, error) {
	sigs := make([]Signature, len(ss))
	for i, s := range ss {
		if err := (&sigs[i]).UnmarshalText([]byte(s)); err != nil {
			return nil, ErrPrefix("SignaturesFromStrings error: ").ErrorIf(err)
		}
	}
	return sigs, nil
}

func Sign(data []byte, priv *ecdsa.PrivateKey) (Signature, error) {
	dh := sha3.Sum256(data)
	return SignHash(dh[:], priv)
}

func SignHash(datahash []byte, priv *ecdsa.PrivateKey) (Signature, error) {
	sig := Signature{}
	errp := ErrPrefix("SignHash error: ")
	data, err := crypto.Sign(datahash, priv)
	if err != nil {
		return sig, errp.ErrorIf(err)
	}
	if len(data) != crypto.SignatureLength {
		return sig, errp.Errorf("invalid signature length, expected 65, got %d", len(data))
	}
	copy(sig[:], data)
	return sig, nil
}

func DeriveSigner(data []byte, sig []byte) (EthID, error) {
	errp := ErrPrefix("DeriveSigner error: ")
	if len(data) == 0 {
		return EthIDEmpty, errp.Errorf("empty data")
	}
	if len(sig) != crypto.SignatureLength {
		return EthIDEmpty, errp.Errorf("invalid signature")
	}
	dh := sha3.Sum256(data)
	pk, err := DerivePublicKey(dh[:], sig)
	if err != nil {
		return EthIDEmpty, errp.ErrorIf(err)
	}
	return EthID(crypto.PubkeyToAddress(*pk)), nil
}

func DeriveSigners(data []byte, sigs []Signature) (EthIDs, error) {
	errp := ErrPrefix("DeriveSigners error: ")
	if len(data) == 0 {
		return nil, errp.Errorf("empty data")
	}
	if len(sigs) == 0 {
		return nil, errp.Errorf("no signature")
	}
	signers := make(EthIDs, len(sigs))
	dh := sha3.Sum256(data)
	for i, sig := range sigs {
		pk, err := DerivePublicKey(dh[:], sig[:])
		if err != nil {
			return nil, errp.ErrorIf(err)
		}
		signers[i] = EthID(crypto.PubkeyToAddress(*pk))
	}
	if err := signers.CheckDuplicate(); err != nil {
		return nil, errp.ErrorIf(err)
	}
	return signers, nil
}

func DerivePublicKey(dh []byte, sig []byte) (*ecdsa.PublicKey, error) {
	if len(sig) != crypto.SignatureLength {
		return nil, fmt.Errorf("invalid signature length, expected 65, got %d", len(sig))
	}
	// Avoid modifying the signature in place in case it is used elsewhere
	sigcpy := make([]byte, crypto.SignatureLength)
	copy(sigcpy, sig)

	// Support signers that don't apply offset (ex: ledger)
	if sigcpy[64] >= 27 {
		sigcpy[64] = sig[64] - 27
	}
	return crypto.SigToPub(dh, sigcpy)
}

func SatisfySigning(threshold uint16, keepers, signers EthIDs, whenZero bool) bool {
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
	t := uint16(0)
	for _, id := range signers {
		if _, ok := set[id]; ok {
			t++
		}
	}
	return t >= threshold
}

// SatisfySigningPlus verify for updating keepers.
func SatisfySigningPlus(threshold uint16, keepers, signers EthIDs) bool {
	if int(threshold) < len(keepers) {
		threshold += 1
	}
	return SatisfySigning(threshold, keepers, signers, false)
}
