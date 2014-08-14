package room
import (
    "testing"
//    "strings"
)

func TestMarkDownText(t *testing.T) {
    md := MarkDownText{}
    if 0 != md.Version{
        t.Fatal("when init , version is 0")
    }
    md.Init("你好！！！hello world！")

    if md.GetMarkDownText().Version != 0{
        t.Fatal("init text, version is 0.")
    }

    md.Update(1, 2, "D")
    if md.Content != "你D！！！hello world！"  {
        t.Fatal("Update Content fail.")
    }

    if md.GetMarkDownText().Version != 1{
        t.Fatal("update text, version is 1.")
    }
    for i := 1; i<10;i++ {
        cmd := TextUpdateCmd{
            Version: int32(i),
            Start: i,
            End: i*2,
            Value:"T",
        }
        stat, err := md.UpdateTextCmd(cmd)
        if err != nil || cmd.Version -1 == stat.Version{
            t.Fatal(err.Error())
        }

    }
    md.UpdateText(0, 2, "BB")

    md.Update(-10, 10, "AA")

    md.Update(0, 0, "AA")

    md.Update(0, 1000, "")
    md.Update(1000, 10, "AA")

}
