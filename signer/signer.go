// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package signer

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
