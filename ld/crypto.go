// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
//
// Much love to the original authors for their work.
// **********
// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	"crypto/ecdsa"
	"errors"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ethereum/go-ethereum/crypto"

	"golang.org/x/crypto/sha3"
)

const (
	vOffset      = 64
	legacySigAdj = 27
)

func Sign(dh []byte, priv *ecdsa.PrivateKey) ([]byte, error) {
	sig, err := crypto.Sign(dh, priv)
	if err != nil {
		return nil, err
	}
	sig[vOffset] += legacySigAdj
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
		sigcpy[vOffset] -= legacySigAdj
	}
	return crypto.SigToPub(dh, sigcpy)
}

func DeriveSigners(data []byte, sigs []Signature) ([]ids.ShortID, error) {
	signers := make([]ids.ShortID, len(sigs))
	dh := sha3.Sum256(data)
	for i, sig := range sigs {
		pk, err := DerivePublicKey(dh[:], sig[:])
		if err != nil {
			return nil, err
		}
		signers[i] = ids.ShortID(crypto.PubkeyToAddress(*pk))
	}
	return signers, nil
}

func SatisfySigning(threshold uint8, keepers, signers []ids.ShortID, whenZero bool) bool {
	if threshold == 0 || len(keepers) == 0 {
		return whenZero
	}
	if len(signers) < int(threshold) {
		return false
	}

	set := ids.NewShortSet(len(keepers))
	set.Add(keepers...)
	t := uint8(0)
	for _, id := range signers[:] {
		if set.Contains(id) {
			t++
		}
	}
	return t >= threshold
}
