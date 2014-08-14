package room

import (
    "io"
    "os"
    "fmt"
    "errors"
    "crypto/md5"
    "encoding/binary"
    "libs/log"
)

const (
    SAVE_DIR="/data/gomm"
)

var (
    ByteOrder = binary.LittleEndian
)

type FileSaver struct {}
const (
    Md5_String_Len = 32
)
func createSaveDir() error{
    err := os.MkdirAll(SAVE_DIR, os.ModePerm)
    if err != nil {
        log.Error("mkdir[%s] err=%v", SAVE_DIR, err.Error()) 
        return err
    }
    return nil
}

func (fs *FileSaver)Save(id string, md MarkDownText) (err error){
    err = createSaveDir()
    if nil != err{
        return err
    }
    if id ==""{
        return errors.New("Empty string [id] not allow.")
    }

    fp, err := os.Create(SAVE_DIR + "/" + id)
    if err !=nil {
        log.Error(err.Error())
        return err
    }
    defer fp.Close()

    err = binary.Write(fp, ByteOrder, md.Version)
    if err !=nil {
        log.Error(err.Error())
        return err
    }

    err = binary.Write(fp, ByteOrder, md.Time)
    if err !=nil {
        log.Error(err.Error())
        return err
    }

    byteContent := []byte(md.Content)
    var lenContent int64 = int64(len(byteContent))

    err = binary.Write(fp, ByteOrder, lenContent)
    if err !=nil {
        log.Error(err.Error())
        return err
    }
    
    var n int
    n, err = fp.Write([]byte(md.Md5))
    if err !=nil {
        log.Error(err.Error())
        return err
    }
	if n != Md5_String_Len{
	    log.Error("Write Md5 to file[%s] fail.", id)
	    return errors.New("Write Md5 to file fail.")
	}

    n, err = fp.Write([]byte(byteContent)) //binary.Write(fp, ByteOrder, md.Content)

    if err != nil {
        log.Error(err.Error())
        return err
    }
    if int64(n) != lenContent {
        log.Error("Write Content-len not equal. %d != %d", n, lenContent)
        return errors.New("Write Content-len not equal")
    }
	log.Info("Write len[%d], content=[%v]",lenContent, md.Content)
    return nil
}

func (fs *FileSaver)Load(id string)(md *MarkDownText, err error){
    md = &MarkDownText{}

    fp, err := os.Open(SAVE_DIR + "/" + id)
    if err !=nil {
        log.Error(err.Error())
        return nil, err
    }
    defer fp.Close()

    err = binary.Read(fp, ByteOrder, &md.Version)
    if err != nil {
        log.Error(err.Error())
        return nil, err
    }

    err = binary.Read(fp, ByteOrder, &md.Time)
    if err != nil {
        log.Error(err.Error())
        return nil, err
    }

    var lenContent int64 = 0
    err = binary.Read(fp, ByteOrder, &lenContent)
    if err != nil {
        log.Error(err.Error())
        return nil, err
    }
    lenContent += Md5_String_Len
	
    var buffer []byte = make([]byte, lenContent, lenContent)
    log.Info("Read-Buffer-len=%v", len(buffer))

    var n int
    n, err = fp.Read(buffer)
    
    if err != nil {
        log.Error(err.Error())
        return nil, err
    }
    if int64(n) != lenContent {
        log.Error("Read Content-len not equal, %d != %d", n, lenContent)
        return nil, errors.New("Read Content-len not equal")
    }
	md.Md5 = string(buffer[:Md5_String_Len])
    md.Content = string(buffer[Md5_String_Len:])
	
	hash := md5.New()
	io.WriteString(hash,md.Content)
	if md.Md5 != fmt.Sprintf("%x", hash.Sum(nil)){
	    log.Error("When read file, check md5 not match.")
	    return nil, errors.New("When read file, check md5 not match.")
	}

    return md, nil
}
