package redis

import (
	"fmt"
	"time"

	"github.com/PIGcanstudy/gorder/common/handler/factory"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

const (
	confName      = "redis"
	localSupplier = "local"
)

var (
	singleton = factory.NewSingleton(supplier)
)

func Init() {
	// 获取配置文件中的所有redis服务器（redis集群）配置
	conf := viper.GetStringMap(confName)
	for supplyName := range conf { // 遍历redis服务器的key
		Client(supplyName)
	}
}

func LocalClient() *redis.Client {
	return Client(localSupplier)
}

// 获取redis客户端实例
func Client(name string) *redis.Client {
	return singleton.Get(name).(*redis.Client)
}

// 供应（创建）一个客户端实例
func supplier(key string) any {
	confKey := confName + "." + key // 对应配置文件的redis.key
	type Section struct {
		IP           string        `mapstructure:"ip"`
		Port         string        `mapstructure:"port"`
		PoolSize     int           `mapstructure:"pool_size"`
		MaxConn      int           `mapstructure:"max_conn"`
		ConnTimeout  time.Duration `mapstructure:"conn_timeout"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout"`
	}
	var c Section
	if err := viper.UnmarshalKey(confKey, &c); err != nil { // 从配置文件中读取指定key的配置并序列化到section结构体上
		panic(err)
	}
	// 新建redis客户端
	return redis.NewClient(&redis.Options{
		Network:         "tcp",
		Addr:            fmt.Sprintf("%s:%s", c.IP, c.Port),
		PoolSize:        c.PoolSize,
		MaxActiveConns:  c.MaxConn,
		ConnMaxLifetime: c.ConnTimeout * time.Millisecond,
		ReadTimeout:     c.ReadTimeout * time.Millisecond,
		WriteTimeout:    c.WriteTimeout * time.Millisecond,
	})
}
