# Linux 常用命令杂记

### linux 内核版本升级操作

linux中安装新内核不会覆盖旧内核，而升级内核会导致新的内核直接替换就的内核，可能会导致系统无法启动

- yum/apt/dnf 方式升级内核

```bash
yum list kernel  --showduplicates  # 查看yum可升级的内核版本  --showduplicates 作用是显示所有可用的相同软件包的不同版本
yum update kerenl-5.14.0-503.11.1.el9_5.el9_5.x86.64

reboot
cat /proc/version

修改默认启动启动的内核

grubby --set-default /boot/vmlinuz-5.14.0-503.21.1.el9_5

```

- centos 上内核降级

```bash
#原内核版本2.6.32-642.el6.x86_64 

yum -y remove kernel kernel-firmware
yum -y install kernel-2.6.32-358.el6.x86_64.rpm kernel-firmware-2.6.32-358.el6.noarch.rpm
reboot

```

- 手动下载包进行升级

```bash


```


### linux 加载/卸载内核模块


### linux 进入紧急救援模式修改root密码（需要能够访问物理机）

进入 grub 界面选择内核

找到 linux 开头的行，按ctrl+e 到行尾，加上 `rd.break` 或者 `init=/bin/bash`   ctrl+d 启动，就可以进入救援模式

```bash
mount -o remount\,rw /sysroot      救援模式切换文件系统
chroot /sysroot
正常修改root密码

```

`rd.break` 是一个内核启动参数，用于在系统启动的早期阶段中断启动过程，进入一个临时的救援 shell,此时，根文件系统尚未挂载。

具体来说：

* **`rd`** ：代表 "ramdisk"，即初始 RAM 磁盘（initramfs）。
* **`break`** ：表示在 initramfs 执行过程中暂停，进入救援模式。

initramfs 是一个临时的根文件系统，包含启动系统所需的基本工具和驱动程序。它的主要任务是挂载真正的根文件系统。在挂载根文件系统之前，系统会执行 initramfs 中的脚本。



* **`rd.break`** ：
* 在 **initramfs 阶段** 中断启动过程。
* initramfs 是一个临时的根文件系统，包含启动系统所需的基本工具和驱动程序。它的主要任务是挂载真正的根文件系统。
* 当使用 `rd.break` 时，系统会在 initramfs 执行过程中暂停，进入一个临时的救援 shell。此时， **真正的根文件系统尚未挂载** 。
* **`init=/bin/bash`** ：
* 在  **内核启动完成后** ，跳过正常的初始化过程（如 systemd 或 init），直接启动一个 Bash shell。
* 此时， **根文件系统已经挂载** ，但系统服务（如网络、登录管理器等）尚未启动。

### linux 磁盘挂载诊断

### linux 磁盘读取速率脚本

### linux 分区扩容

### linux 内存泄露快速检测

### linux rookit 排查

### linux dns服务器配置

### linux ntp时间同步操作

### linux 编译内核模块
