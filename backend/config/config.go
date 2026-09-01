// 项目配置信息包
package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name string
		Port string
	}
	Database struct {
		Dsn          string
		MaxIdleConns int
		MaxOpenCons  int
	}
	DeepSeek struct {
		ApiKey  string `mapstructure:"api_key"`
		BaseUrl string `mapstructure:"base_url"`
		Model   string `mapstructure:"model"`
	}
}

var Appconfig *Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config:%v", err)
	}

	Appconfig = &Config{}

	if err := viper.Unmarshal(Appconfig); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}
	InitDB()
	InitRedis()
}

// InfoTest := Test{
// 	message : "pong",
// }
