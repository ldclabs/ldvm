// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package signer

import (
	"crypto/ed25519"
	"errors"
	"strconv"

	"github.com/ldclabs/ldvm/util/encoding"
)

type ed25519Signer struct {
	priv ed25519.PrivateKey
	key  Key
}

func NewEd25519() (Signer, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, errors.New("signer.NewEd25519: " + err.Error())
	}

	return &ed25519Signer{
		priv: priv,
		key:  Key(pub),
	}, nil
}

func Ed25519From(privateSeed []byte) (Signer, error) {
	if seedLen := len(privateSeed); seedLen != ed25519.SeedSize {
		return nil, errors.New("signer.Ed25519From: invalid seed length, expected 32, got %d" +
			strconv.Itoa(seedLen))
	}

	priv := ed25519.NewKeyFromSeed(privateSeed)
	return &ed25519Signer{
		priv: priv,
		key:  Key(priv.Public().(ed25519.PublicKey)),
	}, nil
}

func (s *ed25519Signer) Kind() Kind {
	return Ed25519
}

func (s *ed25519Signer) Key() Key {
	return s.key
}

func (s *ed25519Signer) PrivateSeed() []byte {
	return s.priv.Seed()
}

func (s *ed25519Signer) SignHash(digestHash []byte) (Sig, error) {
	if hashLen := len(digestHash); hashLen != DigestLength {
		return nil, errors.New("signer.Signer<Ed25519>.SignHash: invalid hash length, expected 32, got " +
			strconv.Itoa(hashLen))
	}

	return ed25519.Sign(s.priv, digestHash), nil
}

func (s *ed25519Signer) SignData(message []byte) (Sig, error) {
	return s.SignHash(encoding.Sum256(message))
}

func (s *ed25519Signer) Sign(message []byte) (Sig, error) {
	return ed25519.Sign(s.priv, message), nil
}
