// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package sync

import (
	ss "sync"
)

// Mutex is reexport sync.Mutex for easy use.
// More info sees https://pkg.go.dev/sync#Mutex
type Mutex = ss.Mutex

// RWMutex is reexport sync.RWMutex for easy use.
// More info sees https://pkg.go.dev/sync#RWMutex
type RWMutex = ss.RWMutex

// WaitGroup is reexport sync.WaitGroup for easy use.
// More info sees https://pkg.go.dev/sync#WaitGroup
type WaitGroup = ss.WaitGroup
