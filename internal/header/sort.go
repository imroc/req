package header

import (
	"net/textproto"
	"sort"
)

type KeyValues struct {
	Key    string
	Values []string
}

type sorter struct {
	order       map[string]int
	defaultRank int
	kvs         []KeyValues
}

func (s *sorter) Len() int      { return len(s.kvs) }
func (s *sorter) Swap(i, j int) { s.kvs[i], s.kvs[j] = s.kvs[j], s.kvs[i] }
func (s *sorter) Less(i, j int) bool {
	iRank, jRank := s.defaultRank, s.defaultRank
	if index, ok := s.order[textproto.CanonicalMIMEHeaderKey(s.kvs[i].Key)]; ok {
		iRank = index
	}
	if index, ok := s.order[textproto.CanonicalMIMEHeaderKey(s.kvs[j].Key)]; ok {
		jRank = index
	}
	return iRank < jRank
}

func SortKeyValues(kvs []KeyValues, orderedKeys []string) {
	order := make(map[string]int)
	for i, key := range orderedKeys {
		order[textproto.CanonicalMIMEHeaderKey(key)] = i
	}
	s := &sorter{
		order:       order,
		defaultRank: len(orderedKeys),
		kvs:         kvs,
	}
	sort.Stable(s)
}
