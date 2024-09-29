package fhl

import (
	"math/rand/v2"
	"strconv"
	"strings"
)

// 解析为诗
func parseArticle(id int, s string) (*Article, string) {
	fields := strings.SplitN(s, "\t", 5)
	if len(fields) < 5 {
		panic("Incorrect dataset format,id:"+strconv.Itoa(id)+"s: "+s)
	}
	return &Article{
		Id:      id,
		Title:   fields[3],
		Dynasty: fields[1],
		Author:  fields[2],
		Content: strings.Split(fields[4], "/"),
	}, fields[0]
}

// 排序用比较器
type KVPair struct {
	string
	int
}
type byValueDesc []KVPair

func (s byValueDesc) Len() int {
	return len(s)
}
func (s byValueDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byValueDesc) Less(i, j int) bool {
	return s[i].int > s[j].int
}

// func forEachPossibleErrHash(s string, fn func(h HashType) bool) {
// 	rs := []rune(s)
// 	if fn(hash(rs)) {
// 		return
// 	}
// 	for i, r := range rs {
// 		rs[i] = -1
// 		if fn(hash(rs)) {
// 			return
// 		}
// 		for j, r := range rs[:i] {
// 			rs[j] = -1
// 			if fn(hash(rs)) {
// 				return
// 			}
// 			rs[j] = r
// 		}
// 		rs[i] = r
// 	}
// }

func randomSample(n, count int) []int {
	picked := map[int]struct{}{}
	ret := []int{}
	for i := n - count; i < n; i++ {
		x := rand.IntN(i + 1)
		if _, dup := picked[x]; dup {
			x = i
		}
		picked[x] = struct{}{}
		ret = append(ret, x)
	}
	return ret
}

type CorrectAnswer struct {
	Text     string
	Keywords []IntPair
}

func CorrectAnswerToStringArr(ca []CorrectAnswer) []string {
	history := []string{}
	for _, a := range ca {
		history = append(history, a.Dump())
	}
	return history
}

func (a CorrectAnswer) Dump() string {
	b := strings.Builder{}
	b.WriteString(a.Text)
	for _, k := range a.Keywords {
		b.WriteRune('/')
		for i := 0; i < k.b; i++ {
			b.WriteString(strconv.FormatInt(int64(k.a+i), 36))
		}
	}
	return b.String()
}

func (f *FHL) CountToArry() {
	f.HotWords2CountKey = []RunePair{}
	f.HotWords2CountVal = []int{}
	for k, v := range f.HotWords2Count {
		f.HotWords2CountKey = append(f.HotWords2CountKey, k)
		f.HotWords2CountVal = append(f.HotWords2CountVal, v)
	}
}

func (f *FHL) ArryToCount(){
 f.HotWords2Count=map[RunePair]int{}
	for k := range f. HotWords2CountKey {
		 f.HotWords2Count[f.HotWords2CountKey[k]]=f.HotWords2CountVal[k]
	}
	clear(f.HotWords2CountKey)
	clear(f.HotWords2CountVal)
}