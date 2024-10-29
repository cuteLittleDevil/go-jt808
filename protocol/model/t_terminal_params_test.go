package model

import (
	"encoding/hex"
	"errors"
	"github.com/cuteLittleDevil/go-jt808/protocol"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"io"
	"os"
	"testing"
)

func TestTerminalParamDetails(t *testing.T) {
	type args struct {
		msg                  string
		paramParseBeforeFunc func(id uint32, content []byte)
	}
	type want struct {
		path string
		err  error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "包含全部已知附加信息的",
			args: args{
				msg:                  "7E010443A20100000000014419999999000500045B00000001040000000A00000002040000003C00000003040000000200000004040000003C00000005040000000200000006040000003C000000070400000002000000100B31333031323334353637300000001105313233343500000012053132333435000000130E3132372E302E302E313A37303030000000140531323334350000001505313233343500000016053132333435000000170531323334350000001A093132372E302E302E310000001B04000004570000001C04000004580000001D093132372E302E302E310000002004000000000000002104000000000000002204000000000000002301300000002401300000002501300000002601300000002704000000000000002804000000000000002904000000000000002C04000003E80000002D04000003E80000002E04000003E80000002F04000003E800000030040000000A0000003102003C000000320416320A1E000000400B3133303132333435363731000000410B3133303132333435363732000000420B3133303132333435363733000000430B3133303132333435363734000000440B3133303132333435363735000000450400000001000000460400000000000000470400000000000000480B3133303132333435363738000000490B313330313233343536373900000050040000000000000051040000000000000052040000000000000053040000000000000054040000000000000055040000003C000000560400000014000000570400003840000000580400000708000000590400001C200000005A040000012C0000005B0200500000005C0200050000005D02000A0000005E02001E00000064040000000100000065040000000100000070040000000100000071040000006F000000720400000070000000730400000071000000740400000072000000751500030190320000002800030190320000002800050100000076130400000101000002020000030300000404000000000077160101000301F43200000028000301F43200000028000500000079032808010000007A04000000230000007B0232320000007C1405000000000000000000000000000000000000000000008004000000240000008102000B000000820200660000008308BEA9415830303031000000840101000000900102000000910101000000920101000000930400000001000000940100000000950400000001000001000400000064000001010213880000010204000000640000010302138800000110080000000000000101F07E",
				paramParseBeforeFunc: func(id uint32, content []byte) {},
			},
			want: want{
				path: "./testdata/0x0200_terminal_param_1.txt",
				err:  nil,
			},
		},
		{
			name: "参数长度不符合 0x001 uint32",
			args: args{
				msg:                  "7E0104000C0144199999990005000401000000010000000001CC7E",
				paramParseBeforeFunc: nil,
			},
			want: want{
				path: "",
				err:  protocol.ErrBodyLengthInconsistency,
			},
		},
		{
			name: "参数长度不符合 0x031 uint16",
			args: args{
				msg:                  "7E0104000C0144199999990005000401000000310000000001fc7E",
				paramParseBeforeFunc: nil,
			},
			want: want{
				path: "",
				err:  protocol.ErrBodyLengthInconsistency,
			},
		},
		{
			name: "参数长度不符合 0x084 byte",
			args: args{
				msg:                  "7E0104000C0144199999990005000401000000840000000001497E",
				paramParseBeforeFunc: nil,
			},
			want: want{
				path: "",
				err:  protocol.ErrBodyLengthInconsistency,
			},
		},
		{
			name: "参数长度不符合 0x032 [4]byte",
			args: args{
				msg:                  "7E0104000C0144199999990005000401000000320000000001ff7E",
				paramParseBeforeFunc: nil,
			},
			want: want{
				path: "",
				err:  protocol.ErrBodyLengthInconsistency,
			},
		},
		{
			name: "参数长度不符合 0x110 [8]byte",
			args: args{
				msg:                  "7E0104000C0144199999990005000401000001100000000001DC7E",
				paramParseBeforeFunc: nil,
			},
			want: want{
				path: "",
				err:  protocol.ErrBodyLengthInconsistency,
			},
		},
		{
			name: "参数不一致",
			args: args{
				msg:                  "7E0104000C0144199999990005000402000000010400000001Cb7E",
				paramParseBeforeFunc: nil,
			},
			want: want{
				path: "",
				err:  protocol.ErrBodyLengthInconsistency,
			},
		},
		{
			name: "body数据缺失",
			args: args{
				msg:                  "7E01040007014419999999000500040200000001C57E",
				paramParseBeforeFunc: nil,
			},
			want: want{
				path: "",
				err:  protocol.ErrBodyLengthInconsistency,
			},
		},
		{
			name: "len长度异常",
			args: args{
				msg:                  "7E0104000C0144199999990005000401000000010F00000001C37E",
				paramParseBeforeFunc: nil,
			},
			want: want{
				path: "",
				err:  protocol.ErrBodyLengthInconsistency,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.args.msg
			data, _ := hex.DecodeString(msg)
			jtMsg := jt808.NewJTMessage()
			if err := jtMsg.Decode(data); err != nil {
				t.Errorf("TerminalParamDetails = %v", err)
				return
			}

			var t0x0104 T0x0104
			if tt.args.paramParseBeforeFunc != nil {
				t0x0104.ParamParseBeforeFunc = tt.args.paramParseBeforeFunc
			}
			if err := t0x0104.Parse(jtMsg); err != nil {
				if !errors.Is(err, tt.want.err) {
					t.Errorf("TerminalParamDetails = %v, want %v", err, tt.want.err)
				}
				return
			}
			got := t0x0104.TerminalParamDetails.String()
			txt := tt.want.path
			f, err := os.Open(txt)
			if err != nil {
				_ = os.WriteFile(txt, []byte(got), os.ModePerm)
			}
			if wantData, _ := io.ReadAll(f); string(wantData) != got {
				_ = os.WriteFile(txt+".tmp", []byte(got), os.ModePerm)
				t.Errorf("TerminalParamDetails =\n%s\n want %s", got, string(wantData))
				return
			}
		})
	}
}
