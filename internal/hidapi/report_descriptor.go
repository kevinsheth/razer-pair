//go:build cgo && (darwin || linux || windows || freebsd)

package hidapi

const (
	itemTypeMain   = 0
	itemTypeGlobal = 1

	tagFeature     = 11
	tagReportSize  = 7
	tagReportID    = 8
	tagReportCount = 9
	tagPush        = 10
	tagPop         = 11
)

type reportGlobals struct {
	size  uint32
	count uint32
	id    uint32
}

func maxFeatureReportSize(descriptor []byte) int {
	globals := reportGlobals{}
	var stack []reportGlobals
	bitsByID := make(map[uint32]uint64)

	for offset := 0; offset < len(descriptor); {
		itemType, tag, value, next, ok := readItem(descriptor, offset)
		if !ok {
			break
		}
		offset = next

		switch {
		case itemType == itemTypeGlobal && tag == tagReportSize:
			globals.size = value
		case itemType == itemTypeGlobal && tag == tagReportID:
			globals.id = value
		case itemType == itemTypeGlobal && tag == tagReportCount:
			globals.count = value
		case itemType == itemTypeGlobal && tag == tagPush:
			stack = append(stack, globals)
		case itemType == itemTypeGlobal && tag == tagPop && len(stack) > 0:
			globals = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
		case itemType == itemTypeMain && tag == tagFeature:
			bitsByID[globals.id] += uint64(globals.size) * uint64(globals.count)
		}
	}

	maximum := 0
	for id, bits := range bitsByID {
		size := int((bits + 7) / 8)
		if id != 0 {
			size++
		}
		maximum = max(maximum, size)
	}
	return maximum
}

func readItem(descriptor []byte, offset int) (itemType, tag byte, value uint32, next int, ok bool) {
	prefix := descriptor[offset]
	if prefix == 0xfe { // Long item: length, tag, data.
		if offset+3 > len(descriptor) {
			return 0, 0, 0, 0, false
		}
		next = offset + 3 + int(descriptor[offset+1])
		return 0, 0, 0, next, next <= len(descriptor)
	}

	size := int(prefix & 0x03)
	if size == 3 {
		size = 4
	}
	next = offset + 1 + size
	if next > len(descriptor) {
		return 0, 0, 0, 0, false
	}
	for i, b := range descriptor[offset+1 : next] {
		value |= uint32(b) << (8 * i)
	}
	return (prefix >> 2) & 0x03, prefix >> 4, value, next, true
}
