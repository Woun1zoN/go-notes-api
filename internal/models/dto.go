package models

import (
    
)

type CreateDTO struct {
	Title      string    `json:"title" validate:"required"`
	Content    string    `json:"content" validate:"min=1"`
}

type UpdateDTO struct {
    Title   *string `json:"title" validate:"required,min=1"`
    Content *string `json:"content" validate:"omitempty,min=1"`
}