package model


type BulkDeleteRequest struct {
	ID []int `json:"id"`
}