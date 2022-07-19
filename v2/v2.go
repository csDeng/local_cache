package v2

import "sync"

type Memo struct {
	f     userFunc
	mu    *sync.Mutex
	cache map[string]result
}

// 私有变量
type userFunc func(key string) (interface{}, error)
type result struct {
	value interface{}
	err   error
}

// 开放接口
func New(f userFunc) *Memo {
	return &Memo{
		f:     f,
		mu:    new(sync.Mutex),
		cache: make(map[string]result),
	}
}

// 返回常见的 einter,error
// 并发安全
func (m *Memo) Get(key string) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	res, has := m.cache[key]
	if !has {
		// 没有函数记忆则调用
		res.value, res.err = m.f(key)
		m.cache[key] = res
	}
	return res.value, res.err
}
