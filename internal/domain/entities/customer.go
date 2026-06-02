package entities

type Customer struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Cpf string `json:"cpf"`
	Email string `json:"email"`
}
