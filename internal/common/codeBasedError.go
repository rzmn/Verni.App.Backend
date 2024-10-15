package common

import (
	"fmt"
)

type ErrorCode interface {
	Message() string
}

type CodeBasedError[Code ErrorCode] struct {
	Code        Code
	Description *string
}

func (e *CodeBasedError[_]) Error() string {
	base := fmt.Sprintf("%v [%s]", e.Code, e.Code.Message())
	if e.Description != nil {
		return fmt.Sprintf("%s - %s", base, *e.Description)
	} else {
		return base
	}
}

func NewError[Code ErrorCode](code Code) *CodeBasedError[Code] {
	return &CodeBasedError[Code]{
		Code: code,
	}
}

func NewErrorWithDescription[Code ErrorCode](code Code, description string) *CodeBasedError[Code] {
	return &CodeBasedError[Code]{
		Code:        code,
		Description: &description,
	}
}
