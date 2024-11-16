package model

import (
	"encoding/binary"
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/utils"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"strings"
)

type P0x9206 struct {
	BaseHandle
	// FTPAddrLen ftp地址长度
	FTPAddrLen byte `json:"ftpAddrLen"`
	// FTPAddrLen ftp地址
	FTPAddr string `json:"ftpAddr"`
	// Port 服务器端口
	Port uint16 `json:"port"`
	// UsernameLen 用户名长度
	UsernameLen byte `json:"usernameLen"`
	// Username 用户名
	Username string `json:"username"`
	// PasswordLen 密码长度
	PasswordLen byte `json:"passwordLen"`
	// Password 密码
	Password string `json:"password"`
	// FileUploadPathLen 文件上传路径长度
	FileUploadPathLen byte `json:"fileUploadPathLen"`
	// FileUploadPath 文件上传路径
	FileUploadPath string `json:"fileUploadPath"`
	// ChannelNo 逻辑通道号
	ChannelNo byte `json:"channelNo"`
	// StartTime YY-MM-DD-HH-MM-SS
	StartTime string `json:"startTime"`
	// EndTime YY-MM-DD-HH-MM-SS
	EndTime string `json:"endTime"`
	// AlarmFlag 报警标志
	AlarmFlag uint64 `json:"alarmFlag"`
	// MediaType 音视频类型(媒体类型) 0-音频和视频 1-音频 2-视频 3-音频或视频
	MediaType byte `json:"mediaType"`
	// StreamType 码流类型 0-主或子码流 1-主码流 2-子码流
	StreamType byte `json:"streamType"`
	// MemoryPosition 存储位置 0-主存储器或灾备存储器 1-主存储器 2-灾备存储器
	MemoryPosition byte `json:"memoryPosition"`
	// TaskExecuteCondition 任务执行条件 用 bit 位表示：
	//bit0:WIFI 为1时表示 WI-FI 下可下载
	//bit1:LAN 为1时表示LAN 连接时可下载
	//bi2:3G/4G 为1时表示3G/4G 连接时可下载
	TaskExecuteCondition byte `json:"taskExecuteCondition"`
}

func (p *P0x9206) Protocol() consts.JT808CommandType {
	return consts.P9206FileUploadInstructions
}

func (p *P0x9206) ReplyProtocol() consts.JT808CommandType {
	return consts.T1206FileUploadCompleteNotice
}

func (p *P0x9206) Parse(jtMsg *jt808.JTMessage) error {
	body := jtMsg.Body
	if len(body) < 1 {
		return protocol.ErrBodyLengthInconsistency
	}

	p.FTPAddrLen = body[0]
	start := 1
	end := start + int(p.FTPAddrLen)
	if len(body) < end+2+1 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.FTPAddr = string(body[start:end])
	p.Port = binary.BigEndian.Uint16(body[end : end+2])
	p.UsernameLen = body[end+2]

	start = end + 2 + 1
	end = start + int(p.UsernameLen)
	if len(body) < end+1 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.Username = string(body[start:end])
	p.PasswordLen = body[end]

	start = end + 1
	end = start + int(p.PasswordLen)
	if len(body) < end+1 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.Password = string(body[start:end])
	p.FileUploadPathLen = body[end]

	start = end + 1
	end = start + int(p.FileUploadPathLen)
	if len(body) != end+1+6+6+8+1+1+1+1 {
		return protocol.ErrBodyLengthInconsistency
	}
	p.FileUploadPath = string(body[start:end])

	start = end
	p.ChannelNo = body[start]
	p.StartTime = utils.BCD2Time(body[start+1 : start+7])
	p.EndTime = utils.BCD2Time(body[start+7 : start+13])
	p.AlarmFlag = binary.BigEndian.Uint64(body[start+13 : start+21])
	p.MediaType = body[start+21]
	p.StreamType = body[start+22]
	p.MemoryPosition = body[start+23]
	p.TaskExecuteCondition = body[start+24]
	return nil
}

func (p *P0x9206) Encode() []byte {
	data := make([]byte, 0, 100)
	data = append(data, byte(len(p.FTPAddr)))
	data = append(data, p.FTPAddr...)
	data = binary.BigEndian.AppendUint16(data, p.Port)
	data = append(data, p.UsernameLen)
	data = append(data, p.Username...)
	data = append(data, p.PasswordLen)
	data = append(data, p.Password...)
	data = append(data, p.FileUploadPathLen)
	data = append(data, p.FileUploadPath...)
	data = append(data, p.ChannelNo)
	data = append(data, utils.Time2BCD(p.StartTime)...)
	data = append(data, utils.Time2BCD(p.EndTime)...)
	data = binary.BigEndian.AppendUint64(data, p.AlarmFlag)
	data = append(data, p.MediaType)
	data = append(data, p.StreamType)
	data = append(data, p.MemoryPosition)
	data = append(data, p.TaskExecuteCondition)
	return data
}

func (p *P0x9206) HasReply() bool {
	return false
}

func (p *P0x9206) String() string {
	return strings.Join([]string{
		"数据体对象:{",
		fmt.Sprintf("\t%s:[%x]", p.Protocol(), p.Encode()),
		fmt.Sprintf("\t[%02x] 服务器IP地址服务:[%d]", p.FTPAddrLen, p.FTPAddrLen),
		fmt.Sprintf("\t[%x] 服务器IP地址:[%s]", p.FTPAddr, p.FTPAddr),
		fmt.Sprintf("\t[%04x] 服务器端口:[%d]", p.Port, p.Port),
		fmt.Sprintf("\t[%02x] 用户名长度:[%d]", p.UsernameLen, p.UsernameLen),
		fmt.Sprintf("\t[%x] 用户名:[%s]", p.Username, p.Username),
		fmt.Sprintf("\t[%02x] 密码长度:[%d]", p.PasswordLen, p.PasswordLen),
		fmt.Sprintf("\t[%x] 密码:[%s]", p.Password, p.Password),
		fmt.Sprintf("\t[%02x] 文件上传路径长度:[%d]", p.FileUploadPathLen, p.FileUploadPathLen),
		fmt.Sprintf("\t[%x] 文件上传路径:[%s]", p.FileUploadPath, p.FileUploadPath),
		fmt.Sprintf("\t[%02x] 逻辑通道号:[%d]", p.ChannelNo, p.ChannelNo),
		fmt.Sprintf("\t[%012x] 开始时间:[%s]", utils.Time2BCD(p.StartTime), p.StartTime),
		fmt.Sprintf("\t[%012x] 结束时间:[%s]", utils.Time2BCD(p.EndTime), p.EndTime),
		fmt.Sprintf("\t[%016x] 告警标志:[%d]", p.AlarmFlag, p.AlarmFlag),
		fmt.Sprintf("\t[%02x] 音视频类型(媒体类型):[%d] 0-音频和视频 1-音频 2-视频 3-音频或视频", p.MediaType, p.MediaType),
		fmt.Sprintf("\t[%02x] 码流类型:[%d] 0-主或子码流 1-主码流 2-子码流", p.StreamType, p.StreamType),
		fmt.Sprintf("\t[%02x] 存储位置:[%d] 0-主存储器或灾备存储器 1-主存储器 2-灾备存储器", p.MemoryPosition, p.MemoryPosition),
		fmt.Sprintf("\t[%02x] 任务执行条件:[%d]", p.TaskExecuteCondition, p.TaskExecuteCondition),
		"}",
	}, "\n")
}
