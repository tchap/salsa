// Copyright (c) 2013 The AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package main

import (
	"strings"
)

func toPackageJsonVersion(ver *string) {
	switch strings.Count(*ver, ".") {
	case 0:
		*ver = *ver + ".0.0"
	case 1:
		*ver = *ver + ".0"
	case 2:
	case 3:
		i := strings.LastIndex(*ver, ".")
		*ver = (*ver)[:i] + "-" + (*ver)[i+1:]
	default:
		panic("invalid version string: " + *ver)
	}
}
