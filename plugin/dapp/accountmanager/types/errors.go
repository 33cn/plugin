package types

import "fmt"

// some errors definition
var (
	ErrAccountIDExist      = fmt.Errorf("%s", "The account ID has been registered!")
	ErrAccountIDNotExist   = fmt.Errorf("%s", "The account ID is not exist")
	ErrAccountIDNotPermiss = fmt.Errorf("%s", "You don't have permission to do that!")
	ErrAssetBalance        = fmt.Errorf("%s", "Insufficient balance!")
	ErrNotAdmin            = fmt.Errorf("%s", "No adiministrator privileges!")
)
