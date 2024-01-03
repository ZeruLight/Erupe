package mhfcid

import (
	"math"
)

// ConvertCID converts a MHF Character ID String to integer
//
// Banned characters: 0, I, O, S
func ConvertCID(ID string) (r uint32) {
	if len(ID) != 6 {
		return
	}

	m := map[rune]uint32{
		'1': 0,
		'2': 1,
		'3': 2,
		'4': 3,
		'5': 4,
		'6': 5,
		'7': 6,
		'8': 7,
		'9': 8,
		'A': 9,
		'B': 10,
		'C': 11,
		'D': 12,
		'E': 13,
		'F': 14,
		'G': 15,
		'H': 16,
		'J': 17,
		'K': 18,
		'L': 19,
		'M': 20,
		'N': 21,
		'P': 22,
		'Q': 23,
		'R': 24,
		'T': 25,
		'U': 26,
		'V': 27,
		'W': 28,
		'X': 29,
		'Y': 30,
		'Z': 31,
	}

	for i, c := range ID {
		r += m[c] * uint32(math.Pow(32, float64(i)))
	}
	return
}
