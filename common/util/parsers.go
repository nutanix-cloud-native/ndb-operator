package util

import "github.com/google/uuid"

func IsValidUUID(u string) error {
	_, err := uuid.Parse(u)
	return err
}
