package utils

import (
	"testing"
)

func TestHash(t *testing.T) {
	hash := stringToHash("foo")
	if hash != 2851307223 {
		t.Fail()
	}
}

func TestHashToRGB(t *testing.T) {
	r, g, b := HashStringToColor("foo")
	if r != 243 {
		t.Fail()
	}
	if g != 14 {
		t.Fail()
	}
	if b != 215 {
		t.Fail()
	}
}
