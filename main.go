package encrypted

import (
	"errors"

	"github.com/tinzenite/shared"
)

/*
Create returns a new Encrypted instance, ready to be connected to an existing
network.
*/
func Create(path, peerName string) (*Encrypted, error) {
	if empty, _ := shared.IsDirectoryEmpty(path); !empty {
		return nil, errors.New("non empty directory as path")
	}
	return nil, nil
}

/*
Load returns the Encrypted structure for an existing instance.
*/
func Load() {}
