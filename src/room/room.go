// online websocket room
package room

import (
	"encoding/json"
	"errors"
	"fmt"
	"libs/log"
	"libs/websocket"
	"math/rand"
	"strings"
	"sync"
)

const (
	MAX_ROOM_CHAN_LEN = 256
)

var Saver IMarkDownSaver = new(FileSaver)
type IMarkDownSaver interface {
	Save(id string, md MarkDownText) (err error)
	Load(id string) (md *MarkDownText, err error)
}

type AllRooms struct {
	RW    *sync.RWMutex
	Rooms map[string]*Room
}

type Room struct {
	Id          string                `json:"room"`
	OnlineUsers map[int64]*OnlineUser `json:"users"`
	RW          *sync.RWMutex         `json:"-"`
	Broadcast   chan *Message         `json:"-"`
	msgIndex    int64                 `json:"-"`
	mdText      *MarkDownText         `json:"-"`
}

type Message struct {
	BaseFrameMsg
	Sender *OnlineUser           `json:"-"`
	SendTo map[int64]*OnlineUser `json:"-"` // if nil { send to the room.}
}

var allRooms = &AllRooms{
	RW:    new(sync.RWMutex),
	Rooms: make(map[string]*Room),
}

var (
	MsgOk    = Message{BaseFrameMsg: BaseFrameMsg{Type: Msg_Type_Sys_Ok}}
	MsgError = Message{BaseFrameMsg: BaseFrameMsg{Type: Msg_Type_Sys_Err}}
)

type UserStatus struct {
    *Session
    MsgLen int		`json:"MsgLen"`
}

type RoomStatus struct {
	Id        string	`json:"id"`
	Peoples    int		`json:"peoples"`
	MsgLen    int		`json:"broadcast"`
	MarkDownStatus
	Text		string	`json:"text"`
	Users []*UserStatus	`json:"users"`
}

func GetRoomStatus(debug bool) (rooms []*RoomStatus) {
	allRooms.RW.RLock()
	defer allRooms.RW.RUnlock()

	for k, room := range allRooms.Rooms {
	    room.RW.Lock()
		nLen := len(room.mdText.Content)
		if nLen >20 {
		    nLen = 20
		}
		status := &RoomStatus{
			Id:        k,
			Peoples:   len(room.OnlineUsers),
			MsgLen:    len(room.Broadcast),
			MarkDownStatus:MarkDownStatus{
			    Version:room.mdText.Version,
			    Md5:	room.mdText.Md5,
			    Time:	room.mdText.Time,
			},
			Text:	room.mdText.Content[:nLen],
			Users: make([]*UserStatus, 0),
		}
		rooms = append(rooms, status)
		if debug {
			for _, user := range room.OnlineUsers {
			    var userStat = &UserStatus{
			        Session: user.Session,
			        MsgLen: len(user.Inbox),
			    }
			    status.Users = append(status.Users, userStat)
			}
		}
		room.RW.Unlock()
	}
	return rooms
}

func (room *Room) SaveMdText() error {
	room.RW.Lock()
	defer room.RW.Unlock()
	return Saver.Save(room.Id, *room.mdText)
}

func (room *Room) UpdateText(onlineUser *OnlineUser, msg *Message) (stat MarkDownStatus, cmds []TextUpdateCmd, err error) {
	if msg.Type&Msg_Type_Text != Msg_Type_Text {
		return stat, nil, errors.New(fmt.Sprintf("Msg Type[%v] not allow in func Room.UpdateText.", msg.TypeS()))
	}

	err = json.Unmarshal([]byte(msg.Content), &cmds)
	if err != nil {
		log.Error("convert Update-Msg-Content err=[%v], Content=[%v]", err.Error(), msg.Content)
		return stat, nil, err
	}

	room.RW.Lock()
	defer room.RW.Unlock()
	for idx, cmd := range cmds {
		cmd.Version += int32(idx)
		stat, err = room.mdText.UpdateTextCmd(cmd)
		if err != nil {
			break
		}
	}
	return stat, cmds, err
}

func (room *Room) GetMarkDownText() (all MarkDownText) {
    room.RW.RLock()
	defer room.RW.RUnlock()
	return room.mdText.GetMarkDownText()
}

func (this *Room) GenMsgIndex() int64 {
	this.msgIndex++
	return this.msgIndex
}

func (this *Room) jsonString() string {
	arrString := make([]string, 0) //, len(this.OnlineUsers))

	for _, onUser := range this.OnlineUsers {
		arrString = append(arrString, jsonString(onUser))
	}
	var users string
	if len(arrString) == 1 {
		users = arrString[0]
	} else {
		users = strings.Join(arrString, ",")
	}
	res := fmt.Sprintf("{\"room\":\"%s\",\"users\":[%s]}", this.Id, users)

	return res
}

func smapleMsg(typ int64, content string) *Message {
	msg := &Message{
		BaseFrameMsg: BaseFrameMsg{
			Type:    typ,
			Index:   0,
			Content: content,
		},
		Sender: nil,
		SendTo: nil,
	}
	return msg
}
func (room *Room) msgSysConn(onlineUser *OnlineUser) *Message {
	return smapleMsg(Msg_Type_Sys_Conn, jsonString(onlineUser))
}

func (this *Room) msgTextInit() *Message {
	return smapleMsg(Msg_Type_Text_Init, jsonString(this.mdText.GetMarkDownText()))
}

func (this *Room) msgListUsers() *Message {
	return smapleMsg(Msg_Type_User_List, this.jsonString())
}

func findRoom(id string) (room *Room) {
	allRooms.RW.Lock()
	defer allRooms.RW.Unlock()
	room, ok := allRooms.Rooms[id]

	if !ok {
		room = &Room{
			Id:          id,
			msgIndex:    0,
			OnlineUsers: make(map[int64]*OnlineUser),
			RW:          new(sync.RWMutex),
			Broadcast:   make(chan *Message, MAX_ROOM_CHAN_LEN),
			mdText:      &MarkDownText{},
		}
		room.mdText.Init("")
		md, err := Saver.Load(id)
		if err == nil && md != nil {
			log.Info("Room[%s] load mdText from HD, md.Version=%v", id, md.Version)
			room.mdText = md
		}

		allRooms.Rooms[id] = room
		go room.run()
		log.Info("Room[%v] Created.", id)
	}

	return room
}

func buildSession(room *Room, conn *websocket.Conn) (session Session) {

	for {
		var rid = rand.Int63()
		_, find := room.OnlineUsers[rid]
		if !find {
			session = Session{
				Id:        rid,
				WebSocket: conn,
				Inbox:     make(chan *Message, MAX_USER_CHAN_LEN),
			}
			break
		}
	}
	return session
}

//---------------------------------------
func JoinRoom(id string, cookie string, conn *websocket.Conn) (*Room, *OnlineUser) {
	log.Info("User want go Join Room[%s].", id)

	var room *Room = findRoom(id)
	var findUser bool = false

	room.RW.Lock()
	defer room.RW.Unlock()

	var session = buildSession(room, conn)
	var user User

	var onlineUser = OnlineUser{
		User:    &user,
		Session: &session,
	}

	for _, oline := range room.OnlineUsers {
		if cookie == oline.Cookie {
			findUser = true
			// 指向同一个user对像
			onlineUser.User = oline.User
			onlineUser.UserId = oline.UserId + 1
		}
	}

	if !findUser {
		onlineUser.Room = room
		onlineUser.Cookie = cookie
		onlineUser.UserId = int64(len(room.OnlineUsers)) + 1

	}
	// TODO: get the Name and userId from the cookie.
	onlineUser.Name = fmt.Sprintf("U-%d", onlineUser.UserId)

	// ADD user to the Room
	room.OnlineUsers[session.Id] = &onlineUser
	room.Broadcast <- onlineUser.msgJoin()

	onlineUser.Session.Inbox <- room.msgSysConn(&onlineUser)
	onlineUser.Session.Inbox <- room.msgListUsers()

	log.Info("User[%v] join Room[%v], peoples=%v", onlineUser.Name, id, len(room.OnlineUsers))
	return room, &onlineUser
}

func (this *Room) isNobody() bool {
	return len(this.OnlineUsers) == 0
}

func (this *Room) msgHandler(msg *Message) (closeMySelf bool) {
	this.RW.Lock()
	defer this.RW.Unlock()
	for _, user := range this.OnlineUsers {
		if user == msg.Sender {
			log.Debug("Sender is user.")
			// TODO: filter msg.
		} else {
			if len(user.Inbox) == MAX_USER_CHAN_LEN {
				log.Error("User[%s] chan is full.", user.Name)
				continue
			}
			user.Inbox <- msg
		}
	}

	return this.isNobody()
}

func (this *Room) run() {
	for {
		select {
		case msg := <-this.Broadcast:
			log.Info("Room[%v]->Broadcast msg T[%v] to [%v] peoples.", this.Id, msg.TypeS(), len(this.OnlineUsers))

			closeMySelf := this.msgHandler(msg)

			if closeMySelf {
				this.close()
				return
			}
		}
	}
}

func (this *Room) close() {
	allRooms.RW.Lock()
	defer allRooms.RW.Unlock()
	delete(allRooms.Rooms, this.Id)
	log.Info("Room[%s] closed.", this.Id)
}
