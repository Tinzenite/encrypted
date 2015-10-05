package encrypted

/*
Storage is the interface a struct must satisfy to allow encrypted to use it as a
storage backend.
*/
type Storage interface {
	/*Store writes the given data to the key.*/
	Store(key string, data []byte) error
	/*Retrieve fetches the data for a key.*/
	Retrieve(key string) ([]byte, error)
	/*Remove is called to remove a key and associated data from storage.*/
	Remove(key string) error
}
