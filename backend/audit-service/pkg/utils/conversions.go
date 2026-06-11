package utils

import "github.com/google/uuid"

func StringToUUID(str string) (uuid.UUID, error) {
	 return uuid.Parse(str)
}