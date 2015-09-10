package encrypted

import "github.com/tinzenite/channel"

type Encrypted struct {
	// root path
	path string
	// internal hidden struct for channel callbacks
	cInterface *chaninterface
	// tox communication channel
	channel *channel.Channel
}
