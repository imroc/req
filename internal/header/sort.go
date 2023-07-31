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
	order map[string]int
	kvs   []KeyValues
}

func (s *sorter) Len() int      { return len(s.kvs) }
func (s *sorter) Swap(i, j int) { s.kvs[i], s.kvs[j] = s.kvs[j], s.kvs[i] }
func (s *sorter) Less(i, j int) bool {
	if index, ok := s.order[textproto.CanonicalMIMEHeaderKey(s.kvs[i].Key)]; ok {
		i = index
	}
	if index, ok := s.order[textproto.CanonicalMIMEHeaderKey(s.kvs[j].Key)]; ok {
		j = index
	}
	return i < j
}

func SortKeyValues(kvs []KeyValues, orderedKeys []string) {
	order := make(map[string]int)
	for i, key := range orderedKeys {
		order[textproto.CanonicalMIMEHeaderKey(key)] = i
	}
	s := &sorter{
		order: order,
		kvs:   kvs,
	}
	sort.Sort(s)
}
