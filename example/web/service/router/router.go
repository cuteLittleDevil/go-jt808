package router

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"net/http"
	"web/internal/shared"
	"web/service/record"
)

func Register(h *server.Hertz) {
	group := h.Group("/api/v1/jt808/")
	{
		group.POST("/8103", p8103)
		group.POST("/8104", p8104)
		group.POST("/8801", p8801)
		group.POST("/9101", p9101)
		group.POST("/9102", p9102)
		group.POST("/9201", p9201)
		group.POST("/9202", p9202)
		group.POST("/9205", p9205)
		group.POST("/9206", p9206)
		group.POST("/9208", p9208)
		group.GET("/details", details)
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
