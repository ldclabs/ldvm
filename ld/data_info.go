// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	cborpatch "github.com/ldclabs/cbor-patch"
	jsonpatch "github.com/ldclabs/json-patch"

	"github.com/ldclabs/ldvm/util"
	"github.com/ldclabs/ldvm/util/signer"
)

var (
	// AAAAAAAAAAAAAAAAAAAAAAAAAADzaDye
	RawModelID = util.ModelIDEmpty
	// AAAAAAAAAAAAAAAAAAAAAAAAAAGIYKah
	CBORModelID = util.ModelID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	}
	// AAAAAAAAAAAAAAAAAAAAAAAAAALZFhrw
	JSONModelID = util.ModelID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 2,
	}
)

type DataInfo struct {
	ModelID util.ModelID `cbor:"m" json:"mid"` // model id
	// data version，the initial value is 1, should increase 1 when updating,
	// 0 indicates that the data is invalid, for example, deleted or punished.
	Version uint64 `cbor:"v" json:"version"`
	// MultiSig: m of n, threshold is m, keepers length is n.
	// The minimum value is 0, means no one can update the data.
	// the maximum value is len(keepers)
	Threshold uint16 `cbor:"th" json:"threshold"`
	// keepers who owned this data, no more than 64
	Keepers     signer.Keys  `cbor:"kp" json:"keepers"`
	Approver    signer.Key   `cbor:"ap" json:"approver,omitempty"`
	ApproveList TxTypes      `cbor:"apl" json:"approveList,omitempty"`
	Payload     util.RawData `cbor:"pl" json:"payload"`
	// data signature claims
	SigClaims *SigClaims `cbor:"sc,omitempty" json:"sigClaims,omitempty"`
	// data signature signing by a certificate authority
	Sig *signer.Sig `cbor:"s,omitempty" json:"sig,omitempty"`

	// external assignment fields
	ID  util.DataID `cbor:"-" json:"id"`
	raw []byte      `cbor:"-" json:"-"`
}

func (t *DataInfo) Clone() *DataInfo {
	x := new(DataInfo)
	*x = *t

	x.Keepers = t.Keepers.Clone()
	if t.Approver != nil {
		x.Approver = t.Approver.Clone()
	}
	if t.ApproveList != nil {
		x.ApproveList = make(TxTypes, len(t.ApproveList))
		copy(x.ApproveList, t.ApproveList)
	}
	x.Payload = make([]byte, len(t.Payload))
	copy(x.Payload, t.Payload)

	if t.SigClaims != nil {
		sc := *t.SigClaims
		x.SigClaims = &sc
	}
	if t.Sig != nil {
		sig := t.Sig.Clone()
		x.Sig = &sig
	}
	x.raw = nil
	return x
}

// SyntacticVerify verifies that a *DataInfo is well-formed.
func (t *DataInfo) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("ld.DataInfo.SyntacticVerify: ")

	switch {
	case t == nil:
		return errp.Errorf("nil pointer")

	case len(t.Keepers) > MaxKeepers:
		return errp.Errorf("too many keepers")

	case int(t.Threshold) > len(t.Keepers):
		return errp.Errorf("invalid threshold")

	case t.SigClaims == nil && t.Sig != nil:
		return errp.Errorf("no sigClaims, signature should be nil")

	case t.SigClaims != nil && t.Sig == nil:
		return errp.Errorf("invalid signature")
	}

	if err = t.Keepers.Valid(); err != nil {
		return errp.Errorf("invalid keepers, %v", err)
	}

	if t.Approver != nil {
		if err = t.Approver.Valid(); err != nil {
			return errp.Errorf("invalid approver, %v", err)
		}
	}

	if t.ApproveList != nil {
		if err = t.ApproveList.CheckDuplicate(); err != nil {
			return errp.Errorf("invalid approveList, %v", err)
		}

		for _, ty := range t.ApproveList {
			if !DataTxTypes.Has(ty) {
				return errp.Errorf("invalid TxType %s in approveList", ty)
			}
		}
	}

	if t.SigClaims != nil {
		if err = t.SigClaims.SyntacticVerify(); err != nil {
			return errp.ErrorIf(err)
		}
		if err = t.Sig.Valid(); err != nil {
			return errp.ErrorIf(err)
		}
	}

	if t.raw, err = t.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

func (t *DataInfo) Verify(digestHash []byte, sigs signer.Sigs) bool {
	return t.Keepers.Verify(digestHash, sigs, t.Threshold)
}

func (t *DataInfo) VerifyPlus(digestHash []byte, sigs signer.Sigs) bool {
	return t.Keepers.VerifyPlus(digestHash, sigs, t.Threshold)
}

// ValidSigClaims should be called after DataInfo.SyntacticVerify.
// ValidSigClaims should be called with DataInfo.ID.
func (t *DataInfo) ValidSigClaims() error {
	if t.SigClaims == nil {
		return nil
	}

	errp := util.ErrPrefix("ld.DataInfo.ValidSigClaims: ")
	switch {
	case t.ID == util.DataIDEmpty:
		return errp.Errorf("invalid data id")

	case t.Sig.Kind() == signer.Unknown:
		return errp.Errorf("invalid signature")

	case t.SigClaims.Subject != t.ID:
		return errp.Errorf("invalid subject, expected %s, got %s",
			t.ID, t.SigClaims.Subject)

	case t.SigClaims.Audience != t.ModelID:
		return errp.Errorf("invalid audience, expected %s, got %s",
			t.ModelID, t.SigClaims.Audience)

	case t.SigClaims.CWTID != util.HashFromData(t.Payload):
		return errp.Errorf("invalid CWT id")
	}

	return nil
}

func (t *DataInfo) MarkDeleted(data []byte) error {
	t.Version = 0
	t.SigClaims = nil
	t.Sig = nil
	t.Payload = data
	return t.SyntacticVerify()
}

type patcher interface {
	Apply(doc []byte) ([]byte, error)
}

// Patch applies a patch to the data, returns the patched data.
// It will not change the data.
func (t *DataInfo) Patch(operations []byte) ([]byte, error) {
	var err error
	var p patcher
	errp := util.ErrPrefix("ld.DataInfo.Patch: ")

	switch t.ModelID {
	case RawModelID:
		return operations, nil

	case CBORModelID:
		p, err = cborpatch.NewPatch(operations)
		if err != nil {
			return nil, errp.Errorf("invalid CBOR patch, %v", err)
		}

	case JSONModelID:
		p, err = jsonpatch.NewPatch(operations)
		if err != nil {
			return nil, errp.Errorf("invalid JSON patch, %v", err)
		}

	default:
		return nil, errp.Errorf("unsupport mid %s", t.ModelID)
	}

	return errp.ErrorMap(p.Apply(t.Payload))
}

func (t *DataInfo) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *DataInfo) Unmarshal(data []byte) error {
	return util.ErrPrefix("ld.DataInfo.Unmarshal: ").
		ErrorIf(util.UnmarshalCBOR(data, t))
}

func (t *DataInfo) Marshal() ([]byte, error) {
	return util.ErrPrefix("ld.DataInfo.Marshal: ").
		ErrorMap(util.MarshalCBOR(t))
}

// SigClaims is a set of claims that used to sign a DataInfo.
// reference to https://www.rfc-editor.org/rfc/rfc8392.html#section-3
type SigClaims struct {
	Issuer     util.DataID  `cbor:"1,keyasint" json:"iss"` // the id of certificate authority
	Subject    util.DataID  `cbor:"2,keyasint" json:"sub"` // the id of DataInfo
	Audience   util.ModelID `cbor:"3,keyasint" json:"aud"` // the model id of DataInfo
	Expiration uint64       `cbor:"4,keyasint" json:"exp"`
	NotBefore  uint64       `cbor:"5,keyasint" json:"nbf"`
	IssuedAt   uint64       `cbor:"6,keyasint" json:"iat"`
	CWTID      util.Hash    `cbor:"7,keyasint" json:"cti"` // the hash of DataInfo.Payload

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

func (s *SigClaims) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("ld.SigClaims.SyntacticVerify: ")

	switch {
	case s == nil:
		return errp.Errorf("nil pointer")

	case s.Issuer == util.DataIDEmpty:
		return errp.Errorf("invalid issuer")

	case s.Subject == util.DataIDEmpty:
		return errp.Errorf("invalid subject")

	case s.Expiration == 0:
		return errp.Errorf("invalid expiration time")

	case s.IssuedAt == 0:
		return errp.Errorf("invalid issued time")

	case s.CWTID == util.HashEmpty:
		return errp.Errorf("invalid CWT id")
	}

	if s.raw, err = s.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}

	return nil
}

func (s *SigClaims) Bytes() []byte {
	if len(s.raw) == 0 {
		s.raw = MustMarshal(s)
	}
	return s.raw
}

func (s *SigClaims) Unmarshal(data []byte) error {
	return util.ErrPrefix("ld.SigClaims.Unmarshal: ").
		ErrorIf(util.UnmarshalCBOR(data, s))
}

func (s *SigClaims) Marshal() ([]byte, error) {
	return util.ErrPrefix("ld.SigClaims.Marshal: ").
		ErrorMap(util.MarshalCBOR(s))
}
