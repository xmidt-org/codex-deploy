package cipher

import (
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

// Sub returns the standard child Viper, using CipherKey, for this package.
// If passed nil, this function returns nil.
func Sub(v *viper.Viper) *viper.Viper {
	if v != nil {
		return v.Sub(CipherKey)
	}

	return nil
}

// FromViper produces an Options from a (possibly nil) Viper instance.
// Callers should use FromViper(Sub(v)) if the standard subkey is desired.
func FromViper(v *viper.Viper) (*Options, error) {
	o := new(Options)
	if v != nil {
		if err := v.Unmarshal(o); err != nil {
			return nil, err
		}
	}

	return o, nil
}
