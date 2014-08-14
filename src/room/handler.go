package room

import (
	"io"
    "fmt"
    "time"
    "net/http"
    "math/rand"
    "crypto/md5"
    "encoding/json"
    "libs/websocket"
    "libs/log"
)

const (
    COOKIE_NAME = "uuid"
    COOKIE_EXPIRE_DAYS = 30
    COOKIE_DOMAIN = ""
)

func RoomStatusHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("%s %s", r.Method, r.URL.Path)
	status := GetRoomStatus(true)
    js, _ := json.Marshal(status)
    w.Header().Set("Content-Type", "application/json")
    io.WriteString(w, string(js))
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("%s %s", r.Method, r.URL.Path)

    status := GetRoomStatus(true)
    io.WriteString(w, fmt.Sprintf("Cookie=%v\n", getCookie(w, r)))

    io.WriteString(w, fmt.Sprintf("Opned Room count=%v\n\r", len(status)))

    for i, room := range status{
        io.WriteString(w, fmt.Sprintf(" ===%d-Room[%v], Peoples[%d], Broadcast[%d] ===\n", i, room.Id, room.Peoples, room.MsgLen))
        io.WriteString(w, fmt.Sprintf("    Text: ver=[%d], md5=[%v],Time=[%v], txt=[%s]\n", room.Version, 
                room.Md5, room.Time, room.Text))

        for _, user := range room.Users {
            io.WriteString(w, fmt.Sprintf("\tsid[%v], User[%v] chan-Len=[%d]\n",user.Id, user.Name, user.MsgLen))
        }
        io.WriteString(w, fmt.Sprintf(" === end - Room[%v] ===\n\n", room.Id))
    }
}

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("%s %s", r.Method, r.URL.Path)
    var roomId = r.URL.Query().Get("room")
    var userId = rand.Int63() % 10
    var cookie string = getCookie(w, r)

    if "" == roomId {
        log.Debug("room-id=%v, userid=%v", roomId,  userId)
        io.WriteString(w, "pls used wss://0.0.0.0.0/comet?room=xxx to connect.")
        return
    }

    conn, err := websocket.Upgrade(w, r, http.Header{})
    if err != nil {
        log.Error(err.Error())
        return
    }

    log.Info("new websocket connect to Room[%v]. cookie=%v", roomId, cookie)

    conn.SetReadDeadline(time.Time{})
    conn.SetWriteDeadline(time.Time{})

    _, onlineUser := JoinRoom(roomId, cookie, conn)

    go onlineUser.WaitForRoomMsg()
    onlineUser.WaitForFrame()
    log.Info("User[%v] end session[%v]", onlineUser.Name, onlineUser.Id)
}


func genCookieValue()string{
    rand.Int63()
    h := md5.New()
    io.WriteString(h, fmt.Sprintf("%v%v", rand.Int63(), time.Now().UnixNano()))
    return fmt.Sprintf("%x", h.Sum(nil))
}

func addCookie(w http.ResponseWriter)(cookie *http.Cookie) {
    val := genCookieValue()
    expire := time.Now().AddDate(0, COOKIE_EXPIRE_DAYS, 0) // 1 months
    cookie = &http.Cookie{
        Name: COOKIE_NAME,
        Value: val,
        Path: "/",
//        Domain: COOKIE_DOMAIN,
        Expires: expire,
        RawExpires: expire.Format(time.UnixDate),
        MaxAge:86400 * COOKIE_EXPIRE_DAYS,
    }

    http.SetCookie(w, cookie)
    return cookie
}

func getCookie(w http.ResponseWriter, req *http.Request) string {
    var cookie, err = req.Cookie(COOKIE_NAME)

    if err != nil{
        cookie = addCookie(w)
        log.Info("Cookie[%v] not found. err=%s, genNewCookie=%s", COOKIE_NAME, err.Error(), cookie.Value)
    }
    return cookie.Value
}