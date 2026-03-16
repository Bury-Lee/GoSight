package main

import (
	"fmt"
	"log"
	"net/http"
)

// 完成后迁移到main.go,在配置中决定是否启用前端
func main() {
	// 路由注册
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/hello", apiHandler)

	port := ":8080"
	fmt.Printf("Server running at http://localhost%s\n", port)

	// 启动服务器
	log.Fatal(http.ListenAndServe(port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Welcome to Go Server</h1>")
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Hello from Go API"}`)
}
