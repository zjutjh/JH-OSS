package config

import (
	"log"

	"github.com/spf13/viper"
)

// Config 全局 Viper 实例
var Config = viper.New()

// OSSFolder 存储文件夹
var OSSFolder string

// init 加载配置文件
func init() {
	Config.SetConfigName("config")
	Config.SetConfigType("yaml")
	Config.AddConfigPath(".")
	err := Config.ReadInConfig()
	if err != nil {
		log.Fatal("Config not found", err)
	}

	OSSFolder = Config.GetString("oss.folder")
}
