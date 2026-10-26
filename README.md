

# HTTP2SMTP - USE HTTP SEND EMAIL 

[![GitHub Downloads](https://img.shields.io/github/downloads/chuccp/http2smtp/total.svg)](https://github.com/chuccp/http2smtp/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/cooge123/http2smtp)](https://hub.docker.com/r/cooge123/http2smtp)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://github.com/chuccp/http2smtp/blob/main/LICENSE)
[![Wechat Group](https://img.shields.io/badge/wechat-group-green.svg?logo=wechat)](#Community)
![GitHub last commit](https://img.shields.io/github/last-commit/chuccp/http2smtp)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/chuccp/http2smtp/release.yml)
![GitHub top language](https://img.shields.io/github/languages/top/chuccp/http2smtp)

⭐️ If it helps you, please give a star.

**English**🌎 | [**简体中文**🀄](./README_zh.md)

## Important Updates

Added scheduled task functionality, which allows using cron expressions to periodically read results from an API or link and send emails. Supports configuring templates for JSON response APIs to convert JSON data into text for email content.

## Project Description

A gateway service that converts SMTP protocol to HTTP interfaces, helping developers:

- Avoid hardcoding SMTP configurations in code
- Dynamically send emails through REST API
- Visually configure multiple SMTP providers
- Use scheduled tasks to reduce email configuration in projects

## Main Features

- 🚀 Configure SMTP servers and recipient emails via Web UI
- 📦 Support for GET/POST/JSON request formats
- 🔒 Token-based API access control
- 📎 Multi-file attachment support (Base64 encoding/form upload)
- 🐳 Ready-to-use Docker image
- 📊 Send record query and statistics
- 📅 Scheduled task support for requesting links and sending emails
- 📧 Email template support

## Community

Welcome to join WeChat group or Telegram for more suggestions.

WeChat group:

<img src="https://github.com/chuccp/http2smtp/blob/main/image/WeChat.png?raw=true" alt="WebChat" width="200">

Telegram:

https://t.me/+JClG9_DojaM0ZGE1

## Quick Start

### Direct Execution

```bash
# Windows (PowerShell environment)
# Download package
Invoke-WebRequest -Uri "https://github.com/chuccp/http2smtp/releases/latest/download/http2smtp-windows-amd64.tar.gz" -OutFile "http2smtp-windows-amd64.tar.gz"
# Extract files
tar -zxvf http2smtp-windows-amd64.tar.gz
# Run program
.\http2smtp.exe

# Linux
# Download package (using wget for compatibility)
wget https://github.com/chuccp/http2smtp/releases/latest/download/http2smtp-linux-amd64.tar.gz
# Extract files
tar -zxvf http2smtp-linux-amd64.tar.gz
# Add execution permissions and run
chmod +x http2smtp
./http2smtp
```


### Docker Execution

```bash
##use docker-compose（recommended)
cd docker
docker compose up -d
```
or
```bash
docker pull cooge123/http2smtp

docker run -p 12566:12566 -p 12567:12567 -it --rm cooge123/http2smtp
```
After startup, access the management page using the default port number: http://127.0.0.1:12566

## Configuration Instructions

After first startup, the configuration file [config.ini](./config.ini) is generated:

```ini
[core]
init      = true   ## Initialization switch, becomes true after initial configuration
dbType    = sqlite  ## Database type, supports sqlite and mysql

[sqlite]
filename = d-mail.db  ## Database path

[manage]
port    = 12566       ## Management port

[api]
port = 12567          ## API port

[mysql]
host     = 127.0.0.1   ## Database address
port     = 3306         ## Database port
dbname   = d_mail      ## Database name
charset  = utf8        ## Database character set
username = root        ## Database username
password = 123456      ## Database password
```

The attachment cache directory is not a `config.ini` setting: it comes from `-storage_root`
(or the `storage_root` environment variable) and defaults to `./`, i.e. dates appear as
subdirectories beside the binary. The value is written back to `config.ini` on every start,
so editing `core.cachepath` by hand has no effect.

The ports may also be given as `-web_port` / `-api_port` (or `web_port` / `api_port` in the
environment). A port passed that way is treated as fixed, which is how the Docker image
keeps the settings page from moving it.

## API Documentation

### Send Email API

`POST /sendMail`

**Parameters**:

| Parameter  | Type     | Required | Description         |
|------------|----------|----------|---------------------|
| token      | string   | Yes      | Authorization token |
| subject    | string   | No       | Email subject       |
| content    | string   | Yes      | Email content       |
| recipients | []string | No       | Additional recipient list |
| files      | []File   | No       | Attachment file list |

**Successful Response**:

```json
ok
```


### Complete Request Examples

**JSON Format (with attachments)**

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


**Form Submission (with multiple attachments)**

```bash
curl -X POST 'http://127.0.0.1:12567/sendMail' \
--form 'token={{token}}' \
--form 'subject=test' \
--form 'content=this is a test' \
--form 'recipients=finance@example.com,sales@example.com' \
--form 'files=@"/data/reports/sales.pdf"' \
--form 'files=@"/data/reports/expenses.xlsx"'
```


**GET Request Example**

```bash
curl 'http://127.0.0.1:12567/sendMail?token={{token}}&subject=test&content=this%20is%20a%20test&recipients=aaa@mail.com,bbb@mail.com'
```


## Build Instructions

The management UI is server-rendered Go templates, embedded in the binary — there is no
frontend build step:

```bash
go build -o http2smtp ./
```

The release workflow does the same; the tarball carries only the binary. The UI is served
at `http://127.0.0.1:12566/` with no separate web server.

## UI Display


### initialize

<img src="https://github.com/chuccp/http2smtp/blob/main/image/001.png?raw=true" alt="WebChat" width="800">

### add SMTP

<img src="https://github.com/chuccp/http2smtp/blob/main/image/002.png?raw=true" alt="WebChat" width="800">

### add email address

<img src="https://github.com/chuccp/http2smtp/blob/main/image/003.png?raw=true" alt="WebChat" width="800">

### add TOEKN

<img src="https://github.com/chuccp/http2smtp/blob/main/image/004.png?raw=true" alt="WebChat" width="800">

### add schedule

<img src="https://github.com/chuccp/http2smtp/blob/main/image/005.png?raw=true" alt="WebChat" width="800">

## License

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](https://github.com/chuccp/http2smtp/blob/main/LICENSE)

This project is licensed under the Apache License 2.0. 

See the [LICENSE](./LICENSE) file for the full license text.