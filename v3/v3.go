package v3

import "sync"

type Memo struct {
	f     userFunc
	mu    *sync.Mutex
	cache map[string]*token
}

// 私有变量
type userFunc func(key string) (interface{}, error)
type result struct {
	value interface{}
	err   error
}

// 包含数据的令牌
type token struct {
	res   result
	ready chan struct{} // 数据准备好之后，关闭
}

// 开放接口
func New(f userFunc) *Memo {
	return &Memo{
		f:     f,
		mu:    new(sync.Mutex),
		cache: make(map[string]*token),
	}
}

// 返回常见的 einter,error
// 非并发安全
func (m *Memo) Get(key string) (interface{}, error) {
	m.mu.Lock()
	data := m.cache[key]
	if data == nil {
		// 获取到零值，说明是第一次访问，也就是快协程
		// 初始化data对象
		data = &token{ready: make(chan struct{})}
		m.cache[key] = data
		m.mu.Unlock()

		// 获取数据
		data.res.value, data.res.err = m.f(key)

		// 关闭通道，其他堵塞的慢协程因为读取到零值，而被释放
		// 同时也意味着，这个数据准备好了
		close(data.ready)
	} else {
		// 慢协程走的逻辑
		m.mu.Unlock()
		<-data.ready
	}
	return data.res.value, data.res.err
}
