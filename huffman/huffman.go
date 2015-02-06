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

package huffman

import (
	"errors"
	"fmt"
	"sort"

	"github.com/mikispag/rosettaflash/charset"
	"github.com/mikispag/rosettaflash/utils"
)

type Huffman struct {
	Code_lengths []byte
	Symbols      map[byte]byte
	Trailer      byte // representation of End-of-block (code 256)
}

type HuffmanEncoder_V1 struct {
	min_allowed_code int
	max_allowed_code int
	valid_codes      []byte
}

type HuffmanEncoder_V2 struct {
	min_allowed_code int
	valid_codes      []byte
}

func InitializeHuffmanEncoder_V1(allowed_charset *charset.Charset) *HuffmanEncoder_V1 {
	var encoder HuffmanEncoder_V1
	encoder.min_allowed_code = 132
	encoder.max_allowed_code = 192
	encoder.valid_codes = encoder.generateValidCodes(allowed_charset)
	return &encoder
}

func InitializeHuffmanEncoder_V2(allowed_charset *charset.Charset) *HuffmanEncoder_V2 {
	var encoder HuffmanEncoder_V2
	encoder.min_allowed_code = 28
	encoder.valid_codes = encoder.generateValidCodes(allowed_charset)
	return &encoder
}

func (encoder *HuffmanEncoder_V1) generateValidCodes(allowed_charset *charset.Charset) []byte {
	valid_codes := make([]byte, 0, 64)
	for code := encoder.min_allowed_code; code < encoder.max_allowed_code; code++ {
		// Take 6 bits and add trailing 10
		code_trimmed_and_appended := utils.Binary(code&63, 6, true)
		code_trimmed_and_appended = append(code_trimmed_and_appended, []byte{1, 0}...)

		code_byte := utils.Bits_to_byte(code_trimmed_and_appended)

		for _, allowed := range (*allowed_charset).Binary {
			if code_byte == allowed {
				valid_codes = append(valid_codes, byte(code))
				break
			}
		}
	}
	return valid_codes
}

func (encoder *HuffmanEncoder_V2) generateValidCodes(allowed_charset *charset.Charset) []byte {
	valid_codes := make([]int, 0, 64)

	for _, allowed := range (*allowed_charset).Binary {
		if allowed >= byte(encoder.min_allowed_code) {
			valid_codes = append(valid_codes, int(allowed))
		}
	}

	sort.Ints(valid_codes)
	valid_codes_bytes := make([]byte, len(valid_codes))
	for i, v := range valid_codes {
		valid_codes_bytes[i] = byte(v)
	}

	return valid_codes_bytes
}

func (encoder *HuffmanEncoder_V1) GenerateHuffman(data []byte, allowed_charset *charset.Charset) (huffman Huffman, err error) {
	var h Huffman
	var map_codes func(symbols []int, codes []byte, valid_codes []byte, allowed_charset *charset.Charset) ([]byte, error)
	map_codes = func(symbols []int, codes []byte, valid_codes []byte, allowed_charset *charset.Charset) ([]byte, error) {
		//fmt.Printf("W map_codes valid_codes = %v\n", valid_codes)
		symbols_length := len(symbols)
		codes_length := len(codes)
		valid_length := len(valid_codes)

		if symbols_length == codes_length {
			return codes, nil
		}

		if valid_length == 0 {
			return []byte{0}, errors.New("No result.")
		}

		prev_code := int(codes[codes_length-1])
		prev_symbol := symbols[codes_length-1]
		symbol := symbols[codes_length]

		max_code_1 := prev_code + symbol - prev_symbol

		max_code_index := -(symbols_length - codes_length) % valid_length

		if max_code_index < 0 {
			max_code_index += valid_length
		}

		max_code_2 := int(valid_codes[max_code_index])
		max_code := byte(0)

		if max_code_1 < max_code_2 {
			max_code = byte(max_code_1)
		} else {
			max_code = byte(max_code_2)
		}

		reachable_codes := make([]byte, 0, valid_length)
		for _, code := range valid_codes {
			if code <= max_code {
				reachable_codes = append(reachable_codes, code)
			}
		}

		if byte(symbol) == data[len(data)-1] {
			reachable_codes_temp := make([]byte, 0, valid_length)
			for _, code := range reachable_codes {
				lsb := utils.Binary(int(code), 8, true)[2:]
				byte_code := append(lsb, []byte{0, 0}...)
				byte_code_byte := utils.Bits_to_byte(byte_code)

				for _, allowed := range (*allowed_charset).Binary {
					if allowed == byte_code_byte {
						reachable_codes_temp = append(reachable_codes_temp, code)
						break
					}
				}
				reachable_codes = reachable_codes_temp
			}
		}

		//fmt.Printf("W reachable_codes = %v\n", reachable_codes)

		for i := len(reachable_codes) - 1; i >= 0; i-- {
			current_code := reachable_codes[i]
			next_codes := make([]byte, 0, valid_length)

			for _, valid := range valid_codes {
				if valid > current_code {
					next_codes = append(next_codes, valid)
				}
			}

			assigned_codes, err := map_codes(symbols, append(codes, current_code), next_codes, allowed_charset)
			if err == nil {
				return assigned_codes, nil
			}
		}

		return []byte{0}, errors.New("No result.")
	}

	fmt.Printf("GenerateHuffman(V1, data = %x)\n", data)

	min_allowed_code := (*encoder).min_allowed_code

	// Generate a sorted list of bytes in data
	sorted_bytes_set := make([]int, len(data))
	for i, b := range data {
		sorted_bytes_set[i] = int(b)
	}
	sorted_bytes_set = utils.RemoveDuplicates(sorted_bytes_set)
	sort.Ints(sorted_bytes_set)

	assigned_codes, err := map_codes(append([]int{-1}, sorted_bytes_set...), []byte{byte(min_allowed_code - 1)}, (*encoder).valid_codes, allowed_charset)
	if err != nil {
		return h, err
	}

	assigned_codes = assigned_codes[1:]

	//fmt.Printf("W assigned_codes = %v\n", assigned_codes)

	symbols := make(map[byte]byte)
	for i := 0; i < len(sorted_bytes_set); i++ {
		symbols[byte(sorted_bytes_set[i])] = assigned_codes[i]
	}

	slack_2 := 0
	slack_6 := 1
	slack_8 := int(assigned_codes[0]) - min_allowed_code
	code_lengths := make([]byte, 0, len(assigned_codes))

	for len(code_lengths) < 257 || slack_2 > 0 || slack_6 > 0 || slack_8 > 0 {
		if len(sorted_bytes_set) > 0 && len(code_lengths) == sorted_bytes_set[0] {
			code_lengths = append(code_lengths, 8)
			current_code := assigned_codes[0]
			assigned_codes = assigned_codes[1:]
			sorted_bytes_set = sorted_bytes_set[1:]
			if len(assigned_codes) > 0 {
				slack_8 = int(assigned_codes[0]) - int(current_code) - 1
			} else {
				slack_8 = 124 - utils.CountOccurrencies(code_lengths, 8)
			}
		} else if len(code_lengths) == 256 {
			code_lengths = append(code_lengths, 2)
			slack_2 = 1
		} else if slack_8 > 0 {
			code_lengths = append(code_lengths, 8)
			slack_8--
		} else if slack_6 > 0 {
			code_lengths = append(code_lengths, 6)
			slack_6--
		} else if slack_2 > 0 {
			code_lengths = append(code_lengths, 2)
			slack_2--
		} else {
			code_lengths = append(code_lengths, 0)
		}
	}

	// Check for HLIT
	extra_code_lengths := 257 - len(code_lengths)
	if extra_code_lengths < 16 && extra_code_lengths > 12 || extra_code_lengths > 28 {
		return h, errors.New("Invalid HLIT.")
	}

	// fmt.Printf("W code_lengths = %v\nW symbols = %v\n", code_lengths, symbols)

	// Populate h
	h.Code_lengths = code_lengths
	h.Symbols = symbols
	h.Trailer = 0

	fmt.Println("Huffman found.")

	return h, nil
}

func (encoder *HuffmanEncoder_V2) GenerateHuffman(data []byte) (huffman Huffman, err error) {
	var h Huffman
	var map_codes func(symbols []int, codes []byte, valid_codes []byte) ([]byte, error)
	map_codes = func(symbols []int, codes []byte, valid_codes []byte) ([]byte, error) {
		symbols_length := len(symbols)
		codes_length := len(codes)
		valid_length := len(valid_codes)

		if symbols_length == codes_length {
			return codes, nil
		}

		prev_code := int(codes[codes_length-1])
		prev_symbol := symbols[codes_length-1]
		symbol := symbols[codes_length]

		max_code := prev_code + symbol - prev_symbol

		reachable_codes := make([]byte, 0, valid_length)
		for _, code := range valid_codes {
			if int(code) <= max_code {
				reachable_codes = append(reachable_codes, code)
			}
		}

		for i := len(reachable_codes) - 1; i >= 0; i-- {
			current_code := reachable_codes[i]
			next_codes := make([]byte, 0, valid_length)

			for _, valid := range valid_codes {
				if valid > current_code {
					next_codes = append(next_codes, valid)
				}
			}

			assigned_codes, err := map_codes(symbols, append(codes, current_code), next_codes)
			if err == nil {
				return assigned_codes, nil
			}
		}

		return []byte{0}, errors.New("No result.")
	}

	fmt.Printf("GenerateHuffman(V2, data = %x)\n", data)

	min_allowed_code := (*encoder).min_allowed_code

	// Generate a sorted list of bytes in data
	sorted_bytes_set := make([]int, len(data))
	for i, b := range data {
		sorted_bytes_set[i] = int(b)
	}
	sorted_bytes_set = utils.RemoveDuplicates(sorted_bytes_set)
	sort.Ints(sorted_bytes_set)

	assigned_codes, err := map_codes(append([]int{-1}, sorted_bytes_set...), []byte{byte(min_allowed_code - 1)}, (*encoder).valid_codes)
	if err != nil {
		return h, err
	}

	assigned_codes = assigned_codes[1:]
	//fmt.Printf("W assigned_codes = %v\n", assigned_codes)

	symbols := make(map[byte]byte)
	for i := 0; i < len(sorted_bytes_set); i++ {
		symbols[byte(sorted_bytes_set[i])] = assigned_codes[i]
	}

	slack_6 := 3
	slack_8 := int(assigned_codes[0]) - min_allowed_code
	code_lengths := make([]byte, 0, len(assigned_codes))

	for len(code_lengths) < 257 || slack_6 > 0 || slack_8 > 0 {
		if len(sorted_bytes_set) > 0 && len(code_lengths) == sorted_bytes_set[0] {
			code_lengths = append(code_lengths, 8)
			current_code := assigned_codes[0]
			assigned_codes = assigned_codes[1:]
			sorted_bytes_set = sorted_bytes_set[1:]
			if len(assigned_codes) > 0 {
				slack_8 = int(assigned_codes[0]) - int(current_code) - 1
			} else {
				slack_8 = 228 - utils.CountOccurrencies(code_lengths, 8)
			}
		} else if len(code_lengths) == 256 {
			if slack_6 > 0 {
				return h, errors.New("No Huffman.")
			} else {
				code_lengths = append(code_lengths, 6)
				slack_6 = 3
			}
		} else if slack_8 > 0 {
			code_lengths = append(code_lengths, 8)
			slack_8--
		} else if slack_6 > 0 {
			code_lengths = append(code_lengths, 6)
			slack_6--
		} else {
			code_lengths = append(code_lengths, 0)
		}
	}

	// fmt.Printf("W code_lengths = %v\nW symbols = %v\n", code_lengths, symbols)

	// Populate h
	h.Code_lengths = code_lengths
	h.Symbols = symbols
	h.Trailer = 3

	fmt.Println("Huffman found.")

	return h, nil
}
