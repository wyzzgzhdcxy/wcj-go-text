# 文本工具箱（Text Toolkit）

一个基于 [Wails v2](https://wails.io/) + Vue 3 的 Windows 桌面工具箱，集成了**文本处理**、**视频下载**、**音乐下载**、**网页打包成桌面应用**四类实用功能。

## 功能特性

### 一、文本工具

| 分组 | 功能 | 说明 |
|------|------|------|
| 常用编码 | URL 编码 / 解码 | 对文本进行 URL 百分号编码或还原 |
| | Base64 编码 / 解码 | 文本与 Base64 互转 |
| | MD5 摘要 | 计算文本 MD5 值 |
| 哈希编码 | SHA1 / SHA256 / SHA512 | 计算文本哈希值 |
| | HEX 编码 / 解码 | 文本与十六进制互转 |
| | ASCII 编码 / 解码 | 文本与 ASCII 码序列互转 |
| 普通文本 | 转 JSON | 多行文本转为 JSON 数组 |
| | 大小写转换 | 全大写 / 全小写 |
| | 删除重复行 / 空行 | 去重、去空行 |
| | 去除不可见字符 | 清除零宽字符等 |
| | 下划线转驼峰 | `snake_case` → `camelCase` |
| 文本两端 | 添加前后缀 | 为每行首尾添加引号、`#` 等字符 |
| 文本排序 | 正向 / 逆向排序 | 支持中文拼音排序 |
| | 倒序 / 交换 | 行顺序反转、交换输入与结果 |
| 去除文本 | 去连续重复字符 | 如 `a//b` → `a/b` |
| 格式转换 | 转 SQL 条件 | 每行一个值 → `('v1', 'v2', ...)` |
| 字符转换 | 转 JSON | 分隔符文本（首行为表头）→ JSON 对象数组 |
| 文件处理 | 文件编码 / 哈希 | 对本地文件做 Base64、MD5、SHA1/256/512 |

### 二、工具

| 分组 | 功能 | 说明 |
|------|------|------|
| 视频下载 | yt-dlp 下载 | 支持 YouTube / B站等，可选格式、代理、Cookie、字幕/缩略图/章节嵌入、限速、时间段、纯音频提取 |
| | B站视频 | B站 / YouTube 下载任务管理 |
| | M3U8 下载 | 从网页提取 m3u8 链接并下载、实时进度 |
| 音乐 | 音乐搜索 | 网易云音乐搜索（带 SQLite 缓存） |
| | 音乐解析 | 音频 / 歌词 / 封面解析下载 |
| 生成 | 应用生成 | 把任意网页打包成独立的 Windows 桌面应用 |

## 技术栈

| 层级 | 技术 |
|------|------|
| 桌面框架 | [Wails v2](https://wails.io/) |
| 后端 | Go |
| 前端 | Vue 3 + [Element Plus](https://element-plus.org/) + [Vite](https://vitejs.dev/) + Vue Router |
| 数据库 | SQLite（`modernc.org/sqlite` 纯 Go 驱动，用于音乐搜索缓存） |
| 打包 | Wails 构建 + NSIS 安装包 |

## 项目结构

```
wcj-go-text/
├── main.go                  # Wails 应用入口
├── app.go                   # 文本处理 + 应用生成器逻辑
├── app_movie.go             # 视频 / 音乐下载逻辑
├── template.go              # 应用生成器模板与图标处理
├── buildtime_keep.go        # 构建时间注入（-ldflags）
├── golang/                  # 核心下载实现
│   ├── download.go          # 批量下载
│   ├── m3u8.go              # M3U8 下载（调用 N_m3u8DL-RE）
│   ├── ytdlp.go             # yt-dlp 集成
│   └── musicDB.go           # 音乐缓存（SQLite）
├── frontend/                # Vue 3 前端
│   ├── src/
│   │   ├── App.vue          # 根组件（侧边栏导航）
│   │   ├── main.js          # 前端入口
│   │   ├── pages/           # 各功能页面
│   │   └── wailsjs/         # Wails 生成的 JS 绑定
│   └── package.json
├── tpl/                     # 应用生成器内嵌模板
├── build/                   # 构建配置与产物
├── .github/workflows/       # CI（GitHub Actions）
├── go.mod
└── wails.json
```

## 环境要求

- **Go** >= 1.27
- **Node.js** >= 24
- **Wails CLI** v2.15

可选（仅视频下载功能需要）：

- [yt-dlp](https://github.com/yt-dlp/yt-dlp)（`yt-dlp 下载`、`B站视频`）
- [N_m3u8DL-RE](https://github.com/nilaoda/N_m3u8DL-RE)（`M3U8 下载`）

将上述可执行文件加入系统 `PATH` 即可。

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

从 [Releases](https://github.com/wyzzgzhdcxy/wcj-go-text/releases) 页面下载最新的 `.exe` 或 NSIS 安装包，解压 / 安装后直接运行即可，无需额外依赖。

## 许可证

未指定。
