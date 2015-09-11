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

func (c *chaninterface) OnFriendRequest(address, message string) {
	log.Println("NewConnection:", address[:8], "ignoring!")
}

func (c *chaninterface) OnMessage(address, message string) {
	log.Println("Received:", message)
}

func (c *chaninterface) OnAllowFile(address, name string) (bool, string) {
	log.Println("Disallowing all file transfers for now.")
	return false, ""
}

func (c *chaninterface) OnFileReceived(address, path, name string) {
	log.Println("OnFileReceived")
}

func (c *chaninterface) OnFileCanceled(address, path string) {
	log.Println("OnFileCanceled")
}

func (c *chaninterface) OnConnected(address string) {
	log.Println("Connected:", address[:8])
}
