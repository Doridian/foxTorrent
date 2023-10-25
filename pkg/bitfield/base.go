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
	if index >= uint64(len(b.data)*8) {
		return
	}
	b.data[index/8] |= 1 << (7 - index%8)
}

func (b *Bitfield) ClearBit(index uint64) {
	if index >= uint64(len(b.data)*8) {
		return
	}
	b.data[index/8] &= ^(1 << (7 - index%8))
}

func (b *Bitfield) GetBit(index uint64) bool {
	if index >= uint64(len(b.data)*8) {
		return false
	}
	return b.data[index/8]&(1<<(7-index%8)) != 0
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

func (b *Bitfield) ForEachMatchingBit(set bool, f func(index uint64) error) error {
	matches := false
	for i := uint64(0); i < uint64(len(b.data)); i++ {
		for j := uint64(0); j < 8; j++ {
			if set {
				matches = b.data[i]&(1<<(7-j)) != 0
			} else {
				matches = b.data[i]&(1<<(7-j)) == 0
			}
			if matches {
				err := f(i*8 + j)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (b *Bitfield) GetData() []byte {
	return b.data
}
