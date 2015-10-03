package encrypted

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/tinzenite/shared"
)

/*
handleLockMessage handles the logic upon receiving a LockMessage. Notably this
includes allowing or disallowing a lock for a specific time frame.
*/
func (c *chaninterface) handleLockMessage(address string, lm *shared.LockMessage) {
	switch lm.Action {
	case shared.LoRequest:
		if c.enc.isLockedAddress(address) {
			// we catch this to avoid having peers trying to sync multiple times at the same time
			log.Println("Relock tried for same address, ignoring!")
			return
		}
		if c.enc.setLock(address) {
			// if successful notify peer of success
			accept := shared.CreateLockMessage(shared.LoAccept)
			c.enc.channel.Send(address, accept.JSON())
		}
		// if not successful send release to signify that peer has no lock
		deny := shared.CreateLockMessage(shared.LoRelease)
		c.enc.channel.Send(address, deny.JSON())
		return
	case shared.LoRelease:
		if c.enc.isLockedAddress(address) {
			c.enc.ClearLock()
			// TODO notify of clear?
			return
		}
		log.Println("handleLockMessage: WARNING: received release request from invalid peer!", address[:8])
	default:
		log.Println("handleLockMessage: Invalid action received!")
	}
}

/*
handleRequestMessage handles the logic upon receiving a RequestMessage. NOTE:
will only be actually handled if Encrypted is currently locked.
*/
func (c *chaninterface) handleRequestMessage(address string, rm *shared.RequestMessage) {
	var data []byte           // data to send
	var identification string // identification for writing temp file
	var err error
	// check file type and fetch data accordingly
	switch rm.ObjType {
	case shared.OtObject:
		// TODO differentiate ORGDIR from path? how? we don't have it... :( FIXME: add more special ID____ things
		// fetch data for normal objects from storage
		data, err = c.enc.storage.Retrieve(rm.Identification)
		identification = rm.Identification
	case shared.OtModel:
		// model is read from specially named file
		data, err = ioutil.ReadFile(c.enc.RootPath + "/" + shared.IDMODEL)
		identification = shared.IDMODEL
	case shared.OtPeer:
		data, err = ioutil.ReadFile(c.enc.RootPath + "/" + shared.ORGDIR + "/" + shared.PEERSDIR + "/" + rm.Identification)
		identification = rm.Identification
	case shared.OtAuth:
		data, err = ioutil.ReadFile(c.enc.RootPath + "/" + shared.ORGDIR + "/" + shared.AUTHJSON)
		identification = rm.Identification
	default:
		log.Println("handleRequestMessage: Invalid ObjType requested!", rm.ObjType.String())
		return
	}
	// if error return
	if err != nil {
		// print error only if not model (because missing model signals that this peer is empty)
		if rm.ObjType != shared.OtModel {
			log.Println("handleRequestMessage: retrieval of", rm.ObjType, "failed:", err)
		}
		// notify sender that it don't exist in any case
		nm := shared.CreateNotifyMessage(shared.NoMissing, identification)
		c.enc.channel.Send(address, nm.JSON())
		return
	}
	// path for temp file
	filePath := c.enc.RootPath + "/" + shared.SENDINGDIR + "/" + c.buildKey(address, identification)
	// write data to temp sending file
	err = ioutil.WriteFile(filePath, data, shared.FILEPERMISSIONMODE)
	if err != nil {
		log.Println("handleRequestMessage: failed to write data to SEDIR:", err)
		return
	}
	// send file
	err = c.enc.channel.SendFile(address, filePath, rm.Identification, func(success bool) {
		// if NOT success, log and keep file for debugging
		if !success {
			log.Println("handleRequestMessage: Failed to send file on request!", filePath)
			return
		}
		// remove file
		err := os.Remove(filePath)
		if err != nil {
			log.Println("handleRequestMessage: failed to remove temp file:", err)
			return
		}
	})
	// if error log
	if err != nil {
		log.Println("handleRequestMessage: SendFile returned error:", err)
	}
}

/*
handlePushMessage handles the logic upon receiving a PushMessage.
*/
func (c *chaninterface) handlePushMessage(address string, pm *shared.PushMessage) {
	// note that file transfer is allowed for when file is received
	key := c.buildKey(address, pm.Identification)
	// if we reach this, allow and store push message too
	c.mutex.Lock()
	c.allowedTransfers[key] = *pm
	c.mutex.Unlock()
	// notify that we have received the push message
	rm := shared.CreateRequestMessage(pm.ObjType, pm.Identification)
	c.enc.channel.Send(address, rm.JSON())
}

/*
buildKey is a helper function that builds the key used to identify transfers.
*/
func (c *chaninterface) buildKey(address, identification string) string {
	return address + ":" + identification
}
