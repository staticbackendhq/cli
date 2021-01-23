package ws

import (
	"fmt"
	"strings"
)

type Validator func(interface{}) (map[string]interface{}, error)

type Hub struct {
	sockets       map[*Socket]string
	ids           map[string]*Socket
	channels      map[*Socket][]chan bool
	subscriptions map[string][]string
	broadcast     chan Command
	register      chan *Socket
	unregister    chan *Socket
	validateAuth  Validator
}

type Command struct {
	SID     string `json:"sid"`
	Type    string `json:"type"`
	Data    string `json:"data"`
	Channel string `json:"channel"`
	Token   string `json:"token"`
}

func NewHub(v Validator) *Hub {
	return &Hub{
		broadcast:     make(chan Command),
		register:      make(chan *Socket),
		unregister:    make(chan *Socket),
		sockets:       make(map[*Socket]string),
		ids:           make(map[string]*Socket),
		channels:      make(map[*Socket][]chan bool),
		subscriptions: make(map[string][]string),
		validateAuth:  v,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case sck := <-h.register:
			h.sockets[sck] = sck.id
			h.ids[sck.id] = sck

			cmd := Command{
				Type: "init",
				Data: sck.id,
			}
			sck.send <- cmd

		case sck := <-h.unregister:
			if _, ok := h.sockets[sck]; ok {
				h.unsub(sck)
				delete(h.sockets, sck)
				delete(h.ids, sck.id)
				delete(h.channels, sck)
				close(sck.send)
			}
		case msg := <-h.broadcast:
			sockets, p := h.getTargets(msg)
			for _, sck := range sockets {
				select {
				case sck.send <- p:
				default:
					fmt.Println("connection closed: broadcast, target socket closed")
					h.unsub(sck)
					close(sck.send)
					delete(h.ids, msg.SID)
					delete(h.sockets, sck)
					delete(h.channels, sck)
				}
			}
		}
	}
}

const (
	MsgTypeError     = "error"
	MsgTypeOk        = "ok"
	MsgTypeEcho      = "echo"
	MsgTypeAuth      = "auth"
	MsgTypeToken     = "token"
	MsgTypeJoin      = "join"
	MsgTypeJoined    = "joined"
	MsgTypeChanIn    = "chan_in"
	MsgTypeChanOut   = "chan_out"
	MsgTypeDBCreated = "db_created"
	MsgTypeDBUpdated = "db_updated"
	MsgTypeDBDeleted = "db_deleted"
)

func (h *Hub) getTargets(msg Command) (sockets []*Socket, payload Command) {
	fmt.Println("recv", msg)

	sender, ok := h.ids[msg.SID]
	if !ok {
		fmt.Println("cannot find sender socket")
		return
	}

	switch msg.Type {
	case MsgTypeEcho:
		sockets = append(sockets, sender)
		payload = msg
		payload.Data = "echo: " + msg.Data
	case MsgTypeAuth:
		sockets = append(sockets, sender)
		if _, err := h.validateAuth(msg.Data); err != nil {
			payload = Command{Type: MsgTypeError, Data: "invalid token"}
			return
		}

		payload = Command{Type: MsgTypeToken, Data: msg.Data}
	case MsgTypeJoin:
		subs, ok := h.channels[sender]
		if !ok {
			subs = make([]chan bool, 0)
		}

		closeSubChan := make(chan bool)
		subs = append(subs, closeSubChan)

		h.channels[sender] = subs

		members, ok := h.subscriptions[msg.Data]
		if !ok {
			members = make([]string, 0)
		}

		members = append(members, msg.SID)
		h.subscriptions[msg.Data] = members

		sockets = append(sockets, sender)
		payload = Command{Type: MsgTypeJoined, Data: msg.Data}
	case MsgTypeChanIn:
		sockets = append(sockets, sender)

		if len(msg.Channel) == 0 {
			payload = Command{Type: MsgTypeError, Data: "no channel was specified"}
			return
		} else if strings.HasPrefix(strings.ToLower(msg.Channel), "db-") {
			payload = Command{
				Type: MsgTypeError,
				Data: "you cannot write to database channel",
			}
			return
		}

		go h.Publish(msg, msg.Channel)

		payload = Command{Type: MsgTypeOk}
	default:
		sockets = append(sockets, sender)

		payload.Type = MsgTypeError
		payload.Data = fmt.Sprintf(`%s command not found`, msg.Type)
	}

	return
}

func (h *Hub) unsub(sck *Socket) {
	id, ok := h.sockets[sck]
	if !ok {
		return
	}

	// remove subscriptions, since we're on a local dev it's ok
	// to use this kind of loop.
	for channel, members := range h.subscriptions {
		for idx, sid := range members {
			if id == sid {
				members = append(members[:idx], members[idx+1:]...)
				h.subscriptions[channel] = members
				break
			}
		}
	}

	subs, ok := h.channels[sck]
	if !ok {
		return
	}

	for _, sub := range subs {
		sub <- true
		close(sub)
	}
}

func (msg Command) IsDBEvent() bool {
	switch msg.Type {
	case MsgTypeDBCreated, MsgTypeDBUpdated, MsgTypeDBDeleted:
		return true
	}
	return false
}

// Publish sends a message to all socket in that channel
func (h *Hub) Publish(msg Command, channel string) {
	msg.Type = MsgTypeChanOut

	members, ok := h.subscriptions[channel]
	if !ok {
		return
	}

	for _, sid := range members {
		sck, ok := h.ids[sid]
		if !ok {
			continue
		}

		select {
		case sck.send <- msg:
		default:
			fmt.Println("again that shit, socket is closed")
			h.unsub(sck)
			close(sck.send)
			delete(h.ids, msg.SID)
			delete(h.sockets, sck)
			delete(h.channels, sck)
		}
	}
}
