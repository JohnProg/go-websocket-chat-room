package room

import (
    "io"
    "encoding/json"
    "libs/log"
    "libs/websocket"
)


type OnlineUser struct {
    User
    //SessionId  int64 // websocket connection session ID
    Room       *Room
    WebSocket  *websocket.Conn
    Inbox      chan Message
}

func (this *OnlineUser) msgJoin() Message {
    msg := Message{
        BaseFrameMsg:BaseFrameMsg{
            Type: Msg_Type_User_Join,
            Index: this.Room.genMsgIndex(),
            Content: this.jsonString(),
        },
        Sender: this,
        SendTo: nil,
    }
    return msg
}

func (this *OnlineUser) WaitForRoomMsg(){
    _info("User[%v] Wait For Room Msg", this.SessionId)
    for {
        select {
        case msg := <- this.Inbox:
            _info("User[%v] get msg[%v] from Inbox, content=%s", this.SessionId, msg.TypeS(), msg.Content)

            switch msg.Type {
            case Msg_Type_Sys_Err:
                _info("User[%v] get msg[Sys_Err], end the User Loop->WaitForRoomMsg.",  this.SessionId)
                this.quit()
                return
            case Msg_Type_User_Quit:
                _info("User[%v] quit Room[%v].", this.SessionId, this.Room.Id)
                return
            case Msg_Type_User_Chat:
                this.WebSocket.Write([]byte(msg.jsonString()), false)
            default:
                _info("WaitForRoomMsg_case_default: json->%v", msg.jsonString())
                this.WebSocket.Write([]byte(msg.jsonString()), false)
//                return
            }
        }
    }
}

func (this *OnlineUser) clientMessageHandler(msg Message){
    // TODO: 处理广播与非广播消息
    if Msg_Type_Sys == Msg_Type_Sys & msg.Type {
        this.Inbox <- msg
    } else {
        this.Room.Broadcast <- msg
    }
}

func (this *OnlineUser) WaitForFrame(){
    _info("User[%v] Wait For Frame", this.SessionId)
    for{
        frameType, bFrame, err := this.WebSocket.Read()
        log.Debug("Handle WebSocket Frame typ=[%v], val=[%s]", frameType, bFrame)

        if err != nil {
            if err != io.ErrUnexpectedEOF {
                log.Error(err.Error())
            } else {
                log.Debug("Client close the socket. EOF.")
                this.quit()
                return 
            }
            this.Inbox <- MsgError
            return
        }
        if frameType == websocket.CloseMessage {
            _info("User=[%v] Close Frame revced. end wait Frame loop", this.SessionId)
            this.quit()
            return

        } else if frameType == websocket.PingMessage {
            this.WebSocket.Pong([]byte(`"pong"`))
            continue 
        } else if frameType == websocket.TextMessage {
            msg := Message{}
            if err := json.Unmarshal(bFrame, &msg); err != nil {
                log.Error("Recve a Frame not JSON format. err=%v, b=%v", err.Error(), bFrame)
                continue
            }
            log.Info("Text Frame -> T=%v, Content=%v, index=%v, ", msg.TypeS(), msg.Content, msg.Index)
            msg.Index = this.Room.genMsgIndex()
            msg.Sender = this
            //TODO msg.SendTo

            this.clientMessageHandler(msg)

            continue
        } else {
            log.Warn("TODO: revce frame-type=%v. can not handler.", frameType)
            continue
        }
    }
}
func (this *OnlineUser) quit() {
    this.WebSocket.Close()
    delete(this.Room.OnlineUsers, this.SessionId)

    _info("User=[%v] Quit Room[%v],send Quit Msg to pepole=[%v].", this.SessionId, this.Room.Id, len(this.Room.OnlineUsers))

    msgQuit := Message{
        BaseFrameMsg:BaseFrameMsg{
            Index: this.Room.genMsgIndex(),
            Type: Msg_Type_User_Quit,
            Content: "I Quit.",
        },
        Sender: this,
        SendTo: nil,
    }
    // 因user已从OnlineUsers中delete掉，所以不会广播到他自己的Inbox
    this.Room.Broadcast <- msgQuit
    this.Inbox <- msgQuit
}
