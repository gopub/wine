package path

import (
	"sort"
	"strings"
)

type sortNodeList []*Node

func (l sortNodeList) Len() int {
	return len(l)
}

func (l sortNodeList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l sortNodeList) Less(i, j int) bool {
	return strings.Compare(l[i].Path(), l[j].Path()) < 0
}

func SortByPath(l []*Node) {
	sort.Sort((sortNodeList)(l))
}
