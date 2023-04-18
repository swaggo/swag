package api

// MyFunc godoc
// @Description My Function
// @Success 200 {object} MyStruct
// @Router /myfunc [get]
func MyFunc() {}

type MyStruct struct {
	URLTemplate string `json:"urltemplate" example:"http://example.org/{{ path }}" swaggertype:"string"`
}
