package model

import (
	"time"
)

type TaskReq struct {
	Tittle      string `form:"tittle"`
	Description string `form:"description"`
	Status      string `form:"status"`
	Date        string `form:"date"`
	CategoryID  int    `form:"category_id"`
	Important   bool   `form:"important"`
}

type TaskRes struct {
	ID           int        `json:"id"`
	Tittle       string     `json:"tittle"`
	Description  string     `json:"description"`
	Status       string     `json:"status"`
	Date         time.Time  `json:"date"`
	Image        *string    `json:"image"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	IdUser       int        `json:"id_user"`
	CategoryID   *int       `json:"category_id"`
	CategoryName *string    `json:"category_name"`
	Important    *bool      `json:"important"`
}

type TaskFull struct {
	ID          int        `json:"id"`
	Tittle      string     `json:"tittle"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Date        time.Time  `json:"date"`
	Image       *string    `json:"image"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}
