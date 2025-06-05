package entity

type UserToken struct {
	Id    int32    `json:"sub"`
	Email string   `json:"email"`
	Roles []string `json:"roles"`
}
