package cipher

import (
	"errors"
	"github.com/spf13/viper"
	"strings"
)

type LocalCerts struct {
	Path     string
	HashName string
}

func LoadPublic(v *viper.Viper) (PublicKeyCipher, error) {

	config := new(LocalCerts)
	cipherViper := v.Sub("cipher")
	if cipherViper == nil {
		return nil, errors.New("no cipher to load")
	}
	err := cipherViper.Unmarshal(config)
	if err != nil {
		return nil, err
	}

	// check for noop
	if strings.ToLower(config.HashName) == "noop" {
		return new(NOOP), nil
	}

	loadConfg := LoadConfig{
		&BasicHashLoader{
			config.HashName,
		},
		&FileLoader{
			Path: config.Path,
		},
	}
	return LoadPublicKey(loadConfg)

}

func LoadPrivate(v *viper.Viper) (PrivateKeyCipher, error) {

	config := new(LocalCerts)
	cipherViper := v.Sub("cipher")
	if cipherViper == nil {
		return nil, errors.New("no cipher to load")
	}
	err := cipherViper.Unmarshal(config)
	if err != nil {
		return nil, err
	}

	// check for noop
	if strings.ToLower(config.HashName) == "noop" {
		return new(NOOP), nil
	}

	loadConfg := LoadConfig{
		&BasicHashLoader{
			config.HashName,
		},
		&FileLoader{
			Path: config.Path,
		},
	}
	return LoadPrivateKey(loadConfg)

}
