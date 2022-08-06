// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ld

import (
	cborpatch "github.com/ldclabs/cbor-patch"
	jsonpatch "github.com/ldclabs/json-patch"

	"github.com/ldclabs/ldvm/util"
)

var (
	// LM111111111111111111116DBWJs
	RawModelID = util.ModelIDEmpty
	// LM1111111111111111111Ax1asG
	CBORModelID = util.ModelID{
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
	}
	// LM1111111111111111111L17Xp3
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
	// keepers who owned this data, no more than 1024
	Keepers     util.EthIDs  `cbor:"kp" json:"keepers"`
	Approver    *util.EthID  `cbor:"ap,omitempty" json:"approver,omitempty"`
	ApproveList TxTypes      `cbor:"apl,omitempty" json:"approveList,omitempty"`
	Data        util.RawData `cbor:"d" json:"data"`
	// data signature claims
	SigClaims *SigClaims `cbor:"sc,omitempty" json:"sigClaims,omitempty"`
	// data signature signing by a certificate authority
	Sig *util.Signature `cbor:"s,omitempty" json:"sig,omitempty"`

	// external assignment fields
	ID  util.DataID `cbor:"-" json:"id"`
	raw []byte      `cbor:"-" json:"-"`
}

func (t *DataInfo) Clone() *DataInfo {
	x := new(DataInfo)
	*x = *t

	x.Keepers = make(util.EthIDs, len(t.Keepers))
	copy(x.Keepers, t.Keepers)
	if t.Approver != nil {
		id := *t.Approver
		x.Approver = &id
	}
	if t.ApproveList != nil {
		x.ApproveList = make(TxTypes, len(t.ApproveList))
		copy(x.ApproveList, t.ApproveList)
	}
	x.Data = make([]byte, len(t.Data))
	copy(x.Data, t.Data)

	if t.SigClaims != nil {
		sc := *t.SigClaims
		x.SigClaims = &sc
	}
	if t.Sig != nil {
		sig := *t.Sig
		x.Sig = &sig
	}
	x.raw = nil
	return x
}

// SyntacticVerify verifies that a *DataInfo is well-formed.
func (t *DataInfo) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("DataInfo.SyntacticVerify error: ")

	switch {
	case t == nil:
		return errp.Errorf("nil pointer")

	case len(t.Keepers) > MaxKeepers:
		return errp.Errorf("too many keepers")

	case int(t.Threshold) > len(t.Keepers):
		return errp.Errorf("invalid threshold")

	case t.Approver != nil && *t.Approver == util.EthIDEmpty:
		return errp.Errorf("invalid approver")

	case t.Sig == nil && t.SigClaims != nil:
		return errp.Errorf("invalid signature")

	case t.Sig != nil && t.SigClaims == nil:
		return errp.Errorf("invalid signature claims")
	}

	if err = t.Keepers.CheckDuplicate(); err != nil {
		return errp.Errorf("invalid keepers, %v", err)
	}

	if err = t.Keepers.CheckEmptyID(); err != nil {
		return errp.Errorf("invalid keepers, %v", err)
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
	}

	if t.raw, err = t.Marshal(); err != nil {
		return errp.ErrorIf(err)
	}
	return nil
}

// Signer returns the signer of the DataInfo.
// Should be called after DataInfo.SyntacticVerify.
// Should be called with DataInfo.ID.
func (t *DataInfo) Signer() (signer util.EthID, err error) {
	errp := util.ErrPrefix("DataInfo.Signer error: ")

	switch {
	case t.Sig == nil || t.SigClaims == nil:
		return signer, errp.Errorf("invalid signature claims")

	case t.SigClaims.Subject != t.ID:
		return signer, errp.Errorf("invalid subject, expected %s, got %s",
			t.ID, t.SigClaims.Subject)

	case t.SigClaims.Audience != t.ModelID:
		return signer, errp.Errorf("invalid audience, expected %s, got %s",
			t.ModelID, t.SigClaims.Audience)

	case t.SigClaims.CWTID != util.HashFromData(t.Data):
		return signer, errp.Errorf("invalid CWT id")
	}

	signer, err = util.DeriveSigner(t.SigClaims.Bytes(), t.Sig[:])
	return signer, errp.ErrorIf(err)
}

func (t *DataInfo) MarkDeleted(data []byte) error {
	t.Version = 0
	t.SigClaims = nil
	t.Sig = nil
	t.Data = data
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
	errp := util.ErrPrefix("DataInfo.Patch error: ")

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

	return errp.ErrorMap(p.Apply(t.Data))
}

func (t *DataInfo) Bytes() []byte {
	if len(t.raw) == 0 {
		t.raw = MustMarshal(t)
	}
	return t.raw
}

func (t *DataInfo) Unmarshal(data []byte) error {
	return util.ErrPrefix("DataInfo.Unmarshal error: ").
		ErrorIf(util.UnmarshalCBOR(data, t))
}

func (t *DataInfo) Marshal() ([]byte, error) {
	return util.ErrPrefix("DataInfo.Marshal error: ").
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
	CWTID      util.Hash    `cbor:"7,keyasint" json:"cti"` // the hash of DataInfo.Data

	// external assignment fields
	raw []byte `cbor:"-" json:"-"`
}

func (s *SigClaims) SyntacticVerify() error {
	var err error
	errp := util.ErrPrefix("SigClaims.SyntacticVerify error: ")

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
	return util.ErrPrefix("SigClaims.Unmarshal error: ").
		ErrorIf(util.UnmarshalCBOR(data, s))
}

func (s *SigClaims) Marshal() ([]byte, error) {
	return util.ErrPrefix("SigClaims.Marshal error: ").
		ErrorMap(util.MarshalCBOR(s))
}
