package fhl_test

import (
	"fhlApi/fhl"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestLookupText(t *testing.T) {
	f := fhl.NewFHL()
	if f.LoadDatasetFile("../"+fhl.DatasetPath).Error != nil {
		logrus.Infoln("缺少诗词数据文件")
		return
	}
	if f.LoadPrecalFile("../"+fhl.PrecalPath).Error != nil || f.LoadPrecalErrCorr("../"+fhl.ErrCorrPath).Error != nil {
		logrus.Infoln(f.Error)
		if f.Init().Calculate().InitPrecal().SavePrecal().Error != nil || f.InitErrCorr().Error != nil {
			logrus.Infoln(f.Error)
			return
		}
		f.DeleteCache()
	}
	//test
	var y bool
	var a, b int
	y, a, b = f.LookupText([]string{"悠哉悠哉", "辗转反侧"})
	logrus.Infoln(y, f.GetArticle(a).Content[b:])
	y, a, b = f.LookupText([]string{"悠哉游哉", "辗转反侧"})
	logrus.Infoln(y, f.GetArticle(a).Content[b:])
	y, a, b = f.LookupText([]string{"辗转反侧", "呜呜"})
	logrus.Infoln(y, a, b)
	y, a, b = f.LookupText([]string{"梳洗罢", "独倚望江楼"})
	logrus.Infoln(y, f.GetArticle(a).Content[b:])
	y, a, b = f.LookupText([]string{"梳洗黑", "独倚望江楼"})
	logrus.Infoln(y, f.GetArticle(a).Content[b:])
	y, a, b = f.LookupText([]string{"江南好"})
	logrus.Infoln(y, f.GetArticle(a).Content[b:])

	logrus.Infoln(f.GenerateD(8))
}
