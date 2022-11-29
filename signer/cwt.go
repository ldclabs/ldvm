// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package signer

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"time"

	"github.com/fxamacker/cbor/v2"

	"github.com/ldclabs/ldvm/ids"
	"github.com/ldclabs/ldvm/util/encoding"
)

const (
	headerLabelAlgorithm = 1
	headerLabelKeyID     = 4
)

// Reference: https://www.iana.org/assignments/cbor-tags/cbor-tags.xhtml#tags
const (
	cborTagCOSESign1 = 18
)

// sign1MessagePrefix represents the fixed prefix of COSE_Sign1_Tagged.
var sign1MessagePrefix = []byte{
	0xd2, // #6.18
	0x84, // array of length 4
}

// Reference:  https://www.iana.org/assignments/cose/cose.xhtml#algorithms
// ed25519ProtectedHeader represents:
//
//	cbor.Marshal(map[int]interface{}{
//		headerLabelAlgorithm: AlgorithmEd25519,
//	})
var ed25519ProtectedHeader = []byte{
	0xa1, // #5.1
	0x01, // key 1
	0x27, // value -8
}

// CWT is a simple wrapper around a CBOR Web Token (CWT) for LDC.
// CBOR Web Token (CWT) https://www.rfc-editor.org/rfc/rfc8392.html
// CBOR Object Signing and Encryption (COSE) https://www.rfc-editor.org/rfc/rfc8152.html
type CWT struct {
	Claims Claims // https://www.rfc-editor.org/rfc/rfc8392.html#section-3
	Key    Key
	Sig    Sig
	ExData cbor.RawMessage // Externally Supplied Data, https://www.rfc-editor.org/rfc/rfc8152.html#section-4.3
}

// Claims is a set of claims that used to sign data.
type Claims struct {
	Issuer     string   `cbor:"1,keyasint" json:"iss"` // OPTIONAL, if not present, the issuer is the same as the subject
	Subject    string   `cbor:"2,keyasint" json:"sub"` // REQUIRED
	Audience   string   `cbor:"3,keyasint" json:"aud"` // OPTIONAL
	Expiration uint64   `cbor:"4,keyasint" json:"exp"` // OPTIONAL, seconds relative to 1970-01-01T00:00Z in UTC time.
	NotBefore  uint64   `cbor:"5,keyasint" json:"nbf"` // OPTIONAL, seconds relative to 1970-01-01T00:00Z in UTC time.
	IssuedAt   uint64   `cbor:"6,keyasint" json:"iat"` // OPTIONAL, seconds relative to 1970-01-01T00:00Z in UTC time.
	CWTID      ids.ID32 `cbor:"7,keyasint" json:"cti"` // REQUIRED, should be SHA3-256 digest of the data.

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

func (c *Claims) SyntacticVerify() error {
	var err error

	switch {
	case c == nil:
		return errors.New("signer.Claims.SyntacticVerify: nil pointer")

	case c.Subject == "":
		return errors.New("signer.Claims.SyntacticVerify: subject required")

	case len(c.CWTID) != 32 || !c.CWTID.Valid():
		return errors.New("signer.Claims.SyntacticVerify: invalid CWT id")
	}

	if c.raw, err = c.Marshal(); err != nil {
		return err
	}

	return nil
}

func (c *Claims) Bytes() []byte {
	if len(c.raw) == 0 {
		var err error
		if c.raw, err = c.Marshal(); err != nil {
			panic(errors.New("signer.Claims.Bytes: " + err.Error()))
		}
	}
	return c.raw
}

func (c *Claims) Unmarshal(data []byte) error {
	if err := encoding.UnmarshalCBOR(data, c); err != nil {
		return errors.New("signer.Claims.Unmarshal: " + err.Error())
	}
	return nil
}

func (c *Claims) Marshal() ([]byte, error) {
	data, err := encoding.MarshalCBOR(c)
	if err != nil {
		return nil, errors.New("signer.Claims.Marshal: " + err.Error())
	}
	return data, nil
}

type sign1Message struct {
	_           struct{} `cbor:",toarray"`
	Protected   cbor.RawMessage
	Unprotected cbor.RawMessage
	Payload     cbor.RawMessage
	Signature   Sig
}

func (c *CWT) WithSign(signer Signer) error {
	if err := c.Claims.SyntacticVerify(); err != nil {
		return err
	}

	if signer.Kind() != Ed25519 {
		return errors.New("signer.Claims.ToCWT: invalid signer kind, expected Ed25519")
	}

	if c.ExData == nil {
		c.ExData = []byte{}
	}

	toBeSigned, err := encoding.MarshalCBOR([]interface{}{
		"Signature1",           // context
		ed25519ProtectedHeader, // body_protected
		c.ExData,               // external_aad
		c.Claims.Bytes(),       // payload
	})
	if err != nil {
		return err
	}

	c.Key = signer.Key()
	c.Sig, err = signer.Sign(toBeSigned)
	return err
}

func (c CWT) MarshalCBOR() ([]byte, error) {
	unprotected, err := encoding.MarshalCBOR(map[int]Key{
		headerLabelKeyID: c.Key,
	})
	if err != nil {
		return nil, err
	}

	if len(c.Sig) == 0 {
		return nil, errors.New("signer.CWT.MarshalCBOR: should call CWT.WithSign(signer)")
	}

	return encoding.MarshalCBOR(cbor.Tag{
		Number: cborTagCOSESign1,
		Content: sign1Message{
			Protected:   ed25519ProtectedHeader,
			Unprotected: unprotected,
			Payload:     c.Claims.Bytes(),
			Signature:   c.Sig,
		},
	})
}

func (c *CWT) UnmarshalCBOR(data []byte) error {
	if c == nil {
		return errors.New("signer.CWT.UnmarshalCBOR: nil pointer")
	}

	if !bytes.HasPrefix(data, sign1MessagePrefix) {
		return errors.New("signer.CWT.UnmarshalCBOR: invalid COSE_Sign1_Tagged object")
	}

	var msg sign1Message
	if err := encoding.UnmarshalCBOR(data[1:], &msg); err != nil {
		return err
	}

	if string(msg.Protected) != string(ed25519ProtectedHeader) {
		return errors.New("signer.CWT.UnmarshalCBOR: invalid protected header")
	}

	unprotected := map[int]Key{}
	if err := encoding.UnmarshalCBOR(msg.Unprotected, &unprotected); err != nil {
		return err
	}

	ok := false
	if c.Key, ok = unprotected[headerLabelKeyID]; !ok || c.Key.Kind() != Ed25519 {
		return errors.New("signer.CWT.UnmarshalCBOR: invalid Ed25519 kid in unprotected header")
	}

	if msg.Signature.Kind() != Ed25519 {
		return errors.New("signer.CWT.UnmarshalCBOR: invalid Ed25519 signature")
	}

	c.Sig = msg.Signature
	if err := c.Claims.Unmarshal(msg.Payload); err != nil {
		return err
	}

	return c.Claims.SyntacticVerify()
}

func (c *CWT) Verify() error {
	if c.Key.Kind() != Ed25519 {
		return errors.New("signer.CWT.Verify: invalid Ed25519 key")
	}

	if c.Sig.Kind() != Ed25519 {
		return errors.New("signer.CWT.Verify: invalid Ed25519 signature")
	}

	now := uint64(time.Now().Unix())
	if c.Claims.Expiration > 0 && c.Claims.Expiration <= now {
		return errors.New("signer.CWT.Verify: expired")
	}

	if c.Claims.NotBefore > 0 && c.Claims.NotBefore > now {
		return errors.New("signer.CWT.Verify: not yet valid")
	}

	if c.ExData == nil {
		c.ExData = []byte{}
	}

	toBeSigned, err := encoding.MarshalCBOR([]interface{}{
		"Signature1",           // context
		ed25519ProtectedHeader, // body_protected
		c.ExData,               // external_aad
		c.Claims.Bytes(),       // payload
	})
	if err != nil {
		return err
	}

	if !ed25519.Verify(ed25519.PublicKey(c.Key.Bytes()), toBeSigned, c.Sig.Bytes()) {
		return errors.New("signer.CWT.Verify: Ed25519 verify failed")
	}

	return nil
}
