package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type MyStruct struct {
	ID       int      `json:"id" example:"1" format:"int64"`
	Name     string   `json:"name" example:"poti"`
	Intvar   int      `json:"myint,string"`                            // integer as string
	Boolvar  bool     `json:",string"`                                 // boolean as a string
	TrueBool bool     `json:"truebool,string" example:"true"`          // boolean as a string
	Floatvar float64  `json:",string"`                                 // float as a string
	UUIDs    []string `json:"uuids" type:"array,string" format:"uuid"` // string array with format
}

// @Summary Call DoSomething
// @Description Does something
// @Accept  json
// @Produce  json
// @Param body body MyStruct true "My Struct"
// @Success 200 {object} MyStruct
// @Failure 500
// @Router /do-something [post]
func DoSomething(w http.ResponseWriter, r *http.Request) {
	objectFromJSON := new(MyStruct)
	if err := json.NewDecoder(r.Body).Decode(&objectFromJSON); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Print(err.Error())
	}
	json.NewEncoder(w).Encode(ojbectFromJSON)
}

// @title Swagger Example API
// @version 1.0
// @description This is a sample server.
// @host localhost:4000
// @basePath /
func main() {
	http.HandleFund("/do-something", DoSomething)
	http.ListenAndServe(":8080", nil)
}
