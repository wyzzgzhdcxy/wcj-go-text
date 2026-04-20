# 文本工具箱

一个基于 Wails + Vue 3 的轻量级桌面文本处理工具，提供丰富的文本编码、转换、格式化功能。

## 目录

- [功能特性](#功能特性)
- [技术栈](#技术栈)
- [项目结构](#项目结构)
- [快速开始](#快速开始)
- [功能详解](#功能详解)
- [编译构建](#编译构建)
- [下载使用](#下载使用)

## 功能特性

### 常用编码
| 功能 | 说明 |
|------|------|
| URL编码 | 将文本进行URL编码处理 |
| URL解码 | 将URL编码的文本还原 |
| Base64编码 | 将文本转换为Base64格式 |
| Base64解码 | 将Base64格式文本还原 |
| MD5 | 计算文本的MD5哈希值 |

### 哈希编码
| 功能 | 说明 |
|------|------|
| SHA1 | 计算文本的SHA1哈希值 |
| SHA256 | 计算文本的SHA256哈希值 |
| SHA512 | 计算文本的SHA512哈希值 |
| HEX编码 | 将二进制数据转换为十六进制字符串 |
| HEX解码 | 将十六进制字符串还原为二进制数据 |
| ASCII编码 | 将文本转换为ASCII码序列 |
| ASCII解码 | 将ASCII码序列还原为文本 |

### 普通文本
| 功能 | 说明 |
|------|------|
| 转JSON | 将多行文本转换为JSON数组格式 |
| 全部大写 | 将英文文本转换为大写 |
| 全部小写 | 将英文文本转换为小写 |
| 删除重复行 | 去除文本中重复的行 |
| 删除空行 | 去除文本中的空行 |
| 去除不可见字符 | 清除零宽字符等不可见字符 |
| 下划线转驼峰 | 将 `snake_case` 转换为 `camelCase` |

### 文本两端
为每行文本的收尾添加指定字符，例如给每行添加引号：
```
输入:
line1
line2

输出:
"line1"
"line2"
```

### 文本排序
| 功能 | 说明 |
|------|------|
| 正向排序 | 按字母/拼音顺序正向排序（支持中文） |
| 逆向排序 | 按字母/拼音顺序逆向排序 |
| 倒序 | 将文本行顺序完全反转 |
| 交换数据 | 交换输入区和结果区的内容 |

### 去除文本
去除连续重复的字符，保留指定的分隔字符。例如去除连续重复的斜杠：`a//b/c` → `a/b/c`

### 格式转换
| 功能 | 说明 |
|------|------|
| 转SQL条件 | 将多行文本转换为SQL的IN条件语句 |
| 转JSON | 将带分隔符的文本转换为JSON对象数组 |

### 文件处理
支持直接选择本地文件进行编码/哈希计算：
| 功能 | 说明 |
|------|------|
| Base64编码 | 对文件内容进行Base64编码 |
| Base64解码 | 将Base64内容还原为文件 |
| MD5 | 计算文件的MD5值 |
| SHA1 | 计算文件的SHA1值 |
| SHA256 | 计算文件的SHA256值 |
| SHA512 | 计算文件的SHA512值 |

## 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| 框架 | [Wails v2](https://wails.io/) | Go语言桌面应用框架 |
| 后端 | Go | 高性能后端逻辑 |
| 前端框架 | [Vue 3](https://vuejs.org/) | 渐进式JavaScript框架 |
| UI组件库 | [Element Plus](https://element-plus.org/) | Vue 3 UI组件库 |
| 构建工具 | [Vite](https://vitejs.dev/) | 下一代前端构建工具 |

## 项目结构

```
wcj-go-text/
├── frontend/                    # 前端项目
│   ├── src/
│   │   ├── App.vue            # 根组件
│   │   ├── main.js            # 前端入口
│   │   ├── pages/             # 页面组件
│   │   │   ├── textCommonEncode.vue   # 常用编码
│   │   │   ├── textHashEncode.vue     # 哈希编码
│   │   │   ├── textNormal.vue          # 普通文本
│   │   │   ├── textBothEnds.vue        # 文本两端
│   │   │   ├── textSort.vue            # 文本排序
│   │   │   ├── textRemove.vue          # 去除文本
│   │   │   ├── textFormat.vue          # 格式转换
│   │   │   ├── textChar.vue            # 字符转换
│   │   │   └── textFile.vue            # 文件处理
│   │   └── wailsjs/             # Wails生成的JS绑定
│   │       └── go/main/App.js   # Go方法的JS封装
│   ├── package.json
│   └── vite.config.js
├── main.go                      # Wails应用入口
├── app.go                       # 应用主逻辑
├── build/                       # 构建相关配置
│   └── linux/
│       └── ...
├── README.md
└── go.mod
```

## 快速开始

### 环境要求

- Go >= 1.18
- Node.js >= 16
- Wails CLI

### 安装依赖

```bash
# 安装前端依赖
cd frontend
npm install

# 安装Wails CLI（如果没有）
go install github.com/wailsframework/wails/v2/cmd/wails@latest
```

### 开发模式

```bash
# 在项目根目录运行
wails dev
```

这会同时启动前端开发服务器和后端Go服务，支持热重载。

### 编译构建

```bash
# 构建生产版本
wails build

# 构建并指定输出目录
wails build -o ./dist
```

### 交叉编译

```bash
# 构建Windows版本
wails build -platform windows

# 构建Linux版本
wails build -platform linux

# 构建macOS版本
wails build -platform darwin
```

## 功能详解

### 文本处理示例

**URL编码/解码**
```
输入: hello world&name=test
URL编码输出: hello%20world%26name%3Dtest
```

**删除重复行**
```
输入:
apple
banana
apple
cherry
banana

输出:
apple
banana
cherry
```

**下划线转驼峰**
```
输入: hello_world_golang
输出: helloWorldGolang
```

**转SQL条件**
```
输入:
北京
上海
广州

输出: ('北京', '上海', '广州')
```

### 文件处理

1. 点击"选择文件"按钮
2. 在文件对话框中选择文件
3. 点击相应的编码/哈希按钮
4. 结果将显示在下方文本框中

支持的文件类型无限制，可处理任意格式的本地文件。

## 下载使用

从 [Releases](https://github.com/wangchaojun/wcj-go-text/releases) 页面下载对应平台的压缩包：

- **Windows**: 下载 `.exe` 或 `.msi` 安装包
- **Linux**: 下载 `.AppImage` 或 `.deb` 包
- **macOS**: 下载 `.dmg` 安装包

下载后直接运行即可，无需安装额外依赖。

## 快捷操作

- **复制结果**: 点击右下角"复制结果"按钮，一键复制输出内容到剪贴板
- **交换数据**: 在文本排序功能中，可交换输入区和结果区的内容
- **热键支持**: 标准键盘快捷键（Ctrl+C/V）可在文本框中使用

## 许可证

MIT License
