package configuration

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	BindIpAddress     string
	BindPort          int
	AppendOnly        bool
	AppendFileName    string
	MaxClients        int
	PeerServerAddress []string
	SelfServerAddress string
)

func InitializeConfigurations() {
	viper.SetConfigFile("redis.conf")
	viper.SetConfigType("properties")
	viper.AddConfigPath(".") // 设置配置文件和可执行二进制文件在用一个目录
	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatal(err) // 读取配置文件失败致命错误
	}

	BindIpAddress = viper.GetString("bind")
	BindPort = viper.GetInt("port")
	if "yes" == viper.GetString("appendonly") {
		AppendOnly = true
	}
	AppendFileName = viper.GetString("appendfilename")
	PeerServerAddress = viper.GetStringSlice("peers")
	SelfServerAddress = viper.GetString("self")
}
