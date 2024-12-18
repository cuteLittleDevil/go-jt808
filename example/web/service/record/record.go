package record

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/service"
	"time"
)

type (
	manager struct {
		terminals map[string]*terminalInfo
	}

	terminalInfo struct {
		images    []string
		joinTime  time.Time
		leaveTime time.Time
		messages  [10]PackageInfo
		next      int
	}

	PackageInfo struct {
		// TerminalSeq 终端流水号
		TerminalSeq uint16 `json:"terminalSeq,omitempty"`
		// PlatformSeq 平台下发的流水号
		PlatformSeq uint16 `json:"platformSeq,omitempty"`
		// TerminalData 终端主动上传的数据 分包合并的情况是全部body合在一起
		TerminalData string `json:"terminalData,omitempty"`
		// PlatformData 平台下发的数据
		PlatformData string `json:"platformData,omitempty"`
		// ActiveSend 是否是平台主动下发的
		ActiveSend bool `json:"activeSend,omitempty"`
		// SubcontractComplete 分包情况是否最终完成了
		SubcontractComplete bool `json:"subcontractComplete,omitempty"`
		// TerminalCommand 终端的指令
		TerminalCommand string `json:"terminalCommand,omitempty"`
		// PlatformCommand 平台的指令
		PlatformCommand string `json:"platformCommand,omitempty"`
	}
)

type operationFunc func(manager *manager)

var (
	operationFuncChan = make(chan operationFunc, 10)
)

func Run() {
	var (
		record = &manager{
			terminals: make(map[string]*terminalInfo),
		}
	)
	defer close(operationFuncChan)
	for opFunc := range operationFuncChan {
		opFunc(record)
	}
}

func Join(msg service.Message) {
	ch := make(chan struct{})
	operationFuncChan <- func(record *manager) {
		defer close(ch)
		key := msg.Header.TerminalPhoneNo
		if _, ok := record.terminals[key]; !ok {
			record.terminals[key] = &terminalInfo{
				next:     0,
				messages: [10]PackageInfo{},
			}
		}
		v := record.terminals[key]
		v.joinTime = time.Now()
		v.messages[v.next] = toPackageInfo(msg)
	}
	<-ch
}

func Leave(key string) {
	ch := make(chan struct{})
	operationFuncChan <- func(record *manager) {
		defer close(ch)
		if v, ok := record.terminals[key]; ok {
			v.leaveTime = time.Now()
		}
	}
	<-ch
}

func AddMessage(msg service.Message) {
	ch := make(chan struct{})
	operationFuncChan <- func(record *manager) {
		defer close(ch)
		sim := msg.Header.TerminalPhoneNo
		if v, ok := record.terminals[sim]; ok {
			if v.next >= len(v.messages) {
				v.next = 0
			}
			v.messages[v.next] = toPackageInfo(msg)
			v.next++
		}
	}
	<-ch
}

func PutImageURL(sim string, savePath string) {
	ch := make(chan struct{})
	operationFuncChan <- func(record *manager) {
		defer close(ch)
		if v, ok := record.terminals[sim]; ok {
			v.images = append(v.images, savePath)
		}
	}
	<-ch
}

func Details() any {
	type Response struct {
		Sim       string          `json:"sim"`
		Images    []string        `json:"images"`
		JoinTime  string          `json:"joinTime"`
		LeaveTime string          `json:"leaveTimes"`
		Messages  [10]PackageInfo `json:"messages"`
		IsOnline  bool            `json:"isOnline"`
	}
	list := make([]Response, 0, 1000)
	ch := make(chan struct{})
	operationFuncChan <- func(record *manager) {
		defer close(ch)
		for k, v := range record.terminals {
			list = append(list, Response{
				Sim:       k,
				Images:    v.images,
				JoinTime:  v.joinTime.Format(time.RFC3339),
				LeaveTime: v.leaveTime.Format(time.RFC3339),
				Messages:  v.messages,
				IsOnline:  v.joinTime.After(v.leaveTime),
			})
		}
	}
	<-ch
	return list
}

func toPackageInfo(msg service.Message) PackageInfo {
	ex := msg.ExtensionFields
	return PackageInfo{
		TerminalSeq:         ex.TerminalSeq,
		PlatformSeq:         ex.PlatformSeq,
		TerminalData:        fmt.Sprintf("%x", ex.TerminalData),
		PlatformData:        fmt.Sprintf("%x", ex.PlatformData),
		ActiveSend:          ex.ActiveSend,
		SubcontractComplete: ex.SubcontractComplete,
		TerminalCommand:     msg.Command.String(),
		PlatformCommand:     ex.PlatformCommand.String(),
	}
}
