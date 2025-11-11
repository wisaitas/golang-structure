package register

type Request struct {
	Name            string `json:"name" validate:"required"`
	Email           string `json:"email" validate:"required,email"`
	Age             int    `json:"age" validate:"required,min=0,max=150"`
	Password        string `json:"password" validate:"required,min=8,max=24"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}
