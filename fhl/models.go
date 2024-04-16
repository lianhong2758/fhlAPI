package fhl

import (
	"math/rand/v2"
	"sort"
	"strconv"
	"strings"
)

// 游戏题目与进度的接口，下面的 SubjectA/B/C/D 都会实现之。
// 包含任意形式的数据结构，但是需要支持与字符串的互转，
// 以及「尝试提交一个句子，若正确则更新状态」的操作。
type Subject interface {
	Parse(string) // 从一个字符串解析出数据
	Dump() string // 将数据表示为一个字符串
	// 尝试用一段文本作答，若与题目匹配则更新数据结构
	// 第一个参数是提交的文本内容，包含多句时用斜杠分隔
	// 第二个参数为当前答题的一方
	// 第一个返回值是关键词的下标集合，若答案不合法则为 nil
	// 第二个返回值是一个字符串，表示变化量
	Answer(string, Side) ([]IntPair, string)
	End() bool // 游戏是否结束
}

// 表示游戏中的一方，A/B,用0/1表示
type Side int

const (
	SideA    Side = 0
	SideB    Side = 1
	SideNone Side = -1
)

type IntPair struct {
	a, b int
}

// 生成普通飞花题目备选列表
// 返回 count1 个不重复的单字与 count2 个不重复的双字
func (f *FHL) GenerateA(count1, count2 int) []string {
	ret := []string{}

	// 单字
	if count1 > 0 {
		n := len(f.HotWords1) / 2
		for _, i := range randomSample(n, count1-1) {
			ret = append(ret, f.HotWords1[i])
		}
		ret = append(ret, f.HotWords1[rand.IntN(n)+n])
	}

	// 双字
	if count2 > 0 {
		n := len(f.HotWords2) / 2
		for _, i := range randomSample(n, count2-1) {
			ret = append(ret, f.HotWords2[i])
		}
		ret = append(ret, f.HotWords2[rand.IntN(n)+n])
	}

	return ret
}

// 生成多字飞花题目，返回一句长度为 length 的句子
// 需确保 ALL_HOT_LEN_MIN <= length <= ALL_HOT_LEN_MAX
func (f *FHL) GenerateB(length int) string {
	collection := f.AllHotSentences[length-ALL_HOT_LEN_MIN]
	return collection[rand.IntN(len(collection))]
}

// 生成超级飞花题目，返回长度为 sizeLeft 和 sizeRight 的字符串列表
func (f *FHL) GenerateC(sizeLeft, sizeRight int) ([]string, []string) {
	var left []string
	var right []string

	leftTopPart := sizeLeft
	leftBottomPart := 0
	if sizeLeft >= 3 {
		leftTopPart -= 1
		leftBottomPart += 1
	}
	n := len(f.HotWords1)

	// 左侧的一个字最多对应右侧的多少字
	singleWordLimit := sizeRight/sizeLeft + 2
	word2Limit := sizeRight / 5

	for len(right) < sizeRight {
		// 所有已选字的集合
		all := map[string]struct{}{}

		// 先选取 n 个高频字
		left = []string{}
		for _, i := range randomSample(n/4, leftTopPart) {
			left = append(left, f.HotWords1[i])
		}
		for _, i := range randomSample(n-n/4, leftBottomPart) {
			left = append(left, f.HotWords1[i+n/4])
		}
		for _, c := range left {
			all[c] = struct{}{}
		}

		// 寻找包含任意选出字的名篇句子，并从中找出一些高频词
		// 如果性能不足可以后续优化
		right = []string{}
		// 记录 left 中每个字已经对应了 right 中多少个字
		count := make([]int, len(left))
		// 有多少个双字
		word2Count := 0
		// 先打乱
		perm := rand.Perm(len(f.HotArticles))
		for _, id := range perm {
			article := f.HotArticles[id]
			sentenceIndex := -1
			for j, t := range left {
				if count[j] >= singleWordLimit {
					continue
				}
				for i, s := range article.Content {
					if len([]rune(s)) >= 4 && strings.Contains(s, t) {
						count[j] += 1
						sentenceIndex = i
						break
					}
				}
				if sentenceIndex != -1 {
					break
				}
			}
			if sentenceIndex == -1 {
				continue
			}
			h1, h2 := f.getHotWords(article.Content[sentenceIndex])
			// 先考虑双字词，如有且不与单字重复，则加入
			word2Valid := false
			if word2Count < word2Limit {
				for _, word := range h2 {
					runes := []rune(word)
					_, picked0 := all[string(runes[0])]
					_, picked1 := all[string(runes[1])]
					_, picked := all[word]
					if !picked0 && !picked1 && !picked {
						right = append(right, word)
						all[string(runes[0])] = struct{}{}
						all[string(runes[1])] = struct{}{}
						all[word] = struct{}{}
						word2Valid = true
						word2Count++
						break
					}
				}
			}
			if !word2Valid {
				// 随机选取一个常见字
				// 先打乱
				for i := range h1 {
					j := rand.IntN(i + 1)
					h1[i], h1[j] = h1[j], h1[i]
				}
				for _, word := range h1 {
					if _, picked := all[word]; !picked {
						right = append(right, word)
						all[word] = struct{}{}
						break
					}
				}
			}
			if len(right) == sizeRight {
				break
			}
		}
	}

	sort.Slice(right, func(i, j int) bool {
		return len(right[i]) < len(right[j])
	})

	return left, right
}

// 生成 2 个长度为 n 的关键词组作为谜之飞花题目
// 其中第二个关键词组全部由双字高频词构成,第一个关键词组由单字或双字高频词构成
func (f *FHL) GenerateD(n int) ([]string, []string) {
	var hotWordsList1 []string
	var hotWordsList2 []string

	count := 0
	length := len(f.HotArticles)
	//储存已被选择过的诗
	var tmp map[int]int
	tmp = make(map[int]int)

	for count < n {
		i := rand.IntN(length)
		//该诗已被选择过
		if _, ok := tmp[i]; ok {
			continue
		}
		tmp[i] = 1

		content := f.HotArticles[i].Content
		//随机选择某一句
		j := rand.IntN(len(content))

		sHotWords, dHotWords := f.getHotWords(content[j])
		//由诗句获得的单字高频词组、双字高频词组都为空
		if len(sHotWords) == 0 && len(dHotWords) == 0 {
			continue
		}

		freqMax := 0
		s1 := 0
		s2 := 0
		flag := 0

		k := rand.IntN(6)
		//有0.6的概率填入（若可行）两个单字词
		if k >= 4 && len(sHotWords) > 1 {
			for i := 0; i < len(sHotWords); i++ {
				for j := i + 1; j < len(sHotWords); j++ {
					if freq, ok := f.HotWordsFreq[sHotWords[i]+sHotWords[j]]; ok {
						//高频词组合频次大于50方记入; 取最大值
						if freq >= 50 && freq > freqMax {
							s1 = i
							s2 = j
							flag = 1
						}
					}
				}
			}
			if flag == 1 {
				hotWordsList1 = append(hotWordsList1, sHotWords[s1])
				hotWordsList2 = append(hotWordsList2, sHotWords[s2])
			}
		}
		//有0.4的概率填入（若可行）单字词 + 双字词
		if k < 4 && len(sHotWords) > 0 && len(dHotWords) > 0 {
			for i := 0; i < len(sHotWords); i++ {
				for j := 0; j < len(dHotWords); j++ {
					//确保单字词不包含在双字词中
					if !strings.Contains(dHotWords[j], sHotWords[i]) {
						if freq, ok := f.HotWordsFreq[sHotWords[i]+dHotWords[j]]; ok {
							//高频词组合频次大于50方记入; 取最大值
							if freq >= 50 && freq > freqMax {
								s1 = i
								s2 = j
								flag = 1
							}
						}
					}
				}
			}
			if flag == 1 {
				hotWordsList1 = append(hotWordsList1, sHotWords[s1])
				hotWordsList2 = append(hotWordsList2, dHotWords[s2])
			}
		}

		if flag == 0 {
			continue
		}

		//计数加1
		count++
		// fmt.Println(content[j])
	}

	sort.Strings(hotWordsList1)
	sort.Strings(hotWordsList2)

	return hotWordsList1, hotWordsList2
}

// 普通飞花 题目与进度
type SubjectA struct {
	Word string
}

func (s *SubjectA) Parse(str string) {
	s.Word = str
}
func (s *SubjectA) Dump() string {
	return s.Word
}
func (s *SubjectA) Answer(str string, from Side) ([]IntPair, string) {
	// 第二个返回值：""
	p := strings.Index(str, s.Word)
	if p != -1 {
		return []IntPair{{runes(str[:p]), runes(s.Word)}}, ""
	} else {
		return nil, ""
	}
}
func (s *SubjectA) End() bool {
	return false
}

// 多字飞花 题目与进度
type SubjectB struct {
	Words    []rune
	CurIndex int
}

func (s *SubjectB) Parse(str string) {
	// 例：春花秋月何时了/3
	fields := strings.SplitN(str, "/", 2)
	s.Words = []rune(fields[0])
	s.CurIndex, _ = strconv.Atoi(fields[1])
}
func (s *SubjectB) Dump() string {
	return string(s.Words) + "/" + strconv.Itoa(s.CurIndex)
}
func (s *SubjectB) Answer(str string, from Side) ([]IntPair, string) {
	// 第二个返回值：下一位轮到的玩家要飞的字的下标，若游戏结束则为 -1
	p := strings.IndexRune(str, s.Words[s.CurIndex])
	if p != -1 {
		if from == SideB {
			s.CurIndex++
			if s.CurIndex == len(s.Words) {
				s.CurIndex = -1
			}
		}
		return []IntPair{{runes(str[:p]), 1}}, strconv.Itoa(s.CurIndex)
	} else {
		return nil, ""
	}
}
func (s *SubjectB) End() bool {
	return (s.CurIndex == -1)
}

// 超级飞花 题目与进度
type SubjectC struct {
	WordsLeft  []string
	WordsRight []string
	UsedRight  []bool
}

func (s *SubjectC) Parse(str string) {
	// 例：古 梦 雁/长 舟 送 寄 事 神 不 生 西风 多少/1000010011
	fields := strings.SplitN(str, "/", 3)
	s.WordsLeft = strings.Split(fields[0], " ")
	s.WordsRight = strings.Split(fields[1], " ")
	s.UsedRight = make([]bool, len(s.WordsRight))
	for i := range s.UsedRight {
		s.UsedRight[i] = (fields[2][i] == '1')
	}
}
func (s *SubjectC) Dump() string {
	used := []rune{}
	for _, b := range s.UsedRight {
		if b {
			used = append(used, '1')
		} else {
			used = append(used, '0')
		}
	}
	return strings.Join(s.WordsLeft, " ") + "/" +
		strings.Join(s.WordsRight, " ") + "/" +
		string(used)
}
func (s *SubjectC) Answer(str string, from Side) ([]IntPair, string) {
	// 第二个返回值：右侧被匹配的关键词下标
	indexLeft, indexRight := -1, -1
	ps := make([]IntPair, 2)
	for i, w := range s.WordsLeft {
		p := strings.Index(str, w)
		if p != -1 {
			indexLeft = i
			ps[0] = IntPair{runes(str[:p]), runes(w)}
			break
		}
	}
	for i, w := range s.WordsRight {
		p := strings.Index(str, w)
		if !s.UsedRight[i] && p != -1 {
			indexRight = i
			ps[1] = IntPair{runes(str[:p]), runes(w)}
			break
		}
	}
	if indexLeft == -1 || indexRight == -1 {
		return nil, ""
	}
	s.UsedRight[indexRight] = true
	return ps, strconv.Itoa(indexRight)
}
func (s *SubjectC) End() bool {
	for _, u := range s.UsedRight {
		if !u {
			return false
		}
	}
	return true
}

// 谜之飞花 题目与进度
type SubjectD struct {
	WordsLeft  []string
	WordsRight []string
	UsedLeft   []bool
	UsedRight  []bool
}

func (s *SubjectD) Parse(str string) {
	// 例：万 书 今 凉 得 来 柳 欲/一片 丝 如此 孤 庭 细 舟 觉/00000000/00000000
	fields := strings.SplitN(str, "/", 4)
	s.WordsLeft = strings.Split(fields[0], " ")
	s.WordsRight = strings.Split(fields[1], " ")
	s.UsedLeft = make([]bool, len(s.WordsLeft))
	s.UsedRight = make([]bool, len(s.WordsRight))
	for i := range s.UsedLeft {
		s.UsedLeft[i] = (fields[2][i] == '1')
	}
	for i := range s.UsedRight {
		s.UsedRight[i] = (fields[3][i] == '1')
	}
}
func (s *SubjectD) Dump() string {
	used := []rune{}
	for _, b := range s.UsedLeft {
		if b {
			used = append(used, '1')
		} else {
			used = append(used, '0')
		}
	}
	used = append(used, '/')
	for _, b := range s.UsedRight {
		if b {
			used = append(used, '1')
		} else {
			used = append(used, '0')
		}
	}
	return strings.Join(s.WordsLeft, " ") + "/" +
		strings.Join(s.WordsRight, " ") + "/" +
		string(used)
}
func (s *SubjectD) Answer(str string, from Side) ([]IntPair, string) {
	// 第二个返回值："a,b"，左右侧被匹配的关键词下标
	indexLeft, indexRight := -1, -1
	ps := make([]IntPair, 2)
	for i, w := range s.WordsLeft {
		p := strings.Index(str, w)
		if !s.UsedLeft[i] && p != -1 {
			indexLeft = i
			ps[0] = IntPair{runes(str[:p]), runes(w)}
			break
		}
	}
	for i, w := range s.WordsRight {
		p := strings.Index(str, w)
		if !s.UsedRight[i] && p != -1 {
			indexRight = i
			ps[1] = IntPair{runes(str[:p]), runes(w)}
			break
		}
	}
	if indexLeft == -1 || indexRight == -1 {
		return nil, ""
	}
	s.UsedLeft[indexLeft] = true
	s.UsedRight[indexRight] = true
	return ps, strconv.Itoa(indexLeft) + "," + strconv.Itoa(indexRight)
}
func (s *SubjectD) End() bool {
	for _, u := range s.UsedLeft {
		if !u {
			return false
		}
	}
	for _, u := range s.UsedRight {
		if !u {
			return false
		}
	}
	return true
}

// 计算一个字符串中的 Unicode 字符数
func runes(s string) int {
	return len([]rune(s))
}
