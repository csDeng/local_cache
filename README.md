# local_cache

​		之前听别人说缓存的时候，第一反应总是 redis ，但是 redis 以分布式、持久化著称，如果说我有个单机服务，仅仅是想要缓存一些热点数据，无脑使用 redis 的话，尽管 redis 基于 epoll 做了多路复用，但是 IO 消耗时间 ，然后获取存在内存中的数据也需要时间，多少还是有点多余操作。那么有没有办法直接使用进程本身堆栈做缓存呢，毕竟，从理论上来说，这样获取数据才是最快的吧。所以尝试一下基于函数记忆的理论进行本地缓存的简单实现。



* 整体思路，进行函数记忆

## v1

因为没有加锁，在单机串行下没有问题，但是在遇到并发时，遇到数据竞态，即多个Goroutine 对同一变量进行读写。



* 缓存效果测试结果

```shell
Running tool: D:\dev\go1.18\bin\go.exe test -timeout 30s -run ^TestV1$ github.com/csDeng/local_cache/v1

=== RUN   TestV1
    d:\Github\local_cache\v1\v1_test.go:37: https://www.baidu.com, 118532 us, 227 bytes
    d:\Github\local_cache\v1\v1_test.go:37: https://cn.bing.com/, 480652 us, 82548 bytes
    d:\Github\local_cache\v1\v1_test.go:37: https://blog.csdn.net/, 562235 us, 249504 bytes
    d:\Github\local_cache\v1\v1_test.go:37: https://www.baidu.com, 0 us, 227 bytes
    d:\Github\local_cache\v1\v1_test.go:37: https://cn.bing.com/, 0 us, 82548 bytes
    d:\Github\local_cache\v1\v1_test.go:37: https://blog.csdn.net/, 0 us, 249504 bytes
    d:\Github\local_cache\v1\v1_test.go:37: https://www.baidu.com, 0 us, 227 bytes
    d:\Github\local_cache\v1\v1_test.go:37: https://cn.bing.com/, 0 us, 82548 bytes
    d:\Github\local_cache\v1\v1_test.go:37: https://blog.csdn.net/, 0 us, 249504 bytes
--- PASS: TestV1 (1.16s)
PASS
ok      github.com/csDeng/local_cache/v1        (cached)
```



```shell
Running tool: D:\dev\go1.18\bin\go.exe test -timeout 30s -run ^TestConcurrence$ github.com/csDeng/local_cache/v1

=== RUN   TestConcurrence
    d:\Github\local_cache\v1\v1_test.go:53: https://www.baidu.com, 131400 us, 227 bytes
    d:\Github\local_cache\v1\v1_test.go:53: https://www.baidu.com, 130879 us, 227 bytes
    d:\Github\local_cache\v1\v1_test.go:53: https://www.baidu.com, 131061 us, 227 bytes
    d:\Github\local_cache\v1\v1_test.go:53: https://cn.bing.com/, 495898 us, 82091 bytes
    d:\Github\local_cache\v1\v1_test.go:53: https://cn.bing.com/, 500095 us, 82568 bytes
    d:\Github\local_cache\v1\v1_test.go:53: https://cn.bing.com/, 536338 us, 82548 bytes
    d:\Github\local_cache\v1\v1_test.go:53: https://blog.csdn.net/, 558143 us, 248229 bytes
    d:\Github\local_cache\v1\v1_test.go:53: https://blog.csdn.net/, 558664 us, 249002 bytes
    d:\Github\local_cache\v1\v1_test.go:53: https://blog.csdn.net/, 575998 us, 248869 bytes
--- PASS: TestConcurrence (0.58s)
PASS
ok      github.com/csDeng/local_cache/v1        (cached)


```

> 可以看到并发测试时，并没有起到缓存效果

* 竞态测试运行结果

```shell
PS D:\Github\local_cache> go test  -timeout 30s -run TestV1 github.com/csDeng/local_cache/v1
ok      github.com/csDeng/local_cache/v1        1.616s
```

```shell
PS D:\Github\local_cache> go test  -timeout 30s -run TestConcurrence github.com/csDeng/local_cache/v1 -race
==================
WARNING: DATA RACE
Write at 0x00c0002048a8 by goroutine 14:


Previous write at 0x00c0002048a8 by goroutine 8:

...

FAIL
FAIL    github.com/csDeng/local_cache/v1        1.649s
FAIL
```



## v2

加互斥锁，解决数据竞态，但是锁的范围太大了，如果慢函数极慢，则会一直堵塞，最后扛不住压力崩掉。



* 竞态测试运行结果

```shell
PS D:\Github\local_cache> go test  -timeout 30s -run TestV2 github.com/csDeng/local_cache/v2 --race 
ok      github.com/csDeng/local_cache/v2        3.086s
PS D:\Github\local_cache> go test  -timeout 30s -run TestConcurrence github.com/csDeng/local_cache/v2 -race
ok      github.com/csDeng/local_cache/v2        2.908s
```

* 缓存效果测试

```shell
Running tool: D:\dev\go1.18\bin\go.exe test -timeout 30s -run ^TestV2$ github.com/csDeng/local_cache/v2

=== RUN   TestV2
    d:\Github\local_cache\v2\v2_test.go:37: https://www.baidu.com, 116500 us, 227 bytes
    d:\Github\local_cache\v2\v2_test.go:37: https://cn.bing.com/, 482912 us, 82594 bytes
    d:\Github\local_cache\v2\v2_test.go:37: https://blog.csdn.net/, 490979 us, 249721 bytes
    d:\Github\local_cache\v2\v2_test.go:37: https://www.baidu.com, 0 us, 227 bytes
    d:\Github\local_cache\v2\v2_test.go:37: https://cn.bing.com/, 0 us, 82594 bytes
    d:\Github\local_cache\v2\v2_test.go:37: https://blog.csdn.net/, 0 us, 249721 bytes
    d:\Github\local_cache\v2\v2_test.go:37: https://www.baidu.com, 0 us, 227 bytes
    d:\Github\local_cache\v2\v2_test.go:37: https://cn.bing.com/, 0 us, 82594 bytes
    d:\Github\local_cache\v2\v2_test.go:37: https://blog.csdn.net/, 0 us, 249721 bytes
--- PASS: TestV2 (1.09s)
PASS
ok      github.com/csDeng/local_cache/v2        (cached)


> 测试运行完成时间: 2022/7/19 11:56:17 <

Running tool: D:\dev\go1.18\bin\go.exe test -timeout 30s -run ^TestConcurrence$ github.com/csDeng/local_cache/v2

=== RUN   TestConcurrence
    d:\Github\local_cache\v2\v2_test.go:53: https://www.baidu.com, 215356 us, 227 bytes
    d:\Github\local_cache\v2\v2_test.go:53: https://blog.csdn.net/, 751572 us, 249715 bytes
    d:\Github\local_cache\v2\v2_test.go:53: https://blog.csdn.net/, 751572 us, 249715 bytes
    d:\Github\local_cache\v2\v2_test.go:53: https://www.baidu.com, 751572 us, 227 bytes
    d:\Github\local_cache\v2\v2_test.go:53: https://cn.bing.com/, 1225879 us, 82091 bytes
    d:\Github\local_cache\v2\v2_test.go:53: https://cn.bing.com/, 1225310 us, 82091 bytes
    d:\Github\local_cache\v2\v2_test.go:53: https://blog.csdn.net/, 1225879 us, 249715 bytes
    d:\Github\local_cache\v2\v2_test.go:53: https://www.baidu.com, 1225879 us, 227 bytes
    d:\Github\local_cache\v2\v2_test.go:53: https://cn.bing.com/, 1225879 us, 82091 bytes
--- PASS: TestConcurrence (1.23s)
PASS
ok      github.com/csDeng/local_cache/v2        (cached)


> 测试运行完成时间: 2022/7/19 11:56:31 <
```

> 可以看到缓存效果雀氏是有，但是由于加了互斥锁，使得本来并发的请求变成了同步串行。



## v3

>  Gopher 常说，不要使用共享内存来通信，而是用通信来共享内存。

到了这里，雀氏也没感受到 `Go` 的魅力，所以还是使用一下 `channel` 吧。

整体思路，就是利用 chan 的非缓存通道读取堵塞，但是读取关闭的通道并不会 panic 的特性，进行减小加锁的区域。

* 缓存效果测试

```shell

=== RUN   TestV3
    d:\Github\local_cache\v3\v3_test.go:37: https://www.baidu.com, 126799 us, 227 bytes
    d:\Github\local_cache\v3\v3_test.go:37: https://cn.bing.com/, 464846 us, 82587 bytes
    d:\Github\local_cache\v3\v3_test.go:37: https://blog.csdn.net/, 529467 us, 250227 bytes
    d:\Github\local_cache\v3\v3_test.go:37: https://www.baidu.com, 0 us, 227 bytes
    d:\Github\local_cache\v3\v3_test.go:37: https://cn.bing.com/, 0 us, 82587 bytes
    d:\Github\local_cache\v3\v3_test.go:37: https://blog.csdn.net/, 0 us, 250227 bytes
    d:\Github\local_cache\v3\v3_test.go:37: https://www.baidu.com, 0 us, 227 bytes
    d:\Github\local_cache\v3\v3_test.go:37: https://cn.bing.com/, 0 us, 82587 bytes
    d:\Github\local_cache\v3\v3_test.go:37: https://blog.csdn.net/, 0 us, 250227 bytes
--- PASS: TestV3 (1.12s)
PASS
ok      github.com/csDeng/local_cache/v3        (cached)


> 测试运行完成时间: 2022/7/19 12:25:34 <

Running tool: D:\dev\go1.18\bin\go.exe test -timeout 30s -run ^TestConcurrence$ github.com/csDeng/local_cache/v3

=== RUN   TestConcurrence
    d:\Github\local_cache\v3\v3_test.go:53: https://www.baidu.com, 282437 us, 227 bytes
    d:\Github\local_cache\v3\v3_test.go:53: https://www.baidu.com, 282437 us, 227 bytes
    d:\Github\local_cache\v3\v3_test.go:53: https://www.baidu.com, 282940 us, 227 bytes
    d:\Github\local_cache\v3\v3_test.go:53: https://blog.csdn.net/, 655623 us, 247062 bytes
    d:\Github\local_cache\v3\v3_test.go:53: https://blog.csdn.net/, 655623 us, 247062 bytes
    d:\Github\local_cache\v3\v3_test.go:53: https://blog.csdn.net/, 655623 us, 247062 bytes
    d:\Github\local_cache\v3\v3_test.go:53: https://cn.bing.com/, 944808 us, 82091 bytes
    d:\Github\local_cache\v3\v3_test.go:53: https://cn.bing.com/, 944808 us, 82091 bytes
    d:\Github\local_cache\v3\v3_test.go:53: https://cn.bing.com/, 944808 us, 82091 bytes
--- PASS: TestConcurrence (0.94s)
PASS
ok      github.com/csDeng/local_cache/v3        (cached)
> 测试运行完成时间: 2022/7/19 12:25:37 <
```



## THE END 

请给时光以生命，给岁月以文明，给技术以思考。

