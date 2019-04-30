package toml

import (
	"fmt"
	"sort"
	"unicode"
	"unicode/utf8"
)

var (
	// Based on https://github.com/golang/lint/blob/32a87160691b3c96046c0c678fe57c5bef761456/lint.go#L702
	commonInitialismMap = map[string]struct{}{
		"API":   {},
		"ASCII": {},
		"CPU":   {},
		"CSRF":  {},
		"CSS":   {},
		"DNS":   {},
		"EOF":   {},
		"GUID":  {},
		"HTML":  {},
		"HTTP":  {},
		"HTTPS": {},
		"ID":    {},
		"IP":    {},
		"JSON":  {},
		"LHS":   {},
		"QPS":   {},
		"RAM":   {},
		"RHS":   {},
		"RPC":   {},
		"SLA":   {},
		"SMTP":  {},
		"SQL":   {},
		"SSH":   {},
		"TCP":   {},
		"TLS":   {},
		"TTL":   {},
		"UDP":   {},
		"UI":    {},
		"UID":   {},
		"UUID":  {},
		"URI":   {},
		"URL":   {},
		"UTF8":  {},
		"VM":    {},
		"XML":   {},
		"XSRF":  {},
		"XSS":   {},
	}
	commonInitialisms = keys(commonInitialismMap)
	commonInitialism  = mustDoubleArray(newDoubleArray(commonInitialisms))
	longestLen        = longestLength(commonInitialisms)
	shortestLen       = shortestLength(commonInitialisms, longestLen)
)

// ToSnakeCase returns a copy of the string s with all Unicode letters mapped to their snake case.
// It will insert letter of '_' at position of previous letter of uppercase and all
// letters convert to lower case.
// ToSnakeCase does not insert '_' letter into a common initialism word like ID, URL and so on.
func ToSnakeCase(s string) string {
	if s == "" {
		return ""
	}
	result := make([]byte, 0, len(s))
	var runeBuf [utf8.UTFMax]byte
	var j, skipCount int
	for i, c := range s {
		if i < skipCount {
			continue
		}
		if unicode.IsUpper(c) {
			if i != 0 {
				result = append(result, '_')
			}
			next := nextIndex(j, len(s))
			if length := commonInitialism.Lookup(s[j:next]); length > 0 {
				for _, r := range s[j : j+length] {
					if r < utf8.RuneSelf {
						result = append(result, toLowerASCII(byte(r)))
					} else {
						n := utf8.EncodeRune(runeBuf[:], unicode.ToLower(r))
						result = append(result, runeBuf[:n]...)
					}
				}
				j += length - 1
				skipCount = i + length
				continue
			}
		}
		if c < utf8.RuneSelf {
			result = append(result, toLowerASCII(byte(c)))
		} else {
			n := utf8.EncodeRune(runeBuf[:], unicode.ToLower(c))
			result = append(result, runeBuf[:n]...)
		}
		j++
	}
	return string(result)
}

func isUpperASCII(c byte) bool {
	return 'A' <= c && c <= 'Z'
}

func isLowerASCII(c byte) bool {
	return 'a' <= c && c <= 'z'
}

func toUpperASCII(c byte) byte {
	if isLowerASCII(c) {
		return c - ('a' - 'A')
	}
	return c
}

func toLowerASCII(c byte) byte {
	if isUpperASCII(c) {
		return c + 'a' - 'A'
	}
	return c
}

func nextIndex(i, maxlen int) int {
	if n := i + longestLen; n < maxlen {
		return n
	}
	return maxlen
}

func keys(m map[string]struct{}) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

func shortestLength(strs []string, shortest int) int {
	for _, s := range strs {
		if candidate := utf8.RuneCountInString(s); candidate < shortest {
			shortest = candidate
		}
	}
	return shortest
}

func longestLength(strs []string) (longest int) {
	for _, s := range strs {
		if candidate := utf8.RuneCountInString(s); candidate > longest {
			longest = candidate
		}
	}
	return longest
}

const (
	terminationCharacter = '#'
)

func mustDoubleArray(da *doubleArray, err error) *doubleArray {
	if err != nil {
		panic(err)
	}
	return da
}

func (da *doubleArray) Build(keys []string) error {
	records := makeRecords(keys)
	if err := da.build(records, 1, 0, make(map[int]struct{})); err != nil {
		return err
	}
	return nil
}

type doubleArray struct {
	bc   []baseCheck
	node []int
}

func newDoubleArray(keys []string) (*doubleArray, error) {
	da := &doubleArray{
		bc:   []baseCheck{0},
		node: []int{-1}, // A start index is adjusting to 1 because 0 will be used as a mark of non-existent node.
	}
	if err := da.Build(keys); err != nil {
		return nil, err
	}
	return da, nil
}

// baseCheck contains BASE, CHECK and Extra flags.
// From the top, 22bits of BASE, 2bits of Extra flags and 8bits of CHECK.
//
//  BASE (22bit) | Extra flags (2bit) | CHECK (8bit)
// |----------------------|--|--------|
// 32                    10  8         0
type baseCheck uint32

func (bc baseCheck) Base() int {
	return int(bc >> 10)
}

func (bc *baseCheck) SetBase(base int) {
	*bc |= baseCheck(base) << 10
}

func (bc baseCheck) Check() byte {
	return byte(bc)
}

func (bc *baseCheck) SetCheck(check byte) {
	*bc |= baseCheck(check)
}

func (bc baseCheck) IsEmpty() bool {
	return bc&0xfffffcff == 0
}

func (da *doubleArray) Lookup(path string) (length int) {
	idx := 1
	tmpIdx := idx
	for i := 0; i < len(path); i++ {
		c := path[i]
		tmpIdx = da.nextIndex(da.bc[tmpIdx].Base(), c)
		if tmpIdx >= len(da.bc) || da.bc[tmpIdx].Check() != c {
			break
		}
		idx = tmpIdx
	}
	if next := da.nextIndex(da.bc[idx].Base(), terminationCharacter); next < len(da.bc) && da.bc[next].Check() == terminationCharacter {
		return da.node[da.bc[next].Base()]
	}
	return -1
}

func (da *doubleArray) LookupByBytes(path []byte) (length int) {
	idx := 1
	tmpIdx := idx
	for i := 0; i < len(path); i++ {
		c := path[i]
		tmpIdx = da.nextIndex(da.bc[tmpIdx].Base(), c)
		if tmpIdx >= len(da.bc) || da.bc[tmpIdx].Check() != c {
			break
		}
		idx = tmpIdx
	}
	if next := da.nextIndex(da.bc[idx].Base(), terminationCharacter); next < len(da.bc) && da.bc[next].Check() == terminationCharacter {
		return da.node[da.bc[next].Base()]
	}
	return -1
}

func (da *doubleArray) build(srcs []record, idx, depth int, usedBase map[int]struct{}) error {
	sort.Stable(recordSlice(srcs))
	base, siblings, leaf, err := da.arrange(srcs, idx, depth, usedBase)
	if err != nil {
		return err
	}
	if leaf != nil {
		da.bc[idx].SetBase(len(da.node))
		da.node = append(da.node, leaf.value)
	}
	for _, sib := range siblings {
		da.setCheck(da.nextIndex(base, sib.c), sib.c)
	}
	for _, sib := range siblings {
		if err := da.build(srcs[sib.start:sib.end], da.nextIndex(base, sib.c), depth+1, usedBase); err != nil {
			return err
		}
	}
	return nil
}

func (da *doubleArray) setBase(i, base int) {
	da.bc[i].SetBase(base)
}

func (da *doubleArray) setCheck(i int, check byte) {
	da.bc[i].SetCheck(check)
}

func (da *doubleArray) findEmptyIndex(start int) int {
	i := start
	for ; i < len(da.bc); i++ {
		if da.bc[i].IsEmpty() {
			break
		}
	}
	return i
}

// findBase returns good BASE.
func (da *doubleArray) findBase(siblings []sibling, start int, usedBase map[int]struct{}) (base int) {
	for idx, firstChar := start+1, siblings[0].c; ; idx = da.findEmptyIndex(idx + 1) {
		base = da.nextIndex(idx, firstChar)
		if _, used := usedBase[base]; used {
			continue
		}
		i := 0
		for ; i < len(siblings); i++ {
			next := da.nextIndex(base, siblings[i].c)
			if len(da.bc) <= next {
				da.bc = append(da.bc, make([]baseCheck, next-len(da.bc)+1)...)
			}
			if !da.bc[next].IsEmpty() {
				break
			}
		}
		if i == len(siblings) {
			break
		}
	}
	usedBase[base] = struct{}{}
	return base
}

func (da *doubleArray) arrange(records []record, idx, depth int, usedBase map[int]struct{}) (base int, siblings []sibling, leaf *record, err error) {
	siblings, leaf, err = makeSiblings(records, depth)
	if err != nil {
		return -1, nil, nil, err
	}
	if len(siblings) < 1 {
		return -1, nil, leaf, nil
	}
	base = da.findBase(siblings, idx, usedBase)
	da.setBase(idx, base)
	return base, siblings, leaf, err
}

type sibling struct {
	start int
	end   int
	c     byte
}

func (da *doubleArray) nextIndex(base int, c byte) int {
	return base ^ int(c)
}

func makeSiblings(records []record, depth int) (sib []sibling, leaf *record, err error) {
	var (
		pc byte
		n  int
	)
	for i, r := range records {
		if len(r.key) <= depth {
			leaf = &r
			continue
		}
		c := r.key[depth]
		switch {
		case pc < c:
			sib = append(sib, sibling{start: i, c: c})
		case pc == c:
			continue
		default:
			return nil, nil, fmt.Errorf("stringutil: BUG: records hasn't been sorted")
		}
		if n > 0 {
			sib[n-1].end = i
		}
		pc = c
		n++
	}
	if n == 0 {
		return nil, leaf, nil
	}
	sib[n-1].end = len(records)
	return sib, leaf, nil
}

type record struct {
	key   string
	value int
}

func makeRecords(srcs []string) (records []record) {
	termChar := string(terminationCharacter)
	for _, s := range srcs {
		records = append(records, record{
			key:   string(s + termChar),
			value: utf8.RuneCountInString(s),
		})
	}
	return records
}

type recordSlice []record

func (rs recordSlice) Len() int {
	return len(rs)
}

func (rs recordSlice) Less(i, j int) bool {
	return rs[i].key < rs[j].key
}

func (rs recordSlice) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}
