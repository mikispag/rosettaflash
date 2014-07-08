package utils

import "fmt"

func Binary(n int, length int, reverse bool) []byte {
	values := make([]byte, length)
	rev_values := make([]byte, length)

	for i := length - 1; i >= 0; i-- {
		value := byte(1 & (uint8(n) >> uint8(i)))
		values[i] = value
		if reverse {
			rev_values[length-i-1] = value
		}
	}

	if !reverse {
		return values
	} else {
		return rev_values
	}
}

func Bits_to_byte(bits []byte) byte {
	var res uint8
	for i := 7; i >= 0; i-- {
		bit := uint8(bits[i]) & 1
		res += bit << uint8(7-i)
	}
	return byte(res)
}

func RemoveDuplicates(in []int) []int {
	found := make(map[int]bool)
	for _, i := range in {
		found[i] = true
	}
	out := make([]int, 0, len(found))
	for i := range found {
		out = append(out, i)
	}
	return out
}

func RemoveBytesDuplicates(in []byte) []byte {
	found := make(map[byte]bool)
	for _, i := range in {
		found[i] = true
	}
	out := make([]byte, 0, len(found))
	for i := range found {
		out = append(out, i)
	}
	return out
}

func CountOccurrencies(s []byte, x byte) int {
	count := 0
	for _, b := range s {
		if b == x {
			count++
		}
	}
	return count
}

func Max(in []byte) byte {
	max_so_far := byte(0)

	for _, b := range in {
		if b > max_so_far {
			max_so_far = b
		}
	}

	return max_so_far
}

func Panic(err error) {
	if err != nil {
		fmt.Errorf("Error: %v\n", err)
		panic(err)
	}
}
