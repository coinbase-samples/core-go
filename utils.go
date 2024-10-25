/**
 * Copyright 2024-present Coinbase Global, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"github.com/shopspring/decimal"
)

func StrToNum(v string) (amount decimal.Decimal, err error) {
	amount, err = decimal.NewFromString(v)
	return
}

// StrSliceDiff returns a new slice of strings that are not present in b, but not
// in a (O(n^2)). This is an exact string match, so case is important.
func StrSliceDiff(a []string, b []string) (diff []string) {

	for _, bv := range b {
		var found bool
		for _, av := range a {
			if av == bv {
				found = true
				break
			}
		}

		if !found {
			diff = append(diff, bv)
		}
	}

	return
}
