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
