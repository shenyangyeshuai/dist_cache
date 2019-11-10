# 版本迭代

1. 基于 HTTP 协议实现基本的单机版缓存功能 , key 为 string , value 为 []byte .

2. 因为 HTTP 协议的本质原因 , 实现另一套基于 TCP 的单机版缓存功能 , 性能提升很多 .

3. 将基于 TCP 的服务端改成异步处理 , 具体方式就是 channel 的 channel .
   一个 conn 可以同时过来很多请求 , 创建一个 channel 的 channel , 然后
   每个请求按照循序创建一个 channel 并放进 channel 的 channel 后异步处理 .
   这样, 有可能先到的请求处理完成的慢 , 但是顺序得以保证 .

4. 将单机版改成分布式版本 .
   协议库使用 github.com/hashicorp/memberlist .
   一致性哈希使用 stathat.com/c/consistent .
   通过在 Server 端保存一个 node 启动多个节点的服务 .
   因为负载均衡需要重定向 , 所以不适合缓存应用的开发 .
   故而通过客户端在启动时随机访问一台缓存节点 , 获取集群所有节点列表并对自己操作的
   每一个键计算一致性哈希来决定访问集群哪个节点 .

5. rebalance 功能 . 通过提供一个 HTTP 接口 , 在 cache 中保存一个 scanner
   来实现当有新节点加入到集群时 , 针对指定的节点遍历 k-v , 将一部分 k-v 迁移到新节点上 .

6. 提供简单的缓存过期功能 . 在 cache 中增加 ttl 字段来保存过期时间 .
   在节点启动时 , 可以设置某一节点的 key 的过期时间 . 方法时异步启动一个 goroutine ,
   先睡一段过期时间 , 然后遍历所有的 key 并判断是否过期来决定删除 .
