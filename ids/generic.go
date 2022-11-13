// (c) 2022-2022, LDC Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ids

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
)

type orderedT interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}

type idT interface {
	~[20]byte | ~[32]byte

	String() string
	Bytes() []byte
	Valid() bool
}

type IDList[T idT] []T

func NewIDList[T idT](cap int) IDList[T] {
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
			return errors.New("ids.IDList.CheckValid: empty id exists")
		}
		if !v.Valid() {
			return errors.New("ids.IDList.CheckValid: invalid id " + v.String())
		}
	}
	return nil
}

func (list IDList[T]) CheckDuplicate() error {
	set := make(map[T]struct{}, len(list))
	for _, v := range list {
		if _, ok := set[v]; ok {
			return errors.New("ids.IDList.CheckDuplicate: duplicate id " + v.String())
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

func (list IDList[T]) Sort()    { sort.Stable(list) }
func (list IDList[T]) Len() int { return len(list) }
func (list IDList[T]) Less(i, j int) bool {
	return bytes.Compare(list[i].Bytes(), list[j].Bytes()) == -1
}
func (list IDList[T]) Swap(i, j int) { list[j], list[i] = list[i], list[j] }

type Set[T orderedT] map[T]struct{}

func NewSet[T orderedT](cap int) Set[T] {
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

func (s Set[T]) CheckAdd(vv ...T) error {
	for _, v := range vv {
		if s.Has(v) {
			return fmt.Errorf("ids.Set.CheckAdd: duplicate value %v", v)
		}
		s[v] = struct{}{}
	}
	return nil
}

func (s Set[T]) List() List[T] {
	list := make(List[T], 0, len(s))
	for v := range s {
		list = append(list, v)
	}
	list.Sort()
	return list
}

type List[T orderedT] []T

func (list List[T]) Sort()              { sort.Stable(list) }
func (list List[T]) Len() int           { return len(list) }
func (list List[T]) Less(i, j int) bool { return list[i] < list[j] }
func (list List[T]) Swap(i, j int)      { list[j], list[i] = list[i], list[j] }
