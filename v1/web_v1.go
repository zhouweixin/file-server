package main

import (
	"fmt"
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
	fmt.Println("file server listening, please visit: http://127.0.0.1:8888")
	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		panic(err)
	}
}
