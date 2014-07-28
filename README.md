## 数据结构
```go
// 用户结构，TODO： 头像
type User struct{
    SessionId int64      // websocket连接唯一ID
    Name string          // 昵称
}
//消息类型
const (
    // 系统消息
    Msg_Type_Sys_Ok     = 0x00001000
    Msg_Type_Sys_Err    = 0x00001001
    Msg_Type_Sys_Conn   = 0x00001002    //当连接建立后，客户端收到的第一个消息类型。
    
    // 用户级消息
    Msg_Type_User_Join  = 0x00010000
    Msg_Type_User_Quit  = 0x00010001
    Msg_Type_User_Chat  = 0x00010002  // 注意: 同是chat消息，上传与下行格式不一样
    Msg_Type_User_Name  = 0x00010003  // user set him nickname, will broadcast to the room
    Msg_Type_User_List  = 0x00010004  // list All user in the room.

    // 文本编辑消息
    Msg_Type_Text_Init   = 0x00100000
    Msg_Type_Text_Update = 0x00100001
    Msg_Type_Text_CurPos = 0x00100002
    Msg_Type_Text_GetAll = 0x00100003
)
 
// 消息格式(即每一个websocket-frame, 类型为Text, 主体为JSON结构)：
{
    "type": Msg_Type_xx_xx,
    "index": <int64>, // 可能为零，每个Room内单调递的一个数值。
    "content": string, // 可转成javascript对像的JSON串, 会因type不同而转出来的结构不同,大部分消息应该该包含"sessionId":<int64>
}

```

## 各种消息结构<省略index字段>
```go
//----------------------
{
    "type": (, // 非广播类型消息，当连接建立后，客户端收到的第一个消息类型。
    "content": "{room:<str>,sessionId:<int64>}"
}
//----------------------
{ 
    "type": Msg_Type_User_List, // 非广播类型消息
    "content": "{room:<str>,users:[<{name:<str>,sessionId:<int64>}>,<user-object>, ...]}"
}
//----------------------
{ 
    "type": Msg_Type_User_Name | Msg_Type_User_Quit | Msg_Type_User_Join, // 广播类型消息
    "content": "{name:<str>,sessionId:<int64>}"
}
//----------上行：Client->Server ------------
{ 
    "type": Msg_Type_User_Chat, //上行：Client->Server
    "content": "<say-some-string>"
}
//----------下行：Server->Client ------------
{ 
    "type": Msg_Type_User_Chat, // 广播类型消息,-下行：Server->Client
    "content": "{sessionId:<int64>, say:<say-some-string>}"
}
```