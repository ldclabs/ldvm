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

func DeriveSender(data []byte, sig []byte) (ids.ShortID, error) {
	dh := crypto.Keccak256Hash(data).Bytes()
	pk, err := DerivePublicKey(dh, sig)
	if err != nil {
		return ids.ShortEmpty, err
	}
	return ids.ShortID(crypto.PubkeyToAddress(*pk)), nil
}

func DeriveSigners(data []byte, sigs []Signature) ([]ids.ShortID, error) {
	signers := make([]ids.ShortID, len(sigs))
	dh := crypto.Keccak256Hash(data).Bytes()
	for i, sig := range sigs {
		pk, err := DerivePublicKey(dh, sig[:])
		if err != nil {
			return nil, err
		}
		signers[i] = ids.ShortID(crypto.PubkeyToAddress(*pk))
	}
	return signers, nil
}
