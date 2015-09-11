package encrypted

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/tinzenite/channel"
	"github.com/tinzenite/shared"
)

/*
Encrypted is the object which is used to control the encrypted Tinzenite peer.
*/
type Encrypted struct {
	rootPath   string
	peer       *shared.Peer
	cInterface *chaninterface
	channel    *channel.Channel
	wg         sync.WaitGroup
	stop       chan bool
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
	return enc.peer.Name
}

/*
Close cleanly closes everything.
*/
func (enc *Encrypted) Close() {
	enc.stop <- true
	enc.wg.Wait()
	enc.channel.Close()
}

/*
run is the background thread for keeping everything running.
*/
func (enc *Encrypted) run() {
	defer func() { enc.log("Background process stopped.") }()
	updateTicker := time.Tick(15 * time.Second)
	for {
		select {
		case <-enc.stop:
			enc.wg.Done()
			return
		case <-updateTicker:
			err := enc.updatePeers()
			if err != nil {
				enc.warn("Failed to update peers:", err.Error())
			}
		}
	}
}

/*
updatePeers should be called regularily to allow the connection of new peers.
*/
func (enc *Encrypted) updatePeers() error {
	// load peers from ORGDIR
	path := enc.rootPath + "/" + shared.ORGDIR + "/" + shared.PEERSDIR
	peersFiles, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	var peers []*shared.Peer
	for _, stat := range peersFiles {
		data, err := ioutil.ReadFile(path + "/" + stat.Name())
		if err != nil {
			log.Println("Error loading peer " + stat.Name() + " from disk!")
			continue
		}
		peer := &shared.Peer{}
		err = json.Unmarshal(data, peer)
		if err != nil {
			log.Println("Error unmarshaling peer " + stat.Name() + " from disk!")
			continue
		}
		peers = append(peers, peer)
	}
	// now update channel accordingly
	for _, peer := range peers {
		// tox will return an error if the address has already been added, so we just ignore it
		_ = enc.channel.AcceptConnection(peer.Address)
	}
	return nil
}

/*
Log function.
*/
func (enc *Encrypted) log(msg ...string) {
	toPrint := []string{"Encrypted:"}
	toPrint = append(toPrint, msg...)
	log.Println(strings.Join(toPrint, " "))
}

/*
Warn function.
*/
func (enc *Encrypted) warn(msg ...string) {
	toPrint := []string{"Encrypted:", "WARNING:"}
	toPrint = append(toPrint, msg...)
	log.Println(strings.Join(toPrint, " "))
}
