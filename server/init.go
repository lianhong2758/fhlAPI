package server

import (
	"github.com/lianhong2758/fhlAPI/fhl"
	"fmt"
	"os"

	"github.com/FloatTech/floatbox/file"
	"github.com/sirupsen/logrus"
)

func FhlInit() (*fhl.FHL, error) {
	defer func() {
		if err := recover(); err != nil {
			logrus.Error("panic:", err)
		}
	}()
	if file.IsNotExist("data") {
		_ = os.MkdirAll("data", 0755)
	}
	f := fhl.NewFHL()
	if f.LoadDatasetFile(fhl.DatasetPath).Error != nil {
		return nil, fmt.Errorf("缺少诗词数据文件")

	}
	if f.LoadPrecalFile(fhl.PrecalPath).Error != nil || f.LoadPrecalErrCorr(fhl.ErrCorrPath).Error != nil {
		logrus.Error(f.Error)
		if f.Init().Calculate().InitPrecal().SavePrecal().Error != nil || f.InitErrCorr().Error != nil {
			return nil, f.Error

		}
		f.DeleteCache()
	}
	return f, nil
}
