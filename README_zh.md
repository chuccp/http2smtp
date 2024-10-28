[**English**🌎 ](./README.md)| **简体中文**🀄

在项目中，经常使用邮件来通知异常日志。然而，这通常需要在项目中配置SMTP，并提供接收邮件的邮箱地址。当邮件地址发生变化时，就需要修改项目的配置文件。此外，由于网络限制，可能无法配置SMTP。

本程序可以用HTTP接口替代SMTP，简化了邮件发送流程。您只需在管理页面配置SMTP和接收邮件的电子邮箱地址，即可实现HTTP发送邮件。

支持GET和POST请求。

**GET请求示例**：

```powershell
curl 'http://127.0.0.1:12567/sendMail?token={{token}}&content=this%20is%20a%20test&recipients=aaa@mail.com,bbb@mail.com'
```

**POST请求示例**：

```powershell
curl -X POST 'http://127.0.0.1:12567/sendMail' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--data-urlencode 'token={{token}}' \
--data-urlencode 'content=this%20is%20a%20test'
```

**发送带附件的邮件示例**：

```powershell
curl -X POST 'http://127.0.0.1:12567/sendMail' \
--form 'files=@"/111111.txt"' \
--form 'files=@"/22222222222222.txt"' \
--form 'token={{token}}"' \
--form 'content=1212'
```

这种方式本质上是简单的表单提交，不同语言和平台都能方便地使用本项目。

**POST提交json格式例子**:

```powershell
curl -X POST 'http://127.0.0.1:12567/sendMail' \
--header 'Content-Type: application/json' \
--data '{
"token": "{{token}}",
"content":"this is a test",
"recipients":["aaa@mail.com","bbb@mail.com"]
}'
```

**参数说明**：

- `token`：在管理界面中手动添加获得，是与SMTP和接收邮箱绑定的唯一值。
- `content`：邮件内容。
- `subject`：邮件主题。如果在生成token时已设置主题，当此参数为空时将使用预设主题。
- `files`：需要发送的附件，支持多个文件。
- `recipients`:补充的邮箱地址，选填

**安装方法**：

您可以直接从以下链接下载编译好的版本：

[Download from GitHub](https://github.com/chuccp/smtp2http/releases)

or

```
curl -uri "https://github.com/chuccp/smtp2http/releases/latest/download/smtp2http-windows-amd64.tar.gz" -o smtp2http-windows-amd64.tar.gz
```

docker 镜像
```
docker pull cooge123/smtp2http

docker run -p 12566:12566 -p 12567:12567 -it --rm cooge123/smtp2http
```
下载并解压后，直接运行即可。默认端口号为12566。程序运行后会生成配置文件，您可以在其中修改端口号，修改后重启程序即可使用新的端口号。

启动后，使用浏览器打开 `http://127.0.0.1:12566` 即可进入管理管理。

**配置文件说明**：

程序运行后自动生成配置文件，包含以下部分：

```
[core]
init      = true   # 是否已经完成初始化，默认为false
cachePath = .cache  # 邮件发送文件的临时缓存路径
dbType    = sqlite  # 数据库类型，目前支持sqlite和mysql

[sqlite]
filename = d-mail.db  # SQLite文件路径

[manage]
port     = 12566      # 后台管理的端口号
username = 111111     # 后台管理的账号
password = 111111     # 后台管理的密码
webPath  = web        # web静态文件路径

[api]
port = 12566          # 发送邮件的端口号，如果不想与管理后台共用端口号，可以改成其它端口号

[mysql]
host     = 127.0.0.1  # MySQL主机地址
port     = 3306       # MySQL端口号
dbname   = d_mail     # MySQL数据库名称
charset  = utf8       # 编码格式，默认为utf8
username = root       # MySQL账号
password = 123456     # MySQL密码
```

---

**编译说明**：

如果想自行编译，除了需要编译本项目的外，还需要编译web页面 https://github.com/chuccp/d-mail-view

**软件操作**：

首次进入管理后台时，需要配置数据库和后台管理账号。目前支持SQLite和MySQL数据库。

![initial](https://github.com/chuccp/smtp2http/blob/main/image/0001.png?raw=true "Initial Configuration")

添加SMTP地址：

![SMTP Configuration](https://github.com/chuccp/smtp2http/blob/main/image/0002.png?raw=true "SMTP Configuration")

添加接收邮件的邮箱地址：

![Mail Configuration](https://github.com/chuccp/smtp2http/blob/main/image/0003.png?raw=true "Mail Configuration")

添加Token：

![Token Configuration](https://github.com/chuccp/smtp2http/blob/main/image/0004.png?raw=true "Token Configuration")

配置完成后，即可使用Token给邮箱发送信息。



