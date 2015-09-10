package encrypted

type chaninterface struct {
	// reference back to encrypted
	enc *Encrypted
}

func createChanInterface(enc *Encrypted) *chaninterface {
	return &chaninterface{
		enc: enc}
}

// TODO implement channel callbacks
