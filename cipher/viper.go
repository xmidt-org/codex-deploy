package cipher

import (
	"encoding/json"
	"github.com/go-kit/kit/log"
	"github.com/goph/emperror"
	"github.com/spf13/viper"
)

type LocalCerts struct {
	Path     string
	HashName string
}

const (
	// CipherKey is the Viper subkey under which logging should be stored.
	// NewOptions *does not* assume this key.
	CipherKey = "cipher"
)

type Options []Config

type Ciphers struct {
	options map[AlgorithmType]map[string]Decrypt
}

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

func PopulateCiphers(o Options, logger log.Logger) Ciphers {
	c := Ciphers{
		options: map[AlgorithmType]map[string]Decrypt{},
	}
	for _, elem := range o {
		elem.Logger = logger
		if decrypter, err := elem.LoadDecrypt(); err == nil {
			if _, ok := c.options[elem.Type]; !ok {
				c.options[elem.Type] = map[string]Decrypt{}
			}
			c.options[elem.Type][elem.KID] = decrypter
		}
	}
	return c
}

func (c *Ciphers) Get(alg AlgorithmType, KID string) (Decrypt, bool) {
	if d, ok := c.options[alg][KID]; ok {
		return d, ok
	}
	return nil, false
}

// FromViper produces an Options from a (possibly nil) Viper instance.
// cipher key is expected
func FromViper(v *viper.Viper) (o Options, err error) {
	obj := v.Get("cipher")
	data, err := json.Marshal(obj)
	if err != nil {
		return []Config{}, emperror.Wrap(err, "failed to load cipher config")
	}

	err = json.Unmarshal(data, &o)
	return
}
