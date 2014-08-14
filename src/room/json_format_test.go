package room

import (
	"encoding/json"
	"fmt"
	"testing"
	//    "fmt"
)

func TestJsonFormat(t *testing.T) {
	var sessionId int64 = 123
	var name string = "my name"
	var userId int64 = 1111

	user := User{
		UserId: userId,
		//        Name: name,
		Cookie: "",
		Room:   nil,
	}
	session := Session{
		Id:   sessionId,
		Name: name,
	}

	content := "I am Content"
	baseMsg := BaseFrameMsg{
		Type:    Msg_Type_Sys_Ok,
		Index:   sessionId,
		Content: content,
	}

	if string(`{"type":4096,"index":123,"content":"I am Content"}`) != jsonString(baseMsg) {
		t.Fatal("BaseFrameMsg Object convert JSON string format did not Correct.")
	}

	roomId := "Im Room Id"
	room := Room{
		Id:          roomId,
		msgIndex:    0,
		OnlineUsers: make(map[int64]*OnlineUser),
		Broadcast:   make(chan *Message, MAX_ROOM_CHAN_LEN),
	}

	jstring_tmplate_room := string(`{"room":"%s","users":[%s]}`)
	jstring_empty_room := fmt.Sprintf(jstring_tmplate_room, roomId, "")
	if jstring_empty_room != room.jsonString() {
		t.Fatal("Empty Room Object convert JSON string format did not Correct.")
	}

	onlineUser := OnlineUser{User: &user, Session: &session}
	//jstring_onlineuser := string(`{"name":"my name"}`)
	if string(`{"name":"my name","sessionId":123}`) != jsonString(onlineUser) {
		t.Fatal("OnlineUser-->JSON:", jsonString(onlineUser))
	}

	room.OnlineUsers[onlineUser.Session.Id] = &onlineUser
	jstring_room := fmt.Sprintf(jstring_tmplate_room, roomId, jsonString(onlineUser))
	if jstring_room != room.jsonString() {
		t.Fatal("Room Object convert JSON string format did not Correct.")
	}

	room.OnlineUsers[456] = &onlineUser
	// test mutil peoples
	room.jsonString()

	sayMsg := ChatBroadcastContent{
		SessionId: onlineUser.Id,
		Say:        "I say something.",
		Index:      0,
	}

	if string(`{"sessionId":123,"say":"I say something.","index":0}`) != jsonString(sayMsg) {
		t.Fatal("chatMsgContent Object convert JSON string format did not Correct.--> %v", jsonString(sayMsg))
	}

	msg := ChatBroadcastContent{}
	err := json.Unmarshal([]byte(jsonString(sayMsg)), &msg)
	if err != nil {
		t.Fatal(err.Error())
	}
	if msg.SessionId != sessionId {
		t.Fatal("json.Unmarshal chatMsgContent fail. sessionId != user.SessionId.")
	}
	
	strTxtUpdate := []byte(`{"index":1964155495,"type":1048577,"content":"[{\"start\":1,\"end\":10,\"version\":100,\"value\":\"AAAAAAAAAAAAAA\"}]"}`)
	var msgTxtUpdate Message
	var cmds []TextUpdateCmd
	if err:= json.Unmarshal(strTxtUpdate, &msgTxtUpdate); err != nil{
	    t.Fatal(err.Error())
	}
	if err:= json.Unmarshal([]byte(msgTxtUpdate.Content), &cmds); err != nil{
	    t.Fatal(err.Error())
	}
	var cmd = cmds[0]
	if cmd.Start != 1 || cmd.End != 10 || cmd.Version != 100||cmd.Value==""{
	    t.Fatal("TextUpdateCmd json.Unmarshal fail", cmd.Start, cmd.End, cmd.Version, cmd.Value)
	}
}
