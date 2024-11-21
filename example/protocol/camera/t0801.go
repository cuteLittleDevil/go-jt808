package main

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"os"
	"sync"
	"time"
)

type meT0801 struct {
	model.T0x0801
	phone               string
	once                sync.Once
	file                *os.File
	num                 uint16
	sum                 uint16
	repairPacketSeq     []uint16
	startTime           time.Time
	subcontractComplete bool
}

func (t *meT0801) savePath() string {
	format := ".jpg"
	switch t.MultimediaFormatEncode {
	case 0:
		format = ".jpeg"
	case 1:
		format = ".tlf"
	case 2:
		format = ".mp3"
	case 3:
		format = ".wav"
	case 4:
		format = ".wmv"
	}
	return fmt.Sprintf("./%s_%d%s", t.phone, t.MultimediaID, format)
}

func (t *meT0801) schedule(num, sum uint16) {
	if num == 1 {
		t.num = num
		t.sum = sum
		t.subcontractComplete = false
		return
	}
	if num == sum {
		return
	}
	if t.num+1 != num {
		fmt.Println(fmt.Sprintf("不连续的情况 之前进度[%d/%d] 现在进度[%d/%d]", t.num, t.sum, num, sum))
		// 不连续的情况 要补传 目前搞不懂用什么id（是流水号还是分包的序号呢）
		//seq := jtMsg.Header.SerialNumber // 流水号
		//seq := jtMsg.Header.SubPackageNo // 分包序号 以下是用分包序号的
		//for i := t.num + 1; i < num; i++ {
		//	t.repairPacketSeq = append(t.repairPacketSeq, i)
		//}
	}
	t.num = num
	t.sum = sum
}

func (t *meT0801) OnReadExecutionEvent(msg *service.Message) {
	if msg.Header.TerminalPhoneNo != phone {
		return
	}
	t.once.Do(func() {
		name := fmt.Sprintf("./%s_0801.log", t.phone)
		t.file, _ = os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	})
	_, _ = t.file.WriteString(fmt.Sprintf("read %x\n", msg.ExtensionFields.TerminalData))

	if msg.Header.SubPackageNo == 1 { // 需要关闭过滤分包的 service.WithHasSubcontract(false),
		_ = t.T0x0801.Parse(msg.JTMessage)
		fmt.Println("包开始", t.String())
		t.startTime = time.Now()
	}
	num := msg.Header.SubPackageNo
	sum := msg.Header.SubPackageSum
	t.schedule(num, sum)
	if num == sum && len(msg.Body) == int(msg.Header.Property.BodyDayaLen) { // 最后一个包
		t.subcontractComplete = true
	}
	if msg.ExtensionFields.SubcontractComplete { // 分包完成
		_ = t.T0x0801.Parse(msg.JTMessage)
		name := t.savePath()
		fmt.Println("包完成", name, time.Since(t.startTime), t.String())
		_ = os.WriteFile(name, t.T0x0801.MultimediaPackage, os.ModePerm)
	}
}

func (t *meT0801) OnWriteExecutionEvent(msg service.Message) {
	if msg.Header.TerminalPhoneNo != t.phone {
		return
	}
	t.once.Do(func() {
		name := fmt.Sprintf("./%s_0801.log", t.phone)
		t.file, _ = os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	})
	_, _ = t.file.WriteString(fmt.Sprintf("write %x\n", msg.ExtensionFields.PlatformData))
}

func (t *meT0801) ReplyProtocol() consts.JT808CommandType {
	return consts.P8800MultimediaUploadRespond
}

func (t *meT0801) HasReply() bool {
	defer func() {
		t.subcontractComplete = false
	}()
	return t.subcontractComplete
}

func (t *meT0801) ReplyBody(jtMsg *jt808.JTMessage) ([]byte, error) {
	_ = t.Parse(jtMsg)
	p8800 := model.P0x8800{
		MultimediaID: t.MultimediaID,
	}
	if len(t.repairPacketSeq) > 0 {
		p8800.AgainPackageCount = byte(len(t.repairPacketSeq))
		p8800.AgainPackageList = append(p8800.AgainPackageList, t.repairPacketSeq...)
		t.repairPacketSeq = t.repairPacketSeq[0:0]
	}
	return p8800.Encode(), nil
}
