package encrypted

import (
	"github.com/tinzenite/channel"
	"github.com/tinzenite/shared"
)

/*
Create returns a new Encrypted instance, ready to be connected to an existing
network.
*/
func Create(path, peerName string, storage Storage) (*Encrypted, error) {
	// must start on empty directory
	if empty, err := shared.IsDirectoryEmpty(path); !empty {
		if err != nil {
			return nil, err
		}
		return nil, ErrNonEmpty
	}
	// ensure valid parameters
	if path == "" || peerName == "" || storage == nil {
		return nil, shared.ErrIllegalParameters
	}
	// flag whether we need to clen up after us
	var failed bool
	// make directories
	err := shared.MakeEncryptedDir(path)
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
		RootPath: path, // rootPath for storing root
		storage:  storage}
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
	// make peer with name, address, and set trusted to false
	peer, err := shared.CreatePeer(peerName, address, false)
	if err != nil {
		failed = true
		return nil, err
	}
	encrypted.Peer = peer
	// run background
	initialize(encrypted)
	// store initial copy
	err = encrypted.Store()
	if err != nil {
		failed = true
		return nil, err
	}
	// return instance
	return encrypted, nil
}

/*
Load returns the Encrypted structure for an existing instance.
*/
func Load(path string, storage Storage) (*Encrypted, error) {
	// TODO missing check whether this is a valid path... FIXME
	// make missing dirs if path ok? createEncryptedDirectories(path)
	// ensure valid parameters
	if path == "" || storage == nil {
		return nil, shared.ErrIllegalParameters
	}
	// build structure
	encrypted := &Encrypted{
		RootPath: path,
		storage:  storage}
	// prepare interface
	encrypted.cInterface = createChanInterface(encrypted)
	// load data
	selfPeer, err := shared.LoadToxDumpFrom(path + "/" + shared.LOCALDIR)
	if err != nil {
		return nil, err
	}
	// set self peer
	encrypted.Peer = selfPeer.SelfPeer
	// build channel
	encrypted.channel, err = channel.Create(encrypted.Peer.Name, selfPeer.ToxData, encrypted.cInterface)
	if err != nil {
		return nil, err
	}
	// run background
	initialize(encrypted)
	// return instance
	return encrypted, nil
}

/*
initialize is used to start the background process.
*/
func initialize(enc *Encrypted) {
	enc.wg.Add(1)
	enc.stop = make(chan bool, 1)
	go enc.run()
}
