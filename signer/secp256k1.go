// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package signer

import (
	"crypto/ecdsa"
	"errors"
	"strconv"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ldclabs/ldvm/util/encoding"
)

type secp256k1Signer struct {
	priv *ecdsa.PrivateKey
	key  Key
}

func NewSecp256k1() (Signer, error) {
	priv, err := crypto.GenerateKey()
	if err != nil {
		return nil, errors.New("signer.NewSecp256k1: " + err.Error())
	}

	addr := crypto.PubkeyToAddress(priv.PublicKey)
	return &secp256k1Signer{
		priv: priv,
		key:  addr[:],
	}, nil
}

func Secp256k1From(privateSeed []byte) (Signer, error) {
	priv, err := crypto.ToECDSA(privateSeed)
	if err != nil {
		return nil, errors.New("signer.Secp256k1From: " + err.Error())
	}

	addr := crypto.PubkeyToAddress(priv.PublicKey)
	return &secp256k1Signer{
		priv: priv,
		key:  addr[:],
	}, nil
}

func (s *secp256k1Signer) Kind() Kind {
	return Secp256k1
}

func (s *secp256k1Signer) Key() Key {
	return s.key
}

func (s *secp256k1Signer) PrivateSeed() []byte {
	return crypto.FromECDSA(s.priv)
}

func (s *secp256k1Signer) SignHash(digestHash []byte) (Sig, error) {
	if hashLen := len(digestHash); hashLen != DigestLength {
		return nil, errors.New(
			"signer.Signer<Secp256k1>.SignHash: invalid hash length, expected 32, got " +
				strconv.Itoa(hashLen))
	}

	sig, err := crypto.Sign(digestHash, s.priv)
	if err != nil {
		return nil, errors.New("signer.Signer<Secp256k1>.SignHash: " + err.Error())
	}
	return sig, nil
}

func (s *secp256k1Signer) SignData(message []byte) (Sig, error) {
	return s.SignHash(encoding.Sum256(message))
}
