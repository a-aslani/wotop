package wotop_util

import (
	"math/rand"
	"time"
)

func ToSliceAny[T any](objs []T) []any {
	datas := make([]any, 0)
	for _, obj := range objs {
		datas = append(datas, obj)
	}
	return datas
}

func GetRandomItem[T any](slice []T) T {
	rand.Seed(time.Now().UnixNano())     // seed or it will be set to 1
	randomIndex := rand.Intn(len(slice)) // generate a random int in the range 0 to 9
	pick := slice[randomIndex]           // get the value from the slice
	return pick
}
