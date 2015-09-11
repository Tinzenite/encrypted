package encrypted

import "errors"

/*
Various errors for encrypted.
*/
var (
	ErrNonEmpty = errors.New("non empty directory as path")
)
