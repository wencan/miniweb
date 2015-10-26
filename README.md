# miniweb

miniweb是一个golang编写的微型HTTP服务框架。接口设计参考了express、beego。  
miniweb需要组合net/http使用。

这是我写的第一个golang程序。代码写得丑勿怪。

## Sample
~~~ go
package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/wencan/miniweb"
)

type HelloFilter struct{}

func (HelloFilter) Get(in *miniweb.Input, out miniweb.Output) bool {
    out.Ok([]byte("Hello\n"))
    return true
}

func main() {
    router := miniweb.NewRouter()
    router.AnyFunc("/*", func(in *miniweb.Input, out miniweb.Output) bool {
        log.Println(in.Request.Method, in.Request.RequestURI, "From", in.Request.RemoteAddr)
        return false
    })

    router.Get("/hello", HelloFilter{})

    router.GetFunc("/player/:playerid(^\\w*$)/item/:itemid", func(in *miniweb.Input, out miniweb.Output) bool {
        out.Ok([]byte(fmt.Sprintf("Hello, %s\n", in.Fields["playerid"])))
        return true
    })

    router.GetFunc("/user/?:user(^\\w*$)", func(in *miniweb.Input, out miniweb.Output) bool {
        user := in.Fields["user"]
        if len(user) > 0 {
            out.Ok([]byte(fmt.Sprintln("user:", user)))
        } else {
            out.Ok([]byte(fmt.Sprintln("No user")))
        }
        return true
    })

    log.Println("Listen *:12345")
    if err := http.ListenAndServe(":12345", router); err != nil {
        log.Fatal(err)
    }
}
~~~

## HTTP方法
miniweb支持OPTIONS、HEAD、GET、POST、PUT、PATCH、DELETE、TRACE、CONNECT几个HTTP方法，同时支持通过 Router.Any方法支持其它HTTP方法。

## Filter
Filter同express的中间件。miniweb支持对单个路由添加多个filter，router依据filters的添加顺序依次调用这些filteer。  
如果filter处理函数返回true，router不再调用后续filter。  
filter支持函数和struct对象两种形态。

## 路由
路由为url的匹配规则
### 精确匹配
>__/hello__ 匹配：  
1. /hello

### *匹配
匹配任意个路径段
>__/user/*/hello__ 匹配：  
1. /user/hello  
2. /user/123/hello  
3. /user/123/456/hello  

>__/*__ 匹配全部url

### :匹配
获取匹配，正则表示式可选
>__/player/:playerid/item/:itemid__ 匹配：  
1. /player/123/item/456    
_in.Fields["playerid"]： 123； in.Fields["itemid"]： 456_  

>__/player/:playerid(\^[\u4e00-\u9fa5]*$)/item/:itemid__ 匹配：  
1. /player/中文/item/456  
_in.Fields["playerid"]： 中文； in.Fields["itemid"]： 456_  

### ?:匹配
可选获取匹配，正则表达式可选
>__/user/?:user(\^\\w*$)__ 匹配：  
1. /user  
2. /user/abc123  

### 正则匹配
>__/zh-cn/(\^[\u4e00-\u9fa5]*$)__ 匹配：  
1. /zh-cn/中文  