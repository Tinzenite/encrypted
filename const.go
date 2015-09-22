package encrypted

import (
	"errors"
	"time"
)

/*lockTimeout if how long a lock is kept if no new messages are received.*/
const lockTimeout = time.Duration(1 * time.Minute)

/*
Various errors for encrypted.
*/
var (
	ErrNonEmpty = errors.New("non empty directory as path")
)
