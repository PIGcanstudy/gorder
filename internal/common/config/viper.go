package config

import "github.com/spf13/viper"

func init() {
	if err := NewViperConfig(); err != nil {
		panic(err)
	}
}

func NewViperConfig() error {
	viper.SetConfigName("global")
	viper.SetConfigType("yaml")             // 指定配置文件类型
	viper.AddConfigPath("../common/config") // 指定路径
	viper.AutomaticEnv()                    // 如果有环境变量就去环境变量上去找
	return viper.ReadInConfig()
}
