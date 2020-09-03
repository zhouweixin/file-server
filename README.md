# 统一异常处理--file-server

代码地址: 

## 1 创建工程file-server, 目录结构为

> /file-server
  ├── README.md
  ├── go.mod
  └── v1
      └── web_v1.go


## 2 开发接口`/list/*`

### 2.1 web_v1.go

```go
package main

import (
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/list/", func(writer http.ResponseWriter,
		request *http.Request) {

		// 1.解析文件路径
		path := request.URL.Path[len("/list/"):]

		// 2.打开文件
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		// 3.读文件的内容
		all, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}

		// 4.写入response
		_, err = writer.Write(all)
		if err != nil {
			panic(err)
		}
	})

	// 5.启动服务
	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		panic(err)
	}
}
```

### 2.2 启动文件服务
启动命令
```shell script
$ go run web.go
file server listening, please visit: http://127.0.0.1:8888
```

### 2.3 浏览器里测试
> http://127.0.0.1:8888/list/README.md

即可看到文件内容

### 2.4 分析

1. 假如请求写错, 服务端会发生panic, 如输入url: http://127.0.0.1:8888/list/aaa.txt

   ![image-20200902132520293](/Users/zhouweixin/Library/Application Support/typora-user-images/image-20200902132520293.png)

2. 用户并不明白为什么出错, 只看到一些不友好的信息

   ![image-20200902132434479](/Users/zhouweixin/Library/Application Support/typora-user-images/image-20200902132434479.png)

3. 如何把预料到的及不可预料的错误, 以友好的方式呈现?

## 3 统一异常处理

### 3.1 web_v2.go

1. 利用go语言的函数式编程, 实现统一的异常处理

2. 将`/list/`路由处理函数提出来, 命名handlerFileList, 并添加返回值error; 修改内容: 遇到错误即抛出, 只做对的事情

3. **创建错误包装函数errWrapper, 入参为handlerFileList, 返回值为http.HandleFunc需要的处理函数**

```go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// "/list/"路由处理函数, 遇到错误即抛出, 只做对的事情
func handlerFileList(writer http.ResponseWriter, request *http.Request) error {

	// 1.解析文件路径
	path := request.URL.Path[len("/list/"):]

	// 2.打开文件
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// 3.读文件的内容
	all, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	// 4.写入response
	_, err = writer.Write(all)
	if err != nil {
		return err
	}

	return nil
}

type appHandler func(writer http.ResponseWriter, request *http.Request) error

// 错误包装函数, 传入函数, 返回函数
func errWrapper(handler appHandler) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		err := handler(writer, request)
		if err != nil {
			code := http.StatusOK
			switch {
			case os.IsNotExist(err):
				log.Printf("Error occurred handling request: %s", err.Error())
				code = http.StatusNotFound
			case os.IsPermission(err):
				code = http.StatusForbidden
			default:
				code = http.StatusInternalServerError
			}

			http.Error(writer, http.StatusText(code), http.StatusNotFound)
		}
	}
}

func main() {
	http.HandleFunc("/list/", errWrapper(handlerFileList))

	// 5.启动服务
	fmt.Println("file server listening, please visit: http://127.0.0.1:8888")
	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		panic(err)
	}
}

```

### 3.2 web_v3.go

1. 添加业务错误

```go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const prefix = "/list/"

type UserError string

func (e UserError) Message() string {
	return string(e)
}

func (e UserError) Error() string {
	return e.Message()
}

// "/list/"路由处理函数, 遇到错误即抛出, 只做对的事情
func handlerFileList(writer http.ResponseWriter, request *http.Request) error {

	if !strings.HasPrefix(request.URL.Path, prefix) {
		return UserError("the prefix of url must be " + prefix)
	}

	// 1.解析文件路径
	path := request.URL.Path[len(prefix):]

	// 2.打开文件
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// 3.读文件的内容
	all, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	// 4.写入response
	_, err = writer.Write(all)
	if err != nil {
		return err
	}

	return nil
}

type appHandler func(writer http.ResponseWriter, request *http.Request) error

type BusinessError interface {
	error
	Message() string
}

// 错误包装函数, 传入函数, 返回函数
func errWrapper(handler appHandler) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic: %v", r)
			}
		}()

		err := handler(writer, request)
		if err != nil {
			log.Printf("Error occurred handling request: %s", err.Error())

			// 业务错误
			if businessErr, ok := err.(BusinessError); ok {
				http.Error(writer, businessErr.Message(), http.StatusBadRequest)
				return
			}

			code := http.StatusOK
			switch {
			case os.IsNotExist(err):
				code = http.StatusNotFound
			case os.IsPermission(err):
				code = http.StatusForbidden
			default:
				code = http.StatusInternalServerError
			}

			http.Error(writer, http.StatusText(code), http.StatusNotFound)
		}
	}
}

func main() {
	http.HandleFunc("/", errWrapper(handlerFileList))

	// 5.启动服务
	fmt.Println("file server listening, please visit: http://127.0.0.1:8888")
	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		panic(err)
	}
}
```

