package storage

import (
	"sort"
	"strings"
	"sync"

	"github.com/cespare/xxhash"
)

const (
	labelSep = ":/:"
)

var labelBufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 1024)
	},
}

type Label struct {
	Name  string
	Value string
}

type LabelSet []Label

func (ls LabelSet) filter() LabelSet {
	mark := make(map[string]struct{})
	var size int
	for _, v := range ls {
		_, ok := mark[v.Name]
		if v.Name != "" && v.Value != "" && !ok {
			ls[size] = v
			size++
		}
		mark[v.Name] = struct{}{}
	}

	return ls[:size]
}

func (ls LabelSet) Len() int           { return len(ls) }
func (ls LabelSet) Less(i, j int) bool { return ls[i].Name < ls[j].Name }
func (ls LabelSet) Swap(i, j int)      { ls[i], ls[j] = ls[j], ls[i] }

func (ls LabelSet) AddMetricName(metric string) LabelSet {
	labels := ls.filter()
	labels = append(labels, Label{
		Name:  metricName,
		Value: metric,
	})
	return labels
}

func (ls LabelSet) Hash() uint64 {
	sort.Sort(ls)
	b := labelBufPool.Get().([]byte)

	const sep = '\xff'
	for _, v := range ls {
		b = append(b, v.Name...)
		b = append(b, sep)
		b = append(b, v.Value...)
		b = append(b, sep)
	}
	h := xxhash.Sum64(b)

	b = b[:0]
	labelBufPool.Put(b) // reuse bytes buffer

	return h
}

func (ls LabelSet) Bytes() []byte {
	bs := make([]byte, 0)
	for _, label := range ls {
		bs = append(bs, label.Name+"="+label.Value+labelSep...)
	}

	return bs
}

func labelBytesTo(bs []byte) []Label {
	ret := make([]Label, 0)
	for _, label := range strings.Split(string(bs), labelSep) {
		pair := strings.SplitN(label, "=", 2)
		if len(pair) < 2 {
			continue
		}

		ret = append(ret, Label{Name: pair[0], Value: pair[1]})
	}

	return ret
}