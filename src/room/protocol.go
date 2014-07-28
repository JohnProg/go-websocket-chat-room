package room
import (
    "libs/log"
    "encoding/json"
)
// 消息格式定义参看 src/www/doc/protocol.md
const (
    Msg_Type_Sys        = 0x00001000
    Msg_Type_Sys_Ok     = 0x00001000    // 4096
    Msg_Type_Sys_Err    = 0x00001001
    Msg_Type_Sys_Conn   = 0x00001002    //当连接建立后，客户端收到的第一个消息类型。

    Msg_Type_User       = 0x00010000
    Msg_Type_User_Join  = 0x00010000    // 65536 
    Msg_Type_User_Quit  = 0x00010001
    Msg_Type_User_Chat  = 0x00010002
    Msg_Type_User_Name  = 0x00010003    // user set him nickname, will broadcast to the room
    Msg_Type_User_List  = 0x00010004    // list All user in the room.

    Msg_Type_Text        = 0x00100000
    Msg_Type_Text_Init   = 0x00100000   // 1048576
    Msg_Type_Text_Update = 0x00100001
    Msg_Type_Text_CurPos = 0x00100002
    Msg_Type_Text_GetAll = 0x00100003
)

var mapMString map[int64] string
func init(){
    mapMString = make(map[int64] string, 256)
    mapMString[Msg_Type_Sys_Ok]     = "OK"
    mapMString[Msg_Type_Sys_Err]    = "Err"
    mapMString[Msg_Type_Sys_Conn]   = "Conn"
    mapMString[Msg_Type_User_Join]  = "Join"
    mapMString[Msg_Type_User_Quit]  = "Quit"
    mapMString[Msg_Type_User_Chat]  = "Chat"
    mapMString[Msg_Type_User_Name]  = "Name"
    mapMString[Msg_Type_User_List]  = "List"
    mapMString[Msg_Type_Text_Init]  = "Init"
    mapMString[Msg_Type_Text_Init]  = "Update"
    mapMString[Msg_Type_Text_Init]  = "CurPos"
    mapMString[Msg_Type_Text_Init]  = "All"
}
//Init()
func (this *BaseFrameMsg) TypeS() string {
    s, ok := mapMString[this.Type]
    if !ok {
        log.Warn("not found msg type[%v]", this.Type)
    }
    return s
}

type User struct {
    SessionId int64 `json:"sessionId"`
    Name string     `json:"Name"`
}

type BaseFrameMsg struct {
    Type    int64 `json:"type"`
    Index   int64 `json:"index"`
    // Content可转成javascript对像的JSON串, 会因type不同而转出来的结构不同,大部分消息应该该包含"sessionId":<int64>
    Content string `json:"content"`
}

func (this *User) jsonString() string {
    bytes, err := json.Marshal(this)
    if err != nil {
        log.Error(err.Error())
        return ""
    }
    return string(bytes)
}

func (this *BaseFrameMsg) jsonString() string {
    bytes, err := json.Marshal(this)
    if err != nil {
        log.Error(err.Error())
        return ""
    }
    return string(bytes)
}

//func makeMsg(msgType int, interface {}) *BaseFrameMsg {
//    switch msgType{
//    
//    }
//
//
//}


