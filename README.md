# wgo 类 Python 的 Golang 脚本化运行工具

wgo 是类似 Python 命令的脚本化运行工具。

## 预览

![wgo](wgo.gif)

## 安装

可以从 [releases](https://github.com/wxnacy/wgo/releases) 页面下载二进制文件运行

也可以直接安装最新版本

```bash
$ go get -u github.com/wxnacy/wgo
```

***暂不支持 windows 平台***

## 使用

```bash
$ wgo
>>> fmt.Println("Hello World")
Hello World

>>>
```

**退出**

`<c-d>` 或者输入 `exit`

**导入包**

脚本内置了一些包，包括 `fmt` `os` `time` `strings`

也可以导入新的包，就像在文件里写代码一样

```bash
>>> import "bytes"
```

**直接输出变量**

可以像 Python 命令行那样，输入变量名，直接打印

```bash
>>> t := time.Now()
>>> t
2019-03-19 17:54:36.626646507 +0800 CST m=+0.000424636

>>>
```

**代码补全**

如果想要代码补全，需要安装 [gocode](https://github.com/mdempsky/gocode)

```bash
$ go get -u github.com/mdempsky/gocode
```

现在的代码补全功能，如果当行代码比较复杂，需要在想要补全的报名前加一个空格，这不影响代码输出，只是稍微有点别扭，比如：

![wgo1](wgo1.gif)

## 更新日志

### 1.0.4
- 支持 `import` 设置别名
- 支持更多的打印命令

### 1.0.3
- 修复没有安装 `gocode` 报错的 bug
- 增加 `var x = 1` 表达式的运行
- 增加运行时的版本信息
- 修改导入包异常的 bug
