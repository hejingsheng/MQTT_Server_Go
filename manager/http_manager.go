package manager

import "net/http"

func sayHello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World!"))
}

func Http_Manager_Server() {

	http.HandleFunc("/test", sayHello)

	//2.设置监听的TCP地址并启动服务
	//参数1：TCP地址(IP+Port)
	//参数2：当设置为nil时表示使用DefaultServeMux
	http.ListenAndServe("127.0.0.1:8080", nil)

}