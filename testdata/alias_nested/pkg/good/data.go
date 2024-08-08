package good

type Gen struct {
	Emb Emb `json:"emb"`
} // @name Gen

type Emb struct {
	Good bool `json:"good"`
}
