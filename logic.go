package encrypted

import (
	"log"

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
		log.Println("WARNING: received release request from invalid peer!")
	default:
		log.Println("Invalid action in LockMessage received!")
	}
}

/*
handleRequestMessage handles the logic upon receiving a RequestMessage. NOTE:
will only be actually handled if Encrypted is currently locked.
*/
func (c *chaninterface) handleRequestMessage(address string, rm *shared.RequestMessage) {
	if !c.enc.checkLock(address) {
		log.Println("DEBUG: not locked to given address!")
		return
	}
	// path of file to send (will be set accordingly depending on ObjType)
	var filePath string
	// check what file to get and set filePath accordingly
	switch rm.ObjType {
	case shared.OtObject:
		filePath = c.enc.RootPath + "/" + rm.Identification
	case shared.OtModel:
		filePath = c.enc.RootPath + "/" + shared.MODELJSON
	default:
		// TODO maybe allow retrieval of this peer too? Need to get peer from PEERSDIR
		log.Println("Invalid ObjType requested!", rm.ObjType.String())
		return
	}
	// check that file exists
	if exists, _ := shared.FileExists(filePath); !exists {
		log.Println("DEBUG: file doesn't exist!", filePath)
		return
	}
	// send file
	err := c.enc.channel.SendFile(address, filePath, rm.Identification, func(success bool) {
		if !success {
			log.Println("Failed to send file on request!", filePath)
		}
	})
	// if error log
	if err != nil {
		log.Println("SendFile returned error:", err)
	}
}

/*
handlePushMessage handles the logic upon receiving a PushMessage. NOTE: will
only be actually handled if Encrypted is currently locked.
*/
func (c *chaninterface) handlePushMessage(address string, pm *shared.PushMessage) {
	if !c.enc.checkLock(address) {
		log.Println("DEBUG: not locked to given address!")
		return
	}
	// TODO implement that file transfer is allowed. Note: write files to temp
	// until transfer successfully completes, THEN overwrite existing (if any)
	log.Println("TODO received push message!", pm.String())
}
