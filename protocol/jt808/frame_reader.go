package jt808

import "bytes"

// FrameSign 帧标识位，JT808 规定每帧以 0x7e 开头和结尾.
const FrameSign byte = 0x7e

// FrameReader 从 TCP 字节流中切出完整的 JT808 帧.
//
// 一帧的形态：[0x7e] + 消息体 + [0x7e]，本类型只负责找边界，不做转义/校验/解码.
//
// 推荐使用（支持中途因解码失败而停止）：
//
//	if frame, ok := r.FeedSingleComplete(data); ok {
//	    // 处理 frame
//	} else {
//	    r.Append(data)
//	    for {
//	        frame, ok := r.PopFrame()
//	        if !ok {
//	            break
//	        }
//	        // 处理 frame；若解码失败可直接 return，未取出的数据仍留在缓冲区
//	    }
//	}
//
// 零拷贝约定：
//   - 返回的 frame 是 data 或内部 historyData 的视图，不会复制字节
//   - 须在下次 conn.Read / Append 之前同步处理完
//   - 若要异步保存或长期持有，调用方自行 copy
type FrameReader struct {
	// historyData 粘包缓冲区，存放尚未凑齐一帧的字节
	historyData []byte
}

// NewFrameReader 创建帧读取器.
func NewFrameReader() *FrameReader {
	return &FrameReader{
		historyData: make([]byte, 0),
	}
}

// Clear 清空粘包缓冲区.
func (r *FrameReader) Clear() {
	clear(r.historyData)
	r.historyData = r.historyData[:0]
}

// Pending 返回粘包缓冲区中的字节数（即尚未组成完整帧的数据量）.
func (r *FrameReader) Pending() int {
	return len(r.historyData)
}

// FeedSingleComplete 快速路径：缓冲区为空且 data 本身恰好是一帧时，直接返回 data.
//
// 典型场景是单次 conn.Read 读到完整单包，可避免写入 historyData.
// 返回的 frame 与 data 共享底层数组，零拷贝.
func (r *FrameReader) FeedSingleComplete(data []byte) (frame []byte, ok bool) {
	const sign = FrameSign
	if len(r.historyData) != 0 || len(data) <= 2 || data[0] != sign || data[len(data)-1] != sign {
		return nil, false
	}
	count := 0
	index := bytes.IndexFunc(data, func(b rune) bool {
		if byte(b) == sign {
			count++
		}
		return count == 2
	})
	if index != len(data)-1 {
		return nil, false
	}
	return data, true
}

// Append 将本次读到的数据追加到粘包缓冲区.
func (r *FrameReader) Append(data []byte) {
	if len(data) == 0 {
		return
	}
	r.historyData = append(r.historyData, data...)
}

// TryPopFrame 从 data 头部尝试取出一完整帧；不够一帧时 ok=false，rest 为原 data.
//
// 取出的 frame 与 rest 均为 data 的子切片（零拷贝），便于在无状态缓冲场景复用切帧逻辑.
func TryPopFrame(data []byte) (frame, rest []byte, ok bool) {
	const sign = FrameSign
	end := -1
	if len(data) > 2 && data[0] == sign {
		for i := 1; i < len(data); i++ {
			if data[i] == sign {
				end = i + 1
				break
			}
		}
	}
	if end == -1 {
		return nil, data, false
	}
	return data[:end], data[end:], true
}

// PopFrame 从粘包缓冲区取出一帧；数据不够一帧时返回 ok=false.
//
// 取出的 frame 是 historyData 的子切片，随后会前移缓冲区偏移，零拷贝.
func (r *FrameReader) PopFrame() (frame []byte, ok bool) {
	frame, rest, ok := TryPopFrame(r.historyData)
	if !ok {
		return nil, false
	}
	if len(rest) == 0 {
		r.historyData = r.historyData[0:0]
	} else {
		r.historyData = rest
	}
	return frame, true
}

// ReadFrames 追加 data 并一次性取出当前所有完整帧.
//
// 便捷方法，适合测试或确定不会因解码失败而中途停止的场景.
// 生产环境建议使用 FeedSingleComplete + Append + PopFrame 组合.
func (r *FrameReader) ReadFrames(data []byte) [][]byte {
	if frame, ok := r.FeedSingleComplete(data); ok {
		return [][]byte{frame}
	}
	r.Append(data)
	var frames [][]byte
	for {
		frame, ok := r.PopFrame()
		if !ok {
			break
		}
		frames = append(frames, frame)
	}
	return frames
}
