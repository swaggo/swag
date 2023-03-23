package api

import (
	"log"
	"net/http"
	"time"

	"github.com/swaggo/swag/testdata/alias_type/data"
)

/*// @Summary Get time as string
// @Description get time as string
// @ID time-as-string
// @Accept  json
// @Produce  json
// @Success 200 {object} data.StringAlias	"ok"
// @Router /testapi/time-as-string [get]
func GetTimeAsStringAlias(w http.ResponseWriter, r *http.Request) {
	var foo data.StringAlias = "test"
	log.Println(foo)
	//write your code
}*/

/*// @Summary Get time as time
// @Description get time as time
// @ID time-as-time
// @Accept  json
// @Produce  json
// @Success 200 {object} data.DateOnly	"ok"
// @Router /testapi/time-as-time [get]
func GetTimeAsTimeAlias(w http.ResponseWriter, r *http.Request) {
	var foo = data.DateOnly(time.Now())
	log.Println(foo)
	//write your code
}*/

// @Summary Get container with time and time alias
// @Description test container with time and time alias
// @ID time-as-time-container
// @Accept  json
// @Produce  json
// @Success 200 {object} data.TimeContainer	"ok"
// @Router /testapi/time-as-time-container [get]
func GetTimeAsTimeContainer(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	var foo = data.TimeContainer{
		Name:      "test",
		Timestamp: now,
		//CreatedAt: &now,
	}
	log.Println(foo)
	//write your code
}
