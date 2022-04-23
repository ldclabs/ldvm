// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"fmt"
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ldclabs/ldvm/util"
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
	ID          string `json:"id"`                // Type: String, required
	Source      string `json:"source"`            // Type: URI-reference, required
	Specversion string `json:"specversion"`       // Type: String, required
	Type        string `json:"type"`              // Type: String, required
	Subject     string `json:"subject,omitempty"` // Type: String, optional
	Time        int64  `json:"time,omitempty"`    // Type: Timestamp, optional
	Data        string `json:"data,omitempty"`    // Type: String, optional
}

func NewEvent(id ids.ShortID, src EventSource, ac EventAction) *Event {
	sid := ""
	switch src {
	case SrcAccount:
		sid = util.EthID(id).String()
	case SrcModel:
		sid = util.ModelID(id).String()
	case SrcData:
		sid = util.DataID(id).String()
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

type EventsCache struct {
	mu     sync.RWMutex
	size   int
	events []*Event
}

func NewEventsCache(size int) *EventsCache {
	return &EventsCache{size: size, events: make([]*Event, 0, size)}
}

func (e *EventsCache) Query() []*Event {
	e.mu.RLock()
	defer e.mu.RUnlock()

	events := make([]*Event, len(e.events))
	copy(events, e.events)
	return events
}

func (e *EventsCache) Add(events ...*Event) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if ln := len(e.events) + len(events); ln > e.size {
		copy(e.events, e.events[ln-e.size:])
		copy(e.events[e.size-len(events):], events)
	} else {
		e.events = append(e.events, events...)
	}
}
