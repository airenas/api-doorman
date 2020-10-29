package cmdapp

import (
	"strings"

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
	InitEnv(res)
	res.SetEnvPrefix(name)
	return res
}

//InitEnv initializes viper for environment variables
func InitEnv(config *viper.Viper) {
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()
}
