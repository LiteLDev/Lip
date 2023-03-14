# 教程：创建一个工具

自从0.5.0，Lip开始支持工具。工具是可以由Lip执行的程序。你可以使用工具来做一些Lip不能做的事情，比如安装BDS服务器，打包一个世界，甚至是安装其他软件包管理器（如npm）的任何其他工具。

## 先决条件

- **一些项目管理经验** 你应该提前学习Git的基本用法，以及JSON的基本语法。

- **一个用于编辑tooth.json文件的工具** 任何你想用的文本编辑器都可以成为选择，最受欢迎的是VSCode和Vim。
- 
- **一个命令行终端** Lip在Windows中的PowerShell和cmd都能很好地工作，你同样可以使用如 Windows Terminal 一类的终端程序。

- **Lip命令行工具** 你应当安装好一个Lip，如需了解更多信息，可查看[安装](installation.md)

## 准备工具分发

一个工具是一个可执行文件。在Windows中，也支持.cmd文件。可执行文件的名称应该是工具的名称。在Windows上，可执行文件应该是tool_name.exe或tool_name.cmd。如果没有找到.exe文件，Lip将尝试找到.cmd文件。然而，在其他平台上，只有与tool_name完全匹配的文件才被支持。

这里我们将把npm（在Windows上）作为一个Lip工具来打包。npm的文件结构是。

```
node_modules/
  ...
npm.cmd
```

## 书写 tooth.json

Lip将把在.lip/tools/tool_name/下的、以工具名作为其名称的可执行文件（在Windows下，以.exe或.cmd结尾）视为Lip工具。因此，你应该把可执行文件放在.lip/tools/tool_name/下，并把它命名为工具名称。

你可以参考下面的样例来创建 tooth.json

```json
{
    "format_version": 1,
    "tooth": "example.com/exampleuser/exampletool",
    "version": "1.0.0",
    "dependencies": {},
    "information": {
        "name": "Example Tool",
        "description": "An example tool",
        "author": "Example User",
        "license": "MIT",
        "homepage": "example.com"
    },
    "placement": [
        {
            "source": "node_modules/*",
            "destination": ".lip/tools/npm/node_modules/*"
        },
        {
            "source": "npm.cmd",
            "destination": ".lip/tools/npm/npm.cmd"
        }
    ],
    "possession": [
        ".lip/tools/npm/node_modules/"
    ]
}
```

## 测试你的工具

除了[创建一个tooth](tutorials/create_a_lip_tooth?id=测试你的tooth)中的指南之外，你还应该运行下面的命令来测试这个工具：

```shell
lip exec npm [args]
```