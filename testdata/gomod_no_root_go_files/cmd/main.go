package main

import "net/http"

// @title Test API
// @version 1.0
// @description Test API for go.mod without root Go files

// @host localhost:8080
// @BasePath /api/v1
func main() {
	http.HandleFunc("/test", testHandler)
	http.ListenAndServe(":8080", nil)
}

// testHandler handles test requests
// @Summary Test endpoint
// @Description Returns a test response
// @Tags test
// @Accept json
// @Produce json
// @Success 200 {string} string "OK"
// @Router /test [get]
func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
