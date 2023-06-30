package util

import (
	"errors"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateUUID(u string) error {
	_, err := uuid.Parse(u)
	return err
}

func ValidateURL(u string) error {
	_, err := url.ParseRequestURI(u)
	if err != nil {
		return err
	}

	return nil
}

func CombineFieldErrors(fieldErrors field.ErrorList) error {

	if len(fieldErrors) == 0 {
		return nil
	}

	var errorStrings []string
	for _, fe := range fieldErrors {
		errorStrings = append(errorStrings, fe.Error())
	}
	return errors.New(strings.Join(errorStrings, "; "))
}
