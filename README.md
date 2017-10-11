# qlimiter-go
限流 go-server &amp; php-client

# 背景
该项目以服务的方式提供更精准的秒级限流支持
服务端使用 golang 实现，同时提供了 php 版本的client sdk 调用示例

# 使用
1. 启动 server (提前编译好 qlimiter)
```shell
./qlimiter
```

2. 客户端调用 （php版本，sdk位于client文件夹内，将该文件夹内的文件引入到项目中即可）
```php
require 'init.php';
define("QLIMITER_PHP_ROOT", "{$path-to-sdk}");
$client = new Qlimiter_Client("127.0.0.1", 9091);
try {
    $limitKey = 'test';   // 根据不同的业务设置不同的key
    $limitMax = 1000;     // 每秒超过该值即触发限流，limit方法返回值将为false
    $retCurrLimitVal = 0; // 当前秒，并发量
    $res = $client->limit($limitKey, 1000, $retCurrLimitVal); // 返回值 true：未触发限流， false：触发限流
} catch (Exception $ex) {
    var_dump($client->getHost(), $ex->getMessage());
}
```
