package main

import (
	"fhlApi/server"

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
