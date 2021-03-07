package controller

type Response struct {
	Title      map[string]string      `json:"title" example:"en:Map,ru:Карта,kk:Карталар"`
	CustomType map[string]interface{} `json:"map_data" swaggertype:"object,string" example:"key:value,key2:value2"`
	Object     Data                   `json:"object"`
}

type Data struct {
	Text string `json:"title" example:"Object data"`
}
