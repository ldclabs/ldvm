// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"encoding/json"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/ld"
)

type EventSource string
type EventAction string

const (
	SrcAccount EventSource = "account"
	SrcModel   EventSource = "model"
	SrcData    EventSource = "data"
	SrcName    EventSource = "name"

	ActionAdd        EventAction = "add"
	ActionUpdate     EventAction = "update"
	ActionUpdateMeta EventAction = "updateMeta"
	ActionDelete     EventAction = "delete"
)

// https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/spec.md

type Event struct {
	ID          string          `json:"id"`                // Type: String, required
	Source      string          `json:"source"`            // Type: URI-reference, required
	Specversion string          `json:"specversion"`       // Type: String, required
	Type        string          `json:"type"`              // Type: String, required
	Subject     string          `json:"subject,omitempty"` // Type: String, optional
	Time        int64           `json:"time,omitempty"`    // Type: Timestamp, optional
	Data        json.RawMessage `json:"data,omitempty"`    // Type: Binary, optional
}

func NewEvent(id ids.ShortID, src EventSource, ac EventAction) *Event {
	sid := ""
	switch src {
	case SrcAccount:
		sid = ld.EthID(id).String()
	case SrcModel:
		sid = ld.ModelID(id).String()
	case SrcData:
		sid = ld.DataID(id).String()
	default:
		sid = id.String()
	}
	return &Event{
		ID:          sid,
		Source:      string(src),
		Specversion: "v1",
		Type:        fmt.Sprintf("ldvm.%s.%s", src, ac),
	}
}
