// go:build test

// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package signer

import (
	"crypto/ecdsa"
	"encoding/hex"
	"strings"

	"github.com/ldclabs/ldvm/util"
)

// signers for testing
var (
	// 0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc
	Signer1 *SignerTester

	// 0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641
	Signer2 *SignerTester

	// 0x6962DD0564Fb1f8459624e5b7c5dD9A38b2F990d
	Signer3 *SignerTester

	// 0x22C05D35Be1305c33810086d3A4dB598c3E1Cf48
	Signer4 *SignerTester
)

const address1 = "0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc"
const privateSeed1 = "56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"

const address2 = "0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641"
const privateSeed2 = "dc3b75ce8741f4ae37b21c8659c28d0842cbd453b00d6b69adc8c34dae3a7644"

func init() {
	pk1 := MustDecodeHex(privateSeed1)
	s1, err := Secp256k1From(pk1)
	if err != nil {
		panic(err)
	}

	Signer1 = &SignerTester{Signer: s1, PK: s1.(*secp256k1Signer).priv}
	if Signer1.Key().Address().String() != address1 {
		panic("invalid Signer1")
	}
	// fmt.Println(Signer1.Key().Address().String())
	// 0x8db97c7cECe249C2b98bdc0226cc4C2A57bF52fc
	// fmt.Println(Signer1.Key().String())
	// jbl8fOziScK5i9wCJsxMKle_UvwKxwPH

	pk2 := MustDecodeHex(privateSeed2)
	s2, err := Secp256k1From(pk2)
	if err != nil {
		panic(err)
	}

	Signer2 = &SignerTester{Signer: s2, PK: s2.(*secp256k1Signer).priv}
	if Signer2.Key().Address().String() != address2 {
		panic("invalid Signer2")
	}
	// fmt.Println(Signer2.Key().Address().String())
	// 0x44171C37Ff5D7B7bb8Dcad5C81f16284A229E641
	// fmt.Println(Signer2.Key().String())
	// RBccN_9de3u43K1cgfFihKIp5kE1lmGG

	seed := util.Sum256(Signer1.Key().Bytes())
	eSigner, err := Ed25519From(seed)
	if err != nil {
		panic(err)
	}

	Signer3 = &SignerTester{Signer: eSigner}
	// fmt.Println(Signer3.Key().Address().String())
	// 0x6962DD0564Fb1f8459624e5b7c5dD9A38b2F990d
	// fmt.Println(Signer3.Key().String())
	// OVlX-75gy0DuaRuz2k5QnlFVSuKOJezRd4CQdkIjkn5pYt0F

	seed = util.Sum256(Signer2.Key().Bytes())
	eSigner, err = BLS12381From(seed)
	if err != nil {
		panic(err)
	}

	Signer4 = &SignerTester{Signer: eSigner}
	// fmt.Println(Signer4.Key().Address().String())
	// 0x22C05D35Be1305c33810086d3A4dB598c3E1Cf48
	// fmt.Println(Signer4.Key().String())
	// hJEADz4AlkZ_NSt41-9x5eTaahzNzgMzd0wOBF-B2kJGSpWTCQutstgl0tXrZKQVIsBdNQ
}

func MustDecodeHex(str string) []byte {
	buf, err := hex.DecodeString(strings.TrimPrefix(str, "0x"))
	if err != nil {
		panic(err)
	}
	return buf
}

type SignerTester struct {
	Signer
	PK    *ecdsa.PrivateKey
	nonce uint64
}

func (s *SignerTester) MustSignData(data []byte) Sig {
	sig, err := s.SignData(data)
	if err != nil {
		panic(err)
	}
	return sig
}

func (s *SignerTester) Nonce() uint64 {
	s.nonce++
	return s.nonce
}

func NewSigner() *SignerTester {
	s, err := NewSecp256k1()
	if err != nil {
		panic(err)
	}

	return &SignerTester{Signer: s}
}
