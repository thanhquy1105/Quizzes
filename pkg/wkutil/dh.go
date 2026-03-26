package wkutil

import (
	"golang.org/x/crypto/curve25519"
)

func GetCurve25519Key(private, public [32]byte) (Key [32]byte) {

	curve25519.ScalarMult(&Key, &private, &public)
	return
}
