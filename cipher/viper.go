package cipher

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type LocalCerts struct {
	Path     string
	HashName string
}

func Load(v *viper.Viper) (LoadConfig, error) {

	config := new(LocalCerts)
	ciperViper := v.Sub("cipher")
	if ciperViper == nil {
		return LoadConfig{}, errors.New("no cipher to load")
	}
	err := ciperViper.Unmarshal(config)
	if err != nil {
		return LoadConfig{}, err
	}

	return LoadConfig{
		&BasicHashLoader{
			config.HashName,
		},
		&FileLoader{
			Path: config.Path,
		},
	}, nil

}
