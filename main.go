package main

import (
	"fhlApi/fhl"
	"fhlApi/server"
	"fmt"
)

func main() {
	f := fhl.NewFHL().LoadPrecalFile("data/2c-precal.json")
	if f.Error != nil {
		fmt.Println(f.Error)
		f.LoadDatasetFile("data/2b-dedup.txt").Init().Calculate().InitPrecal().SavePrecal()
		if f.Error != nil {
			fmt.Println(f.Error)
			return
		}
		if f.InitErrCorr().Error != nil {
			fmt.Println(f.Error)
			return
		}
		f.DeleteCache()
	}
	server.RunGin(f)
}
