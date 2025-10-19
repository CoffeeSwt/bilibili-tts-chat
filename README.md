# Bilibili TTS Chat

一个基于Go语言开发的B站直播间TTS语音播报工具，能够实时监听直播间的弹幕、礼物、关注等事件，并使用火山引擎TTS服务将文本转换为语音播报。

## ✨ 功能特性

- 🎯 **实时监听** - 连接B站直播间WebSocket，实时获取直播间动态
- 🎵 **TTS语音播报** - 使用火山引擎TTS服务，高质量语音合成
- 🎨 **多音色支持** - 支持多种音色配置，个性化语音体验
- 👥 **用户音色记忆** - 自动记住用户的音色偏好设置
- 🎁 **多事件支持** - 支持弹幕、礼物、关注、超级聊天等多种事件
- ⚙️ **灵活配置** - 支持环境变量和配置文件双重配置方式
- 🚀 **一键构建** - 提供自动化构建脚本，快速部署

## 📋 支持的事件类型

- 💬 **弹幕消息** - 实时播报观众弹幕
- 🎁 **礼物打赏** - 播报礼物信息和感谢
- 👥 **用户关注** - 欢迎新关注用户
- 💎 **舰长购买** - 播报舰长购买信息
- 💰 **超级聊天** - 播报SC消息
- 👍 **点赞互动** - 播报点赞信息
- 🚪 **进入直播间** - 欢迎用户进入

## 🛠️ 技术栈

- **语言**: Go 1.25+
- **TTS服务**: 火山引擎TTS API
- **WebSocket**: Gorilla WebSocket
- **配置管理**: YAML + 环境变量
- **音频播放**: 系统音频接口

## 📦 安装和使用

### 前置要求

1. **Go环境**: 需要Go 1.25或更高版本
2. **火山引擎账号**: 需要申请火山引擎TTS服务
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
# 填入您的TTS服务密钥和B站API密钥
```

3. **配置应用设置**
```bash
# 复制配置文件模板
cp config.yaml.example config.yaml

# 编辑配置文件
# 设置您的直播间ID和其他参数
```

4. **构建和运行**
```bash
# 使用构建脚本（推荐）
.\build.ps1

# 或手动构建
go build -o bilibili-tts-chat.exe .

# 运行程序
.\bilibili-tts-chat.exe
```

## ⚙️ 配置说明

### 环境变量配置 (.env)

```env
# TTS服务配置
TTS_ACCESS_KEY=your_tts_access_key_here
TTS_APP_ID=your_tts_app_id_here

# B站开放平台配置
BILI_ACCESS_KEY=your_bili_access_key_here
BILI_ACCESS_KEY_SECRET=your_bili_access_key_secret_here
```

### 应用配置 (config.yaml)

```yaml
# 直播配置
id_code: "your_broadcaster_id_code"

# 音色配置
voices:
  - id: "zh_female_shuangkuaisisi_moon_bigtts"
  - id: "zh_male_jingqiangdaxiaodou_moon_bigtts"
  - id: "zh_female_wanwanxiaohe_moon_bigtts"
  # 更多音色...
```

### 用户音色配置 (user_voices.yaml)

系统会自动生成此文件，记录用户的音色偏好：

```yaml
users:
  "用户名1":
    voice_id: "zh_female_shuangkuaisisi_moon_bigtts"
    last_used: "2024-01-01T12:00:00Z"
  "用户名2":
    voice_id: "zh_male_jingqiangdaxiaodou_moon_bigtts"
    last_used: "2024-01-01T12:00:00Z"
```

## 🎵 音色配置

### 可用音色列表

| 音色ID | 描述 | 性别 |
|--------|------|------|
| `zh_female_shuangkuaisisi_moon_bigtts` | 双快思思 | 女声 |
| `zh_male_jingqiangdaxiaodou_moon_bigtts` | 京腔大小豆 | 男声 |
| `zh_female_wanwanxiaohe_moon_bigtts` | 弯弯小鹤 | 女声 |
| `zh_male_yangqizhengqi_moon_bigtts` | 阳气正气 | 男声 |
| `zh_female_qingxinwenrou_moon_bigtts` | 清新温柔 | 女声 |

### 音色管理命令

程序运行时，用户可以通过弹幕命令管理音色：

- `!voice` - 查看当前音色
- `!voice list` - 查看所有可用音色
- `!voice set <音色ID>` - 设置音色
- `!voice random` - 随机选择音色

## 🔧 构建和部署

### 使用构建脚本

```powershell
# 运行构建脚本
.\build.ps1
```

构建脚本会自动：
- 检查Go环境
- 清理旧的构建文件
- 编译可执行文件
- 复制配置文件模板
- 生成dist目录

### 手动构建

```bash
# 安装依赖
go mod tidy

# 构建可执行文件
go build -o bilibili-tts-chat.exe .

# 创建部署目录
mkdir dist
cp bilibili-tts-chat.exe dist/
cp .env.example dist/.env
cp config.yaml.example dist/config.yaml
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
COPY --from=builder /app/config.yaml.example ./config.yaml
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
2024/01/01 12:00:00 正在加载用户音色配置...
2024/01/01 12:00:00 用户音色配置加载成功
2024/01/01 12:00:00 正在初始化TTS服务...
2024/01/01 12:00:00 TTS服务初始化成功
2024/01/01 12:00:00 WebSocket连接成功
```

### 停止程序

- 使用 `Ctrl+C` 优雅停止程序
- 程序会自动保存用户音色配置
- 清理TTS服务连接

## ❓ 常见问题

### Q: 程序启动失败，提示TTS服务初始化失败？
A: 请检查以下配置：
- 确认 `.env` 文件中的 `TTS_ACCESS_KEY` 和 `TTS_APP_ID` 正确
- 确认火山引擎TTS服务已开通并有足够余额
- 检查网络连接是否正常

### Q: 无法连接到直播间？
A: 请检查以下配置：
- 确认 `config.yaml` 中的 `id_code` 正确
- 确认B站开放平台权限已申请
- 检查 `.env` 文件中的B站API密钥

### Q: 语音播报没有声音？
A: 请检查：
- 系统音量设置
- 音频设备是否正常
- TTS服务是否正常响应

### Q: 如何添加新的音色？
A: 在 `config.yaml` 的 `voices` 数组中添加新的音色ID：
```yaml
voices:
  - id: "新音色ID"
```

### Q: 用户音色设置丢失？
A: 检查 `user_voices.yaml` 文件是否存在写入权限，程序会在退出时自动保存。

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

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [火山引擎TTS](https://www.volcengine.com/product/tts) - 提供高质量TTS服务
- [B站开放平台](https://open-live.bilibili.com/) - 提供直播间API
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket库

## 📞 联系方式

- 作者: CoffeeSwt
- 项目地址: [https://github.com/CoffeeSwt/bilibili-tts-chat](https://github.com/CoffeeSwt/bilibili-tts-chat)
- 问题反馈: [Issues](https://github.com/CoffeeSwt/bilibili-tts-chat/issues)

---

⭐ 如果这个项目对您有帮助，请给个Star支持一下！