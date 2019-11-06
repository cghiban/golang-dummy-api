package model

type Item struct {
	ID        *string `json:"id"`
	Name      *string `json:"title"`
	ShortDesc *string `json:"description"`
	Keywords  *string `json:"tags"`
}
