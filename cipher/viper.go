package cipher

import (
	"github.com/spf13/viper"
)

type LocalCerts struct {
	Path     string
	HashName string
}

func Load(v *viper.Viper) (LoadConfig, error) {

	config := new(LocalCerts)
	err := v.Sub("cipher").Unmarshal(config)
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
