package router

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"web/internal/shared"
	"web/service/conf"
	"web/service/record"
)

func Register(h *server.Hertz) {
	group := h.Group("/api/v1/jt808/")
	{
		group.POST("/8103", p8103)
		group.POST("/8104", p8104)
		group.POST("/8201", p8201)
		group.POST("/8202", p8202)
		group.POST("/8300", p8300)
		group.POST("/8302", p8302)
		group.POST("/8801", p8801)
		group.POST("/9101", p9101)
		group.POST("/9102", p9102)
		group.POST("/9201", p9201)
		group.POST("/9202", p9202)
		group.POST("/9205", p9205)
		group.POST("/9206", p9206)
		group.POST("/9208", p9208)
		group.GET("/details", details)
		group.GET("/images", images)
	}

	apiRtvsV1 := h.Group("/api/")
	{
		apiRtvsV1.GET("/VideoControl", videoControl)
		apiRtvsV1.POST("/WCF0x9105", wcf0x9105)
	}
}

func details(_ context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, shared.Response{
		Code: http.StatusOK,
		Msg:  "查询终端详情",
		Data: record.Details(),
	})
}

func images(_ context.Context, ctx *app.RequestContext) {
	name := ctx.DefaultQuery("name", "")
	type Response struct {
		IsLocal  bool   `json:"isLocal"`
		LocalURL string `json:"localURL"`
		IsMinio  bool   `json:"isMinio"`
		MinioURL string `json:"minioURL"`
	}
	cameraConfig := conf.GetData().FileConfig.CameraConfig
	response := Response{
		IsLocal:  cameraConfig.Enable,
		LocalURL: "",
		IsMinio:  cameraConfig.MinioConfig.Enable,
		MinioURL: "",
	}

	if response.IsLocal {
		_ = filepath.WalkDir(cameraConfig.Dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if strings.Index(d.Name(), name) >= 0 && strings.Index(d.Name(), ".txt") < 0 {
				response.LocalURL = cameraConfig.URLPrefix + d.Name()
				return filepath.SkipAll
			}
			return nil
		})
	}

	if response.IsMinio {
		_ = filepath.WalkDir(cameraConfig.MinioDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if strings.Index(d.Name(), name) >= 0 && strings.HasSuffix(d.Name(), ".txt") {
				b, _ := os.ReadFile(path)
				response.MinioURL = string(b)
				return filepath.SkipAll
			}
			return nil
		})
	}

	ctx.JSON(http.StatusOK, shared.Response{
		Code: http.StatusOK,
		Msg:  "查询终端图片",
		Data: response,
	})
}
