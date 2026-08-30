# 文本工具箱（Text Toolkit）

一个基于 [Wails v2](https://wails.io/) + Vue 3 的 Windows 桌面工具箱，集成**文本处理**、**视频/音乐下载**、**视频处理**、**图片工具**、**系统工具**、**网页打包成桌面应用**等多类实用功能。

## 功能特性

### 一、文本工具

| 分组 | 功能 | 说明 |
|------|------|------|
| 编码 | 常用编码 | URL / Base64 / MD5 / SHA1 / SHA256 / SHA512 / HEX / ASCII / 大小写 / 去重去空行 / 驼峰转换等 |
| | 对称加密 | AES / DES / 3DES 等对称算法加解密 |
| | 公钥算法 | RSA 等非对称算法加解密 |
| 处理 | 简单文本 | 添加前后缀、排序（含拼音）、去连续重复字符、转 SQL 条件等 |
| | 文本差集 | 多行文本求差集 / 交集 |
| 格式 | JSON 工具 / JSON 表格 | JSON 格式化、表格化浏览、分隔符文本转 JSON |
| | SQL 工具 | SQL 相关辅助工具 |
| 文件 | 重命名 / 分割合并 / 文件同步 / 备份还原 / 文件归类 | 批量文件处理 |
| 代码生成 | 模板工具 | 模板化代码生成 |
| | Java 工具 | Java 相关辅助工具 |

### 二、工具

| 分组 | 功能 | 说明 |
|------|------|------|
| 系统工具 | 命令行 | 在应用内执行任意 `cmd` 命令并查看输出 |
| | 环境检测 | 探测 Java / Maven / Gradle / Python / Node.js / Go 是否安装 |
| | 环境变量 | 查看 / 删除 / 备份 / 恢复用户与系统环境变量 |
| | 定时关机 | 定时关机与取消 |
| | 其他 | cron 表达式、表达式计算、工资计算、时间转换、卡号工具 |
| 视频下载 | yt-dlp 下载 | 支持 YouTube / B站等，可选格式、代理、Cookie、字幕/缩略图/章节嵌入、限速、时间段、纯音频提取 |
| | B站视频 | B站 / YouTube 下载任务管理 |
| | M3U8 下载 | 从网页提取 m3u8 链接并下载、实时进度 |
| 视频 | 视频处理 | 视频信息读取、截图、音频提取与转码 |
| 音乐 | 音乐搜索 | 网易云音乐搜索（带 SQLite 缓存） |
| | 音乐解析 | 音频 / 歌词 / 封面解析下载 |
| 图片 | 图片工具 / Emoji 工具 / 文字生图 | 图片格式转换、Emoji 处理、文字渲染成图片 |
| | 文字转语音 | 本地 TTS（无需 API Key） |
| 生成 | 应用生成 | 把任意网页打包成独立的 Windows 桌面应用 |

## 外部命令依赖

本项目部分功能通过命令行方式调用外部程序。**应用本体可直接运行，仅当使用对应功能时才需要安装对应命令**，全部命令解析规则见文末。

### 需要自行安装的第三方命令

| 命令 | 来源 | 依赖它的功能 |
|------|------|--------------|
| `yt-dlp` | [yt-dlp](https://github.com/yt-dlp/yt-dlp)（含 `yt-dlp.exe`） | `yt-dlp 下载`、`B站视频`：视频/音频下载、格式探测、版本检查与自更新（`yt-dlp -U`） |
| `ffmpeg`、`ffprobe` | [FFmpeg](https://ffmpeg.org/) | `视频处理`：视频信息、时长、截图、音频提取与转码；同时被 yt-dlp 用于格式检测与合并 |
| `N_m3u8DL-RE` | [N_m3u8DL-RE](https://github.com/nilaoda/N_m3u8DL-RE) | `M3U8 下载`：m3u8 流下载与合成 |
| `edge-tts` | Python 包：`pip install edge-tts` | `文字转语音`：本地语音合成 |
| `wails` | `go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0` | `应用生成`：调用 `wails build` 编译模板生成独立 exe（需同时具备 Go 工具链） |
| `git` | [Git](https://git-scm.com/) | 内置 Git 工具函数（clone / pull / push / add / commit，SSH 推送时还会调用 `ssh`） |

### Windows 系统自带命令（无需额外安装）

| 命令 | 用途 |
|------|------|
| `cmd` | `命令行` 页执行任意命令（`cmd /c ...`） |
| `reg` | 环境变量删除 / 备份（`reg export`）/ 恢复（`reg import`）、注册表定位 |
| `regedit` | 打开注册表编辑器并跳转到指定键 |
| `taskkill` | 打开 regedit 前结束其单实例进程 |
| `rundll32` | 打开系统环境变量编辑器（`sysdm.cpl,EditEnvironmentVariables`） |
| `shutdown` | 定时关机与取消关机 |
| `explorer` | 用资源管理器打开目录 |
| `where` | 在 PATH 中定位 `yt-dlp` |

另外，`系统设置` 的窗口置顶功能会调用随程序分发的小工具 `TopMost_x64.exe`，需位于程序同目录或 PATH 中。

非 Windows 平台上，"打开目录" 会退回到 `open`（macOS）或 `xdg-open`（Linux）。

### 只做环境探测的命令

`环境检测` 页会尝试执行 `java`、`mvn`、`gradle`、`python`、`node`、`go` 的版本命令。缺失时仅显示"未安装"，不影响应用其他功能。

### 命令查找规则

- 默认全部通过系统 `PATH` 解析，将上述可执行文件加入 `PATH` 即可；
- 唯一例外是 `yt-dlp`：优先在**程序同目录**查找 `yt-dlp.exe` / `yt-dlp`，找不到再通过 `where yt-dlp` 查 `PATH`，便于绿色便携部署。

## 技术栈

| 层级 | 技术 |
|------|------|
| 桌面框架 | [Wails v2](https://wails.io/) |
| 后端 | Go |
| 前端 | Vue 3 + [Element Plus](https://element-plus.org/) + [Vite](https://vitejs.dev/) + Vue Router |
| 数据库 | SQLite（`modernc.org/sqlite` 纯 Go 驱动，用于音乐搜索缓存、命令历史等） |
| 打包 | Wails 构建 + NSIS 安装包 |

## 项目结构

```
wcj-go-text/
├── main.go                  # Wails 应用入口
├── app/                     # Wails 绑定层（前端可调用的方法）
│   ├── app.go               # 文本处理 + 应用生成器逻辑
│   ├── app_cmd.go           # 命令执行 / 环境变量 / 定时关机
│   ├── app_movie.go         # 视频 / 音乐下载
│   ├── app_tts.go           # 文字转语音
│   └── ...                  # 加密、图片、身份证、文件等其他功能
├── golang/                  # 核心实现
│   ├── ytdlp.go             # yt-dlp 集成（含 FindYtDlp 查找逻辑）
│   ├── m3u8.go              # M3U8 下载（调用 N_m3u8DL-RE）
│   ├── cmdWrapper/          # 外部命令封装（reg / regedit / explorer / m3u8 等）
│   ├── myTTS/               # edge-tts 封装
│   ├── myVideo/             # 视频处理（ffmpeg / ffprobe）
│   ├── sqllite/             # SQLite（音乐缓存、命令历史）
│   └── utils/               # git 操作、备份等工具函数
├── frontend/                # Vue 3 前端
│   ├── src/
│   │   ├── App.vue          # 根组件（侧边栏导航）
│   │   ├── pages/           # 各功能页面
│   │   └── wailsjs/         # Wails 生成的 JS 绑定
│   └── package.json
├── tpl/                     # 应用生成器内嵌模板
├── build/                   # 构建配置与 NSIS 安装包脚本
├── .github/workflows/       # CI（GitHub Actions）
├── go.mod
└── wails.json
```

## 环境要求

- **Go** >= 1.27
- **Node.js** >= 24
- **Wails CLI** v2.15

构建与开发仅需要以上三项。各**运行时功能**所需的外部命令（yt-dlp、ffmpeg、N_m3u8DL-RE、edge-tts 等）详见 [外部命令依赖](#外部命令依赖)。

## 快速开始

### 安装依赖

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0

# 安装前端依赖
cd frontend
npm install
cd ..
```

### 开发模式

```bash
wails dev
```

启动后会自动打开前端开发服务器并支持热重载。

## 构建

### 本地构建

```bash
# 构建 Windows 可执行文件 + NSIS 安装包
wails build -nsis
```

产物输出在 `build/bin/` 目录。

### CI 自动构建

打 `v*` 标签或在 Actions 页面手动触发（`workflow_dispatch`）会触发 GitHub Actions 构建（Windows / amd64）。其中：

- 手动触发：仅构建并上传产物到 workflow 的 Artifacts
- 打 `v*` 标签：构建完成后自动创建 GitHub Release，并挂载 `.exe` 安装包

```bash
# 发布新版本
git tag v1.0.1 && git push origin v1.0.1
```

## 下载使用

从 [Releases](https://github.com/wyzzgzhdcxy/wcj-go-text/releases) 页面下载最新的 `.exe` 或 NSIS 安装包，解压 / 安装后直接运行即可。文本类功能开箱即用；视频下载、视频处理、文字转语音等功能的依赖见 [外部命令依赖](#外部命令依赖)。

## 许可证

未指定。
