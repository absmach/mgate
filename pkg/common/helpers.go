// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package common

import "strings"

func AddSuffixSlash(path string) string {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}
