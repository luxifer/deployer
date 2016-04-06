package utils

import (
	"hash/fnv"
)

func HashStringToColor(s string) (r, g, b uint32) {
	hash := stringToHash(s)
	r = hash & 0xFF0000 >> 16
	g = hash & 0x00FF0 >> 8
	b = hash & 0x0000FF
	return
}

func stringToHash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
