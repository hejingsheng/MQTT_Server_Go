package manager

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Product struct {
	Name      string
	ProductID int64
	Number    int
	Price     float64
	IsOnSale  bool
}

func OnAjax(w http.ResponseWriter, r *http.Request) {

	p := &Product{}
	p.Name = "Xiao mi 6"
	p.IsOnSale = true
	p.Number = 10000
	p.Price = 2499.00
	p.ProductID = 1

	jsonResp, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)

	fmt.Println(string(jsonResp))

}

func Http_Manager_Server() {

	http.HandleFunc("/test", OnAjax)

	//2.设置监听的TCP地址并启动服务
	//参数1：TCP地址(IP+Port)
	//参数2：当设置为nil时表示使用DefaultServeMux
	http.ListenAndServe("127.0.0.1:8080", nil)

}