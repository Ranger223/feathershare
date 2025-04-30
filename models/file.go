package models

import "time"

type File struct {
	ID         int       `json:"id"`
	Filename   string    `json:"filename"`
	UploadedAt time.Time `json:"uploaded_at"`
}
