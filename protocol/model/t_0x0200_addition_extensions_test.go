package model

import (
	"encoding/hex"
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"testing"
)

func TestT0x0200AdditionExtension(t *testing.T) {
	type Handler interface {
		Parse(id uint8, content []byte) (AdditionContent, bool)
		String() string
	}
	type args struct {
		msg string
		Handler
		ID consts.JT808LocationAdditionType
	}
	tests := []struct {
		name string
		args args
		want Handler
	}{
		{
			name: "苏标 0x64",
			args: args{
				msg:     "7E0200407D0201000000000202326095590A4F00002000004C100301E0D7F2073E6EAC0064021400142411271542300104000000CC020201D425040000000030010831010914040000007F150400000001160400000001170200011803000709642F0000001F000201323201000035006401E0D40A073E6AC4241127154210FFFF69643030303033241127154217000500767E",
				Handler: T0x0200AdditionExtension0x64{},
				ID:      0x64,
			},
			want: T0x0200AdditionExtension0x64{
				AlarmID:                        31,
				FlagStatus:                     0,
				AlarmEventType:                 2,
				AlarmLevel:                     1,
				PreVehicleSpeed:                50,
				PreVehicleOrPedestrianDistance: 50,
				DeviationType:                  1,
				RoadSignRecognitionType:        0,
				RoadSignRecognitionData:        0,
				VehicleSpeed:                   53,
				Altitude:                       100,
				Latitude:                       31511562,
				Longitude:                      121531076,
				DateTime:                       "2024-11-27 15:42:10",
				VehicleStatus: T0x0200ExtensionVehicleStatus{
					OriginalValue: 65535,
					ACC:           true,
					LeftTurn:      true,
					RightTurn:     true,
					Wipers:        true,
					Brake:         true,
					Card:          true,
					Location:      true,
				},
				P9208AlarmSign: P9208AlarmSign{
					TerminalID:   "id00003",
					Time:         "2024-11-27 15:42:17",
					SerialNumber: 0,
					AttachNumber: 5,
					AlarmReserve: 0,
				},
			},
		},
		{
			name: "苏标 0x65",
			args: args{
				msg:     "7E0200407D0201000000000202326095590A4F00002000004C100301E0D7F2073E6EAC0064021400142411271542300104000000CC020201D425040000000030010831010914040000007F150400000001160400000001170200011803000709652F0000001F000201323201000035006401E0D40A073E6AC4241127154210FFFF69643030303033241127154217000500777E",
				Handler: T0x0200AdditionExtension0x65{},
				ID:      0x65,
			},
			want: T0x0200AdditionExtension0x65{
				AlarmID:        31,
				FlagStatus:     0,
				AlarmEventType: 2,
				AlarmLevel:     1,
				FatigueLevel:   50,
				Reserved:       [4]byte{50, 1, 0, 0},
				VehicleSpeed:   53,
				Altitude:       100,
				Latitude:       31511562,
				Longitude:      121531076,
				DateTime:       "2024-11-27 15:42:10",
				VehicleStatus: T0x0200ExtensionVehicleStatus{
					OriginalValue: 65535,
					ACC:           true,
					LeftTurn:      true,
					RightTurn:     true,
					Wipers:        true,
					Brake:         true,
					Card:          true,
					Location:      true,
				},
				P9208AlarmSign: P9208AlarmSign{
					TerminalID:   "id00003",
					Time:         "2024-11-27 15:42:17",
					SerialNumber: 0,
					AttachNumber: 5,
					AlarmReserve: 0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.args.msg
			data, _ := hex.DecodeString(msg)
			jtMsg := jt808.NewJTMessage()
			if err := jtMsg.Decode(data); err != nil {
				t.Errorf("T0x0200AdditionExtension = %v", err)
				return
			}
			var t0x0200 T0x0200
			if tt.args.Handler != nil {
				t0x0200.CustomAdditionContentFunc = tt.args.Handler.Parse
			}
			_ = t0x0200.Parse(jtMsg)
			v, ok := t0x0200.Additions[tt.args.ID].Content.CustomValue.(Handler)
			if !ok {
				t.Error("T0x0200AdditionExtension ok = false")
				return
			}
			if v.String() != tt.want.String() {
				t.Errorf("T0x0200AdditionExtension got=\n%s\nwant=\n%s", v.String(), tt.want.String())
				return
			}
		})
	}
}
