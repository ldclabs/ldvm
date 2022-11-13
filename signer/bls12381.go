// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package signer

import (
	"errors"
	"strconv"

	"github.com/ava-labs/avalanchego/utils/crypto/bls"

	"github.com/ldclabs/ldvm/util/encoding"
)

type bls12381Signer struct {
	priv *bls.SecretKey
	key  Key
}

func NewBLS12381() (Signer, error) {
	priv, err := bls.NewSecretKey()
	if err != nil {
		return nil, errors.New("signer.NewBLS12381: " + err.Error())
	}

	return &bls12381Signer{
		priv: priv,
		key:  bls.PublicKeyToBytes(bls.PublicFromSecretKey(priv)),
	}, nil
}

func BLS12381From(privateSeed []byte) (Signer, error) {
	if seedLen := len(privateSeed); seedLen != bls.SecretKeyLen {
		return nil, errors.New("signer.BLS12381From: invalid seed length, expected 32, got %d" +
			strconv.Itoa(seedLen))
	}

	priv, err := bls.SecretKeyFromBytes(privateSeed)
	if err != nil {
		return nil, errors.New("signer.BLS12381From: " + err.Error())
	}

	return &bls12381Signer{
		priv: priv,
		key:  bls.PublicKeyToBytes(bls.PublicFromSecretKey(priv)),
	}, nil
}

func (s *bls12381Signer) Kind() Kind {
	return BLS12381
}

func (s *bls12381Signer) Key() Key {
	return s.key
}

func (s *bls12381Signer) PrivateSeed() []byte {
	return bls.SecretKeyToBytes(s.priv)
}

func (s *bls12381Signer) SignHash(digestHash []byte) (Sig, error) {
	if hashLen := len(digestHash); hashLen != DigestLength {
		return nil, errors.New("signer.Signer<BLS12-381>.SignHash: invalid hash length, expected 32, got " +
			strconv.Itoa(hashLen))
	}

	sig := bls.Sign(s.priv, digestHash)
	return bls.SignatureToBytes(sig), nil
}

func (s *bls12381Signer) SignData(message []byte) (Sig, error) {
	return s.SignHash(encoding.Sum256(message))
}
