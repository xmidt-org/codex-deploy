package cipher

import (
	"crypto"
	"strings"
)

type HashFunction string

const (
	Unknown    HashFunction = "unknown"
	BLAKE2B512 HashFunction = "BLAKE2B512"
	SHA1       HashFunction = "SHA1"
	SHA512     HashFunction = "SHA512"
	MD5        HashFunction = "MD5"
)

func (h HashFunction) GetHash() crypto.Hash {
	switch h {
	case BLAKE2B512:
		return crypto.BLAKE2b_512
	case SHA1:
		return crypto.SHA1
	case SHA512:
		return crypto.SHA512
	case MD5:
		return crypto.MD5
	default:
		return crypto.BLAKE2b_512
	}
}

func (h HashFunction) String() string {
	return string(h)
}

func GetHash(hashType string) HashFunction {
	switch strings.ToUpper(hashType) {
	case BLAKE2B512.String():
		return BLAKE2B512
	case SHA1.String():
		return SHA1
	case SHA512.String():
		return SHA512
	case MD5.String():
		return MD5
	default:
		return Unknown
	}
}
