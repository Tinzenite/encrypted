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
	RootPath         string
	Peer             *shared.Peer
	storage          Storage         // storage to use for writing and reading data
	isLocked         bool            // is Encrypted currently locked?
	lockedSince      *time.Time      // time since Encrypted is locked.
	lockedAddress    string          // the address locked to
	allowedTransfers map[string]bool // storage for allowed uploads to encrypted
	cInterface       *chaninterface
	channel          *channel.Channel
	wg               sync.WaitGroup
	stop             chan bool
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
	return enc.Peer.Name
}

/*
Store writes the current state of the structure to disk so that it can be loaded
again later.
*/
func (enc *Encrypted) Store() error {
	// make org directory
	err := shared.MakeEncryptedDir(enc.RootPath)
	if err != nil {
		return err
	}
	// store self peer
	toxData, err := enc.channel.ToxData()
	if err != nil {
		return err
	}
	selfPeer := &shared.ToxPeerDump{
		SelfPeer: enc.Peer,
		ToxData:  toxData}
	return selfPeer.StoreTo(enc.RootPath + "/" + shared.LOCALDIR)
}

/*
IsLocked returns whether this Encrypted is currently locked to a peer. NOTE:
does NOT update the time. That can only be done internally upon receiving valid
messages.
*/
func (enc *Encrypted) IsLocked() bool {
	// if a time is set and enc is locked, check timeout
	if enc.lockedSince != nil && enc.isLocked && time.Since(*enc.lockedSince) < lockTimeout {
		// if time beneath limit still locked
		return true
	}
	// otherwise: reset and return false
	enc.ClearLock()
	return false
}

/*
ClearLock can be used to clear an existing lock. Internally called when a lock is
released. NOTE: this method is public to allow forcing a lock clear.
*/
func (enc *Encrypted) ClearLock() {
	// note address we are clearing
	address := enc.lockedAddress
	// clean lock
	enc.isLocked = false
	enc.lockedAddress = ""
	enc.lockedSince = nil
	// if not valid address we didn't really clear a lock, so we're done
	if address == "" {
		return
	}
	//clean up any outstanding file transfers for cleared address (but not running ones!)
	var toRemove []string
	for key := range enc.allowedTransfers {
		if strings.HasPrefix(key, address) {
			toRemove = append(toRemove, key)
		}
	}
	for _, key := range toRemove {
		delete(enc.allowedTransfers, key)
	}
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
setLock can set the lock. The return value signifies whether the lock was
successful. If not, it most likely means that Encrypted is already locked.
*/
func (enc *Encrypted) setLock(address string) bool {
	// if still validly locked and address mismatches we can't lock
	if enc.IsLocked() && enc.lockedAddress != address {
		return false
	}
	log.Println("DEBUG: LOCK")
	timeStamp := time.Now()
	// otherwise set lock
	enc.isLocked = true
	enc.lockedAddress = address
	enc.lockedSince = &timeStamp
	return true
}

/*
isLockedAddress returns true if the given address is the currently locking one.
*/
func (enc *Encrypted) isLockedAddress(address string) bool {
	if enc.IsLocked() && enc.lockedAddress == address {
		return true
	}
	return false
}

/*
checkLock returns whether the lock is valid. If yes this method will
update the time stamp. NOTE: the timeout is only enforced if another peer tries
to lock Encrypted in the meantime.
*/
func (enc *Encrypted) checkLock(address string) bool {
	// if the address matches update time stamp and return true
	if enc.lockedAddress == address {
		//note: this allows locks to hold for longer than the timeout, as long as
		//      no other peer requested a lock since.
		newStamp := time.Now()
		enc.lockedSince = &(newStamp)
		return true
	}
	return false
}

/*
run is the background thread for keeping everything running.
*/
func (enc *Encrypted) run() {
	defer func() { enc.log("Background process stopped.") }()
	// update peers once every minute
	updateTicker := time.Tick(1 * time.Minute)
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
NOTE: for now will only connect to trusted peers.
*/
func (enc *Encrypted) updatePeers() error {
	// load peers from ORGDIR
	path := enc.RootPath + "/" + shared.ORGDIR + "/" + shared.PEERSDIR
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
		// ignore self peer
		if peer.Address == enc.Peer.Address {
			continue
		}
		// ignore other encrypted peers (there are no operations we can do with them yet)
		if !peer.Trusted {
			// TODO: why CAN'T encrypted peers sync each other? They can just update their encrypted states accordingly... FIXME
			log.Println("DEBUG: ignoring other encrypted peer on peerUpdate!")
			continue
		}
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
