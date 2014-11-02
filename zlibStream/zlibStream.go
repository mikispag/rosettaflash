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

package zlibStream

import (
	"bytes"
	"encoding/binary"

	"github.com/mikispag/rosettaflash/charset"
	"github.com/mikispag/rosettaflash/huffman"
	"github.com/mikispag/rosettaflash/utils"
)

type Run struct {
	code_length byte
	runs        int
}

type ZlibStream struct {
	bits     bytes.Buffer
	bytes    bytes.Buffer
	header   []byte
	checksum []byte
}

func InitializeZlibStream() *ZlibStream {
	var z ZlibStream
	z.bits.Grow(8 * 1024 * 128)
	return &z
}

func (d *ZlibStream) ByteDisalignment() int {
	return d.bits.Len() % 8
}

func (d *ZlibStream) WriteHeader() []byte {
	return []byte("hC")
}

func (d *ZlibStream) WriteChecksum(checksum uint32) []byte {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, checksum)
	return bytes
}

func (d *ZlibStream) WriteBits(bits []byte) {
	d.bits.Write(bits)
}

func (d *ZlibStream) WriteBytes() bytes.Buffer {
	var bytes bytes.Buffer
	bytes.Grow(1024 * 128)

	//fmt.Printf("W BEGIN WriteBytes - len(bits) = %v\n", d.bits.Len())

	/*
		if d.ByteDisalignment() > 0 {
			fmt.Println("Byte disalignment detected. Compensating.")
			pad := make([]byte, d.ByteDisalignment())
			d.bits.Write(pad)
		}
	*/

	for d.bits.Len() > 0 {
		current_byte := make([]byte, 8)
		_, err := d.bits.Read(current_byte)
		utils.Panic(err)

		new_byte := make([]byte, 8)

		for j := 7; j >= 0; j-- {
			bit := current_byte[j]
			new_byte[7-j] = byte(bit)
		}
		b := utils.Bits_to_byte(new_byte)
		bytes.WriteByte(b)
	}

	//fmt.Printf("W END WriteBytes - len(bytes) = %v\n", bytes.Len())
	return bytes
}

func (d *ZlibStream) WriteFullByteStream(checksum uint32) {
	d.header = d.WriteHeader()
	d.bytes = d.WriteBytes()
	d.checksum = d.WriteChecksum(checksum)
}

func (d *ZlibStream) WriteLenOfLen(lengths []int) {
	for _, length := range lengths {
		d.WriteBits(utils.Binary(length, 3, false))
	}
}

func (d *ZlibStream) WritePaddingBlock() {
	lenOfLen := []int{2, 5, 0, 4, 3, 0, 6, 4, 4, 4, 4, 6, 2}
	/*
		+------+--------+
		| Code | Length |
		+------+--------+
		| 16   | 2      |
		| 17   | 5      |
		| 18   | -      |
		| 0    | 4      |
		| 8    | 3      |
		| 7    | -      |
		| 9    | 6      |
		| 6    | 4      |
		| 10   | 4      |
		| 5    | 4      |
		| 11   | 4      |
		| 4    | 6      |
		| 12   | 2      |
		+------+--------+
	*/

	encode := func(code []byte, n int) {
		first := true
		for n > 0 {
			if !first && n > 6 && d.ByteDisalignment() == 0 {
				x := 10
				if n < 10 {
					x = n
				}
				d.WriteBits([]byte{0, 1})
				d.WriteBits(utils.Binary(x-7, 2, false))
				d.WriteBits([]byte{0, 1, 1, 0})
				n -= x
			} else {
				d.WriteBits(code)
				n--
			}
			first = false
		}
	}

	// Header
	d.WriteBits([]byte{0})                  // BFINAL
	d.WriteBits([]byte{0, 1})               // BTYPE = 10 - dynamic Huffman codes
	d.WriteBits(utils.Binary(8, 5, false))  // HLIT
	d.WriteBits(utils.Binary(16, 5, false)) // HDIST
	d.WriteBits(utils.Binary(9, 4, false))  // HCLEN = len(lenOfLen) - 4 = 9

	// Table
	// Lengths of lengths
	d.WriteLenOfLen(lenOfLen)

	// Literal + lengths
	encode([]byte{1, 0, 1, 0}, 197)
	encode([]byte{1, 1, 0, 0}, 64)
	// encode([]byte{1, 0, 1, 0}, 4) // (see below)

	// Distance
	// encode([]byte{1, 0, 1, 0}, 17) // (see below)

	// Specs allow to combine
	encode([]byte{1, 0, 1, 0}, 21)

	// End of block
	d.WriteBits([]byte{1, 1, 1, 0, 1, 1})
}

func (d *ZlibStream) Compress(block []byte, h *huffman.Huffman, is_last bool) {
	lenOfLen := []int{2, 5, 3, 4, 4, 5, 4, 4, 4, 0, 3, 5, 0, 5, 0, 4, 0}
	/*
		+------+--------+
		| Code | Length |
		+------+--------+
		| 16   | 2      |
		| 17   | 5      |
		| 18   | 3      |
		| 0    | 4      |
		| 8    | 4      |
		| 7    | 5      |
		| 9    | 4      |
		| 6    | 4      |
		| 10   | 4      |
		| 5    | -      |
		| 11   | 3      |
		| 4    | 5      |
		| 12   | -      |
		| 3    | 5      |
		| 13   | -      |
		| 2    | 4      |
		| 14   | -      |
		+------+--------+
	*/

	code_lengths := (*h).Code_lengths
	symbols_map := (*h).Symbols

	encode := func(code []byte, n int) {
		//fmt.Printf("W encode(%v, %v)\n", code, n)
		first := true
		for n > 0 {
			if !first && n > 6 && d.ByteDisalignment() == 2 {
				x := n / 6
				for i := 0; i < x; i++ {
					//fmt.Printf("W write 0011\n")
					d.WriteBits([]byte{0, 0, 1, 1})
				}
				n -= x * 6
			} else {
				d.WriteBits(code)
				n--
			}
			first = false
		}
	}

	//fmt.Printf("W BEGIN Compress\n")

	// Header
	if is_last {
		d.WriteBits([]byte{1}) // BFINAL
	} else {
		d.WriteBits([]byte{0}) // BFINAL
	}
	d.WriteBits([]byte{0, 1})                                  // BTYPE = 10 - dynamic Huffman codes
	d.WriteBits(utils.Binary(len(code_lengths)-257, 5, false)) // HLIT
	d.WriteBits(utils.Binary(5, 5, false))                     // HDIST
	d.WriteBits(utils.Binary(13, 4, false))                    // HCLEN = len(lenOfLen) - 4 = 13

	// Table
	// Lengths of lengths
	d.WriteLenOfLen(lenOfLen)

	// Encode
	runs_map := make([]Run, 0, len(code_lengths))
	for _, code_length := range code_lengths {
		map_len := len(runs_map)
		if map_len > 0 && runs_map[map_len-1].code_length == code_length {
			runs_map[map_len-1].runs++
		} else {
			var run Run
			run.code_length = code_length
			run.runs = 1
			runs_map = append(runs_map, run)
		}
	}
	for _, run := range runs_map {
		switch run.code_length {
		case 0:
			encode([]byte{1, 0, 0, 0}, run.runs)
		case 2:
			encode([]byte{1, 0, 0, 1}, run.runs)
		case 6:
			encode([]byte{1, 0, 1, 0}, run.runs)
		case 8:
			encode([]byte{1, 0, 1, 1}, run.runs)
		}
	}

	// Distance
	if d.ByteDisalignment() == 2 {
		d.WriteBits([]byte{1, 0, 0, 0}) // H 0
		d.WriteBits([]byte{1, 0, 0, 0}) // H 0
		d.WriteBits([]byte{0, 0})       // H 16
		d.WriteBits([]byte{1, 0})       // REPEAT 4x

	} else {
		d.WriteBits([]byte{1, 0, 0, 1}) // H 2
		d.WriteBits([]byte{0, 0})       // H 16
		d.WriteBits([]byte{0, 0})       // REPEAT 3x
		d.WriteBits([]byte{1, 0, 0, 0}) // H 0
		d.WriteBits([]byte{1, 0, 0, 0}) // H 0
	}

	// Data
	for _, b := range block {
		new_symbol := symbols_map[b]
		d.WriteBits(utils.Binary(int(new_symbol), 8, true))
	}
	d.WriteBits(utils.Binary(int(h.Trailer), 2, true))
	//fmt.Printf("W END Compress\n")
}

func (d *ZlibStream) CompressVariant(block []byte, h *huffman.Huffman, is_last bool) {
	lenOfLen := []int{2, 4, 3, 4, 4, 5, 4, 4, 4, 0, 3, 5, 4}
	/*
		+------+--------+
		| Code | Length |
		+------+--------+
		| 16   | 2      |
		| 17   | 4      |
		| 18   | 3      |
		| 0    | 4      |
		| 8    | 4      |
		| 7    | 5      |
		| 9    | 4      |
		| 6    | 4      |
		| 10   | 4      |
		| 5    | -      |
		| 11   | 3      |
		| 4    | 5      |
		| 12   | 4      |
		+------+--------+
	*/

	code_lengths := (*h).Code_lengths
	symbols_map := (*h).Symbols

	encode := func(code []byte, n int) {
		//fmt.Printf("W encode(%v, %v)\n", code, n)
		first := true
		for n > 0 {
			if !first && n > 6 && d.ByteDisalignment() == 2 {
				x := n / 6
				for i := 0; i < x; i++ {
					d.WriteBits([]byte{0, 0, 1, 1})
				}
				n -= x * 6
			} else {
				d.WriteBits(code)
				n--
			}
			first = false
		}
	}

	//fmt.Printf("W BEGIN CompressVariant\n")

	// Header
	if is_last {
		d.WriteBits([]byte{1}) // BFINAL
	} else {
		d.WriteBits([]byte{0}) // BFINAL
	}
	d.WriteBits([]byte{0, 1})                                  // BTYPE = 10 - dynamic Huffman codes
	d.WriteBits(utils.Binary(len(code_lengths)-257, 5, false)) // HLIT
	d.WriteBits(utils.Binary(25, 5, false))                    // HDIST
	d.WriteBits(utils.Binary(9, 4, false))                     // HCLEN = len(lenOfLen) - 4 = 9

	// Table
	// Lengths of lengths
	d.WriteLenOfLen(lenOfLen)

	// Encode
	runs_map := make([]Run, 0, len(code_lengths))
	for _, code_length := range code_lengths {
		map_len := len(runs_map)
		if map_len > 0 && runs_map[map_len-1].code_length == code_length {
			runs_map[map_len-1].runs++
		} else {
			var run Run
			run.code_length = code_length
			run.runs = 1
			runs_map = append(runs_map, run)
		}
	}
	for _, run := range runs_map {
		switch run.code_length {
		case 0:
			encode([]byte{1, 0, 0, 0}, run.runs)
		case 6:
			encode([]byte{1, 0, 0, 1}, run.runs)
		case 8:
			encode([]byte{1, 0, 1, 0}, run.runs)
		}
	}

	// Distance
	if d.ByteDisalignment() == 2 {
		d.WriteBits([]byte{0, 1, 1})            // H 18
		d.WriteBits(utils.Binary(11, 7, false)) // REPEAT ZERO 22x
		d.WriteBits([]byte{0, 0, 1, 0})         // H 16 + REPEAT 4x
	} else {
		d.WriteBits([]byte{1, 0, 0, 0})         // H 0
		d.WriteBits([]byte{0, 1, 1})            // H 18
		d.WriteBits(utils.Binary(10, 7, false)) // REPEAT ZERO 21x
		d.WriteBits([]byte{0, 0, 1, 0})         // H 16 + REPEAT 4x
	}

	// Data
	for _, b := range block {
		new_symbol := symbols_map[b]
		d.WriteBits(utils.Binary(int(new_symbol), 8, true))
	}
	d.WriteBits(utils.Binary(int(h.Trailer), 6, true))

	//fmt.Printf("W END CompressVariant\n")
}

func (d *ZlibStream) GetBytes() []byte {
	bytes := make([]byte, 0, 10240)
	bytes = append(bytes, d.header...)
	bytes = append(bytes, d.bytes.Bytes()...)
	bytes = append(bytes, d.checksum...)
	return bytes
}

func (zlibStream *ZlibStream) Start(data []byte, allowed_charset *charset.Charset) {
	var h, h_temp huffman.Huffman
	var err error

	previous_block_main_encoder := true

	encoder_V1 := huffman.InitializeHuffmanEncoder_V1(allowed_charset)
	encoder_V2 := huffman.InitializeHuffmanEncoder_V2(allowed_charset)

	for len(data) > 0 {
		block_main_encoder := true
		i := 1

		// Choose a long chunk for dynamic Huffman encoding
		bytes_set := make([]byte, 0, 10240)
		bytes_set = append(bytes_set, data[0])

		for ; i < len(data) && len(bytes_set) <= 50 && utils.Max(bytes_set) < 216; i++ {
			bytes_set = append(bytes_set, data[i])
			bytes_set = utils.RemoveBytesDuplicates(bytes_set)
		}
		if i != len(data) {
			i--
		}

		// Find the longest chunk the dynamic encoder is able to encode
		for i > 0 {
			h, err = encoder_V1.GenerateHuffman(data[:i], allowed_charset)
			if err == nil {
				break
			}
			i--
		}

		// Also try a different Huffman
		if i == 0 {
			i = 1
		}
		for i <= len(data) {
			h_temp, err = encoder_V2.GenerateHuffman(data[:i])
			if err != nil {
				break
			}
			h = h_temp
			block_main_encoder = false
			i++
		}
		if !block_main_encoder {
			i--
		}

		// Encode data with the Huffman we found
		block := data[:i]
		data = data[i:]
		last_block := len(data) == 0
		if previous_block_main_encoder {
			zlibStream.WritePaddingBlock()
		}
		if block_main_encoder {
			zlibStream.Compress(block, &h, last_block)
		} else {
			zlibStream.CompressVariant(block, &h, last_block)
		}
		previous_block_main_encoder = block_main_encoder
	}
}
