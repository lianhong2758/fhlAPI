package server

import (
	"time"

	"github.com/FloatTech/ttl"
)

var Games = ttl.NewCache[string, *AnswerResq](time.Minute * 15)
