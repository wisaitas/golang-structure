package updateuser

type Request struct {
	Name string `json:"name" validate:"required"`
	Age  int    `json:"age"`
}
