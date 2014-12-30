/*
 * Copyright 2014 Google Inc. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package charset

import (
	"sort"

	"github.com/mikispag/rosettaflash/utils"
)

type Charset struct {
	Allowed      []byte
	Binary       []byte
	Combinations []int
}

func New(allowed_charset []byte) *Charset {
	var charset Charset
	charset.Allowed = allowed_charset
	charset.Binary = generate_binary_charset(allowed_charset)
	charset.Combinations = generate_pairs_of_allowed(allowed_charset)
	return &charset
}

func generate_binary_charset(allowed_charset []byte) []byte {
	binary_charset := make([]byte, len(allowed_charset))
	for i, allowed := range allowed_charset {
		binary := utils.Binary(int(allowed), 8, false)
		binary_charset[i] = utils.Bits_to_byte(binary)
	}
	return binary_charset
}

func generate_pairs_of_allowed(allowed_charset []byte) []int {
	charset_length := len(allowed_charset)
	combinations := make([]int, 0, charset_length*charset_length)
	for _, b1 := range allowed_charset {
		for _, b2 := range allowed_charset {
			high_byte := int(b1) * 256
			combinations = append(combinations, high_byte+int(b2))
		}
	}
	combinations = utils.RemoveDuplicates(combinations)

	sort.Ints(combinations)
	return combinations
}
