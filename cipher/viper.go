package cipher

import (
	"github.com/go-kit/kit/log"
	"github.com/goph/emperror"
	"github.com/spf13/viper"
)

// LocalCerts specify where locally to find the certs for a hash.
type LocalCerts struct {
	Path     string
	HashName string
}

const (
	// CipherKey is the Viper subkey under which logging should be stored.
	// NewOptions *does not* assume this key.
	CipherKey = "cipher"
)

// Options is the list of configurations used to load ciphers.
type Options []Config

// Ciphers provide all of the possibly algorithms that can be used to encrypt
// or decrypt.
type Ciphers struct {
	Options map[AlgorithmType]map[string]Decrypt
}

// GetEncrypter takes options and creates an encrypter out of it.
func (o Options) GetEncrypter(logger log.Logger) (Encrypt, error) {
	var lastErr error
	for _, elem := range o {
		if encrypter, err := elem.LoadEncrypt(); err == nil {
			return encrypter, nil
		} else {
			lastErr = err
		}
	}
	return DefaultCipherEncrypter(), emperror.Wrap(lastErr, "failed to load encrypt options")
}

// PopulateCiphers takes options and a logger and creates ciphers from them.
func PopulateCiphers(o Options, logger log.Logger) Ciphers {
	c := Ciphers{
		Options: map[AlgorithmType]map[string]Decrypt{},
	}
	for _, elem := range o {
		elem.Logger = logger
		if decrypter, err := elem.LoadDecrypt(); err == nil {
			if _, ok := c.Options[elem.Type]; !ok {
				c.Options[elem.Type] = map[string]Decrypt{}
			}
			c.Options[elem.Type][elem.KID] = decrypter
		}
	}
	return c
}

// Get returns a decrypter given an algorithm and KID.
func (c *Ciphers) Get(alg AlgorithmType, KID string) (Decrypt, bool) {
	if d, ok := c.Options[alg][KID]; ok {
		return d, ok
	}
	return nil, false
}

// FromViper produces an Options from a (possibly nil) Viper instance.
// cipher key is expected
func FromViper(v *viper.Viper) (o Options, err error) {
	err = v.UnmarshalKey(CipherKey, &o)
	return
}
