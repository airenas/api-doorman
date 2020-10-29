package cmdapp

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//Config is a viper based application config
var Config = viper.New()

//Log is applications logger
var Log = logrus.New()

//Sub extracts Sub config from viper using env variables
func Sub(config *viper.Viper, name string) *viper.Viper {
	res := config.Sub(name)
	res.SetEnvPrefix(name)
	res.AutomaticEnv()
	return res
}
