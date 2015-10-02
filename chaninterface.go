package encrypted

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/tinzenite/shared"
)

type chaninterface struct {
	enc   *Encrypted // reference back to encrypted
	mutex sync.Mutex // required for map of incomming stuff
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
	// for now only accept connection from myself for testing
	if address[:8] == "ed284a9f" {
		// TODO remove once done debugging / "dev"-ing
		log.Println("OnFriendRequest: Accepting connection from root.")
		c.enc.channel.AcceptConnection(address)
		return
	}
	// TODO usually encrypted should NEVER accept a friend request
	log.Println("OnFriendRequest: Connection request from", address[:8]+", ignoring!")
}

func (c *chaninterface) OnMessage(address, message string) {
	// check if lock message, or request, or send message
	v := &shared.Message{}
	err := json.Unmarshal([]byte(message), v)
	if err == nil {
		// special case for lock messages (can be received if not locked)
		if v.Type == shared.MsgLock {
			msg := &shared.LockMessage{}
			err := json.Unmarshal([]byte(message), msg)
			if err != nil {
				log.Println("OnMessage: failed to parse JSON!", err)
				return
			}
			c.handleLockMessage(address, msg)
			return
		}
		// for all others ensure that we are locked correctly
		if !c.enc.checkLock(address) {
			// if not warn and ignore message
			log.Println("OnMessage: not locked to given address!", address[:8])
			// TODO send notify that they are unlocked back?
			return
		}
		// if correctly locked handle message according to type
		switch msgType := v.Type; msgType {
		case shared.MsgRequest:
			msg := &shared.RequestMessage{}
			err := json.Unmarshal([]byte(message), msg)
			if err != nil {
				log.Println("OnMessage: failed to parse JSON!", err)
				return
			}
			c.handleRequestMessage(address, msg)
		case shared.MsgPush:
			msg := &shared.PushMessage{}
			err := json.Unmarshal([]byte(message), msg)
			if err != nil {
				log.Println("OnMessage: failed to parse JSON!", err)
				return
			}
			c.handlePushMessage(address, msg)
		default:
			log.Println("OnMessage: WARNING: Unknown object received:", msgType.String())
		}
		// in any case return as we are done handling them
		return
	}
	// if unmarshal didn't work check for plain commands:
	// TODO these are temporary until it works, remove them later
	switch message {
	case "push":
		log.Println("Sending example push message.")
		pm := shared.CreatePushMessage("ID_HERE", shared.OtObject)
		c.enc.channel.Send(address, pm.JSON())
	case "lock":
		log.Println("Sending example lock message.")
		lm := shared.CreateLockMessage(shared.LoRequest)
		c.enc.channel.Send(address, lm.JSON())
	case "unlock":
		log.Println("Sending example unlock message.")
		lm := shared.CreateLockMessage(shared.LoRelease)
		c.enc.channel.Send(address, lm.JSON())
	case "request":
		log.Println("Sending example request message.")
		rm := shared.CreateRequestMessage(shared.OtObject, "ID_HERE")
		c.enc.channel.Send(address, rm.JSON())
	default:
		log.Println("Received:", message)
		c.enc.channel.Send(address, "Received non JSON message.")
	}
}

/*
OnAllowFile is called when a file is to be received. Name should be the
file identification!
*/
func (c *chaninterface) OnAllowFile(address, name string) (bool, string) {
	if !c.enc.checkLock(address) {
		log.Println("OnAllowFile: not locked to given address, refusing!")
		return false, ""
	}
	//check against allowed files and allow if ok
	key := c.buildKey(address, name)
	_, exists := c.enc.allowedTransfers[key]
	if !exists {
		log.Println("OnAllowFile: refusing file transfer due to no allowance!")
		return false, ""
	}
	//write to RECEIVINGDIR
	return true, c.enc.RootPath + "/" + shared.RECEIVINGDIR + "/" + key
}

/*
OnFileReceived is called when a file has been successfully received.
*/
func (c *chaninterface) OnFileReceived(address, path, name string) {
	// TODO fix this: NOTE: no lock check so that locks don't have to stay on for long file transfers
	// no matter what, remove temp file
	defer func() {
		err := os.Remove(path)
		if err != nil {
			log.Println("OnFileReceived: failed to remove temp file:", err)
		}
		// remove from allowedTransfers
		c.mutex.Lock()
		delete(c.enc.allowedTransfers, name)
		c.mutex.Unlock()
	}()
	// fetch push message for file
	pm, exists := c.enc.allowedTransfers[name]
	if !exists {
		log.Println("OnFileReceived: no associated push message found!")
		return
	}
	// read data
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("OnFileReceived: failed to read file:", err)
		return
	}
	// depending on the object type write the file to different locations:
	switch pm.ObjType {
	case shared.OtModel:
		// model is not written to storage but to disk directly
		path := c.enc.RootPath + "/" + shared.IDMODEL
		err = ioutil.WriteFile(path, data, shared.FILEPERMISSIONMODE)
		// log.Println("DEBUG: wrote model.")
	case shared.OtPeer:
		// peers are written to disk too, but in correct dir with pm.Name
		path := c.enc.RootPath + "/" + shared.ORGDIR + "/" + shared.PEERSDIR + "/" + pm.Identification
		err = ioutil.WriteFile(path, data, shared.FILEPERMISSIONMODE)
		// log.Println("DEBUG: wrote peer.")
	case shared.OtAuth:
		// auth is also special case
		path := c.enc.RootPath + "/" + shared.ORGDIR + "/" + shared.AUTHJSON
		err = ioutil.WriteFile(path, data, shared.FILEPERMISSIONMODE)
		// log.Println("DEBUG: wrote auth.")
	case shared.OtObject:
		// write to storage
		err = c.enc.storage.Store(pm.Identification, data)
	default:
		log.Println("OnFileReceived: unknown ObjType for received file!", pm.ObjType)
		return
	}
	// this means something failed
	if err != nil {
		log.Println("OnFileReceived: writing file failed:", err)
		return
	}
}

/*
OnFileCanceled is called when a file has failed to be successfully received.
*/
func (c *chaninterface) OnFileCanceled(address, path string) {
	// note: no lock check so that locks don't have to stay on for long file transfers
	log.Println("OnFileCanceled:", path)
	// remove temp file if exists
	err := os.Remove(path)
	if err != nil {
		log.Println("OnFileCanceled: failed to remove temp file:", err)
		return
	}
	// get name of file, aka key
	list := strings.Split(path, "/")
	i := len(list) - 1
	if i < 0 {
		i = 0
	}
	name := list[i]
	// remove from allowedTransfers
	c.mutex.Lock()
	delete(c.enc.allowedTransfers, name)
	c.mutex.Unlock()
}

/*
OnConnected is called when another peer comes online.
*/
func (c *chaninterface) OnConnected(address string) {
	// only notify log, nothing else to do for us here
	log.Println("OnConnected:", address[:8])
}
