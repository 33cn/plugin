package types

import "fmt"

// some errors definition
var (
	ErrKeyExisted  = fmt.Errorf("%s", "The key has already existed!")
	ErrStorageType = fmt.Errorf("%s", "The key has used storage another type!")
)
