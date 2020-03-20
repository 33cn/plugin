package types

import "fmt"

// some errors definition
var (
	ErrAccountNameExist = fmt.Errorf("%s", "The account name has been registered!")
	ErrAccountNameNotExist = fmt.Errorf("%s", "The account name is not exist")

)
