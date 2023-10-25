package bitfield_test

import (
	"errors"
	"testing"

	"github.com/Doridian/foxTorrent/pkg/bitfield"
	"github.com/stretchr/testify/assert"
)

func TestCreateFromBytes(t *testing.T) {
	bf := bitfield.NewBitfieldFromBytes([]byte{0b10000000, 0b00100000})
	for i := 0; i < 16; i++ {
		shouldBe := i == 0 || i == 10
		bitIs := bf.GetBit(uint64(i))
		assert.Equal(t, shouldBe, bitIs)
	}
}

func TestForEachMatchingBit(t *testing.T) {
	bf := bitfield.NewBitfieldFromBytes([]byte{0b10000011, 0b10100000})

	matchingBits := make([]uint64, 0, 16)
	bf.ForEachMatchingBit(0, true, func(index uint64) error {
		matchingBits = append(matchingBits, index)
		return nil
	})
	assert.Equal(t, []uint64{0, 6, 7, 8, 10}, matchingBits)

	matchingBits = matchingBits[:0]
	expectErr := errors.New("test error")
	err := bf.ForEachMatchingBit(0, true, func(index uint64) error {
		matchingBits = append(matchingBits, index)
		if index == 7 {
			return expectErr
		}
		return nil
	})
	assert.ErrorIs(t, err, expectErr)
	assert.Equal(t, []uint64{0, 6, 7}, matchingBits)

	matchingBits = matchingBits[:0]
	bf.ForEachMatchingBit(0, false, func(index uint64) error {
		matchingBits = append(matchingBits, index)
		return nil
	})
	assert.Equal(t, []uint64{1, 2, 3, 4, 5, 9, 11, 12, 13, 14, 15}, matchingBits)

	matchingBits = matchingBits[:0]
	bf.ForEachMatchingBit(6, true, func(index uint64) error {
		matchingBits = append(matchingBits, index)
		return nil
	})
	assert.Equal(t, []uint64{6, 7, 8, 10}, matchingBits)
}

func TestGetBit(t *testing.T) {
	bf := bitfield.NewBitfieldFromBytes([]byte{0b10000000, 0b00100000})
	assert.True(t, bf.GetBit(0))
	assert.False(t, bf.GetBit(1))
	assert.False(t, bf.GetBit(2))
	assert.False(t, bf.GetBit(3))
	assert.False(t, bf.GetBit(4))
	assert.False(t, bf.GetBit(5))
	assert.False(t, bf.GetBit(6))
	assert.False(t, bf.GetBit(7))
	assert.False(t, bf.GetBit(8))
	assert.False(t, bf.GetBit(9))
	assert.True(t, bf.GetBit(10))
	assert.False(t, bf.GetBit(11))
	assert.False(t, bf.GetBit(12))
	assert.False(t, bf.GetBit(13))
	assert.False(t, bf.GetBit(14))
	assert.False(t, bf.GetBit(15))
	assert.False(t, bf.GetBit(16))
	assert.False(t, bf.GetBit(17))
	assert.False(t, bf.GetBit(18))
}

func TestSetBit(t *testing.T) {
	bf := bitfield.NewBitfield(16)
	bf.SetBit(0)
	bf.SetBit(10)
	bf.SetBit(10)
	bf.SetBit(16)
	bf.SetBit(17)
	bf.SetBit(18)
	assert.Equal(t, []byte{0b10000000, 0b00100000}, bf.GetData())
}

func TestClearBit(t *testing.T) {
	bf := bitfield.NewBitfieldFromBytes([]byte{0b10000000, 0b00100000})
	bf.ClearBit(0)
	bf.ClearBit(1)
	bf.ClearBit(2)
	assert.Equal(t, []byte{0b00000000, 0b00100000}, bf.GetData())
	bf.ClearBit(10)
	bf.ClearBit(16)
	bf.ClearBit(17)
	bf.ClearBit(18)
	assert.Equal(t, []byte{0b00000000, 0b00000000}, bf.GetData())
}

func TestDelta(t *testing.T) {
	bf1 := bitfield.NewBitfieldFromBytes([]byte{0b10000000, 0b00100000})
	bf2 := bitfield.NewBitfieldFromBytes([]byte{0b00000100, 0b00100000})
	bf3 := bf1.Delta(bf2)
	assert.Equal(t, []byte{0b10000000, 0b00000000}, bf3.GetData())
}
