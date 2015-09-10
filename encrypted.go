package encrypted

import "github.com/tinzenite/channel"

/*
Encrypted is the object which is used to control the encrypted Tinzenite peer.
*/
type Encrypted struct {
	// root path
	path string
	// internal hidden struct for channel callbacks
	cInterface *chaninterface
	// tox communication channel
	channel *channel.Channel
}
