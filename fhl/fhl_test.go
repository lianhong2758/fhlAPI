package fhl_test

import (
	"fhlApi/fhl"
	"fmt"
	"testing"
)

func TestLookupText(t *testing.T) {
	f := fhl.NewFHL()
	if f.LoadDatasetFile("../data/2b-dedup.txt").Error != nil {
		fmt.Println("缺少诗词数据文件")
		return
	}
	if f.LoadPrecalFile("../data/2c-precal.json").Error != nil || f.LoadPrecalErrCorr("../data/2c-errcorr.bin").Error != nil {
		fmt.Println(f.Error)
		if f.Init().Calculate().InitPrecal().SavePrecal().Error != nil || f.InitErrCorr().Error != nil {
			fmt.Println(f.Error)
			return
		}
		f.DeleteCache()
	}
	//test
	var y bool
	var a, b int
	y, a, b = f.LookupText([]string{"悠哉悠哉", "辗转反侧"})
	fmt.Println(y, f.GetArticle(a).Content[b:])
	y, a, b = f.LookupText([]string{"悠哉游哉", "辗转反侧"})
	fmt.Println(y, f.GetArticle(a).Content[b:])
	y, a, b = f.LookupText([]string{"辗转反侧", "呜呜"})
	fmt.Println(y, a, b)
	y, a, b = f.LookupText([]string{"梳洗罢", "独倚望江楼"})
	fmt.Println(y, f.GetArticle(a).Content[b:])
	y, a, b = f.LookupText([]string{"梳洗黑", "独倚望江楼"})
	fmt.Println(y, f.GetArticle(a).Content[b:])
	y, a, b = f.LookupText([]string{"江南好"})
	fmt.Println(y, f.GetArticle(a).Content[b:])

	fmt.Println(f.GenerateD(8))
}
