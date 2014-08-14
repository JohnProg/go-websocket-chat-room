// gomm-online main

package room

import (
	"encoding/json"
	"fmt"
	"libs/log"
	"libs/websocket"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	HOST_PORT = "127.0.0.1:11187"
)

func testUpdateTextMsg(t *testing.T, ws1, ws2 *websocket.Conn) {

	log.Info("SendGetAllText--> %v", SendGetAllText(t, ws1))
	var getAllMsg = ReadMsg(t, ws1)
	log.Info("Msg_Type_Text_GetAll resp-->%v", jsonString(getAllMsg))
	
	var mdText MarkDownText 
	json.Unmarshal([]byte(getAllMsg.Content), &mdText)
	
	var ver int32 = mdText.Version
	var msgIndex int64 = 123333

	txtUpdateCmd := TextUpdateCmd{
		Version: ver,
		Start:   0,
		End:     0,
		Value:   "A",
	}
	content := fmt.Sprintf("[%v]", jsonString(txtUpdateCmd))
	log.Info("update content=%s", content)
	log.Info("send Msg_Type_Text_Update req:%v", SendUpdateTextMsg(t, ws1, content, msgIndex))
	msg := ReadMsg(t, ws1)
	if msg.Index != msgIndex || msg.Type != Msg_Type_Sys_Ok || msg.Content == ""{
		t.Fatal("msg not ok, Type=", msg.TypeS(), " Content=",msg.Content, "index=", msg.Index)
	}
	log.Info("Msg_Type_Text_Update get resp:%s", jsonString(msg))

	var broadcastMsgUpdateText = ReadMsg(t, ws2)
	if broadcastMsgUpdateText.Type != Msg_Type_Text_Update {
		t.Fatal("User 1 has send Msg_Type_Text_Update msg")
	}
	log.Info("ws2 get broadcast Msg_Type_Text_Update Msg-->%v", jsonString(broadcastMsgUpdateText))

	curpos := "{cursor:0}"
	log.Info("w2 SendCurPosMsg->%v", SendCurPosMsg(t, ws2, curpos))

	var posMsgResp = ReadMsg(t, ws2)
	log.Info("ws2 get broadcast SendCurPosMsg resp---> %v", jsonString(posMsgResp))
	log.Info("ws1 get broadcast SendCurPosMsg resp---> %v", jsonString(ReadMsg(t, ws1)))

	reqSave := SandSaveTextMsg(t, ws1)
	log.Info("wss1 set request SaveTextMsg-->%v", reqSave)
	respSave := ReadMsg(t, ws1)
	broadcastSaveMsg := ReadMsg(t, ws2)

	if respSave.Type != Msg_Type_Sys_Ok {
		t.Fatal("Msg_Type_Text_Save fail", respSave.Content)
	}
	if broadcastSaveMsg.Type != Msg_Type_Text_Save {
		t.Fatal("Msg_Type_Text_Save braodcast fail", broadcastSaveMsg.Type)
	}

	log.Info("end test UpdateTextMsg")
}

func TestWSServer(t *testing.T) {
	go server_main_loop()

	time.Sleep(time.Millisecond * 1)
	room := "testing"

	ws1, user1 := CreateConnection(t, room, 1)
	ws2, user2 := CreateConnection(t, room, 1)
	
	if user1.Id == user2.Id{
	    t.Fatal("session id equal.")
	}
	
	log.Debug("user1 wait for user2 join message.")

	user2JoinMsg := ReadMsg(t, ws1)
	if user2JoinMsg.Type != Msg_Type_User_Join {
		t.Fatal("User Join ? get[%v]", user2JoinMsg.Type)
	}
	if !strings.Contains(user2JoinMsg.Content, strconv.FormatInt(user2.Session.Id, 10)) {
		t.Fatal("user1 must recv user2-join msg.")
	}

	user2SayContent := "user say something."
	var indexSay int64 = 13993402
	log.Info("send Msg_Type_User_Chat :%s", SandChatMsg(t, ws2, user2SayContent, indexSay))
	
	chatBroadcastMsg := ReadMsg(t, ws1)
	var chatContent ChatBroadcastContent
	if err := json.Unmarshal([]byte(chatBroadcastMsg.Content), &chatContent); err != nil {
	    t.Fatal("Msg_Type_User_Chat broadcast fail")
	}
	
	log.Info("Msg_Type_User_Chat resp:%s", jsonString(chatBroadcastMsg))
	
	if chatContent.Say != user2SayContent || chatContent.SessionId != user2.Id{
		t.Fatal("user1 shell recv user2-say-content")
	}

	sayOkRespMsg := ReadMsg(t, ws2)
	if sayOkRespMsg.Type != Msg_Type_Sys_Ok || sayOkRespMsg.Index != indexSay {
		t.Fatal("send say will recv an OK-Msg")
	}

	// test 100 user Join and say
	max := 4
	arr := []*websocket.Conn{ws1, ws2}
	for i := 0; i < max; i++ {
		ws, _ := CreateConnection(t, room, max+i)
		for _, v := range arr {
			if v != nil && ReadMsg(t, v).Type != Msg_Type_User_Join {
				t.Fatal("did not broadcast Join msg.")
			}
		}
		arr = append(arr, ws)
	}

	SandChatMsg(t, ws1, user2SayContent, indexSay)
	for idx, v := range arr[1:max] {
		msg := ReadMsg(t, v)
		if msg.Type != Msg_Type_User_Chat {
			t.Fatal("sat to [%v] peoples fail", max, idx, msg.Type)
		}
	}

	for i, ws := range arr {
		log.Info("User-%d quit.", i)
		ws.Close()
	}

	time.Sleep(time.Millisecond * 500)
	status := GetRoomStatus(true)
	if len(status) != 0 {
		t.Fatal("Room did not closed. len=%v, peoples=%d", len(status), status[0].Peoples)
	}

	ws1, _ = CreateConnection(t, room, 1)
	ws2, user2 = CreateConnection(t, room, 2)

	user2JoinMsg = ReadMsg(t, ws1)
	if user2JoinMsg.Type != Msg_Type_User_Join {
		t.Fatal("User Join ?")
	}
	if !strings.Contains(user2JoinMsg.Content, strconv.FormatInt(user2.Session.Id, 10)) {
		t.Fatal("user1 must recv user2-join msg.")
	}
	var name, color = "SendRenameMsg", "#0044FF"
	log.Info("Send Msg_Type_User_Name Req:%v", SendRenameMsg(t, ws2, name, color))
	if ReadMsg(t, ws2).Type != Msg_Type_Sys_Ok {
		t.Fatal("rename fail")
	}
	
	var renameMsg = ReadMsg(t, ws1)
	if renameMsg.Type != Msg_Type_User_Name {
		t.Fatal("Msg_Type_User_Name broadcast fail.", renameMsg.Type)
	}
	var rb RenameBroadcastContent
	
	err := json.Unmarshal([]byte(renameMsg.Content), &rb)
	if err != nil {
	    t.Fatal("Msg_Type_User_Name", err.Error())
	}
	if rb.NewName != name || rb.Color != color {
	    t.Fatal("Rename fail.")
	}
	    
	log.Info("Msg_Type_User_Name broadcast:%v", jsonString(renameMsg))
	// test WMD-update-text
	testUpdateTextMsg(t, ws1, ws2)
	return
}

func SandSaveTextMsg(t *testing.T, ws *websocket.Conn) string {
	return SendMsg(t, ws, Msg_Type_Text_Save, "", Msg_Type_Text_Save)
}
func SendCurPosMsg(t *testing.T, ws *websocket.Conn, content string) string {
	return SendMsg(t, ws, Msg_Type_Text_CurPos, content, Msg_Type_Text_CurPos)
}

func SendGetAllText(t *testing.T, ws *websocket.Conn) string {
	return SendMsg(t, ws, Msg_Type_Text_GetAll, "", Msg_Type_Text_GetAll)
}

func SendUpdateTextMsg(t *testing.T, ws *websocket.Conn, content string, index int64) string {
	return SendMsg(t, ws, Msg_Type_Text_Update, content, index)
}

func SendRenameMsg(t *testing.T, ws *websocket.Conn, newName string, color string) string {
    content := fmt.Sprintf("{\"name\":\"%s\",\"color\":\"%s\"}", newName, color)
	return SendMsg(t, ws, Msg_Type_User_Name, content, Msg_Type_User_Name)
}

func SandChatMsg(t *testing.T, ws *websocket.Conn, content string, index int64) string {
	return SendMsg(t, ws, Msg_Type_User_Chat, content, index)
}

func SendMsg(t *testing.T, ws *websocket.Conn, typ int64, content string, index int64) string {
	msg := BaseFrameMsg{
		Type:    typ,
		Content: content,
		Index:   index,
	}
	return sendFrame(t, ws, jsonString(msg))
}

func sendFrame(t *testing.T, ws *websocket.Conn, content string) string {
	if err := ws.WriteMessage(websocket.TextMessage, []byte(content)); err != nil {
		t.Fatal("Send Msg err=%v", err.Error())
	}
	return content
}

func ReadMsg(t *testing.T, ws *websocket.Conn) (msg BaseFrameMsg) {
	if ws == nil {
		t.Fatal("nil conn")
	}
	mTyp, frame, err := ws.ReadMessage()
	if err != nil {
		panic(err.Error())
		t.Fatal(err.Error())
	}

	if mTyp != websocket.TextMessage {
		fmt.Printf("frmae not Text, ->%v", mTyp)
		t.Fatal("frame must Text.")
	}
	if err := json.Unmarshal(frame, &msg); err != nil {
		t.Fatal(err.Error())
	}
	return msg
}

func CreateConnection(t *testing.T, room string, userId int) (ws *websocket.Conn, user *OnlineUser) {
	// 建立连接，服务端会发会两个消息，分别是Sys_Conn与 User_List, 还有Text_Init

	conn, err := net.Dial("tcp", HOST_PORT)

	if err != nil || conn == nil {
		t.Fatal(err.Error())
	}

	ws, resp, err := websocket.NewClient(conn, &url.URL{Host: HOST_PORT, Path: "/comet",
		RawQuery: fmt.Sprintf("room=%s&userid=%d", room, userId)}, nil)

	if err != nil || resp.StatusCode != 101 {
		log.Error("resp.StatusCode=%v", resp.StatusCode)
		t.Fatal(err.Error())
	}

	bMsg := ReadMsg(t, ws)
	if bMsg.Type != Msg_Type_Sys_Conn {
		t.Fatal("frist Message Type must Msg_Type_Sys_Conn. =>", bMsg.Type)
	}

	if err := json.Unmarshal([]byte(bMsg.Content), &user); err != nil {
		t.Fatal("frist Message Type must Msg_Type_Sys_Conn and Content can Convert User Object.")
	}

	msg := ReadMsg(t, ws)
	if msg.Type != Msg_Type_User_List {
		t.Fatal("Msg_Type_User_List require")
	}

	log.Info("1.Msg_Type_Sys_Conn-> %v", jsonString(bMsg))
	log.Info("2.Msg_Type_User_List->%v", jsonString(msg))

	return ws, user
}

func server_main_loop() {
	log.SetLevel(log.LevelDebug)
	http.HandleFunc("/comet", WebsocketHandler)

	svr := &http.Server{
		Addr:           HOST_PORT,
		Handler:        nil,
		ReadTimeout:    1 * time.Second,
		WriteTimeout:   1 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	err := svr.ListenAndServe()
	//    err := http.ListenAndServe(HOST_PORT, nil)
	if err != nil {
		log.Error("ListenAndServe[%v], err=[%v]", HOST_PORT, err.Error())
	}
}
