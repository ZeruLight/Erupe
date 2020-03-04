package deltacomp

import (
	"errors"

	"github.com/Andoryuuta/byteframe"
)

func checkReadUint8(bf *byteframe.ByteFrame) (uint8, error) {
	if len(bf.DataFromCurrent()) >= 1 {
		return bf.ReadUint8(), nil
	}
	return 0, errors.New("Not enough data")
}

func checkReadUint16(bf *byteframe.ByteFrame) (uint16, error) {
	if len(bf.DataFromCurrent()) >= 2 {
		return bf.ReadUint16(), nil
	}
	return 0, errors.New("Not enough data")
}

func readCount(bf *byteframe.ByteFrame) (int, error) {
	var count int

	count8, err := checkReadUint8(bf)
	if err != nil {
		return 0, err
	}
	count = int(count8)

	if count == 0 {
		count16, err := checkReadUint16(bf)
		if err != nil {
			return 0, err
		}
		count = int(count16)
	}

	return int(count), nil
}

// ApplyDataDiff applies a delta data diff patch onto given base data.
func ApplyDataDiff(diff []byte, baseData []byte) []byte {
	// Make a copy of the base data to return,
	// (probably just make this modify the given slice in the future).
	baseCopy := make([]byte, len(baseData))
	copy(baseCopy, baseData)

	patch := byteframe.NewByteFrameFromBytes(diff)

	// The very first matchCount is +1 more than it should be, so we start at -1.
	dataOffset := -1
	for {
		// Read the amount of matching bytes.
		matchCount, err := readCount(patch)
		if err != nil {
			// No more data
			break
		}

		dataOffset += matchCount

		// Read the amount of differing bytes.
		differentCount, err := readCount(patch)
		if err != nil {
			// No more data
			break
		}
		differentCount -= 1

		// Apply the patch bytes.
		for i := 0; i < differentCount; i++ {
			b, err := checkReadUint8(patch)
			if err != nil {
				panic("Invalid or misunderstood patch format!")
			}
			baseCopy[dataOffset+i] = b
		}

		dataOffset += differentCount - 1

	}

	return baseCopy
}
