package metainfo

func (m *Metainfo) TotalLength() uint64 {
	var totalLen uint64 = 0
	for _, file := range m.Info.Files {
		totalLen += file.Length
	}
	return totalLen
}
