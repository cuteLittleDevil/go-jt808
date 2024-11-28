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
				Handler: &T0x0200AdditionExtension0x64{},
				ID:      0x64,
			},
			want: &T0x0200AdditionExtension0x64{
				AlarmID:                        31,
				FlagStatus:                     0,
				AlarmEventType:                 2,
				AlarmLevel:                     1,
				PreVehicleSpeed:                50,
				PreVehicleOrPedestrianDistance: 50,
				DeviationType:                  1,
				RoadSignRecognitionType:        0,
				RoadSignRecognitionData:        0,
				T0x0200ExtensionSBBase: T0x0200ExtensionSBBase{
					VehicleSpeed: 53,
					Altitude:     100,
					Latitude:     31511562,
					Longitude:    121531076,
					DateTime:     "2024-11-27 15:42:10",
					VehicleStatus: T0x0200ExtensionTable18{
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
		},
		{
			name: "苏标 0x65",
			args: args{
				msg:     "7E0200407D0201000000000202326095590A4F00002000004C100301E0D7F2073E6EAC0064021400142411271542300104000000CC020201D425040000000030010831010914040000007F150400000001160400000001170200011803000709652F0000001F000201323201000035006401E0D40A073E6AC4241127154210FFFF69643030303033241127154217000500777E",
				Handler: &T0x0200AdditionExtension0x65{},
				ID:      0x65,
			},
			want: &T0x0200AdditionExtension0x65{
				AlarmID:        31,
				FlagStatus:     0,
				AlarmEventType: 2,
				AlarmLevel:     1,
				FatigueLevel:   50,
				Reserved:       [4]byte{50, 1, 0, 0},
				T0x0200ExtensionSBBase: T0x0200ExtensionSBBase{
					VehicleSpeed: 53,
					Altitude:     100,
					Latitude:     31511562,
					Longitude:    121531076,
					DateTime:     "2024-11-27 15:42:10",
					VehicleStatus: T0x0200ExtensionTable18{
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
		},
		{
			name: "苏标 0x66",
			args: args{
				msg:     "7E0200408101000000000202326095590A4F00002000004C100301E0D7F2073E6EAC0064021400142411271542300104000000CC020201D425040000000030010831010914040000007F15040000000116040000000117020001180300070966310000001F0002013232010000350064020020020201001FFF0000000000003620101010103033241101154217000500000001A77E",
				Handler: &T0x0200AdditionExtension0x66{},
				ID:      0x66,
			},
			want: &T0x0200AdditionExtension0x66{
				AlarmID:    31,
				FlagStatus: 0,
				T0x0200ExtensionSBBase: T0x0200ExtensionSBBase{
					VehicleSpeed: 2,
					Altitude:     306,
					Latitude:     838926336,
					Longitude:    889218050,
					DateTime:     "2000-20-02 02:01:00",
					VehicleStatus: T0x0200ExtensionTable18{
						OriginalValue: 8191,
						ACC:           true,
						LeftTurn:      true,
						RightTurn:     true,
						Wipers:        true,
						Brake:         true,
						Card:          true,
						Location:      true,
					},
					P9208AlarmSign: P9208AlarmSign{
						TerminalID:   "6", // 69643030303033
						Time:         "2020-10-10 10:10:30",
						SerialNumber: 51,
						AttachNumber: 36,
						AlarmReserve: 17,
					},
				},
				AlarmOrEventCount: 1,
				AlarmOrEventList: []T0x0200ExtensionTable22{
					{
						TirePressureAlarmLocation: 21,
						AlarmOrEventType:          16919,
						TirePressure:              5,
						TireTemperature:           0,
						BatteryLevel:              1,
					},
				},
			},
		},
		{
			name: "苏标 0x67",
			args: args{
				msg:     "7E0200407801000000000202326095590A4F00002000004C100301E0D7F2073E6EAC0064021400142411271542300104000000CC020201D425040000000030010831010914040000007F15040000000116040000000117020001180300070967290000001F000201323201000035006701E0241127154210FFFF696430303030332411271542170005003F7E",
				Handler: &T0x0200AdditionExtension0x67{},
				ID:      0x67,
			},
			want: &T0x0200AdditionExtension0x67{
				AlarmID:        31,
				FlagStatus:     0,
				AlarmEventType: 2,
				T0x0200ExtensionSBBase: T0x0200ExtensionSBBase{
					VehicleSpeed: 1,
					Altitude:     12850,
					Latitude:     16777269,
					Longitude:    6750688,
					DateTime:     "2024-11-27 15:42:10",
					VehicleStatus: T0x0200ExtensionTable18{
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
						TerminalID:   "id00003", // 69643030303033
						Time:         "2024-11-27 15:42:17",
						SerialNumber: 0,
						AttachNumber: 5,
						AlarmReserve: 0,
					},
				},
			},
		},
		{
			name: "苏标 0x70",
			args: args{
				msg:     "7E0200408001000000000202326095590A4F00002000004C100301E0D7F2073E6EAC0064021400142411271542300104000000CC020201D425040000000030010831010914040000007F150400000001160400000001170200011803000709702F0000001F000201323201000035006700000100000000E0241127154210FFFF696430303030332411271542170005000000D67E",
				Handler: &T0x0200AdditionExtension0x70{},
				ID:      0x70,
			},
			want: &T0x0200AdditionExtension0x70{
				AlarmID:            31,
				FlagStatus:         0,
				AlarmEventType:     2,
				AlarmTimeThreshold: 306,
				AlarmThreshold1:    12801,
				AlarmThreshold2:    0,
				T0x0200ExtensionSBBase: T0x0200ExtensionSBBase{
					VehicleSpeed: 53,
					Altitude:     103,
					Latitude:     256,
					Longitude:    224,
					DateTime:     "2024-11-27 15:42:10",
					VehicleStatus: T0x0200ExtensionTable18{
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
						TerminalID:   "id00003", // 69643030303033
						Time:         "2024-11-27 15:42:17",
						SerialNumber: 0,
						AttachNumber: 5,
						AlarmReserve: 0,
					},
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
