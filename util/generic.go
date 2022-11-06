// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"errors"
	"sort"
)

type ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}

type IDg interface {
	~[20]byte | ~[32]byte

	String() string
	Valid() bool
}

type IDList[T IDg] []T

func NewIDList[T IDg](cap int) IDList[T] {
	return make(IDList[T], 0, cap)
}

func (list IDList[T]) Has(id T) bool {
	for _, v := range list {
		if v == id {
			return true
		}
	}
	return false
}

func (list IDList[T]) CheckValid() error {
	var empty T
	for _, v := range list {
		if v == empty {
			return errors.New("util.IDList.CheckValid: empty id exists")
		}
		if !v.Valid() {
			return errors.New("util.IDList.CheckValid: invalid id " + v.String())
		}
	}
	return nil
}

func (list IDList[T]) CheckDuplicate() error {
	set := make(map[T]struct{}, len(list))
	for _, v := range list {
		if _, ok := set[v]; ok {
			return errors.New("util.IDList.CheckDuplicate: duplicate id " + v.String())
		}
		set[v] = struct{}{}
	}
	return nil
}

func (list IDList[T]) Valid() error {
	if err := list.CheckValid(); err != nil {
		return err
	}
	if err := list.CheckDuplicate(); err != nil {
		return err
	}
	return nil
}

func (list IDList[T]) Equal(target IDList[T]) bool {
	if len(list) != len(target) {
		return false
	}

	for i := range list {
		if list[i] != target[i] {
			return false
		}
	}

	return true
}

type Set[T ordered] map[T]struct{}

func NewSet[T ordered](cap int) Set[T] {
	return make(Set[T], cap)
}

func (s Set[T]) Has(v T) bool {
	_, ok := s[v]
	return ok
}

func (s Set[T]) Add(vv ...T) {
	for _, v := range vv {
		s[v] = struct{}{}
	}
}

func (s Set[T]) List() []T {
	list := make([]T, 0, len(s))
	for v := range s {
		list = append(list, v)
	}
	sort.SliceStable(list, func(i, j int) bool { return list[i] < list[j] })
	return list
}
