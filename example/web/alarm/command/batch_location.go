package command

import (
	"github.com/cuteLittleDevil/go-jt808/protocol/jt808"
	"github.com/cuteLittleDevil/go-jt808/protocol/model"
)

type BatchLocation struct {
	model.T0x0704
	AlarmLocations []Location
}

func (bl *BatchLocation) Parse(jtMsg *jt808.JTMessage) error {
	if err := bl.T0x0704.Parse(jtMsg); err != nil {
		return err
	}
	for _, item := range bl.T0x0704.Items {
		for additionType, addition := range item.T0x0200AdditionDetails.Additions {
			var (
				location Location
				ok       = true
			)
			switch additionType {
			case 0x64:
				location.T0x0200AdditionExtension0x64.Parse(0x64, addition.Content.Data)
			case 0x65:
				location.T0x0200AdditionExtension0x65.Parse(0x65, addition.Content.Data)
			case 0x66:
				location.T0x0200AdditionExtension0x66.Parse(0x66, addition.Content.Data)
			case 0x67:
				location.T0x0200AdditionExtension0x67.Parse(0x67, addition.Content.Data)
			case 0x70:
				location.T0x0200AdditionExtension0x70.Parse(0x70, addition.Content.Data)
			default:
				ok = false
			}
			if ok {
				bl.AlarmLocations = append(bl.AlarmLocations, location)
			}
		}
	}
	return nil
}
