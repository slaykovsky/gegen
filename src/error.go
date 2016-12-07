package src

import (
	"fmt"
	"time"
)

// VirtError is error structure
type VirtError struct {
	When time.Time
	What string
}

func (e VirtError) Error() string {
	return fmt.Sprintf("%v: %v", e.When, e.What)
}

func MakeError(msg string) error {
	return VirtError{When: time.Now(), What: msg}
}
