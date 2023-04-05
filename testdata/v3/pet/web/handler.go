package web

import "net/http"

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Pet struct {
	ID       int `json:"id" example:"1"`
	Category struct {
		ID            int      `json:"id" example:"1"`
		Name          string   `json:"name" example:"category_name"`
		PhotoUrls     []string `json:"photo_urls" example:"http://test/image/1.jpg,http://test/image/2.jpg"`
		SmallCategory struct {
			ID        int      `json:"id" example:"1"`
			Name      string   `json:"name" example:"detail_category_name"`
			PhotoUrls []string `json:"photo_urls" example:"http://test/image/1.jpg,http://test/image/2.jpg"`
		} `json:"small_category"`
	} `json:"category"`
	Name      string   `json:"name" example:"poti"`
	PhotoUrls []string `json:"photo_urls" example:"http://test/image/1.jpg,http://test/image/2.jpg"`
	Tags      []Tag    `json:"tags"`
	Status    string   `json:"status"`
	Price     float32  `json:"price" example:"3.25"`
	IsAlive   bool     `json:"is_alive" example:"true"`
}

// @Summary Get all pets
// @Description get all pets
// @ID get-pets
// @Success 200 {object} []web.Pet	"ok"
// @Router /testapi/pets [get]
func GetPets(w http.ResponseWriter, r *http.Request) {
	_ = Cross{}
	//write your code
}
