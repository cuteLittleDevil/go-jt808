package record

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/service"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"sort"
	"time"
)

type (
	manager struct {
		terminals map[string]*terminalInfo
	}

	terminalInfo struct {
		images     []string
		minioUrls  []string
		joinTime   time.Time
		leaveTime  time.Time
		updateTime time.Time
		messages   [10]PackageInfo
		next       int
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
		TerminalCommand consts.JT808CommandType `json:"terminalCommand,omitempty"`
		// PlatformCommand 平台的指令
		PlatformCommand consts.JT808CommandType `json:"platformCommand,omitempty"`
		// Remark 备注
		Remark string `json:"remark,omitempty"`
		// UpdateTime 更新时间
		UpdateTime time.Time `json:"updateTime,omitempty"`
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
			v.updateTime = time.Now()
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

func PutMinioURL(sim string, url string) {
	ch := make(chan struct{})
	operationFuncChan <- func(record *manager) {
		defer close(ch)
		if v, ok := record.terminals[sim]; ok {
			v.minioUrls = append(v.minioUrls, url)
		}
	}
	<-ch
}

func Details() any {
	type Response struct {
		Sim        string        `json:"sim"`
		Images     []string      `json:"images"`
		MinioUrls  []string      `json:"minioUrls"`
		JoinTime   string        `json:"joinTime"`
		LeaveTime  string        `json:"leaveTime"`
		UpdateTime string        `json:"updateTime"`
		Messages   []PackageInfo `json:"messages"`
		IsOnline   bool          `json:"isOnline"`
	}
	list := make([]Response, 0, 1000)
	ch := make(chan struct{})
	operationFuncChan <- func(record *manager) {
		defer close(ch)
		for k, v := range record.terminals {
			msgs := make([]PackageInfo, 0, 10)
			for _, info := range v.messages {
				if info.Remark != "" {
					msgs = append(msgs, info)
				}
			}
			sort.Slice(msgs, func(i, j int) bool {
				return msgs[i].UpdateTime.After(msgs[j].UpdateTime)
			})
			list = append(list, Response{
				Sim:        k,
				Images:     v.images,
				MinioUrls:  v.minioUrls,
				JoinTime:   v.joinTime.Format(time.DateTime),
				LeaveTime:  v.leaveTime.Format(time.DateTime),
				UpdateTime: v.updateTime.Format(time.DateTime),
				Messages:   msgs,
				IsOnline:   v.joinTime.After(v.leaveTime),
			})
		}
		sort.Slice(list, func(i, j int) bool {
			return list[i].JoinTime > list[j].JoinTime
		})
	}
	<-ch
	return list
}

func toPackageInfo(msg service.Message) PackageInfo {
	ex := msg.ExtensionFields
	source, target := msg.Command, ex.PlatformCommand
	if msg.ExtensionFields.ActiveSend {
		source, target = target, source
	}
	return PackageInfo{
		TerminalSeq:         ex.TerminalSeq,
		PlatformSeq:         ex.PlatformSeq,
		TerminalData:        fmt.Sprintf("%x", ex.TerminalData),
		PlatformData:        fmt.Sprintf("%x", ex.PlatformData),
		ActiveSend:          ex.ActiveSend,
		SubcontractComplete: ex.SubcontractComplete,
		TerminalCommand:     msg.Command,
		PlatformCommand:     ex.PlatformCommand,
		Remark:              fmt.Sprintf("%s -> %s", source, target),
		UpdateTime:          time.Now(),
	}
}
