// gomm-online main

package main

import (
    "time"
    "flag"
    "strings"
    "math/rand"
    "net/http"
    "libs/log"
    "room"
)

const (
    HOST_PORT = "127.0.0.1:80"
)

var host_port *string = flag.String("h", HOST_PORT, "host:port for HTTP service Listen to.")
var level *string = flag.String("l", "[debug|info|warn|error]", "log level")

func main(){
    flag.Parse()
	if *host_port == ""{
	    *host_port = HOST_PORT 
	}

	var default_level = log.LevelInfo

	*level = strings.ToLower(*level)
	
	if strings.HasPrefix(*level, "info"){
	    default_level = log.LevelInfo
	} else if strings.HasPrefix(*level, "warn"){
	    default_level = log.LevelWarn
	} else if strings.HasPrefix(*level, "err"){
	    default_level = log.LevelError
	} else if strings.HasPrefix(*level, "debug"){
	    default_level = log.LevelDebug
	}else{
	 	default_level = log.LevelInfo
	}

    rand.Seed(time.Now().UnixNano())

    log.SetLevel(default_level)
    http.HandleFunc("/comet",  room.WebsocketHandler)
    http.HandleFunc("/status",  room.StatusHandler)
    http.HandleFunc("/api/rooms/status",  room.RoomStatusHandler)

	log.Info("Server Listen to: %v", *host_port)
    svr :=&http.Server{
        Addr:           *host_port,
        Handler:        nil,
        ReadTimeout:    0 * time.Second,
        WriteTimeout:   0 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }

    err := svr.ListenAndServe()

    if err != nil {
        log.Error("ListenAndServe[%v], err=[%v]", HOST_PORT, err.Error())
    }
    
}

