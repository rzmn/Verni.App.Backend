package auth

import (
	"fmt"
	"net/mail"
	"strings"
)

func validateEmailFormat(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("email is invalid: %v", err)
	}
	if strings.TrimSpace(email) != email {
		return fmt.Errorf("email is invalid: leading or trailing spaces")
	}
	return nil
}
