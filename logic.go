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
}

/*
handleRequestMessage handles the logic upon receiving a RequestMessage. NOTE:
will only be actually handled if Encrypted is currently locked.
*/
func (c *chaninterface) handleRequestMessage(address string, rm *shared.RequestMessage) {
	// TODO check for lock
	log.Println("TODO: received request message!", rm.String())
}

/*
handlePushMessage handles the logic upon receiving a PushMessage. NOTE: will
only be actually handled if Encrypted is currently locked.
*/
func (c *chaninterface) handlePushMessage(address string, pm *shared.PushMessage) {
	// TODO check for lock
	log.Println("TODO received push message!", pm.String())
}
