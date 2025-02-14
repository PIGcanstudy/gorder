package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

func init() {
	if err := NewViperConfig(); err != nil {
		panic(err)
	}
}

var once sync.Once

func NewViperConfig() (err error) {
	once.Do(func() {
		err = newViperConfig()
	})
	return
}

func newViperConfig() error {
	relPath, err := getRelativePathFromCaller()
	if err != nil {
		return err
	}
	viper.SetConfigName("global")
	viper.SetConfigType("yaml")  // 指定配置文件类型
	viper.AddConfigPath(relPath) // 指定路径
	viper.EnvKeyReplacer(strings.NewReplacer("-", "_"))
	//_ = viper.BindEnv("stripe-key", "STRIPE_KEY", "endpoint-stripe-secret", "ENDPOINT_STRIPE_SECRET") // 绑定环境变量
	viper.AutomaticEnv() // 如果有环境变量就去环境变量上去找
	return viper.ReadInConfig()
}

func getRelativePathFromCaller() (relPath string, err error) {
	callerPwd, err := os.Getwd() // 获取当前工作目录
	if err != nil {
		return
	}
	_, here, _, _ := runtime.Caller(0)                         // 获取调用者信息（只用到文件目录）
	relPath, err = filepath.Rel(callerPwd, filepath.Dir(here)) // 获取相对路径
	fmt.Printf("caller from: %s, here: %s, relpath: %s", callerPwd, here, relPath)
	return
}
