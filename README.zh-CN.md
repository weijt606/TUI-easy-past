# tep — TUI Easy Paste

[English](README.md) | **简体中文**

从终端 / TUI(Claude Code、Codex CLI 等)里复制文本,粘到别处往往是乱的:每行
都带着界面的缩进、正文被硬折行截断、盒状边框和 ANSI 颜色码混进来。`tep` 把这些
清理干净,让你直接粘进 Reddit、X、文档或聊天框。

它会**自动识别 Markdown 和纯文本**,并分别处理:

- **纯文本** → 把被终端折行的句子重新拼回成干净的段落。
- **Markdown** → 保留结构:标题、列表项、引用、表格、围栏代码块都保持原有边界,
  只把块内被折行的正文重新拼接。代码块逐字保留。

## 为什么又造一个?

已经有一些清理工具(比如 `ai-clean`)。`tep` 的侧重点是**纯文本 / Markdown 区分
的准确性**,以及**宽度感知的重排**——只有当下一个词在检测到的折行宽度下确实放不下
时才合并行,所以作者手写的换行会被保留,只撤销终端的强制折行。它也**不会把 `>`
当成边框删掉**(那是 Markdown 的引用标记)。

## 安装

```sh
go install github.com/weijt606/TUI-easy-past@latest   # 安装 `tep` 可执行文件
```

或从源码构建:

```sh
git clone https://github.com/weijt606/TUI-easy-past
cd TUI-easy-past
go build -o tep .
```

无 cgo、无第三方依赖。剪贴板读写通过系统自带工具完成:`pbcopy`/`pbpaste`
(macOS)、`wl-copy`/`xclip`/`xsel`(Linux)、`clip`/`Get-Clipboard`(Windows)。

## 用法

```sh
tep                 # 读取剪贴板、清理、写回(最常用)
tep --dry-run       # 打印清理结果,不改动剪贴板
tep --stdin         # 从标准输入读取,清理后写到标准输出
cat session.log | tep -      # 等同于 --stdin
tep --explain       # 额外把做了什么打印到 stderr
```

典型流程:在 TUI 里选中文本,复制,运行 `tep`,粘贴。

### 参数

| 参数 | 作用 |
|---|---|
| `-n`, `--dry-run` | 把结果打印到 stdout;不改动剪贴板。 |
| `--stdin`, `-` | 从标准输入读、写到标准输出。 |
| `--explain` | 把检测到的格式、剥离的边框、缩进等打印到 stderr。 |
| `--no-rejoin` | 只清理边框/空白,保留原有换行。 |
| `--keep-ansi` | 保留 ANSI 转义序列。 |
| `--markdown` | 强制 Markdown 模式(跳过自动识别)。 |
| `--plain` | 强制纯文本模式(跳过自动识别)。 |

## 处理流程(按顺序)

1. 规范化换行符,剥离 ANSI 转义。
2. 删除横向盒边框(`┌──┐`、`└──┘`);保留盒内空行作为空白行,避免丢失段落分隔。
3. 剥离作为边框的左右竖线(`│`、`|` 等)。
4. 去掉行尾空白,并去除公共的左缩进。
5. 识别 Markdown 还是纯文本。
6. 宽度感知地重排折行,同时尊重 Markdown 结构。
7. 合并连续的空行。

## 局限

重排是启发式的——终端折行时会丢掉"这个换行是作者写的还是被强制折的"这一信息,
所以 `tep` 只能从检测到的折行宽度去推断。由此带来:

- 宽度小于约 40 列的内容不做重排,以免误合并刻意写短的行(诗句、窄列表)。真实
  TUI 输出(折行在 80 列以上)能正常重排。
- 如果你只想无损复制 Claude Code 的输出,它内置的 `/copy` 命令会直接复制原始
  Markdown。`tep` 是面向任意 TUI 的通用兜底方案,从你选中的内容出发处理。

如果你更想保留每一处换行、只剥离边框,用 `--no-rejoin`。

## 许可证

MIT
