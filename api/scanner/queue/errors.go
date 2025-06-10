package queue

import "fmt"

type Errors []error

func (e Errors) Error() string {
	return fmt.Sprintf("%v", e.Unwrap())
}

func (e Errors) Unwrap() []error {
	return []error(e)
}
