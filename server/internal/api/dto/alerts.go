package dto

type CreateVerifyEmail struct {
	Email   string `json:"email"`
	Project string `json:"project_name"`
}
