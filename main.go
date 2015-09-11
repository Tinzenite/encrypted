package encrypted

import (
	"github.com/tinzenite/channel"
	"github.com/tinzenite/shared"
)

/*
Create returns a new Encrypted instance, ready to be connected to an existing
network.
*/
func Create(path, peerName string) (*Encrypted, error) {
	// must start on empty directory
	if empty, err := shared.IsDirectoryEmpty(path); !empty {
		if err != nil {
			return nil, err
		}
		return nil, ErrNonEmpty
	}
	// flag whether we need to clen up after us
	var failed bool
	// make dot Tinzenite
	err := shared.MakeDotTinzenite(path)
	if err != nil {
		return nil, err
	}
	// if failed was set --> clean up by removing everything
	defer func() {
		if failed {
			shared.RemoveDotTinzenite(path)
		}
	}()
	// build
	encrypted := &Encrypted{
		path:     path,
		selfName: peerName}
	// prepare chninterface
	encrypted.cInterface = createChanInterface(encrypted)
	// build channel
	channel, err := channel.Create(peerName, nil, encrypted.cInterface)
	if err != nil {
		failed = true
		return nil, err
	}
	encrypted.channel = channel
	// return instance
	return encrypted, nil
}

/*
Load returns the Encrypted structure for an existing instance.
*/
func Load() {}
