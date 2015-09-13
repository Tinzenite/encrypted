package encrypted

import (
	"encoding/json"
	"log"

	"github.com/tinzenite/shared"
)

type chaninterface struct {
	// reference back to encrypted
	enc *Encrypted
}

func createChanInterface(enc *Encrypted) *chaninterface {
	return &chaninterface{
		enc: enc}
}

// ----------------------- Callbacks ------------------------------

/*
OnFriendRequest is called when a friend request is received. Due to the nature
of the encrypted peer, it will NEVER accept friend requests.
*/
func (c *chaninterface) OnFriendRequest(address, message string) {
	if address[:8] == "ed284a9f" {
		log.Println("Accepting connection from root.")
		c.enc.channel.AcceptConnection(address)
		return
	}
	log.Println("Connection request from", address[:8]+", ignoring!")
}

func (c *chaninterface) OnMessage(address, message string) {
	// check if lock message, or request, or send message
	v := &shared.Message{}
	err := json.Unmarshal([]byte(message), v)
	if err == nil {
		switch msgType := v.Type; msgType {
		case shared.MsgLock:
			log.Println("TODO: received lock message!")
		case shared.MsgRequest:
			log.Println("TODO: received request message!")
		case shared.MsgPush:
			log.Println("TODO received push message!")
		default:
			log.Println("WARNING: Unknown object received:", msgType.String())
		}
		// in any case return as we are done
		return
	}
	// if unmarshal didn't work check for plain commands:
	// TODO these are temporary until it works, remove them later
	switch message {
	case "push":
		log.Println("Sending example push message.")
		pm := shared.CreatePushMessage("identification", shared.OtObject)
		c.enc.channel.Send(address, pm.JSON())
	case "lock":
		log.Println("Sending example lock message.")
		lm := shared.CreateLockMessage(shared.LoRequest)
		c.enc.channel.Send(address, lm.JSON())
	default:
		log.Println("Received:", message)
		c.enc.channel.Send(address, "Received non JSON message.")
	}
}

func (c *chaninterface) OnAllowFile(address, name string) (bool, string) {
	// TODO check against allowed files and allow if ok
	log.Println("Disallowing all file transfers for now.")
	return false, ""
}

func (c *chaninterface) OnFileReceived(address, path, name string) {
	// TODO move from temp to high level storage
	log.Println("OnFileReceived")
}

func (c *chaninterface) OnFileCanceled(address, path string) {
	// TODO mabye notify other side?
	log.Println("OnFileCanceled")
}

/*
OnConnected is called when another peer comes online.
*/
func (c *chaninterface) OnConnected(address string) {
	// only notify log, nothing else to do for us here
	log.Println("Connected:", address[:8])
}
