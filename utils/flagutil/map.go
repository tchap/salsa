// Copyright (c) 2013 The AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package flagutil

import (
	"errors"
	"fmt"
	"strings"
)

type MapValue struct {
	M map[string]string
}

func NewMapValue() *MapValue {
	return &MapValue{make(map[string]string)}
}

func (mv *MapValue) Set(kw string) error {
	parts := strings.SplitN(kw, ":", 2)
	if len(parts) != 2 {
		return errors.New("Multiple colons encountered")
	}
	var (
		key   = parts[0]
		value = parts[1]
	)
	if v, ok := mv.M[key]; ok {
		return fmt.Errorf("Value for key %v already set to %v", key, v)
	}
	mv.M[key] = value
	return nil
}

func (mv *MapValue) Get() interface{} {
	return mv.M
}

func (mv *MapValue) String() string {
	return fmt.Sprintf("%v", mv.M)
}
