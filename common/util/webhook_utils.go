package util

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
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

func IsFeatureEnabled(key string) bool {
	val, ok := os.LookupEnv(key)

	if !ok {
		fmt.Printf("error reading %s env variable", key)
		// safer to return "true" as default, since we want the webhooks
		// to be ENABLED everywhere outside the local development process
		return true
	} else {

		val, err := strconv.ParseBool(val)

		if err != nil {
			return true
		}

		return val
	}
}
