package attachment

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"os"
	"strings"
)

type fileEvent struct {
}

func newFileEvent() *fileEvent {
	return &fileEvent{}
}

func (f *fileEvent) OnEvent(progress *PackageProgress) {
	str := fmt.Sprintf("当前进度: [%s] ", progress.ProgressStage.String())
	defer func() {
		fmt.Println(str)
	}()
	extension := progress.ExtensionFields
	switch progress.ProgressStage {
	case ProgressStageInit:
		var t0x1210 model.T0x1210
		_ = t0x1210.Parse(extension.RecentTerminalMessage)
		str += strings.Join([]string{
			t0x1210.String(),
			fmt.Sprintf("\t平台回复的 [%x]", extension.RecentPlatformData),
			"\n",
		}, "\n")
	case ProgressStageStart:
		var t0x1211 model.T0x1211
		_ = t0x1211.Parse(extension.RecentTerminalMessage)
		str += strings.Join([]string{
			t0x1211.String(),
			fmt.Sprintf("\t平台回复的 [%x]", extension.RecentPlatformData),
			"",
		}, "\n")
	case ProgressStageStreamData:
		str += fmt.Sprintf(" 文件传输中[%s] 进度[%d/%d]", extension.CurrentPackage.FileName,
			extension.CurrentPackage.CurrentSize, extension.CurrentPackage.FileSize)
	case ProgressStageSupplementary:
		str += fmt.Sprintf(" 文件补传传输中[%s] 进度[%d/%d]", extension.CurrentPackage.FileName,
			extension.CurrentPackage.CurrentSize, extension.CurrentPackage.FileSize)
	case ProgressStageStreamDataComplete:
		str += " 目前传输文件整体进度:\n"
		for name, v := range progress.Record {
			str += fmt.Sprintf("name=[%s] progres=[%d/%d]\n", name, v.CurrentSize, v.FileSize)
		}
	case ProgressStageComplete:
		var t0x1212 model.T0x1212
		_ = t0x1212.Parse(extension.RecentTerminalMessage)
		str += strings.Join([]string{
			t0x1212.String(),
			fmt.Sprintf("\t平台回复的 [%x]", extension.RecentPlatformData),
			"",
		}, "\n")
	case ProgressStageFailQuit:
		str += fmt.Sprintf(" 文件传输异常 [%v]", extension.Err)
	case ProgressStageSuccessQuit:
		phone := progress.ExtensionFields.RecentTerminalMessage.Header.TerminalPhoneNo
		str += fmt.Sprintf(" 文件传输成功 开始保存 保存数量[%d]\n", len(progress.Record))
		for name, pack := range progress.Record {
			savePath := fmt.Sprintf("./%s/%s", phone, name)
			err := os.WriteFile(savePath, pack.StreamBody, os.ModePerm)
			str += fmt.Sprintf("保存文件[%s] 文件大小[%d byte] 保存情况[%v]\n",
				savePath, len(pack.StreamBody), err)
		}
	}
}
