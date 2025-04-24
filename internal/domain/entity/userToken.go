package entity

type UserToken struct {
	Id    string   `json:"sub"`
	Email string   `json:"email"`
	Roles []string `json:"roles"`
}
