package attachment

import (
	"bufio"
	"encoding/hex"
	"errors"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"os"
	"testing"
)

func Test_iter_su_biao(t *testing.T) {
	file, _ := os.Open("./testdata/su_biao/file.log")
	scanner := bufio.NewScanner(file)
	fileTexts := make([]string, 0)
	for scanner.Scan() {
		fileTexts = append(fileTexts, scanner.Text())
	}
	_ = file.Close()

	var (
		command7e1210        string
		command7e1210Mistake string
		command7e1211        string
		command7e1212        string
	)
	{
		b, _ := os.ReadFile("./testdata/su_biao/7e1210.log")
		command7e1210 = string(b)
	}
	{
		b, _ := os.ReadFile("./testdata/su_biao/7e1210_mistake.log")
		command7e1210Mistake = string(b)
	}
	{
		b, _ := os.ReadFile("./testdata/su_biao/7e1211.log")
		command7e1211 = string(b)
	}
	{
		b, _ := os.ReadFile("./testdata/su_biao/7e1212.log")
		command7e1212 = string(b)
	}

	tests := []struct {
		name string
		args []string
		want error
	}{
		{
			name: "正常流程",
			args: []string{
				command7e1210,
				fileTexts[0],
				fileTexts[1],
				fileTexts[2],
				command7e1211,
				command7e1212,
			},
			want: nil,
		},
		{
			name: "指令少部分",
			args: []string{
				command7e1210[:2],
				command7e1210[2:],
				fileTexts[0],
				fileTexts[1],
				fileTexts[2],
				command7e1211[:2],
				command7e1211[2:],
				command7e1212[:2],
				command7e1212[2:],
			},
			want: nil,
		},
		{
			name: "文件流少部分",
			args: []string{
				command7e1210,
				fileTexts[0],
				fileTexts[1],
				fileTexts[2][:2],
				fileTexts[2][2:],
				command7e1211,
				command7e1212,
			},
			want: nil,
		},
		{
			name: "数据一起发送了",
			args: []string{
				command7e1210 + fileTexts[0] + fileTexts[1] + fileTexts[2] + command7e1211 + command7e1212,
			},
			want: nil,
		},
		{
			name: "错误的文件名",
			args: []string{
				command7e1210Mistake,
				fileTexts[0],
				fileTexts[1],
				fileTexts[2],
				command7e1211,
				command7e1212,
			},
			want: ErrDataInconsistency,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := &PackageProgress{
				ProgressStage: ProgressStageInit,
				Record:        map[string]*Package{},
				historyData:   make([]byte, 0),
				handle:        newStandardJT808DataHandle(consts.ActiveSafetyJS),
				ExtensionFields: ExtensionFields{
					CurrentPackage:        nil,
					RecentTerminalMessage: nil,
					RecentPlatformData:    nil,
					ActiveSafetyType:      consts.ActiveSafetyJS,
					Err:                   nil,
				},
			}

			fileEventer := newFileEvent()
			for _, v := range tt.args {
				data, _ := hex.DecodeString(v)
				progress.historyData = append(progress.historyData, data...)
				for err := range progress.iter() {
					switch {
					case err == nil:
						if progress.hasJT808Reply() {
							data, err := progress.handle.ReplyData()
							progress.ExtensionFields.RecentPlatformData = data
							progress.ExtensionFields.Err = err
						}
						fileEventer.OnEvent(progress)
					case errors.Is(err, ErrInsufficientDataLen):
					default:
						if !errors.Is(err, tt.want) {
							t.Errorf("iter() error = %v, want = %v", err, tt.want)
						}
					}
				}
			}
		})
	}
}
