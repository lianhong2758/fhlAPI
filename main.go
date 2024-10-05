package main

import (
	"github.com/lianhong2758/fhlAPI/server"

	"github.com/sirupsen/logrus"
)

func main() {
	f, err := server.FhlInit()
	if err != nil {
		logrus.Warn("ERROR: ", err)
		return
	}
	server.RunGin(f)
}
