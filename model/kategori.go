package model

import (
	"time"

)

type Kategori struct {
	ID int `json:"id"`
	CategoryName string `json:"category_name"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type KategoriReq struct {
	CategoryName string `form:"category_name"`
}