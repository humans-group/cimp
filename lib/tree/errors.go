package tree

import "fmt"

var (
	ErrorNotFound    = fmt.Errorf("not found")
	ErrorUnsupported = fmt.Errorf("method is not supported")
)
