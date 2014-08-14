package room
import (
    "testing"
)

func TestFileSaver(t *testing.T){
    var saver  = FileSaver{}
    var md = MarkDownText{}
    var id = "for-test-file-saver"

    md.Init(id)
    md.Update(0, 10, "sfallwn,n")
    md.Update(0, 10, "sfdddddddddd")

    err := saver.Save(id, md)
    if err != nil{
        t.Fatal("file saver err=", err.Error())
    }

    var md2 *MarkDownText

    md2, err2 := saver.Load(id)
    if err2 != nil {
        t.Fatal("file saver load err=", err2.Error())
    }
    if md2 == nil || md2.Md5 == "" {
        t.Fatal("file saver load an nil MD")
    }

    if md.Version != md2.Version || md.Md5 != md2.Md5 ||  md.Content != md2.Content {
        t.Fatal("file saver load fail.....")
    }
    if md.Time != md2.Time{
        t.Fatal("file saver load fail.....")
    }
}
