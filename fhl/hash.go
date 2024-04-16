package fhl

import (
	"sort"
)

type HashType uint64
type ArticleIdxType uint32
type ContentIdxType uint32
type ErrCorrRecord struct {
	Hash       HashType
	ArticleIdx ArticleIdxType
	ContentIdx ContentIdxType
}

const HASH_W = 5
const ART_IDX_W = 3
const CON_IDX_W = 2
const RECORD_W = HASH_W + ART_IDX_W + CON_IDX_W

func hash(rs []rune) HashType {
	h := HashType(0)
	for _, r := range rs {
		if r > 0 {
			h = h*100003 + HashType(r)
		}
	}
	return h % (1 << (HASH_W * 8))
}

func (f *FHL) InitErrCorr() *FHL {
	x := []ErrCorrRecord{}
	for i, article := range f.Articles {
		for j, s := range article.Content {
			forEachPossibleErrHash(s, func(h HashType) bool {
				x = append(x, ErrCorrRecord{
					Hash:       h,
					ArticleIdx: ArticleIdxType(i),
					ContentIdx: ContentIdxType(j),
				})
				return false
			})
		}
	}
	println("err len: ", len(x))
	sort.Slice(x, func(i, j int) bool {
		return x[i].Hash < x[j].Hash
	})

	if err := SavePrecalErrCorr(x); err != nil {
		f.Error = err
	}
	return f
}

func forEachPossibleErrHash(s string, fn func(h HashType) bool) {
	rs := []rune(s)
	if fn(hash(rs)) {
		return
	}
	for i, r := range rs {
		rs[i] = -1
		if fn(hash(rs)) {
			return
		}
		for j, r := range rs[:i] {
			rs[j] = -1
			if fn(hash(rs)) {
				return
			}
			rs[j] = r
		}
		rs[i] = r
	}
}

// 在纠错数据库中查找某个 hash 值
// 返回 >= 此 hash 的最小记录位置，即 lower_bound
func (f *FHL) lookupErrCorr(x HashType) int64 {
	lo := int64(-1)
	hi := f.ErrCorrNums
	for lo < hi-1 {
		mid := (lo + hi) / 2
		rec := f.readErrCorrRecord(mid)
		if rec.Hash < x {
			lo = mid
		} else {
			hi = mid
		}
	}
	return hi
}
func (f *FHL) readErrCorrRecord(index int64) ErrCorrRecord {
	buf := [RECORD_W]byte{}
	f.ErrCorrFile.ReadAt(buf[:], 0+index*RECORD_W)
	rec := ErrCorrRecord{0, 0, 0}
	for i, b := range buf[0:HASH_W] {
		rec.Hash += (HashType(b) << (i * 8))
	}
	for i, b := range buf[HASH_W : HASH_W+ART_IDX_W] {
		rec.ArticleIdx += (ArticleIdxType(b) << (i * 8))
	}
	for i, b := range buf[HASH_W+ART_IDX_W:] {
		rec.ContentIdx += (ContentIdxType(b) << (i * 8))
	}
	return rec
}
