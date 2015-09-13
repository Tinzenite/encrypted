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
	log.Println("TODO: received lock message!", lm.String())
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
	log.Println("TODO: received request message!", rm.String())
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
	log.Println("TODO received push message!", pm.String())
}
