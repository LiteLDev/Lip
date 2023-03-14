# 教程: 创建一个tooth

这是教程的第一部分，介绍了Lip的一些基本功能。如果你刚开始使用Lip，一定要看看[入门](getting_started.md)和[创造者指南](creator_s_guide.md)，其中介绍了Lip的基本命令。

在本教程中，你将创建一个包含LiteLoaderBDS插件的tooth。

## 先决条件

- **有一定的项目管理经验** 你应该提前学习Git的基本用法，以及JSON的基本语法。

- **一个用于编辑tooth.json文件的工具** 任何你想用的文本编辑器都可以成为选择，最受欢迎的是VSCode和Vim。

- **一个命令行终端** Lip在Windows中的PowerShell和cmd都能很好地工作，你同样可以使用如 Windows Terminal 一类的终端程序。

- **Lip命令行工具** 你应当安装好一个Lip，如需了解更多信息，可查看[安装](installation.md)

## 准备要分发的插件

Lip 会获取 Git 仓库的一个版本的所有内容进行安装。因此，你应该在Git的管理下准备好所有要安装的文件。

如果你只使用文本文件 (如脚本插件 Addons)，你可以使用你用于开发的仓库来创建tooth。

然而，当你需要分发二进制文件时 （如 原生插件，世界）， 你最好创建一个专门的存储库来存放他们。 否则，二进制文件可能会使你的存储库太大，无法管理。

在这个例子中，我们假设存储库的结构如下所示。

```
exampleplugin.dll
exampleplugin/
  config.json
  libexample.dll
  data/
```

## 初始化tooth

1. 打开一个命令提示符，cd到存储库根目录。如果你使用的是Windows，你可以直接按 *shift* 并在文件资源管理器中点击右键，然后点击 "在这里打开PowerShell窗口"。

2. 运行下面的命令来初始化tooth。该命令将在版本库的根目录下创建一个 tooth.json。

   ```shell
   lip tooth init
   ```

3. 编辑 tooth.json 。填写尖括号内的内容。 

   ```json
   {
       "format_version": 1,
       "tooth": "example.com/exampleuser/exampleplugin",
       "version": "1.0.0",
       "dependencies": {
           "github.com/liteloaderbds-hub/liteloaderbds": [
               [
                   "2.9.x"
               ]
           ]
       },
       "information": {
           "name": "Example Plugin",
           "description": "An example plugin",
           "author": "Example User",
           "license": "MIT",
           "homepage": "example.com"
       },
       "placement": [
           {
               "source": "exampleplugin.dll",
               "destination": "plugins/exampleplugin.dll"
           },
           {
               "source": "exampleplugin/*",
               "destination": "plugins/exampleplugin/*"
           }
       ],
       "possession": [
           "plugins/exampleplugin/data/"
       ]
   }
   ```

   `tooth`字段表示该tooth的路径。如果你想发布这一tooth，它必须是不含协议前缀（如https:// 或 http://）的小写的存储库URL。同时，该字段也是您的包的唯一标识。

   `placement` 字段向Lip提供如何将tooth中的文件复制到BDS目录的信息。`source` 属性所表示的源路径是基于tooth（或 tooth所在存储库）的根目录的相对路径，`destination` 属性则是基于BDS根目录（即 bedrock_server.exe 所在目录）的相对路径。
   
   `possession` 字段用于表示这一tooth的私有路径。当使用`lip uninstall`卸载tooth时，这些路径将会被一同删除。在使用升级和重新安装tooth时，这些文件夹会被保留。请注意，在 `possession` 字段中的路径应为以BDS的根目录为基础的相对路径。每个路径都应该以"/"结束。

## 测试你的tooth

在发布齿包之前，你需要测试这个齿包，来确定它是否按预期工作。

1. 将仓库根目录文件全部打成 zip 压缩包，然后可以把它重命名为 `exampleplugin.tth`，当然其他名称也是可以的

2. 把这个文件扔到一个合适的文件夹，然后执行一下下面的命令来安装你刚刚打好的齿包

   ```shell
   lip install exampleplugin.tth
   ```

3. 运行下面的命令把你刚刚装好的齿包卸载掉

   ```shell
   lip uninstall exampleplugin.tth
   ```

4. 运行下面的命令再把这个齿包装上

   ```shell
   lip install exampleplugin.tth
   ```

5. 检查一下，看看你的齿包是否按你想要的方式工作

## 发布你的tooth

1. 储存并提交修改，然后推送到公共Git服务。

2. 添加一个标签并以版本名发布一个版本。标签名称应该是添加了前缀 "v "的版本名称，例如："v1.0.0"。

3. 如您向将您的tooth提交给lip注册表，可以参见[教程：将你的tooth提交给lip注册表](tutorials/submit_your_tooth_to_lip_registry.md)

## 另一个例子：让一个Minecraft的世界成为一个tooth

一般来说，一个Minecraft世界有以下文件结构。

```
Bedrock level/
  level.dat
  level.dat_old
  levelname.txt
  db/
```

你可以像下面这样创建一个 tooth.json：

```json
{
    "format_version": 1,
    "tooth": "example.com/exampleuser/exampleworld",
    "version": "1.0.0",
    "dependencies": {},
    "information": {
        "name": "Example World",
        "description": "An example world",
        "author": "Example User",
        "license": "MIT",
        "homepage": "example.com"
    },
    "placement": [
        {
            "source": "Bedrock level/*",
            "destination": "worlds/exampleworld/*"
        }
    ],
    "possession": [
        "worlds/exampleworld/"
    ]
}
```

## 下一步

你可以阅读 [tooth.json 文件参考](../tooth_json_file_reference.md) 供进一步参考。
