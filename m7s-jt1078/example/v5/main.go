package main

import (
	"context"
	"fmt"
	"io"
	"m7s.live/v5"
	_ "m7s.live/v5/plugin/mp4"
	_ "m7s.live/v5/plugin/preview"
	"net/http"
)

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// 回调接口 获取音频端口 用于对讲
		fmt.Println(r.URL.String(), string(body))
		//{
		//	"audioPort": 10005,
		//	"sim": "295696659617",
		//	"channel": 1,
		//	"streamPath": "live/jt1078-295696659617-1"
		//}
	})
	go func() {
		_ = http.ListenAndServe(":10011", nil)
	}()
}

func main() {
	ctx := context.Background()
	// 使用自定义模拟器推流 读取本地文件的
	fmt.Println("preview", "http://127.0.0.1:8080/preview")
	fmt.Println("模拟实时视频流地址", "http://127.0.0.1:8080/mp4/live/jt1078-295696659617-1.mp4")
	fmt.Println("模拟回放视频流地址", "http://127.0.0.1:8080/mp4/live/jt1079-295696659617-1.mp4")
	_ = m7s.Run(ctx, "./config.yaml")
}
