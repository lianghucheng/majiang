package util

import (
	"math/rand"
	"time"
)

func Random(num int) int {
	return rand.New(rand.NewSource(time.Now().UnixNano() + rand.Int63n(1000))).Intn(num)
}
