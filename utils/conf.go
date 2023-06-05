package utils

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	EnvPrefix = "FS"
)

func LoadConfigByViper(in string, v interface{}) error {
	viper.SetConfigFile(in)
	viper.SetEnvPrefix(EnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.AllowEmptyEnv(true)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	log.Debugf("using config: %s \n", viper.ConfigFileUsed())

	if err := viper.Unmarshal(v); err != nil {
		panic(err)
	}
	return nil
}
