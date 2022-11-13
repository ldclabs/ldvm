// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package signer

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"errors"
	"strconv"

	"github.com/ava-labs/avalanchego/utils/crypto/bls"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ldclabs/ldvm/util/encoding"
)

const DigestLength = 32

const (
	Unknown Kind = iota
	Ed25519
	Secp256k1
	BLS12381
)

type Kind uint8

type Signer interface {
	Kind() Kind
	Key() Key
	PrivateSeed() []byte
	SignHash(digestHash []byte) (Sig, error)
	SignData(message []byte) (Sig, error)
}

type bls12381Signer struct {
	priv *bls.SecretKey
	key  Key
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
		return nil, errors.New("signer.Signer<BLS12381>.SignHash: invalid hash length, expected 32, got " +
			strconv.Itoa(hashLen))
	}

	sig := bls.Sign(s.priv, digestHash)
	return bls.SignatureToBytes(sig), nil
}

func (s *bls12381Signer) SignData(message []byte) (Sig, error) {
	return s.SignHash(encoding.Sum256(message))
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

type ed25519Signer struct {
	priv ed25519.PrivateKey
	key  Key
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

type secp256k1Signer struct {
	priv *ecdsa.PrivateKey
	key  Key
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
