package encrypted

import "log"

type chaninterface struct {
	// reference back to encrypted
	enc *Encrypted
}

func createChanInterface(enc *Encrypted) *chaninterface {
	return &chaninterface{
		enc: enc}
}

// ----------------------- Callbacks ------------------------------

/*
OnFriendRequest is called when a friend request is received. Due to the nature
of the encrypted peer, it will NEVER accept friend requests.
*/
func (c *chaninterface) OnFriendRequest(address, message string) {
	log.Println("Connection request from", address[:8]+", ignoring!")
}

func (c *chaninterface) OnMessage(address, message string) {
	// TODO check if lock message, or request, or send message
	log.Println("Received:", message)
}

func (c *chaninterface) OnAllowFile(address, name string) (bool, string) {
	// TODO check against allowed files and allow if ok
	log.Println("Disallowing all file transfers for now.")
	return false, ""
}

func (c *chaninterface) OnFileReceived(address, path, name string) {
	// TODO move from temp to high level storage
	log.Println("OnFileReceived")
}

func (c *chaninterface) OnFileCanceled(address, path string) {
	// TODO mabye notify other side?
	log.Println("OnFileCanceled")
}

/*
OnConnected is called when another peer comes online.
*/
func (c *chaninterface) OnConnected(address string) {
	// only notify log, nothing else to do for us here
	log.Println("Connected:", address[:8])
}
