/*
Package service 实现 JT/T 808 TCP 服务端：连接生命周期、无锁会话管理与平台主动下发。

# 设计不变量

  - 进程常驻：GoJT808.Run 不提供 Shutdown，生命周期等于进程。
  - 会话表单协程：join/leave/write 经 operationFuncChan 串行，无互斥锁。
  - 主动下发与会话查找原子：sessionManager.write 在 actor 闭包内投递到连接队列，
    再由调用方阻塞等待 replyChan；业务为一问一答，不推荐对同一终端高频连发。
  - 读循环同步：OnReadExecutionEvent 在下次 conn.Read 前完成，高频热路径零拷贝；
    业务如需异步持有原始 Body/TerminalData，必须自行深拷贝。
  - 写协程单写者：record、platformSerialNumber 与超时错误仅由 write() 协程修改；
    OnWriteExecutionEvent 与自定义 ReplyBody 不应读取高频上行未被拷贝的 Body/TerminalData。

# 连接并发模型与数据流

	                         SendActiveMessage(key)
	                                   │
	                                   ▼
	                      ┌─────────────────────────┐
	                      │   sessionManager.run()  │  单协程管理 map[key]*session
	                      │  join / leave / write   │  首包保存独立深拷贝 Header
	                      └────────────┬────────────┘
	                                   │ activeMsgChan
	                                   ▼
	┌───────────────┐  uplink  ┌─────────────────────────┐  timeout seq  ┌────────────────┐
	│   reader()    ├─────────►│         write()         │◄──────────────┤ time.AfterFunc │
	│  (1023B 复用) │  reissue │  record[seq]*activeRec  │               │  (仅传 uint16) │
	└───────┬───────┘          └────────────┬────────────┘               └───────▲────────┘
	        │                               │                                    │
	        │ EOF/超时 → defer stop()       │ 退出时 defer close(writeDoneChan)   │
	        │  1. onLeaveEvent(key)         └────────────────────────────────────┘
	        │  2. conn.Close()              ※ activeTimeoutChan 永不 close，避免关闭竞态 panic
	        │  3. close(stopWriteChan) ────► write() 执行 shutdownWrite() 后返回
	        │  4. <-writeDoneChan      ◄──── 等待收尾完成（无在途请求时立即返回）

# 热路径零拷贝与按需拷贝边界

	[TCP Read → 1023B data 缓冲复用]
	        │
	        ├─ 首包 joinHandle()：深拷贝 1 次独立 Header（含 bcdTerminalPhoneNo），
	        │                     同时存入 sessionManager 与 connection.sessionHeader，
	        │                     后续所有下行默认回复/补传复用该 Header，不再碰读缓冲。
	        │
	        ├─ 高频上行（0x0200 位置 / 0x0002 心跳 / 0x0704 批量位置）：
	        │   100% 零拷贝 → OnReadExecutionEvent 同步消费 → write() 仅用 Header 字段回 0x8001。
	        │
	        └─ 低频按需拷贝（copyOnDemandForWrite）：
	            1) activePending > 0 且命中 activeRespondHandles（主动下发应答）：
	               拷贝 Body + TerminalData（activePending.Add(1) happens-before conn.Write）。
	            2) 内置 ReplyBody 会解析 Body 的指令（0x0102 / 0x0801 单包 / 0x1212）：
	               拷贝 Body。

# 断线关闭时序（方案 B：保留在途主动下发等待超时）

	reader() 退出             sessionManager           write()
	   │                            │                     │
	   ├── onLeaveEvent(key) ──────►│ delete(record, key) │  (之后不再有新 activeMsg 入队)
	   │◄────────── 返回 ───────────┤                     │
	   ├── conn.Close()             │                     │
	   ├── close(stopWriteChan) ────┼────────────────────►│ 进入 shutdownWrite():
	   │                            │                     │  1. 匹配已入队但未消费的终端应答
	   │                            │                     │  2. 排空未发出 activeMsg → ErrWriteDataFail
	   │                            │                     │  3. 等待已发出在途 record 超时 → ErrWriteDataOverTime
	   │◄─────── close(writeDoneChan) ────────────────────┤ defer 停定时器并退出
*/
package service
