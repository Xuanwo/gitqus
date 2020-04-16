package model

type Comment struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}
