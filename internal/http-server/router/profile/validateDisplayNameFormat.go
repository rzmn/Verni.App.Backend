package profile

import (
	"fmt"
	"regexp"
)

func validateDisplayNameFormat(name string) error {
	if !regexp.MustCompile(`^[A-Za-z]+$`).MatchString(name) {
		return fmt.Errorf("display name is invalid: should contain latin characters only")
	}
	if len(name) < 4 {
		return fmt.Errorf("display name is invalid: should contain at least 4 characters")
	}
	return nil
}
