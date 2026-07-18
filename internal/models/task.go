package models

import (
	"image"
)

type Task struct {
	ID     string
	Img    image.Image
	Status string
	Err    error
}

