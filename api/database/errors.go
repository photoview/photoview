package database

import "fmt"

type Errors struct {
	errors []error
}

func (e Errors) Error() string {
	return fmt.Sprintf("%v", e.Unwrap())
}

func (e Errors) Unwrap() []error {
	return e.errors
}

func (e *Errors) Append(err ...error) {
	e.errors = append(e.errors, err...)
}

func (e *Errors) IsNil() bool {
	return len(e.errors) == 0
}
