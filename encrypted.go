package encrypted

import "github.com/tinzenite/channel"

/*
Encrypted is the object which is used to control the encrypted Tinzenite peer.
*/
type Encrypted struct {
	path       string
	selfName   string
	cInterface *chaninterface
	channel    *channel.Channel
}

/*
Address returns this peers full address.
*/
func (enc *Encrypted) Address() (string, error) {
	return enc.channel.ConnectionAddress()
}

/*
Name returns this peers name.
*/
func (enc *Encrypted) Name() string {
	return enc.selfName
}
