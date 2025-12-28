# Bilibili TTS Chat

![License](https://img.shields.io/badge/license-MIT-blue.svg) ![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8.svg) ![Wails](https://img.shields.io/badge/Wails-v2-red.svg)

**Bilibili TTS Chat** 是一个基于 **Go** 和 **Wails** 开发的桌面端 B 站直播间智能助手。它集成了 **火山引擎** 的高质量 TTS（语音合成）和 LLM（大语言模型）服务，能够实时监听直播间弹幕、礼物、关注等事件，并进行智能语音播报和 AI 互动回复。

该项目旨在为 B 站主播提供一个无需复杂配置、开箱即用的智能场控机器人，增强直播间互动体验。

---

## ✨ 功能特性

- **🖥️ 现代化桌面 UI** - 基于 Wails 构建，提供直观的配置界面和实时日志监控。
- **🎯 实时事件监听** - 秒级响应弹幕、礼物、关注、SC（超级聊天）、舰长等直播间动态。
- **🎵 高质量 TTS** - 集成火山引擎 `seed-tts` 引擎，内置 **286 种** 不同风格的音色。
- **🤖 AI 智能回复** - 基于大模型理解弹幕上下文，自动生成幽默、拟人的回复内容。
- **⚙️ 零门槛配置** - 首次启动只需输入直播间身份码，自动保存配置。
- **📦 单文件运行** - 所有依赖（包括前端资源、配置文件模板）打包为一个 exe，绿色免安装。
- **📝 沉浸式日志** - 界面内实时展示运行日志，关键信息一目了然。

---

## 🚀 快速使用

### 1. 下载与安装
从 [Releases](https://github.com/CoffeeSwt/bilibili-tts-chat/releases) 页面下载最新版本的 `bilibili-tts-chat.exe`。

> 注意：本程序目前仅支持 Windows 平台。

### 2. 首次启动配置
1. 双击运行程序。
2. 首次启动会自动进入**设置界面**。
3. 输入您的 **B 站直播间身份码**（在 B 站直播开放平台获取）。
4. （可选）调整音量、修改助手名称、填写直播间描述（帮助 AI 更好地理解人设）。
5. 点击 **"保存并启动"**。

### 3. 开始直播
程序启动后会自动连接直播间，并跳转到 **日志监控页**。此时：
- 收到的弹幕会实时显示在日志中。
- 助手会根据配置自动播报欢迎语、感谢礼物或回复弹幕。
- 您可以随时点击右上角 **"设置"** 调整参数。

### 4. 高级配置 (.env)
如果您需要使用自己的 API Key（默认使用内置 Key），请在程序同级目录下创建 `.env` 文件：

```env
# 运行模式 (dev/release)
mode=release

# 火山引擎 TTS 配置
tts_x_api_app_id=YOUR_APP_ID
tts_x_api_access_key=YOUR_ACCESS_KEY

# 火山引擎 LLM 配置
llm_volcengine_api_key=YOUR_API_KEY
llm_volcengine_model=YOUR_MODEL_ID

# B 站开放平台配置
bili_app_id=YOUR_BILI_APP_ID
bili_access_key=YOUR_ACCESS_KEY
bili_secret_key=YOUR_SECRET_KEY
```

---

## 🛠️ 二次开发

如果您是开发者，想要修改源码或贡献功能，请参考以下指南。

### 前置要求
- **Go**: 1.21+
- **Node.js**: 18+ (用于前端构建)
- **Wails**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### 项目结构
```
bilibili-tts-chat/
├── bili/           # B 站 WebSocket 与 API 交互逻辑
├── config/         # 配置管理 (Env, User, Voice)
├── handler/        # 业务逻辑处理器 (弹幕, 礼物, 关注等)
├── llm/            # LLM 大模型接口封装
├── logger/         # 日志系统
├── task_manager/   # 任务队列与调度
├── tts_api/        # 火山引擎 TTS 接口
├── voice/          # 音频播放控制
├── wails/          # Wails 桌面端主入口
│   ├── app.go      # 后端与前端交互的 Bridge
│   └── frontend/   # Vue3 + Vite 前端源码
└── build.ps1       # 自动化构建脚本
```

### 开发流程

1. **克隆仓库**
   ```powershell
   git clone https://github.com/CoffeeSwt/bilibili-tts-chat.git
   cd bilibili-tts-chat
   ```

2. **安装依赖**
   ```powershell
   # 后端依赖
   go mod tidy
   
   # 前端依赖
   cd wails/frontend
   npm install
   ```

3. **配置开发环境**
   在项目根目录复制 `.env.example` 为 `.env`，并填入您的开发用 API Key。

4. **启动开发模式**
   在 `wails` 目录下运行：
   ```powershell
   wails dev
   ```
   这将同时启动后端服务和前端开发服务器，支持热重载。

### 构建发布

项目提供了一键构建脚本，会自动处理资源嵌入、编译优化和文件打包。

在项目根目录运行 PowerShell：
```powershell
.\build.ps1
```

构建产物将生成在 `dist/` 目录下，包含：
- `bilibili-tts-chat.exe` (主程序)
- `user.json` (用户配置模板)
- `voices.json` (音色库配置)

---

## 📋 支持的事件

| 事件类型 | 描述 | 处理逻辑 |
| :--- | :--- | :--- |
| **弹幕** | 观众发送的聊天内容 | 触发 AI 回复（如果启用）或直接 TTS 播报 |
| **礼物** | 观众赠送礼物 | 播报感谢语，如 "感谢 xx 送出的 xx" |
| **关注** | 新观众关注直播间 | 播报欢迎关注 |
| **SC** | Super Chat (醒目留言) | 优先播报 SC 内容 |
| **舰长** | 开通/续费大航海 | 播报感谢开通信息 |
| **进场** | 观众进入直播间 | (可选) 播报欢迎进入 |

---

## 📄 许可证

本项目采用 [MIT License](LICENSE) 开源。

## 🙏 致谢

- [Wails](https://wails.io/) - 构建跨平台桌面应用的 Go 框架
- [火山引擎](https://www.volcengine.com/) - 提供优秀的 TTS 和 LLM 服务
- [Bilibili Open Live](https://open-live.bilibili.com/) - B 站直播开放平台

---

⭐ **如果觉得好用，请给个 Star 支持一下！**
