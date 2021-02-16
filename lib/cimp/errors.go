package cimp

import "fmt"

var (
	ErrorNotFoundInKV       = fmt.Errorf("value is not found in KV")
	ErrorParentNotFoundInKV = fmt.Errorf("parent value is not found in KV")
	ErrorTypeIncorrect      = fmt.Errorf("type is incorrect")
)
