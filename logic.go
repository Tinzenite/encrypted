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
		if c.enc.setLock(address) {
			// if successful notify peer of success
			accept := shared.CreateLockMessage(shared.LoAccept)
			c.enc.channel.Send(address, accept.JSON())
		}
		// if not successful we don't do anything, peer will retry
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
	if !c.enc.checkLock(address) {
		log.Println("handleRequestMessage: not locked to given address!", address[:8])
		return
	}
	// key to retrieve from storage
	var identification string
	// check file type and set identification accordingly
	switch rm.ObjType {
	case shared.OtObject:
		identification = rm.Identification
	case shared.OtModel:
		identification = shared.MODELJSON
	default:
		// TODO maybe allow retrieval of this peer too? Need to get peer from PEERSDIR
		log.Println("handleRequestMessage: Invalid ObjType requested!", rm.ObjType.String())
		return
	}
	// fetch data
	data, err := c.enc.storage.Retrieve(identification)
	if err != nil {
		log.Println("handleRequestMessage: retrieval from storage failed:", err)
		return
	}
	// path for temp file
	filePath := c.enc.RootPath + "/" + SEDIR + "/" + c.buildKey(address, identification)
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
handlePushMessage handles the logic upon receiving a PushMessage. NOTE: will
only be actually handled if Encrypted is currently locked.
*/
func (c *chaninterface) handlePushMessage(address string, pm *shared.PushMessage) {
	if !c.enc.checkLock(address) {
		log.Println("handlePushMessage: not locked to given address!", address[:8])
		return
	}
	// note that file transfer is allowed for when file is received
	var key string
	switch pm.ObjType {
	case shared.OtObject:
		key = c.buildKey(address, pm.Identification)
	case shared.OtModel:
		// TODO how do we notice and allow model? FIXME
		log.Println("DEBUG: WARNING model not yet cleanly implemented, check key on push!")
		key = c.buildKey(address, shared.MODELJSON)
	default:
		log.Println("handlePushMessage: Invalid ObjType pushed!", pm.ObjType.String())
		return
	}
	// if we reach this, allow
	c.enc.allowedTransfers[key] = true
}

/*
buildKey is a helper function that builds the key used to identify transfers.
*/
func (c *chaninterface) buildKey(address, identification string) string {
	return address + ":" + identification
}
