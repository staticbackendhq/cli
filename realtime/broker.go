/*
Copyright © 2020 Focus Centric inc. <dominicstpierre@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package realtime

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const (
	SystemID = "sb"

	MsgTypeError     = "error"
	MsgTypeOk        = "ok"
	MsgTypeEcho      = "echo"
	MsgTypeInit      = "init"
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

type Command struct {
	SID     string `json:"sid"`
	Type    string `json:"type"`
	Data    string `json:"data"`
	Channel string `json:"channel"`
	Token   string `json:"token"`
}

type Validator func(interface{}) (map[string]interface{}, error)

type Broker struct {
	Broadcast          chan Command
	newConnections     chan chan Command
	closingConnections chan chan Command
	clients            map[chan Command]string
	ids                map[string]chan Command
	subscriptions      map[string][]string
	validateAuth       Validator
}

func NewBroker(v Validator) *Broker {
	b := &Broker{
		Broadcast:          make(chan Command, 1),
		newConnections:     make(chan chan Command),
		closingConnections: make(chan chan Command),
		clients:            make(map[chan Command]string),
		ids:                make(map[string]chan Command),
		subscriptions:      make(map[string][]string),
		validateAuth:       v,
	}

	go b.start()

	return b
}

func (b *Broker) start() {
	for {
		select {
		case c := <-b.newConnections:
			id, err := uuid.NewUUID()
			if err != nil {
				log.Println(err)
			}

			b.clients[c] = id.String()
			b.ids[id.String()] = c

			msg := Command{
				Type: MsgTypeInit,
				Data: id.String(),
			}

			c <- msg
		case c := <-b.closingConnections:
			b.unsub(c)
		case msg := <-b.Broadcast:
			clients, payload := b.getTargets(msg)
			for _, c := range clients {
				c <- payload
			}
		}
	}
}

func (b *Broker) unsub(c chan Command) {
	defer delete(b.clients, c)

	id, ok := b.clients[c]
	if !ok {
		fmt.Println("cannot find connection id")
	}

	delete(b.ids, id)
}

func (b *Broker) Accept(w http.ResponseWriter, r *http.Request) {
	// check if writer handles flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming is unsupported with your connection.", http.StatusBadRequest)
		return
	}

	// set headers related to event streaming
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	//w.Header().Set("Access-Control-Allow-Origin", "*")

	// each connection has their own message channel
	messages := make(chan Command)
	b.newConnections <- messages

	// make sure we'r removing this connection
	// when the handler completes.
	defer func() {
		b.closingConnections <- messages
	}()

	// handles the client-side disconnection
	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		b.closingConnections <- messages
	}()

	// broadcast messages
	for {
		// write Server Sent Event data
		msg := <-messages
		b, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("error converting to JSON", err)
			continue
		}
		fmt.Fprintf(w, "data: %s\n\n", b)

		// flush immediately.
		flusher.Flush()
	}
}

func (b *Broker) getTargets(msg Command) (sockets []chan Command, payload Command) {
	if msg.SID != SystemID {
		sender, ok := b.ids[msg.SID]
		if !ok {
			fmt.Println("cannot find sender socket", msg.SID)
			return
		}
		sockets = append(sockets, sender)
	}

	switch msg.Type {
	case MsgTypeEcho:
		payload = msg
		payload.Data = "echo: " + msg.Data
	case MsgTypeAuth:
		if _, err := b.validateAuth(msg.Data); err != nil {
			payload = Command{Type: MsgTypeError, Data: "invalid token"}
			return
		}

		payload = Command{Type: MsgTypeToken, Data: msg.Data}
	case MsgTypeJoin:
		members, ok := b.subscriptions[msg.Data]
		if !ok {
			members = make([]string, 0)
		}

		members = append(members, msg.SID)
		b.subscriptions[msg.Data] = members

		payload = Command{Type: MsgTypeJoined, Data: msg.Data}
	case MsgTypeChanIn:
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

		go b.Publish(msg, msg.Channel)

		payload = Command{Type: MsgTypeOk}
	default:
		payload.Type = MsgTypeError
		payload.Data = fmt.Sprintf(`%s command not found`, msg.Type)
	}

	return
}

// Publish sends a message to all socket in that channel
func (b *Broker) Publish(msg Command, channel string) {
	if msg.Type == MsgTypeChanIn {
		msg.Type = MsgTypeChanOut
	}

	members, ok := b.subscriptions[channel]
	if !ok {
		return
	}

	for _, sid := range members {
		c, ok := b.ids[sid]
		if !ok {
			continue
		}

		c <- msg
	}
}
