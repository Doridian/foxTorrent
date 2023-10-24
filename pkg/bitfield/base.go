package bitfield

type Bitfield struct {
	data []byte
}

func NewBitfield(size uint64) *Bitfield {
	return &Bitfield{
		data: make([]byte, (size+7)/8),
	}
}

func NewBitfieldFromBytes(data []byte) *Bitfield {
	return &Bitfield{
		data: data,
	}
}

func (b *Bitfield) SetBit(index uint64) {
	b.data[index/8] |= 1 << (7 - index%8)
}

func (b *Bitfield) ClearBit(index uint64) {
	b.data[index/8] &= ^(1 << (7 - index%8))
}

func (b *Bitfield) GetBit(index uint64) bool {
	return b.data[index/8]&(1<<(7-index%8)) != 0
}

func (b *Bitfield) IsEmpty() bool {
	for _, v := range b.data {
		if v != 0 {
			return false
		}
	}
	return true
}

func (b *Bitfield) Delta(other *Bitfield) *Bitfield {
	newBitfield := &Bitfield{
		data: make([]byte, len(b.data)),
	}
	for i := range b.data {
		newBitfield.data[i] = b.data[i] & (^other.data[i])
	}
	return newBitfield
}

func (b *Bitfield) Nand(other *Bitfield) *Bitfield {
	newBitfield := NewBitfield(uint64(len(b.data)) * 8)
	for i := range b.data {
		newBitfield.data[i] = b.data[i] &^ other.data[i]
	}
	return newBitfield
}

func (b *Bitfield) GetSetBits(start uint64, setBits []uint64) []uint64 {
	setBitsEnd := uint64(len(b.data))

	setBitsCap := uint64(cap(setBits))
	setBitsNeeded := setBitsEnd - start
	if setBitsCap < setBitsNeeded {
		setBitsEnd = start + setBitsCap
	}

	for i := start; i < setBitsEnd; i++ {
		for j := uint64(0); j < 8; j++ {
			if b.data[i]&(1<<uint(7-j)) != 0 {
				setBits = append(setBits, i*8+j)
			}
		}
	}
	return setBits
}

func (b *Bitfield) GetData() []byte {
	return b.data
}
