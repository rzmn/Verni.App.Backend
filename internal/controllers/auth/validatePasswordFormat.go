package auth

import (
	"fmt"
)

func validatePasswordFormat(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password should contain more than 6 characters")
	}
	return nil
}
