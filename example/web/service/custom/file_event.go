package custom

import (
	"fmt"
	"github.com/cuteLittleDevil/go-jt808/attachment"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
	"os"
	"strings"
)

type fileEvent struct {
	dir  string
	file *os.File
}

func NewFileEvent(dir string, logFile string) attachment.FileEventer {
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	return &fileEvent{dir: dir, file: f}
}

func (f *fileEvent) OnEvent(progress *attachment.PackageProgress) {
	str := fmt.Sprintf("当前进度: [%s] ", progress.ProgressStage.String())
	defer func() {
		if f.file != nil {
			_, _ = f.file.WriteString(str + "\n")
			_ = f.file.Sync()
		}
	}()
	extension := progress.ExtensionFields
	switch progress.ProgressStage {
	case attachment.ProgressStageInit:
		var t0x1210 model.T0x1210
		_ = t0x1210.Parse(extension.RecentTerminalMessage)
		str += strings.Join([]string{
			t0x1210.String(),
			fmt.Sprintf("\t平台回复的 [%x]", extension.RecentPlatformData),
			"\n",
		}, "\n")
	case attachment.ProgressStageStart:
		var t0x1211 model.T0x1211
		_ = t0x1211.Parse(extension.RecentTerminalMessage)
		str += strings.Join([]string{
			t0x1211.String(),
			fmt.Sprintf("\t平台回复的 [%x]", extension.RecentPlatformData),
			"",
		}, "\n")
	case attachment.ProgressStageStreamData:
		curPack := extension.CurrentPackage
		str += fmt.Sprintf(" 文件传输中[%s] 进度[%d/%d] 偏移[%d]", curPack.FileName,
			curPack.CurrentSize, curPack.FileSize, curPack.Offset)
	case attachment.ProgressStageSupplementary:
		curPack := extension.CurrentPackage
		str += fmt.Sprintf(" 文件补传传输中[%s] 进度[%d/%d] 偏移[%d]", curPack.FileName,
			curPack.CurrentSize, curPack.FileSize, curPack.Offset)
	case attachment.ProgressStageStreamDataComplete:
		str += " 目前传输文件整体进度:\n"
		for name, v := range progress.Record {
			str += fmt.Sprintf("name=[%s] progres=[%d/%d]\n", name, v.CurrentSize, v.FileSize)
		}
	case attachment.ProgressStageComplete:
		var t0x1212 model.T0x1212
		_ = t0x1212.Parse(extension.RecentTerminalMessage)
		curPack := extension.CurrentPackage
		str += strings.Join([]string{
			t0x1212.String(),
			fmt.Sprintf("本次上传的文件[%s] [%d/%d]", // 放这里保存文件也行
				curPack.FileName, curPack.CurrentSize, curPack.FileSize),
			fmt.Sprintf("\t平台回复的 [%x]", extension.RecentPlatformData),
			"",
		}, "\n")
	case attachment.ProgressStageFailQuit:
		str += fmt.Sprintf(" 文件传输异常 [%v]", extension.Err)
	case attachment.ProgressStageSuccessQuit:
		str += fmt.Sprintf(" 文件传输成功 开始保存 保存数量[%d]\n", len(progress.Record))
		_ = os.MkdirAll(f.dir, os.ModePerm)
		for name, pack := range progress.Record {
			savePath := fmt.Sprintf("./%s/%s", f.dir, name)
			err := os.WriteFile(savePath, pack.StreamBody, os.ModePerm)
			str += fmt.Sprintf("保存文件[%s] 文件大小[%d byte] 保存情况[%v]\n",
				savePath, len(pack.StreamBody), err)
		}
	}
}
