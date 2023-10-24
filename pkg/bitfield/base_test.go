package bitfield_test

import (
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

func TestGetSetBits(t *testing.T) {
	bf := bitfield.NewBitfieldFromBytes([]byte{0b10000000, 0b00100000})

	setBits := make([]uint64, 0, 2)
	setBits = bf.GetSetBits(0, setBits)
	assert.Equal(t, []uint64{0, 10}, setBits)

	setBits = make([]uint64, 0, 1)
	setBits = bf.GetSetBits(0, setBits)
	assert.Equal(t, []uint64{0}, setBits)

	setBits = make([]uint64, 0, 1)
	setBits = bf.GetSetBits(1, setBits)
	assert.Equal(t, []uint64{10}, setBits)
}
