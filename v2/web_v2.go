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
