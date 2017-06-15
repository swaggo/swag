package gen

import ("github.com/easonlin404/gin-swagger/parse")

type Gen struct {

}

func New() *Gen {
	return &Gen{
	}
}

func (g *Gen) Build() {
	searchDir:="./"
	mainApiFile:="./main.go"
	parser:= parse.New()
	// get
	parser.GetAllGoFileInfo(searchDir)
	parser.ParseGeneralApiInfo(mainApiFile)

}