# DLL 劫持 

## Dll的显示加载和隐式加载

* 隐式加载又叫载入时加载，指在主程序载入内存时搜索DLL，并将DLL载入内存（程序无法进行校验）。隐式加载也会有静态链接库的问题，如果程序稍大，加载时间就会过长，用户不能接受。启动时就全部加载
* 显式加载又叫运行时加载，指主程序在运行过程中需要DLL中的函数时再加载（程序可以进行校验）。显式加载是将较大的程序分开加载的，程序运行时只需要将主程序载入内存，软件打开速度快，用户体验好。要用到功能的时候才加载

要通过显式链接使用 DLL，应用程序必须进行函数调用以在运行时显式加载 DLL。要显式链接到 DLL，应用程序必须：

1. 调用 `LoadLibraryEx`或类似函数来加载 DLL 并获取模块句柄。
2. 调用 `GetProcAddress`以获取指向应用程序调用的每个导出函数的函数指针。由于应用程序通过指针调用 DLL 函数，编译器不会生成外部引用，因此无需与导入库链接。但是，您必须有一个typedeforusing语句来定义您调用的导出函数的调用签名。
3. 完成 DLL 后调用 `FreeLibrary`。

## 安全加载dll 选项

系统使用的标准 DLL 搜索顺序取决于是否启用了安全 DLL 搜索模式。安全 DLL 搜索模式将用户的当前目录放在搜索顺序的后面。

默认情况下启用安全 DLL 搜索模式。要禁用此功能，请创建

`HKEY_LOCAL_MACHINE\System\CurrentControlSet\Control\Session Manager\SafeDllSearchMode` 注册表值并将其设置为 0。

如果启用

`SafeDllSearchMode` ，则搜索顺序如下：

1. 加载应用程序的目录。
2. 系统目录。使用 `GetSystemDirectory`函数获取该目录的路径。
3. 16 位系统目录。
4. Windows 目录。使用 `GetWindowsDirectory`函数获取该目录的路径。
5. 当前目录。
6. PATH 环境变量中列出的目录。请注意，这不包括 `App Paths`注册表项指定的每个应用程序路径。计算DLL 搜索路径时不使用 `App Paths`键。

如果禁用 `SafeDllSearchMode` ，则搜索顺序如下：

1. 加载应用程序的目录。
2. 当前目录。
3. 系统目录。使用 `GetSystemDirectory`函数获取该目录的路径。
4. 16 位系统目录。
5. Windows 目录。使用 `GetWindowsDirectory`函数获取该目录的路径。
6. PATH 环境变量中列出的目录。请注意，这不包括 `App Paths`注册表项指定的每个应用程序路径。计算DLL 搜索路径时不使用 `App Paths`键。

动态载入DLL所需要的三个函数详解（LoadLibrary,GetProcAddress,FreeLibrary）[https://www.gandalf.site/2022/07/dll.html](https://www.gandalf.site/2022/07/dll.html)


在自动化挖掘中，一般是找exe调用的非微软exe dll进行劫持，因为微软的dll会有签名

### 实践操作


### 自动化挖掘脚本
