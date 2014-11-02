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

package adler32_fuzzer

import (
	"fmt"

	"github.com/mikispag/rosettaflash/adler32_mod"
	"github.com/mikispag/rosettaflash/charset"
)

func Get_appended_bytes(checksum uint32, allowed_charset *charset.Charset) []byte {
	bytes := make([]byte, 0, 128)
	S1, S2 := adler32_mod.S1(checksum), adler32_mod.S2(checksum)
	byte_to_add := byte(0)

	min_allowed := (*allowed_charset).Combinations[0]

	for !adler32_mod.S_allowed(S1, allowed_charset) {
		byte_to_add = 254 // 0xff leads to bad compression (?)
		if 65521-S1+min_allowed < 255 {
			// Add the correct byte
			byte_to_add = byte(65521 - S1 + min_allowed)
		}

		bytes = append(bytes, byte_to_add)
		S1 += int(byte_to_add)
		S2 += S1
		S1 %= 65521
		S2 %= 65521
	}

	// Now S1 has the minimum allowed value, add \x00 or a better byte till S2 is allowed
	fmt.Println("Finding final bytes to append...")

	for !adler32_mod.S_allowed(S2, allowed_charset) {
		byte_to_add = 0
		for _, b := range (*allowed_charset).Combinations {
			if 65521-S2+b < 255 && adler32_mod.S_allowed(S1+(65521-S2+b), allowed_charset) {
				// Add the correct byte
				byte_to_add = byte(65521 - S2 + b)
				fmt.Println("Found final byte!")
				break
			}
		}

		bytes = append(bytes, byte_to_add)
		S1 += int(byte_to_add)
		S2 += S1
		S1 %= 65521
		S2 %= 65521
	}

	return bytes
}
