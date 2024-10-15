package shared

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	SubLocation      = "0x0200"
	SubLocationBatch = "0x0704"
)

type Location struct {
	Phone string `json:"phone"`
	// Latitude 纬度 以度为单位的纬度值乘以 10 的 6 次方，精确到百万分之一度
	Latitude uint32 `json:"latitude"`
	// Longitude 经度 以度为单位的经度值乘以 10 的 6 次方，精确到百万分之一度
	Longitude      uint32 `json:"longitude"`
	TimestampMicro int64  `json:"timestamp"` // 微🐱
}

type LocationBatch struct {
	Locations []*Location `json:"locations"`
}

func NewLocationBatch(locations ...*Location) *LocationBatch {
	return &LocationBatch{Locations: locations}
}

func NewLocation(phone string, latitude uint32, longitude uint32) *Location {
	timestamp := time.Now().UnixMicro()
	return &Location{Phone: phone, Latitude: latitude, Longitude: longitude, TimestampMicro: timestamp}
}

func (l *Location) Encode() []byte {
	return []byte(fmt.Sprintf("%s,%d,%d,%d", l.Phone, l.Latitude, l.Longitude, l.TimestampMicro))
}

func (l *Location) Decode(data []byte) error {
	strs := strings.Split(string(data), ",")
	if len(strs) != 4 {
		return fmt.Errorf(fmt.Sprintf("invalid location data [%s]", string(data)))
	}
	l.Phone = strs[0]
	lat, _ := strconv.ParseInt(strs[1], 10, 32)
	lon, _ := strconv.ParseInt(strs[2], 10, 32)
	l.Latitude = uint32(lat)
	l.Longitude = uint32(lon)
	micro, _ := strconv.ParseInt(strs[3], 10, 64)
	l.TimestampMicro = micro
	return nil
}

func (lb *LocationBatch) Encode() []byte {
	b, _ := json.Marshal(lb)
	return b
}

func (lb *LocationBatch) Decode(data []byte) error {
	return json.Unmarshal(data, lb)
}
