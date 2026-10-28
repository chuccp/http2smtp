# HTTP2SMTP - 使用http发送邮件

[![GitHub Downloads](https://img.shields.io/github/downloads/chuccp/http2smtp/total.svg)](https://github.com/chuccp/http2smtp/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/cooge123/http2smtp)](https://hub.docker.com/r/cooge123/http2smtp)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://github.com/chuccp/http2smtp/blob/main/LICENSE)
[![Wechat Group](https://img.shields.io/badge/wechat-group-green.svg?logo=wechat)](#社区)
![GitHub last commit](https://img.shields.io/github/last-commit/chuccp/http2smtp)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/chuccp/http2smtp/release.yml)
![GitHub top language](https://img.shields.io/github/languages/top/chuccp/http2smtp)


⭐️ If it helps you, please give a star.

[**English**🌎](./README.md) | **简体中文**🀄

## 重要更新

增加定时任务功能，可以使用cron表达式来定时读取某一个接口或者链接结果发送邮件。支持对json响应的接口配置一个模板，将json数据转换成文本，作为邮件内容发送。

## 项目描述

将SMTP协议转换为HTTP接口的网关服务，帮助开发者：

- 无需在代码中硬编码SMTP配置
- 通过REST API动态发送邮件
- 可视化配置多个SMTP服务商
- 使用定时任务，减少项目中邮件配置

## 主要特性

- 🚀 通过Web UI配置SMTP服务器和接收邮箱
- 📦 支持GET/POST/JSON多种请求格式
- 🔒 基于Token的API访问控制
- 📎 多文件附件支持（Base64编码/表单上传）
- 🐳 开箱即用的Docker镜像
- 📊 发送记录查询与统计
- 🧾 操作日志审计（登录、增删改、清理等，管理员可清理）
- 📅 支持定时任务，请求链接，发送邮件
- 📧 支持邮件模板

## 社区
欢迎加入微信群或者telegram，提供更多意见。

微信群：

<img src="https://github.com/chuccp/http2smtp/blob/main/image/WeChat.png?raw=true" alt="WebChat" width="200">

telegram：

https://t.me/+JClG9_DojaM0ZGE1

## 快速开始

### 直接运行

```bash

# Windows 系统（PowerShell 环境）
# 下载程序包
Invoke-WebRequest -Uri "https://github.com/chuccp/http2smtp/releases/latest/download/http2smtp-windows-amd64.tar.gz" -OutFile "http2smtp-windows-amd64.tar.gz"
# 解压文件
tar -zxvf http2smtp-windows-amd64.tar.gz
# 运行程序 
.\http2smtp.exe
# Linux 系统
# 下载程序包（使用wget确保兼容性）
wget https://github.com/chuccp/http2smtp/releases/latest/download/http2smtp-linux-amd64.tar.gz
# 解压文件
tar -zxvf http2smtp-linux-amd64.tar.gz
# 添加执行权限并运行
chmod +x http2smtp
./http2smtp
```

### Docker运行

```bash
##使用 docker-compose(推荐)
cd docker
docker compose up -d
```
or

```bash

docker pull cooge123/http2smtp

docker run -p 12566:12566 -p 12567:12567 -it --rm cooge123/http2smtp

```
启动后，使用默认端口号访问管理页面：http://127.0.0.1:12566

## 配置说明

首次启动后生成配置文件 `config.ini`：

```ini
[core]
init      = true   ##初始化开关，初始化配置完成后变为true 
dbType    = sqlite  ##数据库类型，支持sqlite和mysql
maxBodySize = 10M  ##请求体上限，支持 K/M/G 后缀，-1 表示不限制

[sqlite]
filename = d-mail.db  ##数据库路径

[manage]
port    = 12566       ##管理端口

[api]
port = 12567          ##API端口    

[mysql]
host     = 127.0.0.1   ##数据库地址
port     = 3306         ##数据库端口
dbname   = d_mail      ##数据库名称
charset  = utf8        ##数据库字符集
username = root        ##数据库用户名
password = 123456      ##数据库密码
```

邮件附件缓存目录不是 `config.ini` 的配置项：它来自 `-storage_root` 参数（或 `storage_root`
环境变量），默认是 `./`，即附件按日期作为子目录落在程序旁边。每次启动都会把这个值写回
`config.ini`，所以手工修改 `core.cachepath` 不会生效。

端口也可以用 `-web_port` / `-api_port`（或环境变量 `web_port` / `api_port`）指定；用这种方式
给出的端口会被视为固定端口——Docker 镜像正是靠这一点防止设置页改动端口。

## API文档

### 发送邮件接口

`POST /sendMail`

**参数**：

| 参数名        | 类型       | 必填 | 说明      |
|------------|----------|----|---------|
| token      | string   | 是  | 授权令牌    |
| subject    | string   | 否  | 邮件主题    |
| content    | string   | 是  | 邮件内容    |
| recipients | []string | 否  | 额外收件人列表 |
| files      | []File   | 否  | 附件文件列表  |

`recipients` 会追加到令牌上配置的收件人（地址重复只发送一次）。格式非法的地址会被
跳过，只有在没有可用收件人时请求才会失败。JSON、表单和 multipart 请求体均由
`core.maxBodySize` 限制（默认 10 MB，支持 K/M/G 后缀，`-1` 表示不限制），修改后需重启生效。

**成功响应**：

```json
ok
```

### 完整请求示例

**JSON格式（含附件）**

```bash
curl -X POST 'http://127.0.0.1:12567/sendMail' \
--header 'Content-Type: application/json' \
--data-raw '{
  "token": "{{token}}",
  "subject": "test",
  "content": "this is a test",
  "recipients": ["ops@example.com"],
  "files": [
    {
      "name": "alert.log",
      "data": "{{base64_content}}"
    }
  ]
}'
```

**表单提交（含多个附件）**

```bash
curl -X POST 'http://127.0.0.1:12567/sendMail' \
--form 'token={{token}}' \
--form 'subject=test' \
--form 'content=this is a test' \
--form 'recipients=finance@example.com,sales@example.com' \
--form 'files=@"/data/reports/sales.pdf"' \
--form 'files=@"/data/reports/expenses.xlsx"'
```

**GET请求示例**

```bash
curl 'http://127.0.0.1:12567/sendMail?token={{token}}&subject=test&content=this%20is%20a%20test&recipients=aaa@mail.com,bbb@mail.com'
```





## 构建说明

管理界面是 Go 模板服务端渲染，已嵌入二进制，没有单独的前端构建步骤：

```bash
go build -o http2smtp ./
```

发布流程同样只构建二进制，压缩包内不再包含 `web/` 目录。界面直接访问
`http://127.0.0.1:12566/` 即可，无需额外的 Web 服务器。

## 界面展示

### 初始化

<img src="https://github.com/chuccp/http2smtp/blob/main/image/001.png?raw=true" alt="WebChat" width="800">

### 添加SMTP

<img src="https://github.com/chuccp/http2smtp/blob/main/image/002.png?raw=true" alt="WebChat" width="800">

### 添加邮箱地址

<img src="https://github.com/chuccp/http2smtp/blob/main/image/003.png?raw=true" alt="WebChat" width="800">

### 添加TOEKN

<img src="https://github.com/chuccp/http2smtp/blob/main/image/004.png?raw=true" alt="WebChat" width="800">

### 添加定时任务

<img src="https://github.com/chuccp/http2smtp/blob/main/image/005.png?raw=true" alt="WebChat" width="800">













        