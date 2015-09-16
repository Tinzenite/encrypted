package encrypted

/*
TODO abstract storage away so that we can write to hadoop AND disk.
*/
type storage struct {
}

func createStorage() *storage {
	return &storage{}
}

/*
TODO check how data should be passed to here
*/
func (s *storage) WriteData() {
	// TODO write to whatever has been selected for data
}

/*
TODO check how data should be passed to here
*/
func (s *storage) WriteOrg() {
	// TODO write non data stuff (ORDFIR, TEMPDIR, etc)
}
