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
	// make org directory
	err := shared.MakeDirectories(path, shared.ORGDIR, shared.ORGDIR+"/"+shared.PEERSDIR)
	if err != nil {
		return nil, err
	}
	// if failed was set --> clean up by removing everything
	defer func() {
		if failed {
			shared.RemoveDirContents(path)
		}
	}()
	// build
	encrypted := &Encrypted{
		rootPath: path} // rootPath for storing root
	// prepare chaninterface
	encrypted.cInterface = createChanInterface(encrypted)
	// build channel
	channel, err := channel.Create(peerName, nil, encrypted.cInterface)
	if err != nil {
		failed = true
		return nil, err
	}
	encrypted.channel = channel
	// get address for peer
	address, err := encrypted.channel.Address()
	if err != nil {
		failed = true
		return nil, err
	}
	// make peer (at correct location!)
	peer, err := shared.CreatePeer(peerName, address)
	if err != nil {
		failed = true
		return nil, err
	}
	encrypted.peer = peer
	// run background stuff
	encrypted.wg.Add(1)
	encrypted.stop = make(chan bool, 1)
	go encrypted.run()
	// return instance
	return encrypted, nil
}

/*
Load returns the Encrypted structure for an existing instance.

TODO write
*/
func Load(path string) (*Encrypted, error) {
	return nil, shared.ErrUnsupported
}
