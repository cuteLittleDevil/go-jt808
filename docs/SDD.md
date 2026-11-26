# go-jt808 软件设计说明（SDD）

| 项 | 内容 |
| --- | --- |
| 文档版本 | 1.2 |
| 依据代码 | `github.com/cuteLittleDevil/go-jt808`（本地路径分析） |
| Go 版本 | 1.23.2+ |
| 许可证 | MIT |
| 生成方式 | 从现有源码反向提炼（as-built SDD） |

---

## 1. 文档目的与范围

### 1.1 目的

本文档描述 go-jt808 的系统边界、模块职责、关键功能规格（输入/输出/行为）、扩展点与验收标准，用于：

- 二次开发对接（平台服务、模拟器、附件服务、GB28181 互通）
- 协议实现对照与回归验收
- 功能扩展时不侵入核心路径

### 1.2 范围

**在范围内：**

| 包 | 职责 |
| --- | --- |
| `service` | JT808 TCP 服务端：连接、会话、上下行消息、主动下发 |
| `protocol` | JT808/JT1078 帧编解码与消息体模型 |
| `terminal` | 终端侧报文构造 / 模拟 / 协议详情 |
| `adapter` | 多后端适配（Leader/Follower） |
| `attachment` | 主动安全报警附件接收服务 |
| `gb28181` | GB28181 SIP 客户端与 JT1078→PS 流转换 |
| `shared/consts` | 指令 ID、版本、主动安全标准等常量 |

**不在范围内（示例工程）：** `example/*` 中的 Web、MQ、数据库、流媒体对接业务实现（仅作参考场景）。

### 1.3 设计目标（代码与 README 一致）

1. 以二次开发与生产使用为首要目标  
2. 通过事件/Handler 扩展业务，不修改核心协议链路  
3. 无锁会话管理（单 goroutine + channel 串行操作 map）  
4. 支持 JT/T 808（2011/2013/2019）+ JT/T 1078 + 主动安全附件（默认苏标）

---

## 2. 系统总览

### 2.1 逻辑架构

```
                    ┌─────────────────────────────────────┐
                    │           业务二次开发层              │
                    │  Custom Handler / TerminalEventer   │
                    │  FileEventer / GB Invite 回调等      │
                    └──────────────▲──────────────────────┘
                                   │ 事件回调
┌──────────┐   TCP    ┌────────────┴────────────┐
│ 车载终端  │◄────────►│      service.GoJT808    │
└──────────┘          │  帧切分/解码/分包/会话/下发 │
                      └────────────┬────────────┘
                                   │ 使用
                      ┌────────────▼────────────┐
                      │ protocol (jt808/model)  │
                      └─────────────────────────┘

  并行能力包：
  adapter  ── 一真多仿，对接多个 808 平台
  attachment ── 主动安全附件 TCP 接收
  gb28181    ── 模拟 IPC：注册/目录/点播 + 流转换
  terminal   ── 构造终端/平台报文、联调
```

### 2.2 模块依赖关系

```
shared/consts
    ▲
    │
protocol  ◄── service, terminal, attachment, adapter, gb28181
    │
service   ◄── example/* 业务
terminal  ◄── 联调/模拟
adapter / attachment / gb28181  ◄── 独立场景
```

各子目录独立 `go.mod`，可按需引用。

### 2.3 运行时数据流（服务端主路径）

```
AcceptTCP
  → connection.run
      → reader: FrameReader 切 0x7e 帧
      → packageParse: 帧级 Decode（Header/Body 切片）+ 分包合并
      → join(sessionManager)  // Key 默认手机号
      → 按 Command 查找 Handler，绑定到 Message
      → OnReadExecutionEvent（Handler + TerminalEventer）
         ※ 核心路径不会调用 Handler.Parse
         ※ 业务若要结构化字段，须在 OnRead 内自行 msg.Handler.Parse / model.Parse
      → write 协程：需要则 Handler.ReplyBody → 编码写回终端
      → 若匹配主动下发应答 → replyChan 返回给 SendActiveMessage
  → leave(sessionManager) on 断开/超时
```

**Parse 职责划分（易混淆，以源码为准）：**

| 层级 | 谁做 | 内容 |
| --- | --- | --- |
| 帧解码 | `service` 固定路径（`packageParse` → `jt808.Decode`） | 反转义、校验、Header、原始 Body 切片 |
| 消息体解析 | **不由 service 自动调用** | `model.T0x0200.Parse` 等把 Body 解成业务字段 |
| 默认平台应答 | `ReplyBody`（write 侧，在 `HasReply` 时） | 多数走 `BaseHandle` 用 Header 组 0x8001；个别 model 在 `ReplyBody` 内自行 Parse |

因此：**二次开发读经纬度、报警附加等，必须在自定义 `OnReadExecutionEvent`（或等价业务代码）中显式 `Parse`。** 见 `service/example_test.go` 中 `meLocation`。

---

## 3. 非功能需求（NFR）

| ID | 类别 | 要求 | 来源/验收线索 |
| --- | --- | --- | --- |
| NFR-01 | 语言/运行时 | Go 1.23+ | go.mod |
| NFR-02 | 并发 | 高连接数：每连接独立 goroutine；会话 map 单线程串行 | session_manager.go；README 10w+ 连接测试 |
| NFR-03 | 性能 | 位置等报文路径尽量少拷贝；Body 复用需深拷贝后才能长期保存 | handle.go 注释 |
| NFR-04 | 可扩展 | 业务通过 Option / Handler / Eventer 扩展，不改 service 核心 | option.go / handle.go |
| NFR-05 | 协议兼容 | 自动识别 2013/2019 头（Version 位、手机号长度） | jt808.go Header.decode |
| NFR-06 | 可测试 | 核心协议与 service 路径有单测；README 声称核心协议覆盖率 100% | protocol/*_test.go, service/*_test.go |
| NFR-07 | 网络 | 服务端当前设计为 TCP（`WithNetwork` 注释：暂不支持 UDP） | service/option.go |

---

## 4. 功能规格（按模块）

说明：每条规格包含 **功能描述 / 输入 / 输出 / 验收标准（AC）**。

---

### 4.1 `protocol` — 协议编解码与消息模型

#### F-PROT-01 帧边界切分（FrameReader）

| 项 | 说明 |
| --- | --- |
| **描述** | 从 TCP 字节流中按 `0x7e ... 0x7e` 切出完整 JT808 帧；只负责边界，不做转义/校验/解码 |
| **输入** | `[]byte` 网络读缓冲；内部 `historyData` 粘包缓冲 |
| **输出** | 完整帧 `[]byte`（零拷贝视图：与 data/history 共享底层数组） |
| **关键 API** | `NewFrameReader()`、`FeedSingleComplete`、`Append`、`PopFrame`、`Clear`、`Pending` |

**验收标准：**

| AC ID | 给定 / 当 / 则 |
| --- | --- |
| AC-PROT-01-1 | 给定缓冲区空且 data 恰好一帧；当 `FeedSingleComplete`；则返回该帧且 ok=true |
| AC-PROT-01-2 | 给定半包分两次到达；当 Append+PopFrame；则凑齐后返回完整帧 |
| AC-PROT-01-3 | 给定粘包（多帧）；当循环 PopFrame；则按序取出每一帧 |
| AC-PROT-01-4 | 调用方若长期持有 frame；必须自行 copy（文档约定） |
| AC-PROT-01-5 | 严格帧边界约定：不自动丢弃非 `0x7e` 前导字节，也不跳过连续冗余 `0x7e`；出现 `7e7e` 等异常空帧时原样弹出，由上层 `Decode` 判定错误并断开连接 |

#### F-PROT-02 报文 Decode / Encode

| 项 | 说明 |
| --- | --- |
| **描述** | 对完整帧做反转义、校验码校验、消息头解析、Body 切片；以及按协议组包（含分包 EncodePackets） |
| **输入** | 完整帧 `[]byte`（含起止 0x7e） |
| **输出** | `*JTMessage{Header, Body, VerifyCode}` 或错误：`ErrCheckCode`、`ErrBodyLengthInconsistency`、`ErrHeaderLength2Short` 等 |
| **头字段** | ID、属性（长度/分包/加密/版本）、手机号 BCD、流水号、分包总包/序号；协议版本 2013/2019 |
| **不变量约定** | `JTMessage` 统一经 `NewJTMessage()`（JT1078 经 `NewPacket()` / `NewCustomPacket()`）创建并在首包解码时唯一确定 `Header`、`Property` 与手机号 BCD；运行期不存在 `Header == nil` 或中途修改手机号的场景 |

**验收标准：**

| AC ID | 给定 / 当 / 则 |
| --- | --- |
| AC-PROT-02-1 | 给定合法心跳/注册等 hex 报文；当 Decode；则 Header.ID、手机号、Body 正确 |
| AC-PROT-02-2 | 给定校验码错误；当 Decode；则返回校验错误 |
| AC-PROT-02-3 | 给定 2019 版本属性；当 decode；则 ProtocolVersion=2019，手机号 10 字节 BCD |
| AC-PROT-02-4 | 给定 Body > 1000；当 EncodePackets；则拆为多包且每包 body ≤ 1000 |

#### F-PROT-03 消息体模型（model）

| 项 | 说明 |
| --- | --- |
| **描述** | 各指令消息体 `Parse` / `Encode` / `Protocol` / `ReplyBody` / `ReplyProtocol`；默认嵌入 `BaseHandle`（通用应答 0x8001 成功） |
| **输入** | `*jt808.JTMessage` 或业务字段 |
| **输出** | 结构化字段 或 body 字节；注册类回复 0x8100 等 |

**关于 `model` 的核心设计取舍与不变量约定：**

- **单连接独立协议实例（无并发共享）**：每个连接协程独立创建并持有自己的 `handles map[JT808CommandType]Handler`，每个协议号对应唯一实例，不存在跨协程并发调用同一模型实例的场景（例如 `P0x8300` / `P0x8302` 在 `Encode()` 中回填 `p.Flag` 是单协程安全的设计）。
- **显式长度/数量字段严格按输入编码**：凡消息体结构体显式定义了协议中的长度或数量字段（如 `P0x9101.ServerIPLen`、`P0x8103.ParamTotal` 与 `ParamContent.Len`、`P0x8302.QuestionContentLen` / `AnswerContentLen`、`P0x9206` 各 `*Len`、`P0x9208.ServerIPLen`、`P0x9212.FileNameLen` / `RetransmitPacketNumber`、`T0x0102.AuthCodeLen`、`T0x0104.RespondParamCount`、`T0x0704.Num`、`T0x0805.MultimediaIDNumber`、`T0x1205.AudioVideoResourceTotal`、`T0x1210.AttachCount` / `FileNameLen`、`T0x1211.FileNameLen` 等），`Encode()` **严格按输入字段值原样组包**，不做隐式 `len(...)` 重算或覆盖；若传入值有误，严格按输入处理，由调用方保证一致性。例外：`T0x0704` 单条位置数据项的长度按该条位置体加附加信息的实际字节数写入（不取 `item.Len`），`Num` 仍原样写入。
- **既有协议细节约定**：`P0x8300` 超过 1000 字节直接按字节切片截断；主动安全湘标（`ActiveSafetyHN`）报警标识号长度按 32 字节处理。

**已实现并有测试对照的消息（与 README/shared 一致）：**

| 类别 | 消息 ID | 名称 |
| --- | --- | --- |
| 基础 JT808 | 0x0001/0x8001 | 终端/平台通用应答 |
| | 0x0002 | 终端心跳 |
| | 0x8003 | 补传分包请求 |
| | 0x0100/0x8100 | 注册 / 注册应答 |
| | 0x0102 | 鉴权 |
| | 0x8103/0x8104/0x0104 | 设置/查询参数 |
| | 0x0200/0x8201/0x0201 | 位置汇报 / 查询 / 查询应答 |
| | 0x8202 | 临时位置跟踪 |
| | 0x8300/0x8302/0x0302 | 文本下发 / 提问 / 提问应答 |
| | 0x0704 | 定位批量上传 |
| | 0x0800/0x0801/0x8800 | 多媒体事件/数据/应答 |
| | 0x8801/0x0805 | 立即拍照 / 应答 |
| JT1078 | 0x1003/0x1005/0x1205/0x1206 | 音视频属性/客流/资源列表/上传完成 |
| | 0x9003/0x9101/0x9102/0x9105 | 查询属性/实时音视频/控制/状态 |
| | 0x9201/0x9202/0x9205/0x9206/0x9207 | 回放/控制/资源查询/文件上传 |
| 主动安全 | 0x1210/0x1211/0x1212/0x9208/0x9212 | 附件信息/文件信息/完成/上传指令/完成应答 |

**验收标准：**

| AC ID | 给定 / 当 / 则 |
| --- | --- |
| AC-PROT-03-1 | 给定标准 0x0200 body；当 T0x0200.Parse；则经纬度(WGS84)、速度、报警/状态标志可解析 |
| AC-PROT-03-2 | 给定 0x0200 附加信息区；当 parse；则附加 ID 按扩展解析（见 addition_extensions） |
| AC-PROT-03-3 | 给定 0x0100 不同版本报文；当 Parse；则 bit14=1 按 2019，bit14=0 且 body<37 按 2011，更长时按 `body[9:29]` 型号 `0x00` 补齐识别 2013 |
| AC-PROT-03-4 | 默认 HasReply=true 的上行；ReplyBody 为 0x8001 成功（注册等特殊覆盖 ReplyProtocol） |
| AC-PROT-03-5 | `protocol/model` 与 `protocol/jt808` 单元测试通过 |

---

### 4.2 `service` — JT808 TCP 服务端

#### F-SVC-01 服务启动与连接接入

| 项 | 说明 |
| --- | --- |
| **描述** | 创建 `GoJT808`，监听 TCP，为每个连接启动 `connection` |
| **输入** | Options：`Addr`(默认 `0.0.0.0:808`)、`Network`(默认 `tcp`)、各类自定义 Func |
| **输出** | 持续 Accept；日志记录 resolve/listen/accept 错误 |
| **API** | `New(opts...)`、`Run()` |

**验收标准：**

| AC ID | 给定 / 当 / 则 |
| --- | --- |
| AC-SVC-01-1 | 给定 WithHostPorts 可用端口；当 Run + 终端 Dial；则连接建立且可收发 |
| AC-SVC-01-2 | 给定非法地址；当 Run；则记录 error 并返回（不 panic） |
| AC-SVC-01-3 | 给定多并发连接；当各自发心跳；则各自独立处理 |

#### F-SVC-02 会话管理（join / leave / 路由）

| 项 | 说明 |
| --- | --- |
| **描述** | 单 goroutine + `operationFuncChan` 管理 `map[key]*session`；Key 默认 `Header.TerminalPhoneNo` |
| **输入** | 首包/后续包 Message；`keyFunc(message) (string, bool)` |
| **输出** | join: key 或 `key invalid` / `key exist`；leave 删除；write 按 key 转发 ActiveMessage |
| **错误** | `ErrNotExistKey`（终端不在线） |

**验收标准：**

| AC ID | 给定 / 当 / 则 |
| --- | --- |
| AC-SVC-02-1 | 给定新终端首包；当 join；则 key 在线且 OnJoinEvent 触发 |
| AC-SVC-02-2 | 给定同一 key 已在线；当另一连接 join 同 key；则返回 key exist 相关错误 |
| AC-SVC-02-3 | 给定 keyFunc 返回 ok=false；当 join；则失败（invalid） |
| AC-SVC-02-4 | 给定断开连接；当 leave；则 SendActiveMessage 对该 key 返回 ErrNotExistKey |
| AC-SVC-02-5 | Key 可自定义（WithKeyFunc），须全局唯一 |

#### F-SVC-03 上行消息处理与默认应答

| 项 | 说明 |
| --- | --- |
| **描述** | 帧解码后按 Command 查找 Handler 并绑定到 `Message`；触发 `OnReadExecutionEvent`；若需应答则在 write 侧调用 `ReplyBody` 编码下行。**核心不会调用 `Handler.Parse`。** |
| **输入** | 终端 TCP 字节流（JT808 帧） |
| **输出** | 平台应答帧（默认 0x8001 等）；事件回调；未支持指令走 OnNotSupportedEvent |
| **默认 Handler 集合** | service.createDefaultHandle：JT808 基础 + JT1078 + 主动安全（苏标）；默认包装为 `defaultHandle`（事件空实现，仍可自动通用应答） |

**实际上行处理顺序（`connection.handleMessages` / `onReadExecutionEvent` / `defaultReplyEvent`）：**

1. 帧级已完成：`JTMessage` 含 Header 与原始 `Body`（可能与读缓冲复用，见 NFR-03）。  
2. `handles[Command]` 命中 → `msg.Handler = handler`；未命中 → `OnNotSupportedEvent`，结束。  
3. **`OnReadExecutionEvent(msg)`**（先 Handler，再 TerminalEventer）——业务解析入口。  
4. 消息进入 `terminalUplinkMsgChan`；write 侧若 `HasReply()` 则 **`ReplyBody` → 写回**，再 `OnWriteExecutionEvent`。  
5. 若正等待主动下发应答且匹配规则命中，则走 complete 路径，可不走默认应答。

**Handler 契约：**

```
// JT808Handler：协议体能力（service 核心仅在「需要应答」时用到 Reply* / HasReply）
Parse(jtMsg) error                 // 消息体 → 结构体；须业务在 OnRead 等处主动调用
Protocol() JT808CommandType
HasReply() bool
ReplyBody(jtMsg) ([]byte, error)   // 平台自动应答 body；write 路径调用
ReplyProtocol() JT808CommandType

// Eventer
OnReadExecutionEvent(*Message)     // 核心必定调用（完整包或未过滤分包时）
OnWriteExecutionEvent(Message)     // 写回/主动下发完成等场景；注意按值传递
```

**关于 `Parse` 的规范约定：**

- 接口上保留 `Parse`，便于自定义 Handler 嵌入 `model.T0x*` 后复用官方解析。  
- **service 运行时不会在上行路径调用 `Parse`**（与历史文档「Parse → OnRead」表述不符；以本版为准）。  
- 需要业务字段时，推荐在 `OnReadExecutionEvent` 内：`_ = h.Parse(msg.JTMessage)`（或解析到嵌入的 model）。  
- 仅依赖默认 0x8001 应答、不读 Body 字段时，可不调用 `Parse`。  
- 部分 model 的 `ReplyBody` 内部会自行 `Parse`（实现细节）；**不能**据此假设核心已统一 Parse，业务读字段仍应显式 Parse。

**验收标准：**

| AC ID | 给定 / 当 / 则 |
| --- | --- |
| AC-SVC-03-1 | 给定 0x0002 心跳；当处理完成；则会话 join 成功且默认 0x8001 成功应答写回 |
| AC-SVC-03-2 | 给定已注册默认指令（含 JT1078/主动安全）；当 createDefaultHandle；则映射存在 |
| AC-SVC-03-3 | 给定 WithCustomHandleFunc 覆盖 0x0200；当上行位置；则调用自定义 OnReadExecutionEvent |
| AC-SVC-03-4 | 给定未实现指令；当处理；则 OnNotSupportedEvent，不崩溃 |
| AC-SVC-03-5 | 给定长期保存 Body 的业务；当 OnRead 中使用；则业务侧深拷贝（规范约束） |
| AC-SVC-03-6 | 给定自定义 0x0200 Handler 且 OnRead 内未 Parse；当仅依赖 model 字段；则字段保持零值（证明核心未自动 Parse） |
| AC-SVC-03-7 | 给定自定义 0x0200 Handler 且 OnRead 内 `Parse(msg.JTMessage)`；当上行标准位置包；则经纬度等业务字段可用 |

#### F-SVC-04 分包合并与补传

| 项 | 说明 |
| --- | --- |
| **描述** | `packageParse` 缓存分包，凑齐后合并 Body；可过滤中间分包事件；支持超时清理与补传请求触发 |
| **输入** | 带 SubPackageSum/No 的多帧 |
| **输出** | 完整 Message（`SubcontractComplete=true`）；可选 0x8003 补传流程 |
| **配置** | `WithHasSubcontract(filter)`，默认 true（过滤中间包，仅完整后业务可见） |

**验收标准：**

| AC ID | 给定 / 当 / 则 |
| --- | --- |
| AC-SVC-04-1 | 给定 N 个分包全部到达；当 completePack；则合并后 body 完整且 hasComplete=true |
| AC-SVC-04-2 | 给定 filter=true；当中间分包到达；则业务 Handler 不被中间包误触发完整逻辑 |
| AC-SVC-04-3 | 给定缺包超时；当 deleteTimeoutPackage / 补传路径；则行为符合单测 |

#### F-SVC-05 平台主动下发（SendActiveMessage）

| 项 | 说明 |
| --- | --- |
| **描述** | 按 Key 找到会话，注入 Header，编码下发，按超时等待终端应答匹配 |
| **输入** | `NewActiveMessage(key, command, body, overTimeDuration)`；超时默认业务侧常传 3s |
| **输出** | `*Message`：成功时 ExtensionFields.Err==nil 且含终端应答；失败 ErrNotExistKey / ErrWriteDataOverTime / 写失败 |
| **应答匹配** | createDefaultRespondHandle：流水号匹配（0x0001/0x0104/0x0201/...）；部分指令跳过中间 0x0001 等待最终应答（如 0x8801→0x0805） |

**验收标准：**

| AC ID | 给定 / 当 / 则 |
| --- | --- |
| AC-SVC-05-1 | 给定在线终端；当下发 0x8201；则终端收到查询包且平台收到 0x0201 匹配流水号 |
| AC-SVC-05-2 | 给定终端离线；当 SendActiveMessage；则 ErrNotExistKey |
| AC-SVC-05-3 | 给定终端不回；当超时；则 ErrWriteDataOverTime，OnWriteExecutionEvent 可感知 |
| AC-SVC-05-4 | 给定 0x8801；当终端先回 0x0001 再回 0x0805；则等待最终 0x0805（isWaitingCommand） |
| AC-SVC-05-5 | 给定自定义 ActiveRespondHandler；当匹配规则命中；则按自定义判定完成 |

#### F-SVC-06 终端事件与空闲超时

| 项 | 说明 |
| --- | --- |
| **描述** | TerminalEventer：Join/Leave/NotSupported + 读写事件；可选 IdleTimeout |
| **输入** | `WithCustomTerminalEventer`；`WithTerminalTimeout(idle, callback)` |
| **输出** | 回调 `TerminalTimeout{Key, Address, ConnectionStartTime, First/LastPacketTime, IdleTimeout}` |

**验收标准：**

| AC ID | 给定 / 当 / 则 |
| --- | --- |
| AC-SVC-06-1 | 给定自定义 TerminalEventer；当连接加入/离开；则 OnJoin/OnLeave 各触发一次 |
| AC-SVC-06-2 | 给定 IdleTimeout=30s 且无报文；当超时；则 onTimeout 回调且连接清理 |
| AC-SVC-06-3 | 给定 IdleTimeout<=0；当长期空闲；则不启用超时检测 |

#### F-SVC-07 统一消息结构 Message

| 项 | 说明 |
| --- | --- |
| **字段** | 嵌入 `*JTMessage`；Key；Command；ExtensionFields{TerminalSeq, PlatformSeq, TerminalData, PlatformData, ActiveSend, SubcontractComplete, Terminal/PlatformCommand, CustomData, Err} |
| **用途** | 上行、下行、主动发送结果、错误结果统一载体 |
| **设计约定** | `Message.ExtensionFields` 与 `ActiveMessage.ExtensionFields` 刻意使用匿名结构体而非导出的命名类型，避免在包级 API 中暴露 `service.ExtensionFields` 顶级符号；调用方直接通过字段访问（如 `msg.ExtensionFields.Err`）。 |

**验收标准：**

| AC ID | 给定 / 当 / 则 |
| --- | --- |
| AC-SVC-07-1 | 给定上行；当 newTerminalMessage；则 TerminalCommand/Seq/Data 填充 |
| AC-SVC-07-2 | 给定主动下发结果；当成功；则 ActiveSend 相关 Platform 字段与终端应答可区分 |

---

### 4.3 `terminal` — 终端模拟与报文构造

#### F-TERM-01 默认指令与分包生成

| 项 | 说明 |
| --- | --- |
| **描述** | 按协议版本与手机号构造 Header，生成默认/自定义指令字节流；覆盖 `protocol v1.18.1` 全部 41 个默认协议；内部使用 `Header.EncodePackets` 支持超长消息体分包 |
| **输入** | `WithHeader(version, phone)`；`CreateDefaultCommandData(cmd)` / `CreateCommandData(cmd, body)` / `CreateDefaultCommandPackets(cmd)` / `CreateCommandPackets(cmd, body)` |
| **输出** | 可直接 TCP 写出的完整 JT808 帧 `[]byte` 或分包切片 `[][]byte`（按包数递增 `PlatformSerialNumber`） |

**验收标准：**

| AC ID | 给定 / 当 / 则 |
| --- | --- |
| AC-TERM-01-1 | 给定 2011/2013/2019 + 手机号；当 CreateDefaultCommandData(全部 41 个默认协议)；则服务端可 Decode 与 Parse 成功 |
| AC-TERM-01-2 | 给定自定义 Handler Encode；当 CreateCustomMessageFunc 命中；则优先用自定义 body |
| AC-TERM-01-3 | 给定未知命令；当 CreateDefaultCommandData / CreateDefaultCommandPackets；则返回 nil 并 warn |
| AC-TERM-01-4 | 给定 Body > 1000 字节；当 CreateCommandPackets；则按 1000 字节拆分多包并按包数递增流水号 |

#### F-TERM-02 预期应答与协议详情

| 项 | 说明 |
| --- | --- |
| **描述** | `ExpectedReply(seq, hexMsg)` 模拟平台/终端应答；`ProtocolDetails(hex)` 打印字段级详情 |
| **输入** | 十六进制报文串、平台流水号 |
| **输出** | 应答帧或人类可读多行详情 |

**验收标准：**

| AC ID | 给定 / 当 / 则 |
| --- | --- |
| AC-TERM-02-1 | 给定合法 hex；当 ProtocolDetails；则含 Header/消息体/校验码说明 |
| AC-TERM-02-2 | 给定非法 hex；当 ProtocolDetails；则返回空串并日志 warn |

---

### 4.4 `adapter` — 多平台适配器

#### F-ADP-01 一真多仿转发

| 项 | 说明 |
| --- | --- |
| **描述** | 真实设备连适配器；适配器对每个配置的后端 808 建立模拟连接 |
| **模式** | **Leader**：读写全通（等同直连）；**Follower**：上行转发，下行默认拒绝，仅 `AllowCommands` 放行 |
| **输入** | `WithHostPorts`、`WithTerminals(...)`、`WithAllowCommand`、`WithTimeoutRetry` |
| **输出** | 真实设备与各后端会话镜像；设备断开→模拟全断；模拟断→按 TimeoutRetry 重连 |

**验收标准：**

| AC ID | 给定 / 当 / 则 |
| --- | --- |
| AC-ADP-01-1 | 给定 Leader+Follower 两后端；当设备注册；则两边均可见上行 |
| AC-ADP-01-2 | 给定仅 Follower 在线；当设备发业务包；则设备收不到 Leader 默认应答（测试案例约定） |
| AC-ADP-01-3 | 给定 Follower AllowCommands 含 0x9205；当该后端下发 0x9205；则真实设备收到 |
| AC-ADP-01-4 | 给定 Follower 下发非允许命令；当转发策略执行；则不写真实设备 |
| AC-ADP-01-5 | 给定后端断开；当超过 TimeoutRetry；则尝试重连 |

---

### 4.5 `attachment` — 主动安全附件服务

#### F-ATT-01 附件上传会话

| 项 | 说明 |
| --- | --- |
| **描述** | 独立 TCP 服务接收报警附件：0x1210 → 0x1211 → 文件流 → 0x1212（补传）→ 成功落盘/失败退出 |
| **默认** | 地址 `0.0.0.0:10808`；标准 `ActiveSafetyJS`（苏标）；可切黑标/广东/湖南/四川等 |
| **输入** | 终端附件通道 TCP 数据；可选 `WithDataHandleFunc` / `WithFileEventerFunc` / `WithActiveSafetyType` / `WithReadTimeout` |
| **输出** | 平台应答帧；`FileEventer.OnEvent(*PackageProgress)`；默认按手机号目录安全落盘（带路径穿越校验），流程见 `docs/attachment-flow.svg` |

**进度阶段 ProgressStage：**

| 阶段 | 含义 |
| --- | --- |
| Init | 收到 0x1210 |
| Start | 收到 0x1211 |
| StreamData | 码流收集中 |
| Supplementary | 补传中 |
| StreamDataComplete | 码流收集完成 |
| Complete | 收到 0x1212 且无需补传 |
| SuccessQuit | 全部文件成功 |
| FailQuit | 未全部完成异常退出 |
| UnexpectedExit | 未解析到手机号等意外退出 |

**验收标准：**

| AC ID | 给定 / 当 / 则 |
| --- | --- |
| AC-ATT-01-1 | 给定完整苏标附件流程；当上传结束；则 ProgressStage 经历 Init→…→SuccessQuit |
| AC-ATT-01-2 | 给定缺块；当 0x1212；则进入补传并最终完整 |
| AC-ATT-01-3 | 给定自定义 FileEventer；当各阶段；则 OnEvent 均被调用 |
| AC-ATT-01-4 | 给定默认事件实现；当 SuccessQuit；则 `./{phone}/{filename}` 写入且大小正确 |
| AC-ATT-01-5 | 给定 ActiveSafetyType 切换；当 0x1210 报警标识解析；则按对应地方标准解析 |

---

### 4.6 `gb28181` — GB28181 客户端与流转换

#### F-GB-01 SIP 注册与保活

| 项 | 说明 |
| --- | --- |
| **描述** | 以设备身份向 GB 平台注册，UDP/TCP 传输可选，默认 30s 心跳 |
| **输入** | sim、`PlatformInfo{ID,IP,Port,Password,Domain}`、`DeviceInfo{ID,IP,Port}` |
| **输出** | 在线状态；心跳无响应计数 |

**验收标准：**

| AC ID | 给定 / 当 / 则 |
| --- | --- |
| AC-GB-01-1 | 给定正确平台参数；当 Init+注册流程；则 online 成功 |
| AC-GB-01-2 | 给定 KeepAlive 间隔；当运行；则周期性心跳 |

#### F-GB-02 目录查询与点播（Invite）

| 项 | 说明 |
| --- | --- |
| **描述** | 处理 MESSAGE 目录等；Invite 仅支持 `SessionName=Play` 点播；SDP 协商后推流 |
| **输入** | SIP MESSAGE / INVITE / ACK / BYE |
| **输出** | MANSCDP XML 应答；SDP 200 OK；JT1078→PS（或 RelayPS）RTP/PS 发送 |
| **流模式** | `JT1078ToPS`（默认）、`JT1078ToPSFilterPacket`、`RelayPS` |

**验收标准：**

| AC ID | 给定 / 当 / 则 |
| --- | --- |
| AC-GB-02-1 | 给定 Catalog 查询；当 message 处理；则返回通道列表（通道 ID 约定：设备 ID 末位换 1–4） |
| AC-GB-02-2 | 给定 Play Invite；当 decodeSDP 成功；则 SubmitInvite 并响应 SDP |
| AC-GB-02-3 | 给定非 Play 会话；当 invite；则 5xx 且提示仅支持点播 |
| AC-GB-02-4 | 给定 JT1078 输入；当 ToGB28181 转换；则平台可播（对接 monibuca/ZLM/WVP 示例） |
| AC-GB-02-5 | 给定 WithInviteEventFunc；当 Invite；则回调可改写 InviteInfo |

---

## 5. 核心接口一览（对接契约）

### 5.1 服务端扩展点

| 扩展点 | Option | 用途 |
| --- | --- | --- |
| 指令处理 | `WithCustomHandleFunc` | 覆盖/增加上行业务（每连接独立 map 实例） |
| 主动应答匹配 | `WithCustomActiveRespondHandlerFunc` | 自定义「下发-应答」完成判定 |
| 会话 Key | `WithKeyFunc` | 非手机号唯一标识 |
| 终端生命周期 | `WithCustomTerminalEventer` | 上线/离线/未支持/读写审计 |
| 空闲超时 | `WithTerminalTimeout` | 僵尸连接清理 |
| 监听 | `WithHostPorts` / `WithNetwork` | 地址与协议 |
| 分包过滤 | `WithHasSubcontract` | 是否对业务隐藏中间分包 |

### 5.2 最小二次开发示例（规范级）

```go
// 1) 启动服务 + 自定义位置
s := service.New(
    service.WithHostPorts("0.0.0.0:808"),
    service.WithCustomHandleFunc(func() map[consts.JT808CommandType]service.Handler {
        return map[consts.JT808CommandType]service.Handler{
            consts.T0200LocationReport: &myLocation{},
        }
    }),
)
go s.Run()

// 2) 主动查询位置
reply := s.SendActiveMessage(
    service.NewActiveMessage(phone, consts.P8201QueryLocation, nil, 3*time.Second),
)
if reply.ExtensionFields.Err != nil { /* 离线或超时 */ }
```

自定义 Handler 推荐：**嵌入官方 model + 在 `OnReadExecutionEvent` 中显式 Parse**（核心不会代调 Parse）：

```go
type myLocation struct{ model.T0x0200 }

func (l *myLocation) OnReadExecutionEvent(msg *service.Message) {
    // 必须自行 Parse，否则 l 上业务字段仍为零值
    if err := l.Parse(msg.JTMessage); err != nil {
        return
    }
    // 使用 l.Latitude / l.Longitude ...
    // 若异步落库：先 copy Body / 所需字段，勿直接长期持有 msg.JTMessage.Body
}

func (l *myLocation) OnWriteExecutionEvent(msg service.Message) {}
```

对照实现：`service/example_test.go` 的 `meLocation`。

---

## 6. 错误与边界行为

| 错误 / 场景 | 行为 |
| --- | --- |
| `ErrNotExistKey` | 主动下发时终端不在线 |
| `ErrWriteDataOverTime` | 等待终端应答超时 |
| `ErrWriteDataFail` | 写连接失败 |
| 校验码错误 | 帧级 Decode 失败，连接侧记日志并断开该连接（见 connection.reader） |
| Body 复用 | `msg.JTMessage.Body` 多数与读缓冲复用；异步/长期保存必须 copy；仅分包合并完成的总包 Body 为独立切片 |
| 未调用 Parse | 业务 model 字段保持零值；不影响默认 ReplyBody 通用应答（若 ReplyBody 不依赖 Body 字段） |
| 同 Key 重复 join | 拒绝并返回 key exist |
| 未支持指令 | OnNotSupportedEvent，不中断连接 |
| 注册应答 | 0x0100 → 0x8100（非 0x8001） |

---

## 7. 部署与示例场景映射

| 场景 | 入口 | 依赖模块 |
| --- | --- | --- |
| 最小服务 | `example/quick_start` | service + terminal |
| 完整 Web 平台 | `example/web` | service + attachment + 业务 |
| 音视频 JT1078 | `example/jt1078` | service + 流媒体 |
| JT808 模拟 GB28181 | `example/jt808_to_gb28181` | adapter + gb28181 + service |
| 多 808 适配 | `example/adapter` | adapter |
| 主动安全附件 | `example/attachment` | attachment + service |
| 分布式 | `example/distributed_cluster` | 多 service + MQ 广播下发 |
| Docker | 根 `Dockerfile` | 构建 jt808_to_gb28181，暴露 808/20000–20002 |

---

## 8. 验收测试矩阵（建议执行）

### 8.1 自动化（仓库内）

| 层级 | 命令建议 | 覆盖目标 |
| --- | --- | --- |
| 协议 | `cd protocol && go test -race -cover -count=1 ./...` | 编解码、model、帧切分（100% 覆盖） |
| 服务 | `cd service && go test -race -cover -count=1 ./...` | 心跳入网、主动下发成功/超时、未支持指令、默认 Handler 集合、分包、流水号匹配 |
| 终端 | `cd terminal && go test -race -cover -count=1 ./...` | 2011/2013/2019 头部构造、全量 41 协议往返、分包编码 |
| 附件 | `cd attachment && go test -race -cover -count=1 ./...` | 附件全流程、补传、路径安全、读超时（100% 覆盖） |
| 适配 | `cd adapter && go test -race -cover -count=1 ./...` | 扇出转发、断连重连、并发竞态 |
| 国标 | `cd gb28181 && go test -race -cover -count=1 ./...` | SIP 处理与流转换 |

### 8.2 场景验收（手工/联调）

| ID | 场景 | 通过准则 |
| --- | --- | --- |
| UAT-01 | 模拟器注册+心跳+位置 | 服务自定义 Handler 打印经纬度；应答码成功 |
| UAT-02 | 平台 0x8201 查询 | 同步返回 0x0201 或明确超时错误 |
| UAT-03 | 分包多媒体/大包 | 合并后 Parse 成功 |
| UAT-04 | 9208 触发附件上传 | attachment 进度到 SuccessQuit 且文件可读 |
| UAT-05 | 9101 实时音视频 | 流媒体可播（按所选 lal/zlm 等） |
| UAT-06 | adapter 双后端 | Leader/Follower 权限符合 README 测试案例 |
| UAT-07 | GB28181 点播 | 目录可见、Invite 后出流 |
| UAT-08 | 空闲超时 | 超时回调触发，会话移除 |

### 8.3 性能基线（README 历史数据，非强制 SLA）

| 项 | 参考结果 |
| --- | --- |
| 连接数 | 10w+（2 核 4G，约 120%+ CPU / 1.7G 内存） |
| 位置入库 | ~5000 条/秒；约 1 亿点成功率约 99.999%（含 MQ 队列溢出丢弃） |

---

## 9. 数据字典（核心类型）

### 9.1 协议版本

| 值 | 含义 |
| --- | --- |
| 2011 / 2013 / 2019 | Header 解析与消息体字段长度分支 |

### 9.2 会话 Key

| 默认 | 说明 |
| --- | --- |
| `TerminalPhoneNo` | 头中 BCD 手机号转十进制字符串 |

### 9.3 主动安全标准（attachment / 0x9208 相关）

| 常量（consts） | 说明 |
| --- | --- |
| ActiveSafetyJS 等 | 默认苏标；可配置黑标/省标 |

### 9.4 位置坐标

| 消息 | 坐标系 |
| --- | --- |
| 0x0200 | WGS84（consts 注释约定） |

---

## 10. 已知限制（as-built）

1. service 网络层以 TCP 为主，UDP 未作为生产路径支持。  
2. 并非 JT808 全量消息均有 model 实现；仅 README/默认 Handler 列表内为「已完成+测试」。  
3. `Header.Encode` 旧 API 已 Deprecated，分包场景应使用 `EncodePackets`。  
4. gb28181 侧重注册、目录、点播；复杂国标信令能力有限。  
5. adapter Follower 写权限依赖 AllowCommand 列表，需业务显式配置。  
6. 分布式下发需业务侧 MQ 广播（示例方案），核心库本身单机会话。  
7. **上行消息体 `Handler.Parse` 不由 service 自动调用**；业务读结构化字段须在事件回调中显式 Parse（见 §2.3、F-SVC-03）。默认 Handler 仅保证帧处理与通用/约定应答，不保证业务字段已填充。

---

## 11. 追溯矩阵（规格 → 源码）

| 规格 ID | 主要源码 |
| --- | --- |
| F-PROT-01 | `protocol/jt808/frame_reader.go` |
| F-PROT-02 | `protocol/jt808/jt808.go`, `packet_codec.go` |
| F-PROT-03 | `protocol/model/*`, `base_handle.go` |
| F-SVC-01 | `service/service.go` Run/New |
| F-SVC-02 | `service/session_manager.go` |
| F-SVC-03 | `service/connection.go`, `handle.go`, `createDefaultHandle` |
| F-SVC-04 | `service/packet_parse.go` |
| F-SVC-05 | `service/active_message.go`, `SendActiveMessage`, respond handlers |
| F-SVC-06 | `service/option.go`, TerminalEventer |
| F-TERM-* | `terminal/terminal.go`, `handle.go`, `option.go` |
| F-ADP-* | `adapter/service.go`, `session.go`, `backend.go`, `option.go` |
| F-ATT-* | `attachment/service.go`, `connection.go`, `package.go`, `file_event.go`, `option.go`, `progress_stage.go` |
| F-GB-* | `gb28181/client.go`, `server.go`, `sdp.go`, `internal/stream` |

---

## 12. 变更记录

| 版本 | 日期 | 说明 |
| --- | --- | --- |
| 1.0 | 2026-07-13 | 基于本地仓库源码与 README 反向生成首版 SDD |
| 1.1 | 2026-07-30 | 修正 A1：澄清 service 上行路径不自动 `Parse`；更新 §2.3、F-SVC-03、§5.2 示例、§6/§10 与验收 AC-SVC-03-6/03-7 |
| 1.2 | 2026-11-26 | 对齐 `protocol v1.18.1`（0x0100 版本判定、0x0704 数据项长度例外）、`terminal`（全量 41 协议与分包 API）、`attachment v0.5.0`（读超时与路径安全）、§8.1 测试矩阵与 §11 追溯矩阵 |

---

## 附录 A 快速检查清单（交付验收用）

- [ ] `service` 可监听并完成注册/心跳/位置闭环  
- [ ] 自定义 Handler 在 `OnReadExecutionEvent` 内显式 `Parse` 后能读取 0x0200 业务字段  
- [ ] 未 Parse 时不误以为核心已填充 model 字段  
- [ ] `SendActiveMessage` 在线成功 / 离线 ErrNotExistKey / 超时 ErrWriteDataOverTime  
- [ ] 默认 Handler 含文档所列 JT1078 与主动安全指令  
- [ ] 分包合并后消息完整  
- [ ] attachment 全流程进度与落盘  
- [ ] terminal 生成报文与服务互通  
- [ ] （可选）adapter / gb28181 场景按示例通过  
- [ ] `go test` 在 protocol、service 模块通过  
