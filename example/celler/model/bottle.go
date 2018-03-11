package model

type Bottle struct {
	ID      int     `json:"id" example:"1"`
	Name    string  `json:"name" example:"bottle_name"`
	Account Account `json:"account"`
}

func BottlesAll() ([]Bottle, error) {
	return bottles, nil
}

func BottleOne(id int) (*Bottle, error) {
	for _, v := range bottles {
		if id == v.ID {
			return &v, nil
		}
	}
	return nil, ErrNoRow
}

var bottles = []Bottle{
	Bottle{ID: 1, Name: "bottle_1", Account: Account{ID: 1, Name: "accout_1"}},
	Bottle{ID: 2, Name: "bottle_2", Account: Account{ID: 2, Name: "accout_2"}},
	Bottle{ID: 3, Name: "bottle_3", Account: Account{ID: 3, Name: "accout_3"}},
}
