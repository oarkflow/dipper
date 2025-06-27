package dipper

// attributeSplitter offers methods to iterate the substrings of a string
// using a given separator.
type attributeSplitter struct {
	s         string
	sep       string
	index     int
	hasMore   bool
	scanIndex int
}

// newAttributeSplitter returns a new attributeSplitter instance.
func newAttributeSplitter(s, sep string) *attributeSplitter {
	return &attributeSplitter{sep: sep, index: -1, s: s, hasMore: true}
}

// HasMore returns true if the iterated string has more fields.
func (s *attributeSplitter) HasMore() bool {
	return s.hasMore
}

// Next returns the next field of the iterated string and the position of the
// field in the string (or an empty string and -1 if the string does not have
// more fields).
func (s *attributeSplitter) Next() (string, int) {
	if !s.hasMore {
		return "", -1
	}
	remain := s.s[s.scanIndex:]
	sepLen := len(s.sep)
	enclosure := 0
	index := -1

	for i := 0; i <= len(remain)-sepLen; i++ {
		switch remain[i] {
		case '[':
			enclosure++
		case ']':
			if enclosure > 0 {
				enclosure--
			}
		}
		if enclosure == 0 && sepLen > 0 && remain[i:i+sepLen] == s.sep {
			index = i
			break
		}
	}
	if index == -1 {
		s.hasMore = false
		res := remain
		s.index++
		s.scanIndex = len(s.s)
		return res, s.index
	}
	res := remain[:index]
	s.index++
	s.scanIndex += index + sepLen
	return res, s.index
}

// CountRemaining returns the number of remaining fields in the string.
func (s *attributeSplitter) CountRemaining() int {
	remain := s.s[s.scanIndex:]
	sepLen := len(s.sep)
	enclosure := 0
	count := 0
	for i := 0; i <= len(remain)-sepLen; i++ {
		switch remain[i] {
		case '[':
			enclosure++
		case ']':
			if enclosure > 0 {
				enclosure--
			}
		}
		if enclosure == 0 && sepLen > 0 && remain[i:i+sepLen] == s.sep {
			count++
		}
	}
	return count + 1
}
