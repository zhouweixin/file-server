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
