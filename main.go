package main

import (
	"fhlApi/fhl"
	"fhlApi/server"
	"fmt"
)

func main() {
	defer func() {
        if err := recover(); err != nil {
            fmt.Println("panic:", err)
        }
    }()
	f := fhl.NewFHL()
	if f.LoadDatasetFile("data/2b-dedup.txt").Error != nil {
		fmt.Println("缺少诗词数据文件")
		return
	}
	if f.LoadPrecalFile("data/2c-precal.json").Error != nil || f.LoadPrecalErrCorr("data/2c-errcorr.bin").Error != nil {
		fmt.Println(f.Error)
		if f.Init().Calculate().InitPrecal().SavePrecal().Error != nil || f.InitErrCorr().Error != nil {
			fmt.Println(f.Error)
			return
		}
		f.DeleteCache()
	}
	server.RunGin(f)
}
