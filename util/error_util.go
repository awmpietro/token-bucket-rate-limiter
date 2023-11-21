package util

import "fmt"

type CustomError struct {
	Message string
	Code    int
}

func (c *CustomError) Error() string {
	return fmt.Sprintf("error: %s\ncode: %d", c.Message, c.Code)
}
