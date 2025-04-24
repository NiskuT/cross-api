package utils

import "github.com/google/uuid"

func GenerateState() string {
	return uuid.New().String()
}
