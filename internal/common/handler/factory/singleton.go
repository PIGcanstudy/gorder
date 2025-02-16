package factory

import "sync"

// 本函数是一个创建redis客户端的工厂，通过配置文件的redis.xx 的.后xx为key获取value（redis客户端实例），其中的value同一时间只允许一个协程访问，
// 并且redis客户端实例只会创建一次（单例模式）

type Supplier func(string) any // Supplier的名字

type Singleton struct {
	cache    map[string]any // 以配置文件的redis.xx 的.后xx为key，value为redis客户端实例
	locker   *sync.Mutex    // 互斥访问锁
	supplier Supplier       // 用来创建一个客户端实例
}

func NewSingleton(supplier Supplier) *Singleton {
	return &Singleton{
		cache:    make(map[string]any),
		locker:   &sync.Mutex{},
		supplier: supplier,
	}
}

func (s *Singleton) Get(key string) any {
	if value, hit := s.cache[key]; hit {
		return value
	}
	s.locker.Lock()
	defer s.locker.Unlock()
	if value, hit := s.cache[key]; hit {
		return value
	}
	s.cache[key] = s.supplier(key) // 创建redis客户端实例
	return s.cache[key]
}
