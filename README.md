## ���ݽṹ
```go
// �û��ṹ��TODO�� ͷ��
type User struct{
    SessionId int64      // websocket����ΨһID
    Name string          // �ǳ�
}
//��Ϣ����
const (
    // ϵͳ��Ϣ
    Msg_Type_Sys_Ok     = 0x00001000
    Msg_Type_Sys_Err    = 0x00001001
    Msg_Type_Sys_Conn   = 0x00001002    //�����ӽ����󣬿ͻ����յ��ĵ�һ����Ϣ���͡�
    
    // �û�����Ϣ
    Msg_Type_User_Join  = 0x00010000
    Msg_Type_User_Quit  = 0x00010001
    Msg_Type_User_Chat  = 0x00010002  // ע��: ͬ��chat��Ϣ���ϴ������и�ʽ��һ��
    Msg_Type_User_Name  = 0x00010003  // user set him nickname, will broadcast to the room
    Msg_Type_User_List  = 0x00010004  // list All user in the room.

    // �ı��༭��Ϣ
    Msg_Type_Text_Init   = 0x00100000
    Msg_Type_Text_Update = 0x00100001
    Msg_Type_Text_CurPos = 0x00100002
    Msg_Type_Text_GetAll = 0x00100003
)
 
// ��Ϣ��ʽ(��ÿһ��websocket-frame, ����ΪText, ����ΪJSON�ṹ)��
{
    "type": Msg_Type_xx_xx,
    "index": <int64>, // ����Ϊ�㣬ÿ��Room�ڵ����ݵ�һ����ֵ��
    "content": string, // ��ת��javascript�����JSON��, ����type��ͬ��ת�����Ľṹ��ͬ,�󲿷���ϢӦ�øð���"sessionId":<int64>
}

```

## ������Ϣ�ṹ<ʡ��index�ֶ�>
```go
//----------------------
{
    "type": (, // �ǹ㲥������Ϣ�������ӽ����󣬿ͻ����յ��ĵ�һ����Ϣ���͡�
    "content": "{room:<str>,sessionId:<int64>}"
}
//----------------------
{ 
    "type": Msg_Type_User_List, // �ǹ㲥������Ϣ
    "content": "{room:<str>,users:[<{name:<str>,sessionId:<int64>}>,<user-object>, ...]}"
}
//----------------------
{ 
    "type": Msg_Type_User_Name | Msg_Type_User_Quit | Msg_Type_User_Join, // �㲥������Ϣ
    "content": "{name:<str>,sessionId:<int64>}"
}
//----------���У�Client->Server ------------
{ 
    "type": Msg_Type_User_Chat, //���У�Client->Server
    "content": "<say-some-string>"
}
//----------���У�Server->Client ------------
{ 
    "type": Msg_Type_User_Chat, // �㲥������Ϣ,-���У�Server->Client
    "content": "{sessionId:<int64>, say:<say-some-string>}"
}
```