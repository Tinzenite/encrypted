package encrypted

type chaninterface struct {
	// reference back to encrypted
	enc *Encrypted
}

func createChanInterface(enc *Encrypted) *chaninterface {
	return &chaninterface{
		enc: enc}
}

// ----------------------- Callbacks ------------------------------

func (c *chaninterface) OnNewConnection(address, message string) {}

func (c *chaninterface) OnMessage(address, message string) {}

func (c *chaninterface) OnAllowFile(address, name string) (bool, string) {
	return false, ""
}

func (c *chaninterface) OnFileReceived(address, path, name string) {}

func (c *chaninterface) OnFileCanceled(address, path string) {}

func (c *chaninterface) OnConnected(address string) {}
