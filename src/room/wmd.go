/*
*  wmd - The Wysiwym Markdown Editor  http://www.wmd-editor.com/
 */
package room

import (
	"io"
	"fmt"
	"time"
	"errors"
	"crypto/md5"
	"libs/log"
)

var (
	ErrCmdVersionNotEqual = errors.New("Update text Cmd Version not Equal.")
)

//------------

type MarkDownStatus struct {
	Version int32  `json:"version"`
	Md5 	string `json:"md5"`
	Time    int64  `json:"time"`
}

type MarkDownText struct {
	MarkDownStatus
	Content string `json:"text"`
}

type TextUpdateCmd struct {
	Version int32  `json:"version"`
	Start   int    `json:"start"`
	End     int    `json:"end"`
	Value   string `json:"value"`
}

func (md *MarkDownText) GetMarkDownText() MarkDownText {
	//对像会自动copy，新对像改变，并不会影响此md
	return *md
}

func (md *MarkDownText) UpdateTextCmd(cmd TextUpdateCmd) (stat MarkDownStatus, err error) {
	if cmd.Version != md.Version {
		return stat, ErrCmdVersionNotEqual
	}
	log.Info("Update Text, start=%v, end=%v, Text=%v", cmd.Start, cmd.End, cmd.Value)
	stat = md.UpdateText(cmd.Start, cmd.End, cmd.Value)
	return stat, nil
}

func (md *MarkDownText) UpdateText(start int, end int, text string) (stat MarkDownStatus) {
	md.Update(start, end, text)
	return md.MarkDownStatus
}

func (md *MarkDownText) Init(content string) string {
	return md.setText(content, 0)
}

func (md *MarkDownText) Update(start int, end int, text string) string {
	return md.setText(repalce(md.Content, text, start, end), md.Version+1)
}

func (md *MarkDownText) setText(content string, version int32) string {
	// TODO: lock
	md.Content = content
	hash := md5.New()
	io.WriteString(hash, content)
	md.Md5 = fmt.Sprintf("%x", hash.Sum(nil))
	md.Version = version
	md.Time = time.Now().Unix()

	return md.Md5
}

func repalce(src string, newStr string, start int, end int) string {
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}
	prefix := substr(src, 0, start)
	suffix := substr(src, end, int(len(src)))
	return prefix + newStr + suffix
}

func substr(s string, start int, length int) string {
	runes := []rune(s)
	l := start + length
	sLen := len(runes)
	if l > sLen {
		l = sLen
	}
	if start > sLen {
		return ""
	}
	return string(runes[start:l])
}
