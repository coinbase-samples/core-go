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

import "testing"

func TestStrSliceDiff(t *testing.T) {

	cases := []struct {
		description string
		a           []string
		b           []string
		expected    []string
	}{
		{
			description: "TestStrSliceDiff0",
			a:           []string{"a", "b", "c"},
			b:           []string{"a", "b", "c", "d"},
			expected:    []string{"d"},
		},
		{
			description: "TestStrSliceDiff1",
			a:           []string{"a", "b", "c"},
			b:           []string{"a", "b", "c"},
			expected:    []string{},
		},
		{
			description: "TestStrSliceDiff2",
			a:           []string{},
			b:           []string{"a", "b", "c"},
			expected:    []string{"a", "b", "c"},
		},
		{
			description: "TestStrSliceDiff3",
			a:           []string{"a", "b", "c"},
			b:           []string{},
			expected:    []string{},
		},
		{
			description: "TestStrSliceDiff4",
			a:           []string{"a", "b", "c"},
			b:           []string{"d"},
			expected:    []string{"d"},
		},
		{
			description: "TestStrSliceDiff5",
			a:           []string{"a", "b", "c"},
			b:           []string{"A"},
			expected:    []string{"A"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			result := StrSliceDiff(tt.a, tt.b)

			if len(result) != len(tt.expected) {
				t.Fatalf("test: %s - expected: %v - received: %v", tt.description, tt.expected, result)
			}

			for i, _ := range result {
				if tt.expected[i] != result[i] {
					t.Fatalf("test: %s - expected: %v - received: %v", tt.description, tt.expected, result)
				}
			}
		})
	}
}
