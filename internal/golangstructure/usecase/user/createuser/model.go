package createuser

type Request struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}
