package v2

import (
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"
)

var urls []string

func init() {
	urls = []string{
		"https://www.baidu.com",
		"https://cn.bing.com/",
		"https://blog.csdn.net/",

		"https://www.baidu.com",
		"https://cn.bing.com/",
		"https://blog.csdn.net/",

		"https://www.baidu.com",
		"https://cn.bing.com/",
		"https://blog.csdn.net/",
	}
}

func TestV2(t *testing.T) {
	m := New(httpGet)
	for _, url := range urls {
		start := time.Now()
		value, err := m.Get(url)
		if err != nil {
			t.Log(err)
		}
		t.Logf("%s, %d us, %d bytes\n", url, time.Since(start).Microseconds(), len(value.([]byte)))
	}
}

func TestConcurrence(t *testing.T) {
	m := New(httpGet)
	wg := new(sync.WaitGroup)

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			start := time.Now()
			value, err := m.Get(u)
			if err != nil {
				t.Log(err)
			}
			t.Logf("%s, %d us, %d bytes\n", u, time.Since(start).Microseconds(), len(value.([]byte)))
			wg.Done()
		}(url)

	}
	wg.Wait()
}

func httpGet(url string) (interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	return ioutil.ReadAll(resp.Body)
}
