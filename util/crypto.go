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

	"golang.org/x/crypto/sha3"
)

const (
	vOffset      = 64
	legacySigAdj = 27
)

var SignatureEmpty = Signature{}

type Signature [crypto.SignatureLength]byte

func (s Signature) MarshalJSON() ([]byte, error) {
	str, err := formatting.EncodeWithChecksum(formatting.Hex, s[:])
	if err != nil {
		return nil, err
	}
	buf := make([]byte, len(str)+2)
	buf[0] = '"'
	copy(buf[1:], []byte(str))
	buf[len(buf)-1] = '"'
	return buf, nil
}

func SignaturesFromStrings(ss []string) ([]Signature, error) {
	sigs := make([]Signature, len(ss))
	for i, s := range ss {
		d, err := formatting.Decode(formatting.Hex, s)
		if err != nil {
			return nil, fmt.Errorf("invalid signature %s, decode failed: %v", strconv.Quote(s), err)
		}
		if len(d) != crypto.SignatureLength {
			return nil, fmt.Errorf("invalid signature %s", strconv.Quote(s))
		}
		sigs[i] = Signature{}
		copy(sigs[i][:], d)
	}
	return sigs, nil
}

func Sign(dh []byte, priv *ecdsa.PrivateKey) (Signature, error) {
	sig := Signature{}
	data, err := crypto.Sign(dh, priv)
	if err != nil {
		return sig, err
	}
	if len(data) != crypto.SignatureLength {
		return sig, fmt.Errorf("invalid signature")
	}
	data[vOffset] = data[vOffset] + legacySigAdj
	copy(sig[:], data)
	return sig, nil
}

func DerivePublicKey(dh []byte, sig []byte) (*ecdsa.PublicKey, error) {
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

func SignData(data []byte, priv *ecdsa.PrivateKey) (Signature, error) {
	dh := sha3.Sum256(data)
	return Sign(dh[:], priv)
}

func DeriveSigner(data []byte, sig []byte) (EthID, error) {
	if len(data) == 0 || len(sig) != crypto.SignatureLength {
		return EthIDEmpty, fmt.Errorf("no data or signatures to derive")
	}
	dh := sha3.Sum256(data)
	pk, err := DerivePublicKey(dh[:], sig)
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
		pk, err := DerivePublicKey(dh[:], sig[:])
		if err != nil {
			return nil, err
		}
		signers[i] = EthID(crypto.PubkeyToAddress(*pk))
	}
	return signers, nil
}

func SatisfySigning(threshold uint8, keepers, signers []EthID, whenZero bool) bool {
	if threshold == 0 || len(keepers) == 0 {
		return whenZero
	}
	if len(signers) < int(threshold) {
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
