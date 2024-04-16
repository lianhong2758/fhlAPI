package fhl

import "github.com/agnivade/levenshtein"

// 检查句子是否在诗词库中
// 返回：(是否完全一致, 篇目编号, 句子下标)
// 找不到时，返回的篇目编号与句子下标均为 -1
func (f *FHL)LookupText(text []string) (bool, int, int) {
	// 找到第一个至少含四字的子句；若无则选择第一个
	pivot := 0
	for i, s := range text {
		if len([]rune(s)) >= 4 {
			pivot = i
			break
		}
	}

	// 是否有接近
	bestDist := 3 // 最大允许的距离 + 1
	bestArticle := -1
	bestContent := -1

	forEachPossibleErrHash(text[pivot], func(h HashType) bool {
		index := f.lookupErrCorr(h)
		for index <  f.ErrCorrNums{
			rec := f.readErrCorrRecord(index)
			if rec.Hash != h {
				break
			}

			article := f.GetArticle(int(rec.ArticleIdx))
			i := int(rec.ContentIdx) - pivot
			if i >= 0 && i+len(text) <= len(article.Content) {
				// 检查两段文字是否相同或接近
				templ := article.Content[i : i+len(text)]
				totalDist := 0
				for j, s := range text {
					dist := levenshtein.ComputeDistance(s, templ[j])
					totalDist += dist
					if totalDist >= bestDist {
						break
					}
				}
				if totalDist < bestDist {
					bestDist = totalDist
					bestArticle = int(rec.ArticleIdx)
					bestContent = i
					if totalDist == 0 {
						return true
					}
				}
			}

			index++
		}
		return false
	})

	return (bestDist == 0), bestArticle, bestContent
}
