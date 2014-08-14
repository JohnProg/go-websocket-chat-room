## 数据结构
```c
// 用户结构，TODO： 头像
type User struct{
    SessionId int64      //websocket连接唯一ID
    Name string          // 昵称
}
//消息类型
const (
    // 系统消息
    Msg_Type_Sys_Ok     = 0x00001000    // 4096
    Msg_Type_Sys_Err    = 0x00001001
    Msg_Type_Sys_Conn   = 0x00001002    //当连接建立后，客户端收到的第一个消息类型。
    
    // 用户级消息
    Msg_Type_User_Join  = 0x00010000    // 65536
    Msg_Type_User_Quit  = 0x00010001
    Msg_Type_User_Chat  = 0x00010002    // 注意: 同是chat消息，上传与下行格式不一样
    Msg_Type_User_Name  = 0x00010003    // user set him nickname, will broadcast to the room
    Msg_Type_User_List  = 0x00010004    // list All user in the room.

    
    // 文本编辑消息
    Msg_Type_Text_Init   = 0x00100000   // 1048576
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


消息交互流程：
====
1.建立连接
---

###前端发服务服发起websocket连接后，会收到以下2条消息

- Msg_Type_Sys_Conn 格式如下(表示服务器分配给当前连接sessionId)：

```c
{"type":4098,"index":0,"content":"{\"name\":\"\",\"sessionId\":4037200794235010051}"}
```

- Msg_Type_User_List 格式如下(表示此room中的在线人员)：

```c
{"type":65540,"index":0,"content":"{\"room\":\"testing\",\"users\":[{\"name\":\"U-1\",\"sessionId\":3916589616287113937}]}"}
```


2.获取markdown文档（同时编辑）
---

###2.1 发送 Msg_Type_Text_GetAll，前端需主动获取当前 markdown档本状态：
```c
{"type":1048579,"index":1048579,"content":""}  //发给服务器的Msg_Type_Text_GetAll消息格式
```

###2.2 返回 Msg_Type_Text_GetAll 消息格式如下：
```c
{"type":4096,"index":1048579,"content":"{\"version\":9,\"md5\":\"<xxoo>\",\"time\":1407827265,\"text\":\"AAAAAAAAA\"}"}
```
    表示当前文档版本是9， Msg_Type_Text_Update消息必需以此版本为基础上发给服务器。


3.更新markdown文档（同时编辑）
---

###3.1 发送 Msg_Type_Text_Update(上行)：
```c
{"type":1048577,"index":123333,"content":"[{\"version\":15,\"start\":0,\"end\":0,\"value\":\"A\"}]"}
```

###3.2 返回 Msg_Type_Text_Update 成功时格式(下行)：
```c
{"type":4096,"index":123333,"content":"{\"version\":16,\"md5\":\"<xxoo>\",\"time\":1407828912}"}
```

###3.3 下行广播 Msg_Type_Text_Update, 当更新文档成功，其它在当用户收到广播消息格式如下：
```c
{"type":1048577,"index":4,"content":"{\"name\":\"U-1\",\"sessionId\":6426100070888298971,\"status\":
{\"version\":16,\"md5\":\"<xxoo>\",\"time\":1407828912},\"Update\":[{\"version\":15,\"start\":0,\"end\":0,\"value\":\"A\"}]}"}
```
###3.4 如果Msg_Type_Text_Update不成功，会收到 Msg_Type_Sys_Err消息，Content中有错误内容：
```c
//最常见的是更新时文档的version值不等，返回形如
{"type":4097,"index":5,"content":"Update text Cmd Version not Equal."}
```
	此时应该发送Msg_Type_Text_GetAll消息，更新本地文档

-------------------

4.聊天与修改个性化名字
---
###4.1 聊天：Msg_Type_User_Chat消息（上行格式）
```c
{"type":65538,"index":13993402,"content":"user say something."}
```
###4.2 聊天广播：Msg_Type_User_Chat（下行格式）
```c
{"type":65538,"index":3,"content":"{\"sessionId\":1443635317331776148,\"say\":\"user say something.\",\"index\":13993402}"}
```
###4.3 改名：Msg_Type_User_Name消息（上行格式）
```c
{"type":65539,"index":65539,"content":"{\"name\":\"new-name\",\"color\":\"#0044FF\"}"}
```
###4.4 改名广播：Msg_Type_User_Name消息（下行格式）
```c
{"type":65539,"index":3,"content":"{\"name\":\"new-name\",\"color\":\"#0044FF\"}"}
```
