# Bilibili TTS Chat

一个基于Go语言开发的B站直播间智能TTS语音播报工具，集成了火山引擎TTS和LLM服务，能够实时监听直播间的弹幕、礼物、关注等事件，并提供智能语音播报和AI对话功能。

## ✨ 功能特性

- 🎯 **实时监听** - 连接B站直播间WebSocket，实时获取直播间动态
- 🎵 **TTS语音播报** - 使用火山引擎TTS服务，支持seed-tts-1.0和seed-tts-2.0引擎
- 🤖 **AI智能对话** - 集成火山引擎LLM服务，支持多种AI模型（OpenAI、Claude、Gemini等）
- 🎨 **丰富音色库** - 支持286种不同音色，个性化语音体验
- 👥 **用户音色记忆** - 自动记住用户的音色偏好设置
- 🎁 **多事件支持** - 支持弹幕、礼物、关注、超级聊天等多种事件
- 📝 **完整日志系统** - 自动记录应用运行日志，便于调试和监控
- ⚙️ **灵活配置** - 支持环境变量和JSON配置文件
- 🔄 **任务管理** - 内置工作流处理器，支持复杂任务编排
- 🚀 **一键构建** - 提供PowerShell构建脚本，快速部署

## 📋 支持的事件类型

- 💬 **弹幕消息** - 实时播报观众弹幕
- 🎁 **礼物打赏** - 播报礼物信息和感谢
- 👥 **用户关注** - 欢迎新关注用户
- 💎 **舰长购买** - 播报舰长购买信息
- 💰 **超级聊天** - 播报SC消息
- 👍 **点赞互动** - 播报点赞信息
- 🚪 **进入直播间** - 欢迎用户进入
- 🔚 **直播结束** - 播报直播状态变化

## 🛠️ 技术栈

- **语言**: Go 1.25.3
- **TTS服务**: 火山引擎TTS API (seed-tts-1.0, seed-tts-2.0)
- **LLM服务**: 火山引擎LLM API (支持OpenAI、Claude、Gemini、OpenRouter)
- **WebSocket**: Gorilla WebSocket
- **配置管理**: JSON + 环境变量
- **日志系统**: 自定义文件日志系统
- **音频播放**: 系统音频接口
- **任务管理**: 内置工作流处理器

## 📦 安装和使用

### 前置要求

1. **Go环境**: 需要Go 1.25.3或更高版本
2. **火山引擎账号**: 需要申请火山引擎TTS和LLM服务
3. **B站开放平台**: 需要申请B站开放平台权限

### 快速开始

1. **克隆项目**
```bash
git clone https://github.com/CoffeeSwt/bilibili-tts-chat.git
cd bilibili-tts-chat
```

2. **配置环境变量**
```bash
# 复制环境变量模板
cp .env.example .env

# 编辑环境变量文件
# 填入您的TTS服务密钥、LLM API密钥和B站API密钥
```

3. **配置应用设置**
```bash
# 编辑用户配置文件
# user.json - 设置您的直播间ID和描述
# voices.json - 音色配置（已包含286种音色）
```

4. **构建和运行**
```powershell
# 使用构建脚本（推荐）
.\build.ps1

# 或手动构建
go build -o app.exe .

# 运行程序
.\app.exe
```

## ⚙️ 配置说明

### 环境变量配置 (.env)

```env
# 运行模式
mode=dev

# LLM Mock模式（开发时使用）
llm_mock_enabled=false

# 火山引擎TTS服务配置
TTS_APP_ID=your_tts_app_id_here
TTS_ACCESS_KEY=your_tts_access_key_here

# 火山引擎LLM服务配置
VOLCENGINE_API_KEY=your_volcengine_api_key_here
VOLCENGINE_MODEL=your_model_name_here

# B站开放平台配置
BILI_APP_ID=your_bili_app_id_here
BILI_ACCESS_KEY=your_bili_access_key_here
BILI_ACCESS_KEY_SECRET=your_bili_access_key_secret_here
```

### 用户配置 (user.json)

```json
{
    "room_id_code": "D4V3X00YW7I80",
    "room_description": "这是一个直播房间，主播是一杯甜的苦咖啡，他正在直播写代码，创造一个AI语音直播助手"
}
```

### 音色配置 (voices.json)

系统内置286种音色配置，支持多种语言和风格：

```json
{
    "voices": [
        {
            "id": 1,
            "name": "vivi",
            "voice_type": "zh_female_vv_uranus_bigtts",
            "gender": "female",
            "api_resource_id": "seed-tts-2.0"
        },
        {
            "id": 2,
            "name": "大壹",
            "voice_type": "zh_male_dayi_saturn_bigtts",
            "gender": "male",
            "api_resource_id": "seed-tts-2.0"
        }
        // ... 更多音色配置
    ]
}
```

## 🎵 音色管理

### 音色特性

- **丰富选择**: 286种不同音色
- **多种风格**: 包含男声、女声、不同年龄段和风格
- **双引擎支持**: 支持seed-tts-1.0和seed-tts-2.0引擎
- **智能记忆**: 自动记住用户音色偏好

### 音色管理命令

程序运行时，用户可以通过弹幕命令管理音色：

- `!voice` - 查看当前音色
- `!voice list` - 查看所有可用音色
- `!voice set <音色ID>` - 设置音色
- `!voice random` - 随机选择音色

## 🤖 AI功能

### LLM服务支持

- **火山引擎LLM**: 主要LLM服务提供商
- **OpenAI**: 支持GPT系列模型
- **Claude**: 支持Anthropic Claude模型
- **Gemini**: 支持Google Gemini模型
- **OpenRouter**: 支持多种开源模型

### AI对话功能

- **智能回复**: 根据弹幕内容生成智能回复
- **上下文理解**: 维护对话上下文
- **流式响应**: 支持实时流式对话
- **Mock模式**: 开发时支持模拟响应

## 📝 日志系统

### 日志功能

- **自动记录**: 应用运行时自动记录详细日志
- **文件存储**: 日志保存在`logs/`目录下
- **按时间分割**: 按日期和时间段自动分割日志文件
- **优雅关闭**: 应用退出时自动刷新和关闭日志文件

### 日志文件格式

```
logs/
└── app_2025-01-26_AM.log  # 按日期和时间段命名
```

日志内容包括：
- 应用启动和关闭信息
- WebSocket连接状态
- TTS和LLM服务调用
- 事件处理详情
- 错误和异常信息

## 🔧 项目结构

```
bilibili-tts-chat/
├── bili/                    # B站相关功能
│   ├── manager.go          # 应用管理器
│   ├── request.go          # 请求结构
│   └── websocket.go        # WebSocket客户端
├── config/                  # 配置管理
│   ├── const.go            # 常量定义
│   ├── env_config.go       # 环境变量配置
│   ├── error.go            # 错误定义
│   ├── user_config.go      # 用户配置
│   └── voice_config.go     # 音色配置
├── handler/                 # 事件处理器
│   ├── dm/                 # 弹幕处理
│   ├── guard/              # 舰长处理
│   ├── like/               # 点赞处理
│   ├── send_gift/          # 礼物处理
│   ├── super_chat/         # 超级聊天处理
│   └── ...                 # 其他事件处理
├── llm/                     # LLM服务
│   ├── client.go           # LLM客户端
│   └── prompt.go           # 提示词管理
├── logger/                  # 日志系统
│   ├── logger.go           # 日志接口
│   └── writer.go           # 文件写入器
├── response/                # 响应结构
├── task_manager/            # 任务管理
│   ├── manager.go          # 任务管理器
│   └── task.go             # 任务定义
├── tts_api/                 # TTS服务
│   └── tts_http.go         # TTS HTTP客户端
├── voice/                   # 音频引擎
│   └── engine.go           # 音频播放引擎
├── workflow/                # 工作流处理
│   └── processor.go        # 工作流处理器
├── .env.example            # 环境变量模板
├── user.json               # 用户配置
├── voices.json             # 音色配置
├── build.ps1               # 构建脚本
└── main.go                 # 主程序入口
```

## 🔧 构建和部署

### 使用构建脚本

```powershell
# 运行构建脚本
.\build.ps1
```

构建脚本会自动：
- 检查Go环境
- 清理旧的构建文件
- 编译可执行文件到dist目录
- 复制配置文件模板
- 创建日志目录

### 手动构建

```bash
# 安装依赖
go mod tidy

# 构建可执行文件
go build -o app.exe .

# 创建部署目录
mkdir dist
cp app.exe dist/
cp .env.example dist/.env
cp user.json dist/
cp voices.json dist/
```

### Docker部署（可选）

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o bilibili-tts-chat .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bilibili-tts-chat .
COPY --from=builder /app/user.json .
COPY --from=builder /app/voices.json .
CMD ["./bilibili-tts-chat"]
```

## 🚀 使用指南

### 启动程序

1. 确保配置文件已正确设置
2. 运行可执行文件
3. 程序会自动连接直播间并开始监听

### 程序日志

程序运行时会输出详细日志：

```
2025/01/26 12:00:00 [INFO] 日志系统初始化成功
2025/01/26 12:00:00 [INFO] 应用管理器启动成功
2025/01/26 12:00:00 [INFO] TTS服务初始化成功
2025/01/26 12:00:00 [INFO] LLM服务初始化成功
2025/01/26 12:00:00 [INFO] 成功连接到 wss://zj-cn-live-comet.chat.bilibili.com:443/sub
2025/01/26 12:00:00 [INFO] 事件驱动任务处理器启动成功
```

### 停止程序

- 使用 `Ctrl+C` 优雅停止程序
- 程序会自动保存用户配置
- 清理TTS和LLM服务连接
- 刷新并关闭日志文件

## ❓ 常见问题

### Q: 程序启动失败，提示TTS服务初始化失败？
A: 请检查以下配置：
- 确认 `.env` 文件中的 `TTS_ACCESS_KEY` 和 `TTS_APP_ID` 正确
- 确认火山引擎TTS服务已开通并有足够余额
- 检查网络连接是否正常

### Q: LLM功能无法使用？
A: 请检查以下配置：
- 确认 `.env` 文件中的 `VOLCENGINE_API_KEY` 和 `VOLCENGINE_MODEL` 正确
- 确认火山引擎LLM服务已开通
- 检查是否开启了 `llm_mock_enabled=true`（开发模式）

### Q: 无法连接到直播间？
A: 请检查以下配置：
- 确认 `user.json` 中的 `room_id_code` 正确
- 确认B站开放平台权限已申请
- 检查 `.env` 文件中的B站API密钥

### Q: 语音播报没有声音？
A: 请检查：
- 系统音量设置
- 音频设备是否正常
- TTS服务是否正常响应
- 检查日志文件中的错误信息

### Q: 如何查看程序运行日志？
A: 程序会自动在 `logs/` 目录下生成日志文件，文件名格式为 `app_YYYY-MM-DD_AM/PM.log`

### Q: 如何添加新的音色？
A: 在 `voices.json` 文件中添加新的音色配置：
```json
{
    "id": 287,
    "name": "新音色名称",
    "voice_type": "音色类型ID",
    "gender": "male/female",
    "api_resource_id": "seed-tts-1.0或seed-tts-2.0"
}
```

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### 开发环境设置

```bash
# 克隆仓库
git clone https://github.com/CoffeeSwt/bilibili-tts-chat.git
cd bilibili-tts-chat

# 安装依赖
go mod tidy

# 运行测试
go test ./...

# 代码格式化
go fmt ./...
```

### 代码规范

- 使用 `go fmt` 格式化代码
- 添加必要的注释
- 编写单元测试
- 遵循Go语言最佳实践
- 确保日志记录完整

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [火山引擎TTS](https://www.volcengine.com/product/tts) - 提供高质量TTS服务
- [火山引擎LLM](https://www.volcengine.com/product/llm) - 提供智能对话服务
- [B站开放平台](https://open-live.bilibili.com/) - 提供直播间API
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket库

## 📞 联系方式

- 作者: CoffeeSwt
- 项目地址: [https://github.com/CoffeeSwt/bilibili-tts-chat](https://github.com/CoffeeSwt/bilibili-tts-chat)
- 问题反馈: [Issues](https://github.com/CoffeeSwt/bilibili-tts-chat/issues)

---

⭐ 如果这个项目对您有帮助，请给个Star支持一下！