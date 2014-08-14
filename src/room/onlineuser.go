package room

import (
	"encoding/json"
	"io"
	"libs/log"
	"libs/websocket"
)

const (
	MAX_USER_CHAN_LEN = 64
)

type User struct {
	UserId int64 `json:"-"`
	//    Name    string              `json:"name"`
	Cookie string `json:"-"`
	Room   *Room  `json:"-"`
}

type Session struct {
	Name      string          `json:"name"`
	Id        int64           `json:"sessionId"`
	WebSocket *websocket.Conn `json:"-"`
	Inbox     chan *Message    `json:"-"`
}

type OnlineUser struct {
	*User
	*Session
}

//广播聊天
type ChatBroadcastContent struct {
	SessionId int64  `json:"sessionId"`
	Say       string `json:"say"`
	Index     int64  `json:"index"`
}

type RenameBroadcastContent struct {
	SessionId int64  `json:"sessionId"`
	NewName   string `json:"name"`
	Color     string `json:"color"`
}

//广播的文档更新命令消息
type TextUpdateCmdsBroadcastContent struct{
    SessionId int64   		`json:"sessionId"`
    Stat	MarkDownStatus	`json:"status"`
    Update	[]TextUpdateCmd	`json:"Update"`
}

func (this *OnlineUser) smapleMsg(typ int64, content string) *Message {
	msg := &Message{
		BaseFrameMsg: BaseFrameMsg{
			Type:    typ,
			Index:   this.Room.GenMsgIndex(),
			Content: content,
		},
		Sender: this,
		SendTo: nil,
	}
	return msg
}

func (this *OnlineUser) msgJoin() *Message {
	return this.smapleMsg(Msg_Type_User_Join, jsonString(this))
}

func (this *OnlineUser) msgChat(srcMsg *Message) *Message {
	sayContent := jsonString(ChatBroadcastContent{
		SessionId: this.Id,
		Say:       srcMsg.Content,
		Index:     srcMsg.Index,
	})
	return this.smapleMsg(Msg_Type_User_Chat, sayContent)
}

func (this *OnlineUser) renameMsg(renameBroadcast RenameBroadcastContent) *Message {
	return this.smapleMsg(Msg_Type_User_Name, jsonString(renameBroadcast))
}

func (this *OnlineUser) updateTextMsg(stat MarkDownStatus, cmds []TextUpdateCmd) *Message {
	var msg = TextUpdateCmdsBroadcastContent{
		SessionId: this.Id,
		Update:    cmds,
		Stat:      stat,
	}
	content, _ := json.Marshal(msg)
	return this.smapleMsg(Msg_Type_Text_Update, string(content))
}

func (this *OnlineUser) saveTextMsg() *Message {
	return this.smapleMsg(Msg_Type_Text_Save, "")
}

func (user *OnlineUser) respMsgOk(msg *Message, content string) {
	//收到用户Chat消息后，返回一个OK，保持index一致
	msg.Type = Msg_Type_Sys_Ok
	msg.Content = content
	user.Inbox <- msg
}

func (this *OnlineUser) respMsgError(msg *Message, content string) {
	//收到用户Chat消息后，返回一个OK，保持index一致，Content为空
	msg.Type = Msg_Type_Sys_Err
	msg.Content = content
	this.Inbox <- msg
}

func (this *OnlineUser) handleUserChat(msg *Message) {
	chatMsg := this.msgChat(msg)
	this.Room.Broadcast <- chatMsg
	this.respMsgOk(msg, "")
}
func (this *OnlineUser) handleUserRename(msg *Message) {
	var renameContent RenameBroadcastContent
	if err := json.Unmarshal([]byte(msg.Content), &renameContent); err != nil {
		log.Error("Msg_Type_User_Name:%v", err.Error())
		this.respMsgError(msg, err.Error())
		return
	}
	renameContent.SessionId = this.Id
	this.Name = renameContent.NewName
	this.Room.Broadcast <- this.renameMsg(renameContent)
	this.respMsgOk(msg, "")
}

func (this *OnlineUser) handleTextUpdate(msg *Message) {
	stat, cmds, err := this.Room.UpdateText(this, msg)
	if err != nil {
		log.Error("User[%v] Text_Update err=%v", this.Name, err.Error())
		this.respMsgError(msg, err.Error())
	} else {
		this.respMsgOk(msg, jsonString(stat))
		this.Room.Broadcast <- this.updateTextMsg(stat, cmds)
	}
}

func (this *OnlineUser) handleTextSave(msg *Message) {
	err := this.Room.SaveMdText()
	if err != nil {
		log.Error("User[%v] Text_Save err=%v", this.Name, err.Error())
		this.respMsgError(msg, err.Error())
	} else {
		this.respMsgOk(msg, "")
		this.Room.Broadcast <- this.saveTextMsg()
	}
}

func (this *OnlineUser) handleTextGetAll(msg *Message) {
	resMsg := this.Room.GetMarkDownText()
	msg.Type = Msg_Type_Sys_Ok
	msg.Content = jsonString(resMsg)
	this.Inbox <- msg
}

func (this *OnlineUser) handleWebsocket(msg *Message) {

	if msg.Type == Msg_Type_Text_Update {
		this.handleTextUpdate(msg)
	} else if msg.Type == Msg_Type_Text_Save {
		this.handleTextSave(msg)
	} else if msg.Type == Msg_Type_User_Chat {
		this.handleUserChat(msg)
	} else if msg.Type == Msg_Type_Text_GetAll{
		this.handleTextGetAll(msg)
	} else if  msg.Type == Msg_Type_Text_Init {
	    log.Info("TODO: Init the Text and broadcast to the room.")
	} else if msg.Type == Msg_Type_User_Name {
		this.handleUserRename(msg)
	}else if Msg_Type_Sys == (Msg_Type_Sys & msg.Type) {
		this.Inbox <- msg
	} else {
		this.Room.Broadcast <- msg
		this.respMsgOk(msg, "")
	}
}

func (this *OnlineUser) WaitForFrame() {
	log.Info("User[%v] Wait For Frame", this.Name)
	for {
		frameType, bFrame, err := this.WebSocket.Read()
		log.Debug("User[%v] recv WebSocket Frame typ=[%v], val=[%s]", this.Name, frameType, bFrame)

		if err != nil {
			if err != io.ErrUnexpectedEOF {
				log.Error("User[%v] close Unexpected err=%v", this.Name, err.Error())
			} else {
				log.Debug("User[%v] close the socket. EOF.", this.Name)
			}

			this.quit()
			return
		}
		if frameType == websocket.CloseMessage {
			log.Info("User=[%v] close Frame revced. end wait Frame loop", this.Name)
			this.quit()
			return

		} else if frameType == websocket.PingMessage {
			this.WebSocket.Pong([]byte{})
			continue
		} else if frameType == websocket.TextMessage {
			msg := Message{}
			if err := json.Unmarshal(bFrame, &msg); err != nil {
				log.Error("Recve a Text-Frame not JSON format. err=%v, frame=%v", err.Error(), string(bFrame))
				continue
			}
			log.Info("User[%v] Frame -> T=%v, Content=%v, index=%v, ", this.Name, msg.TypeS(), msg.Content, msg.Index)
			msg.Sender = this
			//TODO msg.SendTo

			this.handleWebsocket(&msg)

			continue
		} else {
			log.Warn("TODO: revce frame-type=%v. can not handler.", frameType)
			continue
		}
	}
}

func (this *OnlineUser) WaitForRoomMsg() {
	// 发送消息到用户websocket
	log.Info("User[%v] Wait For Room Msg", this.Name)
	for {
		select {
		case msg := <-this.Inbox:
			log.Info("User[%v] get msg[%v] from Inbox, content=%v, index=%v", this.Name, msg.TypeS(), msg.Content, msg.Index)

			switch msg.Type {
			case Msg_Type_User_Quit:
				if msg.Sender == this {
					log.Info("User[%v] quit Room[%v]. end loop[WaitForRoomMsg]", this.Name, this.Room.Id)
					return
				} else {
					log.Info("User[%v], Other quit Room[%v].", this.Name, this.Room.Id)
					this.WebSocket.Write([]byte(jsonString(msg)), false)
				}
			default:
				log.Debug("User[%v] write frame to socket: json->%v", this.Name, jsonString(msg))
				this.WebSocket.Write([]byte(jsonString(msg)), false)
			}
		}
	}
}

func (this *OnlineUser) quit() {
	this.WebSocket.Close()
	this.Room.RW.Lock()
	defer this.Room.RW.Unlock()
	delete(this.Room.OnlineUsers, this.Session.Id)

	msgQuit := &Message{
		BaseFrameMsg: BaseFrameMsg{
			Index:   this.Room.GenMsgIndex(),
			Type:    Msg_Type_User_Quit,
			Content: jsonString(this), //fmt.Sprintf("User[%v] Quit.", this.Name),
		},
		Sender: this,
		SendTo: nil,
	}
	// 因user已从OnlineUsers中delete掉，所以不会广播到他自己的Inbox
	this.Room.Broadcast <- msgQuit
	this.Inbox <- msgQuit 
	log.Info("User=[%v] Quit Room[%v],send quit Msg to peoples=[%v].", this.Name, this.Room.Id, len(this.Room.OnlineUsers))
}
