package attachment

import (
	"errors"
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"io"
	"net"
	"time"
)

type connection struct {
	conn             net.Conn
	activeSafetyType consts.ActiveSafetyType
	dataHandleFunc   func() DataHandler
	fileEventer      FileEventer
	readTimeout      time.Duration
}

func newConnection(conn net.Conn, activeSafetyType consts.ActiveSafetyType,
	customDataHandleFunc func() DataHandler, fileEventer FileEventer, readTimeout time.Duration) *connection {
	if fileEventer == nil {
		fileEventer = newFileEvent()
	}
	return &connection{
		conn:             conn,
		activeSafetyType: activeSafetyType,
		dataHandleFunc: func() DataHandler {
			if customDataHandleFunc == nil {
				return newStandardJT808DataHandle(activeSafetyType)
			}
			return customDataHandleFunc()
		},
		fileEventer: fileEventer,
		readTimeout: readTimeout,
	}
}

func (c *connection) run() {
	var (
		curData  = make([]byte, 100*1024)
		progress = &PackageProgress{
			ProgressStage: ProgressStageInit,
			Record:        map[string]*Package{},
			handle:        c.dataHandleFunc(),
			ExtensionFields: ExtensionFields{
				ActiveSafetyType: c.activeSafetyType,
				Debug: struct {
					RemoteAddr  string
					NetErr      error
					HistoryData []byte
				}{
					RemoteAddr: c.conn.RemoteAddr().String(),
				},
			},
		}
	)
	defer func() {
		if progress.ExtensionFields.TerminalPhoneNo == "" {
			progress.ProgressStage = ProgressStageUnexpectedExit
		} else {
			if progress.ExtensionFields.Err == nil && !progress.isAllComplete() {
				progress.ExtensionFields.Err = ErrDataInconsistency
			}
			if progress.ExtensionFields.Err != nil {
				progress.ProgressStage = ProgressStageFailQuit
			} else {
				progress.ProgressStage = ProgressStageSuccessQuit
			}
		}
		c.fileEventer.OnEvent(progress)
		clear(curData)
		_ = c.conn.Close()
	}()

	for {
		if c.readTimeout > 0 {
			_ = c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
		}
		n, err := c.conn.Read(curData)
		if err != nil {
			progress.ExtensionFields.Debug.NetErr = err
			progress.ExtensionFields.Debug.HistoryData = progress.historyData
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				return
			}
			progress.ExtensionFields.Err = err
			return
		}
		if n > 0 && !c.handleReadData(progress, curData[:n]) {
			return
		}
	}
}

func (c *connection) handleReadData(progress *PackageProgress, data []byte) bool {
	progress.historyData = append(progress.historyData, data...)
	for err := range progress.iter() {
		switch {
		case err == nil:
			if progress.hasJT808Reply() {
				reply, replyErr := progress.handle.ReplyData()
				progress.ExtensionFields.RecentPlatformData = reply
				if replyErr != nil {
					progress.ExtensionFields.Err = replyErr
					return false
				}
				if _, writeErr := c.conn.Write(reply); writeErr != nil {
					progress.ExtensionFields.Err = writeErr
					return false
				}
			}
			c.fileEventer.OnEvent(progress)
		case errors.Is(err, ErrInsufficientDataLen):
			// 是数据长度不够的错误情况 就结束
			return true
		default:
			progress.ExtensionFields.Err = err
			return false
		}
	}
	return true
}
