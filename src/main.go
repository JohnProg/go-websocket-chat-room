// gomm-online main

package main

import (
	"io"
    "time"
    "strconv"
    "net/http"
    "libs/websocket"
    "libs/log"
    "room"
)

const (
    HOST_PORT = ":10086"
)

var _info = log.Info
var _err = log.Error

func websocketHandle(w http.ResponseWriter, r *http.Request) {
    sessionId := time.Now().UnixNano()

    roomId := r.URL.Query().Get("room")
    userId, err := strconv.ParseInt(r.URL.Query().Get("userid"), 10, 0)

    if "" == roomId || err != nil {
        log.Debug("no room id.")
        io.WriteString(w, "pls used wss://0.0.0.0.0/comet?room=xxx&userid=[int32] to connect.")
        return
    }

    conn, err := websocket.Upgrade(w, r, nil)
    if err != nil {
        _err(err.Error())
        return
    }
    _info("new websocket connect. userId=%v", userId)

    user := room.User {
        SessionId: sessionId,
        Name: "name+"+ strconv.FormatInt(userId, 10),
    }

    onlineUser := room.OnlineUser {
        User: user,
        WebSocket: conn,
        Inbox: make(chan room.Message, 16),
    }

    room.JoinRoom(roomId, &onlineUser)
    go onlineUser.WaitForRoomMsg()
    onlineUser.WaitForFrame()
    _info("end session[%v]", sessionId)
}

func main(){
    log.SetLevel(log.LevelDebug)
    http.HandleFunc("/comet",  websocketHandle)
    err := http.ListenAndServe(HOST_PORT, nil)
    if err != nil {
        _err("ListenAndServe[%v], err=[%v]", HOST_PORT, err.Error())
    }
}
