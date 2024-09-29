package fhl

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
)

// 一篇诗词
type Article struct {
	Id      int      // 编号
	Title   string   // 标题
	Dynasty string   // 朝代
	Author  string   // 作者
	Content []string // 内容，由标点拆分出的句子列表
}

type FHL struct {
	//error
	Error error `json:"-"`

	DatasetFile *os.File `json:"-"`
	// 所有诗词的列表
	// 每一项的 Id 等于此列表中的下标
	// 后续建立了一个 LRU 缓存，某篇目不在缓存中时，对应项为 nil
	Articles []*Article `json:"-"`
	// 每个篇目在文件中的偏移值
	Offset []int64
	// 名篇的列表
	HotArticles []*Article
	// 所有高频词组成的列表，单字和双字分开，各自按频率降序排序
	HotWords1 []string //1
	HotWords2 []string //2
	// 高频词的出现频数，单字和双字分开
	HotWords1Count map[rune]int

	HotWords2Count    map[RunePair]int `json:"-"`
	HotWords2CountKey []RunePair
	HotWords2CountVal []int

	// 全部都是高频字的句子
	AllHotSentences [][]string

	// 高频词组合（单字＆单字／单字＆双字／双字＆双字）在诗句中出现的频次
	// 用于优化谜之飞花令的题目选择
	HotWordsFreq map[string]int

	//纠错数据
	//	ECR []ErrCorrRecord `json:"-"`
	//纠错文件
	ErrCorrFile *os.File `json:"-"`

	ErrCorrNums int64 `json:"-"`
}
type RunePair struct{ A, B rune }

const ALL_HOT_LEN_MIN = 5
const ALL_HOT_LEN_MAX = 9

func NewFHL() *FHL {
	return new(FHL)
}

// 如果需要,初始化成员
func (f *FHL) Init() *FHL {
	f.Articles = []*Article{}
	f.Offset = []int64{}
	f.HotArticles = []*Article{}
	f.HotWords1 = []string{}              //1
	f.HotWords2 = []string{}              //2
	f.HotWords1Count = map[rune]int{}     //1
	f.HotWords2Count = map[RunePair]int{} //2
	return f
}

// 计算,需要前置init
func (f *FHL) Calculate() *FHL {
	i := 0 //计数
	sc := bufio.NewScanner(f.DatasetFile)
	offs := int64(0) //位移
	p := 0
	q := 0
	t := 0
	for sc.Scan() {
		prevOffs := offs
		offs += int64(len(sc.Text())) + 1
		i++

		// 将篇目加入列表
		article, flag := parseArticle(len(f.Articles), sc.Text())
		f.Offset = append(f.Offset, prevOffs)
		f.Articles = append(f.Articles, article)
		weight := 1
		if flag == "!" {
			// 名篇
			f.HotArticles = append(f.HotArticles, article)
			weight = 10
		}

		for _, s := range article.Content {
			n := len([]rune(s))
			p += 1
			q += n
			t += n*(n+1)/2 + 1
		}

		// 若不是重复篇目，则计入高频词
		if flag != " " {
			for _, s := range article.Content {
				s := []rune(s)
				for i, c := range s {
					f.HotWords1Count[c] += weight
					if i < len(s)-1 {
						f.HotWords2Count[RunePair{c, s[i+1]}] += weight
					}
				}
			}
		}
	}

	if err := sc.Err(); err != nil {
		panic(err)
	}

	fmt.Printf("dataset: %d articles\n", len(f.Articles))
	fmt.Printf("诗句总行数: %d, 总诗句字数: %d, 组合数%d\n", p, q, t)

	hotWords1List := byValueDesc{}
	hotWords2List := byValueDesc{}
	for k, v := range f.HotWords1Count {
		hotWords1List = append(hotWords1List, KVPair{string(k), v})
	}
	for k, v := range f.HotWords2Count {
		hotWords2List = append(hotWords2List, KVPair{
			string(k.A) + string(k.B),
			v,
		})
	}
	sort.Sort(hotWords1List)
	sort.Sort(hotWords2List)

	fmt.Println("高频单字，每行五十个")
	for i := 0; i < 400; i++ {
		fmt.Print(hotWords1List[i].string)
		if (i+1)%50 == 0 {
			fmt.Println()
		}
		f.HotWords1 = append(f.HotWords1, hotWords1List[i].string)
	}
	fmt.Println("高频双字，每行二十个")
	for i := 0; i < 200; i++ {
		fmt.Print(hotWords2List[i].string, " ")
		if (i+1)%20 == 0 {
			fmt.Println()
		}
		f.HotWords2 = append(f.HotWords2, hotWords2List[i].string)
	}

	// 删去非高频词
	for i := len(f.HotWords1); i < len(hotWords1List); i++ {
		runes := []rune(hotWords1List[i].string)
		delete(f.HotWords1Count, runes[0])
	}
	for i := len(f.HotWords2); i < len(hotWords2List); i++ {
		runes := []rune(hotWords2List[i].string)
		delete(f.HotWords2Count, RunePair{runes[0], runes[1]})
	}

	// 找出仅由不重复的高频字组成的句子
	f.AllHotSentences = make([][]string, ALL_HOT_LEN_MAX-ALL_HOT_LEN_MIN+1)
	allHotSentencesSet := make(map[string]struct{})
	for i := range f.AllHotSentences {
		f.AllHotSentences[i] = []string{}
	}
	for _, article := range f.Articles {
		for _, s := range article.Content {
			runes := []rune(s)
			if len(runes) < ALL_HOT_LEN_MIN || len(runes) > ALL_HOT_LEN_MAX {
				continue
			}
			// 确认每个字是否是高频字
			minCount := hotWords1List[len(f.HotWords1)/2-1].int
			if len(runes) >= ALL_HOT_LEN_MAX-2 {
				minCount = hotWords1List[len(f.HotWords1)-1].int
			}
			allHot := true
			for _, r := range runes {
				if n, has := f.HotWords1Count[r]; !has || n < minCount {
					allHot = false
					break
				}
			}
			if allHot {
				i := len(runes) - ALL_HOT_LEN_MIN
				// 确认是否有重复字
				// 因为 runes 至多只有九个元素，所以用朴素算法
				for i, r := range runes {
					for j := 0; j < i; j++ {
						if runes[j] == r {
							goto out
						}
					}
				}
				// 去重
				if _, existing := allHotSentencesSet[s]; existing {
					goto out
				}
				// 通过检查，加入列表
				f.AllHotSentences[i] = append(f.AllHotSentences[i], s)
				allHotSentencesSet[s] = struct{}{}
			out:
			}
		}
	}
	fmt.Println("仅由不重复的高频字组成的句子")
	for i, c := range f.AllHotSentences {
		fmt.Printf("%d 字：%d 句\n", i+ALL_HOT_LEN_MIN, len(c))
	}
	return f
}

func (f *FHL) InitPrecal() *FHL {
	fmt.Println("InitPrecal")
	f.HotWordsFreq = make(map[string]int)

	// 初始化高频词组合频次表，令其包括所有单字&单字、单字&双字、双字&双字的组合
	for i := 0; i < len(f.HotWords1); i++ {
		for j := i + 1; j < len(f.HotWords1); j++ {
			f.HotWordsFreq[f.HotWords1[i]+f.HotWords1[j]] = 0
		}
		for j := 0; j < len(f.HotWords2); j++ {
			//若单字词为双字词的一部分,忽略
			if !strings.Contains(f.HotWords2[j], f.HotWords1[i]) {
				f.HotWordsFreq[f.HotWords1[i]+f.HotWords2[j]] = 0
			}
		}
	}

	for _, article := range f.Articles {
		content := article.Content
		for i := 0; i < len(content)-1; i++ {
			//拼接两句为一“联”
			sentence := content[i] + content[i+1]
			sHotWords, dHotWords := f.getHotWords(sentence)

			//根据每一“联”诗词中的高频词，更新高频词组合频次表
			for k := 0; k < len(sHotWords); k++ {
				for j := k + 1; j < len(sHotWords); j++ {
					f.HotWordsFreq[sHotWords[k]+sHotWords[j]]++
				}
				for j := 0; j < len(dHotWords); j++ {
					if !strings.Contains(dHotWords[j], sHotWords[k]) {
						f.HotWordsFreq[sHotWords[k]+dHotWords[j]]++
					}
				}
			}
		}
	}
	for k, v := range f.HotWordsFreq {
		if v < 50 {
			delete(f.HotWordsFreq, k)
		}
	}
	return f
}

// 返回一句诗词中的所有高频词，按出现次数降序排序；若无，返回空列表
// 单字高频词与双字高频词分别返回，第一个返回值为单字
func (f *FHL) getHotWords(sentence string) ([]string, []string) {
	count1 := []KVPair{}
	count2 := []KVPair{}

	// 单字词
	s := []rune(sentence)
	for i, c := range s {
		if n, has := f.HotWords1Count[c]; has {
			count1 = append(count1, KVPair{string(c), n})
		}
		if i < len(s)-1 {
			if n, has := f.HotWords2Count[RunePair{c, s[i+1]}]; has {
				count2 = append(count2, KVPair{
					string(c) + string(s[i+1]),
					n,
				})
			}
		}
	}

	sort.Sort(byValueDesc(count1))
	sort.Sort(byValueDesc(count2))

	singleHotWords := []string{}
	doubleHotWords := []string{}
	for _, p := range count1 {
		singleHotWords = append(singleHotWords, p.string)
	}
	for _, p := range count2 {
		doubleHotWords = append(doubleHotWords, p.string)
	}
	return singleHotWords, doubleHotWords
}

func (f *FHL) DeleteCache() *FHL {
	f.Articles=nil
	runtime.GC()
	return f
}
