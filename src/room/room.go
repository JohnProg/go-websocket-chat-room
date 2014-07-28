// online websocket room
package room

import (
    "fmt"
    "strings"
//    "encoding/json"
    "libs/log"
    //"libs/websocket"
)

var _info = log.Info
//var _err = log.Error

var allRooms = make(map[string]*Room)

type Message struct {
    BaseFrameMsg
    Sender  *OnlineUser
    SendTo  map[int64] *OnlineUser // if nil { send to the room.}
}
var (
    baseMsgOk   = BaseFrameMsg{Type: Msg_Type_Sys_Ok}
    baseMsgError  = BaseFrameMsg{Type: Msg_Type_Sys_Err}
//    basemsgquit = baseframemsg{type: msg_type_user_quit}
)

var (
    MsgOk = Message {BaseFrameMsg: baseMsgOk}
    MsgError = Message {BaseFrameMsg: baseMsgError}
)

type Room struct {
    Id          string                  `json:"room"`
    OnlineUsers map[int64] *OnlineUser  `json:"users"`
    Broadcast   chan Message
    CloseSign   chan bool  // when close this room will read an signel
    msgIndex    int64
}

func (this *Room) genMsgIndex() int64 {
    this.msgIndex++
    log.Info("Room[%v]---genRoom.Index=%v", this.Id, this.msgIndex)
    return this.msgIndex
}

func (this *Room) jsonString() string {
    arrString := make([]string, 0) //, len(this.OnlineUsers))

    for _, user := range this.OnlineUsers {
        arrString = append(arrString, user.jsonString())
    }
    var users string
    if len(arrString) <= 1 {
        users = arrString[0] 
    }else{
        users = strings.Join(arrString,",")
    }
    res := fmt.Sprintf("{\"room\":\"%s\",\"users\":[%s]}", this.Id, users)

    _info("RoomJsonString: %v", res)
    return res
}

func (this *Room) msgListUsers() Message {
    msg := Message{
        BaseFrameMsg:BaseFrameMsg{
            Type: Msg_Type_User_List,
            Index: this.genMsgIndex(),
            Content: this.jsonString(),
        },
        Sender: nil,
        SendTo: nil,
    }
    return msg
}
//---------------------------------------
func JoinRoom(id string, user *OnlineUser) (*Room, /*isCreated bool*/){
    room, ok := allRooms[id]

    if !ok {
        allRooms[id] = &Room{
            Id: id,
            msgIndex: 0,
            OnlineUsers: make(map[int64] *OnlineUser),
            Broadcast:   make(chan Message, 256),
            CloseSign:   make(chan bool),
        }
        room, _ = allRooms[id]
        go room.run()
        _info("Create Room=%v", id)
    }

    room.OnlineUsers[user.SessionId] = user
    user.Room = room

    user.Inbox <- room.msgListUsers()
    room.Broadcast <- user.msgJoin() 

    _info("User[%v] join room=[%v], pepole=%v", user.SessionId, id, len(room.OnlineUsers))
    return room
}
func (this *Room) isNobody() bool {
    return len(this.OnlineUsers) == 0
}

func (this *Room) msgHandler(msg Message) {
    mType := msg.Type

    for _, user := range this.OnlineUsers {
        if user == msg.Sender {
            // TODO: filter msg.
            user.Inbox <- msg
        } else {
            user.Inbox <- msg
        }
    }

    if mType == Msg_Type_User_Quit && this.isNobody() {
        this.CloseSign <- true
    }
}

func (this *Room) run(){
    for {
        select {
        case msg := <-this.Broadcast:
            _info("Room[%v]->Broadcast msg T[%v] to [%v] pepole.", this.Id, msg.TypeS(), len(this.OnlineUsers))
            for _, user := range this.OnlineUsers {
                if user == msg.Sender {
                    user.Inbox <- Message{
                        BaseFrameMsg: BaseFrameMsg{
                            Index: msg.Index,
                            Type: Msg_Type_Sys_Ok,
                            Content:"OK",
                        },
                        Sender: nil,
                        SendTo: nil,
                    }
                }else{
                    user.Inbox <- msg
                }
            }

        case sign := <-this.CloseSign:
            if sign == true {
                delete(allRooms, this.Id)
                close(this.Broadcast)
                close(this.CloseSign)
                _info("Room[%s] closed", this.Id)
                return
            }
        }
    }
}

