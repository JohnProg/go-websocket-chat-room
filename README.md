## ���ݽṹ
```c
// �û��ṹ��TODO�� ͷ��
type User struct{
    SessionId int64      //websocket����ΨһID
    Name string          // �ǳ�
}
//��Ϣ����
const (
    // ϵͳ��Ϣ
    Msg_Type_Sys_Ok     = 0x00001000    // 4096
    Msg_Type_Sys_Err    = 0x00001001
    Msg_Type_Sys_Conn   = 0x00001002    //�����ӽ����󣬿ͻ����յ��ĵ�һ����Ϣ���͡�
    
    // �û�����Ϣ
    Msg_Type_User_Join  = 0x00010000    // 65536
    Msg_Type_User_Quit  = 0x00010001
    Msg_Type_User_Chat  = 0x00010002    // ע��: ͬ��chat��Ϣ���ϴ������и�ʽ��һ��
    Msg_Type_User_Name  = 0x00010003    // user set him nickname, will broadcast to the room
    Msg_Type_User_List  = 0x00010004    // list All user in the room.

    
    // �ı��༭��Ϣ
    Msg_Type_Text_Init   = 0x00100000   // 1048576
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


��Ϣ�������̣�
====
1.��������
---

###ǰ�˷����������websocket���Ӻ󣬻��յ�����2����Ϣ

- Msg_Type_Sys_Conn ��ʽ����(��ʾ�������������ǰ����sessionId)��

```c
{"type":4098,"index":0,"content":"{\"name\":\"\",\"sessionId\":4037200794235010051}"}
```

- Msg_Type_User_List ��ʽ����(��ʾ��room�е�������Ա)��

```c
{"type":65540,"index":0,"content":"{\"room\":\"testing\",\"users\":[{\"name\":\"U-1\",\"sessionId\":3916589616287113937}]}"}
```


2.��ȡmarkdown�ĵ���ͬʱ�༭��
---

###2.1 ���� Msg_Type_Text_GetAll��ǰ����������ȡ��ǰ markdown����״̬��
```c
{"type":1048579,"index":1048579,"content":""}  //������������Msg_Type_Text_GetAll��Ϣ��ʽ
```

###2.2 ���� Msg_Type_Text_GetAll ��Ϣ��ʽ���£�
```c
{"type":4096,"index":1048579,"content":"{\"version\":9,\"md5\":\"<xxoo>\",\"time\":1407827265,\"text\":\"AAAAAAAAA\"}"}
```
    ��ʾ��ǰ�ĵ��汾��9�� Msg_Type_Text_Update��Ϣ�����Դ˰汾Ϊ�����Ϸ�����������


3.����markdown�ĵ���ͬʱ�༭��
---

###3.1 ���� Msg_Type_Text_Update(����)��
```c
{"type":1048577,"index":123333,"content":"[{\"version\":15,\"start\":0,\"end\":0,\"value\":\"A\"}]"}
```

###3.2 ���� Msg_Type_Text_Update �ɹ�ʱ��ʽ(����)��
```c
{"type":4096,"index":123333,"content":"{\"version\":16,\"md5\":\"<xxoo>\",\"time\":1407828912}"}
```

###3.3 ���й㲥 Msg_Type_Text_Update, �������ĵ��ɹ��������ڵ��û��յ��㲥��Ϣ��ʽ���£�
```c
{"type":1048577,"index":4,"content":"{\"name\":\"U-1\",\"sessionId\":6426100070888298971,\"status\":
{\"version\":16,\"md5\":\"<xxoo>\",\"time\":1407828912},\"Update\":[{\"version\":15,\"start\":0,\"end\":0,\"value\":\"A\"}]}"}
```
###3.4 ���Msg_Type_Text_Update���ɹ������յ� Msg_Type_Sys_Err��Ϣ��Content���д������ݣ�
```c
//������Ǹ���ʱ�ĵ���versionֵ���ȣ���������
{"type":4097,"index":5,"content":"Update text Cmd Version not Equal."}
```
	��ʱӦ�÷���Msg_Type_Text_GetAll��Ϣ�����±����ĵ�

-------------------

4.�������޸ĸ��Ի�����
---
###4.1 ���죺Msg_Type_User_Chat��Ϣ�����и�ʽ��
```c
{"type":65538,"index":13993402,"content":"user say something."}
```
###4.2 ����㲥��Msg_Type_User_Chat�����и�ʽ��
```c
{"type":65538,"index":3,"content":"{\"sessionId\":1443635317331776148,\"say\":\"user say something.\",\"index\":13993402}"}
```
###4.3 ������Msg_Type_User_Name��Ϣ�����и�ʽ��
```c
{"type":65539,"index":65539,"content":"{\"name\":\"new-name\",\"color\":\"#0044FF\"}"}
```
###4.4 �����㲥��Msg_Type_User_Name��Ϣ�����и�ʽ��
```c
{"type":65539,"index":3,"content":"{\"name\":\"new-name\",\"color\":\"#0044FF\"}"}
```
