package model

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
)

func TestT0x0200EncodeAdditions(t *testing.T) {
	loc := T0x0200{
		T0x0200LocationItem: T0x0200LocationItem{DateTime: "2024-10-01 00:00:00"},
		T0x0200AdditionDetails: T0x0200AdditionDetails{
			Additions: map[consts.JT808LocationAdditionType]Addition{
				consts.A0x31GNSSPositionNum: {
					Len:     1,
					Content: AdditionContent{GNSSPositionNum: 7},
				},
				consts.A0x01Mile: {
					Len:     4,
					Content: AdditionContent{Mile: 11},
				},
				consts.A0x11OverSpeedAlarm: {
					Len: 5,
					Content: AdditionContent{OverSpeedAlarm: AdditionOverSpeedAlarm{
						LocationType: 1,
						AreaID:       0x12345678,
					}},
				},
			},
		},
	}
	got := loc.Encode()
	wantTail := []byte{
		0x01, 0x04, 0x00, 0x00, 0x00, 0x0b,
		0x11, 0x05, 0x01, 0x12, 0x34, 0x56, 0x78,
		0x31, 0x01, 0x07,
	}
	if len(got) != 28+len(wantTail) || !bytes.Equal(got[28:], wantTail) {
		t.Fatalf("additions\n got %x\nwant %x", got[28:], wantTail)
	}

	loc.Additions = map[consts.JT808LocationAdditionType]Addition{
		consts.A0x01Mile: {
			Len:     2,
			Content: AdditionContent{Data: []byte{0x0a, 0x0b}, Mile: 11},
		},
	}
	got = loc.Encode()
	wantTail = []byte{0x01, 0x02, 0x0a, 0x0b}
	if !bytes.Equal(got[28:], wantTail) {
		t.Fatalf("data preference\n got %x\nwant %x", got[28:], wantTail)
	}

	loc.Additions = map[consts.JT808LocationAdditionType]Addition{
		consts.A0x01Mile: {Len: 2, Content: AdditionContent{Mile: 0x01020304}},
	}
	got = loc.Encode()
	wantTail = []byte{0x01, 0x02, 0x01, 0x02, 0x03, 0x04}
	if !bytes.Equal(got[28:], wantTail) {
		t.Fatalf("len kept\n got %x\nwant %x", got[28:], wantTail)
	}

	loc.Additions = nil
	if got = loc.Encode(); len(got) != 28 {
		t.Fatalf("no additions len = %d, want 28", len(got))
	}
}

func TestT0x0200EncodeAdditionFields(t *testing.T) {
	tire := make([]byte, 30)
	tire[1] = 0x37
	loc := T0x0200{
		T0x0200LocationItem: T0x0200LocationItem{DateTime: "2024-10-01 00:00:00"},
		T0x0200AdditionDetails: T0x0200AdditionDetails{
			Additions: map[consts.JT808LocationAdditionType]Addition{
				0x03: {Len: 2, Content: AdditionContent{Speed: 0x21}},
				0x04: {Len: 2, Content: AdditionContent{ManualAlarm: 0x2c}},
				0x05: {Len: 30, Content: AdditionContent{TirePressure: AdditionTirePressure{Values: map[uint8]uint8{1: 0x37, 40: 1}}}},
				0x06: {Len: 2, Content: AdditionContent{CarTemperature: 0x0102}},
				0x11: {Len: 1, Content: AdditionContent{OverSpeedAlarm: AdditionOverSpeedAlarm{}}},
				0x12: {Len: 6, Content: AdditionContent{AreaAlarm: AdditionAreaAlarm{LocationType: 1, AreaID: 0x4d, Direction: 1}}},
				0x13: {Len: 7, Content: AdditionContent{DrivingTimeInsufficientAlarm: AdditionDrivingTimeInsufficientAlarm{
					RoadSectionID: 0x58, RoadSectionDrivingTimeSecond: 0x5858, Result: 1,
				}}},
				0x25: {Len: 4, Content: AdditionContent{ExtendVehicleStatus: AdditionExtendVehicleStatus{Value: 0x63}}},
				0x2A: {Len: 2, Content: AdditionContent{IOStatus: AdditionIOStatus{Value: 0x0a}}},
				0x2B: {Len: 4, Content: AdditionContent{Analog: 0x14}},
				0x33: {Len: 1},
			},
		},
	}
	want := []byte{0x03, 0x02, 0x00, 0x21, 0x04, 0x02, 0x00, 0x2c, 0x05, 0x1e}
	want = append(want, tire...)
	want = append(want,
		0x06, 0x02, 0x01, 0x02,
		0x11, 0x01, 0x00,
		0x12, 0x06, 0x01, 0x00, 0x00, 0x00, 0x4d, 0x01,
		0x13, 0x07, 0x00, 0x00, 0x00, 0x58, 0x58, 0x58, 0x01,
		0x25, 0x04, 0x00, 0x00, 0x00, 0x63,
		0x2a, 0x02, 0x00, 0x0a,
		0x2b, 0x04, 0x00, 0x00, 0x00, 0x14,
		0x33, 0x01,
	)
	got := loc.Encode()
	if len(got) < 28 || !bytes.Equal(got[28:], want) {
		t.Fatalf("fields\n got %x\nwant %x", got[28:], want)
	}
}

func TestT0x0200EncodeAdditionsRoundTrip(t *testing.T) {
	msg := "7e020000800123456789017fff000004000000080006eeb6ad02633df701380003006320070719235901040000000b02020016030200210402002c051e3737370000000000000000000000000000000000000000000000000000001105420000004212064d0000004d4d1307000000580058582504000000632a02000a2b040000001430011e31012806020001927e"
	data, err := hex.DecodeString(msg)
	if err != nil {
		t.Fatal(err)
	}
	jtMsg := jt808.NewJTMessage()
	if err = jtMsg.Decode(data); err != nil {
		t.Fatal(err)
	}
	var loc T0x0200
	if err = loc.Parse(jtMsg); err != nil {
		t.Fatal(err)
	}
	encoded := loc.Encode()
	mile := loc.Additions[consts.A0x01Mile].Content.Mile
	oil := loc.Additions[consts.A0x02Oil].Content.Oil
	temp := loc.Additions[consts.A0x06CarTemperature].Content.CarTemperature
	area := loc.Additions[consts.A0x12AreaAlarm].Content.AreaAlarm.AreaID
	var again T0x0200
	if err = again.Parse(&jt808.JTMessage{Body: encoded}); err != nil {
		t.Fatal(err)
	}
	if again.Additions[consts.A0x01Mile].Content.Mile != mile ||
		again.Additions[consts.A0x02Oil].Content.Oil != oil ||
		again.Additions[consts.A0x06CarTemperature].Content.CarTemperature != temp ||
		again.Additions[consts.A0x12AreaAlarm].Content.AreaAlarm.AreaID != area {
		t.Fatalf("reparse mismatch: %+v", again.Additions)
	}
	id06 := bytes.Index(encoded[28:], []byte{0x06, 0x02})
	id11 := bytes.Index(encoded[28:], []byte{0x11, 0x05})
	if id06 < 0 || id11 < 0 || id06 > id11 {
		t.Fatalf("addition order id06=%d id11=%d body=%x", id06, id11, encoded[28:])
	}
}

func TestT0x0704EncodeAdditions(t *testing.T) {
	loc := T0x0200{
		T0x0200LocationItem: T0x0200LocationItem{DateTime: "2024-10-01 00:00:00"},
		T0x0200AdditionDetails: T0x0200AdditionDetails{
			Additions: map[consts.JT808LocationAdditionType]Addition{
				consts.A0x30WIFISignalStrength: {Len: 1, Content: AdditionContent{WIFISignalStrength: 0x1f}},
				consts.A0x02Oil:                {Len: 2, Content: AdditionContent{Oil: 0x16}},
			},
		},
	}
	body := loc.Encode()
	raw := binary.BigEndian.AppendUint16(nil, 1)
	raw = append(raw, 0x01)
	raw = binary.BigEndian.AppendUint16(raw, uint16(len(body)))
	raw = append(raw, body...)

	var batch T0x0704
	if err := batch.Parse(&jt808.JTMessage{Body: raw}); err != nil {
		t.Fatal(err)
	}
	batch.Num = 9
	batch.Items[0].Len = 28
	got := batch.Encode()
	if binary.BigEndian.Uint16(got[:2]) != 9 {
		t.Fatalf("Num = %d, want 9", binary.BigEndian.Uint16(got[:2]))
	}
	if got[2] != 0x01 {
		t.Fatalf("LocationType = %d", got[2])
	}
	if itemLen := binary.BigEndian.Uint16(got[3:5]); itemLen != uint16(len(body)) {
		t.Fatalf("item len = %d, want %d", itemLen, len(body))
	}
	if !bytes.Equal(got[5:], body) {
		t.Fatalf("item\n got %x\nwant %x", got[5:], body)
	}
}

func TestT0x0104Encode(t *testing.T) {
	reply := T0x0104{
		RespondSerialNumber: 4,
		RespondParamCount:   2,
		TerminalParamDetails: TerminalParamDetails{
			T0x001HeartbeatInterval: ParamContent[uint32]{ID: 0x001, Len: 4, Value: 10},
		},
	}
	got := reply.Encode()
	want := []byte{0x00, 0x04, 0x02, 0x00, 0x00, 0x00, 0x01, 0x04, 0x00, 0x00, 0x00, 0x0a}
	if !bytes.Equal(got, want) {
		t.Fatalf("got %x\nwant %x", got, want)
	}
}
