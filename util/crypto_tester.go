// go:build test

// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
//
// Much love to the original authors for their work.

package util

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

var Signer1, Signer2 *Signer

func init() {
	pk1, err := crypto.HexToECDSA("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")
	if err != nil {
		panic(err)
	}
	Signer1 = &Signer{PK: pk1}
	if EthID(Signer1.Address()).String() != "0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC" {
		panic("invalid Signer1")
	}
	pk2, err := crypto.HexToECDSA("dc3b75ce8741f4ae37b21c8659c28d0842cbd453b00d6b69adc8c34dae3a7644")
	if err != nil {
		panic(err)
	}
	Signer2 = &Signer{PK: pk2}
	if EthID(Signer2.Address()).String() != "0x44171C37Ff5D7B7bb8dcad5C81f16284A229e641" {
		panic("invalid Signer2")
	}
}

type Signer struct {
	PK    *ecdsa.PrivateKey
	nonce uint64
}

func (s *Signer) SignData(data []byte) (Signature, error) {
	dh := sha3.Sum256(data)
	return SignHash(dh[:], s.PK)
}

func (s *Signer) SignHash(digestHash []byte) (Signature, error) {
	return SignHash(digestHash, s.PK)
}

func (s *Signer) MustSignData(data []byte) *Signature {
	sig, err := s.SignData(data)
	if err != nil {
		panic(err)
	}
	return &sig
}

func (s *Signer) Address() EthID {
	return EthID(crypto.PubkeyToAddress(s.PK.PublicKey))
}

func (s *Signer) Nonce() uint64 {
	s.nonce++
	return s.nonce
}

func NewSigner() *Signer {
	pk, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	return &Signer{PK: pk}
}
