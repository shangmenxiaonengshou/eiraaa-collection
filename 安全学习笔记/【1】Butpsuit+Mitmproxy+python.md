# Burpsuite+Mitmproxy+Python

**1.不需要其他语言重写js加密编码算法**

**2.burp中可以直接查看和修改明文数据，最好能支持下intruder**

**3.为其他漏洞利用工具提供支持（如sqlmap）**

**可以用java写burpsuit插件进行实现**

**python**

**mitmproxy  burpsuit   pyexecjs**

* **pyexecjs：使用python调用js加密编码函数，避免js编码算法的重写**
* **burpsuite：一级代理，修改response中js，去除发送请求前的编码动作使请求以明文形式发送**
* **mitmproxy：二级代理，在addon中加载pyexec完成对请求体的加密编码**
  **实现效果：**
  ![](.\resources\images\01-1.png)

## mitmproxy 工具入门

**下载地址**[https://mitmproxy.org/](https://mitmproxy.org/)
**安装之后有这几个主要的exe文件：**

* **mitmproxy** is an interactive, SSL/TLS-capable intercepting proxy with a console interface for HTTP/1, HTTP/2, and WebSockets.**
  **mitmproxy 是一个交互式、支持 SSL/TLS 的拦截代理，具有用于 HTTP/1、HTTP/2 和 WebSocket 的控制台接口。
  **mitmweb** is a web-based interface for mitmproxy.**
  **MitmWeb 是 mitmproxy 的基于 Web 的界面。
  **mitmdump** is the command-line version of mitmproxy. Think tcpdump for HTTP.**
  **mitmdump 是 mitmproxy 的命令行版本。想想 http 的 tcpdump。
* ![](.\resources\images\01-2.png)
  **用法：**
  **先创建一个php文件，接受请求后会把请求头和post的data 处理为json格式echo 出来**

  ```
  <?php
  // 获取请求头
  $headers = getallheaders();

  // 获取 POST 数据
  $postData = file_get_contents("php://input");

  // 创建一个数组存储请求信息
  $responseData = [
      'Request Headers' => $headers,
      'POST Data' => $postData
  ];

  // 将数组编码为 JSON 格式
  $responseJson = json_encode($responseData, JSON_PRETTY_PRINT);

  // 设置响应头为 JSON
  header('Content-Type: application/json');

  // 返回响应数据
  echo $responseJson;
  ?>

  ```

  **之前的包：**

  ![](.\resources\images\01-3.png)

**python脚本**

```
from mitmproxy import http
from mitmproxy import ctx
import base64
import json

def request(flow: http.HTTPFlow) -> None:
    # 检查是否是POST请求
    if flow.request.method == "POST":
        # 获取请求的Content-Type
        content_type = flow.request.headers.get("Content-Type", "")
        
        if "application/x-www-form-urlencoded" in content_type:
            # 解析表单数据
            data = flow.request.urlencoded_form
            ctx.log.info(f"data: {data}")
        
            if "passwd" in data:
                # 对password字段进行Base64加密
                ctx.log.info(f"Original password: {data['passwd']}")
                encoded_password = base64.b64encode(data["passwd"].encode())
                data["passwd"] = encoded_password
                # 更新请求的内容
                flow.request.urlencoded_form = data

addons = [
    # 注册脚本
    request
]

```

**脚本的作用是将密码进行base64编码**

`D:\Apps\mitmproxy\bin\mitmweb.exe  --listen-port 8888 -s ./change-post.py`

![image-20240605230836147](.\resources\images\01-4.png)

**重放包**

![image-20240605230750252](.\resources\images\01-5.png)

**可以看到服务器接受的包已近被理了 ，这是最基本的用法。**

**遇到加密解密的时候，可以用python调用js加密，然后传输，这样我们在burp里面爆破明文，就可以自动加密爆破**
