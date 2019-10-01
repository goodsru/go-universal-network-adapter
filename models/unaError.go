package models

import "fmt"

type UnaError struct {
	Code    int
	Message string
}

func (e *UnaError) Error() string {
	return fmt.Sprintf("%d:%s", e.Code, e.Message)
}
