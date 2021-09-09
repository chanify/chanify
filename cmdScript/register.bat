@echo off

@REM 管理员方式打开
%1 mshta vbscript:CreateObject("Shell.Application").ShellExecute("cmd.exe","/c %~s0 ::","","runas",1)(window.close)&&exit

@REM 进入当前目录
cd /d "%~dp0"

@REM 创建文件夹
mkdir C:\"Program Files"\chanify

@REM 复制程序到文件夹
xcopy /QY chanify.exe C:\"Program Files"\chanify\

@REM 设置系统环境变量
setx path "%path%;C:\Program Files\chanify" /m

@REM 删除自定义注册表项
reg delete HKEY_CLASSES_ROOT\*\shell\ChanifyFile\command /ve /f
reg delete HKEY_CLASSES_ROOT\*\shell\ChanifyAudio\command /ve /f
reg delete HKEY_CLASSES_ROOT\*\shell\ChanifyImage\command /ve /f

@REM 注册文件右键自定义注册表项
reg delete HKEY_CLASSES_ROOT\*\shell\ChanifyFile /ve /f
reg delete HKEY_CLASSES_ROOT\*\shell\ChanifyAudio /ve /f
reg delete HKEY_CLASSES_ROOT\*\shell\ChanifyImage /ve /f

@REM 注册文件右键自定义注册表项命令，这里是作为三种方式发送，根据实际使用需求进行定制
reg add HKEY_CLASSES_ROOT\*\shell\ChanifyFile\command /ve /t REG_SZ /d "cmd.exe /c chanify send --endpoint=https://<address>:<port> --token=<token> --file="%%L""
reg add HKEY_CLASSES_ROOT\*\shell\ChanifyAudio\command /ve /t REG_SZ /d "cmd.exe /c chanify send --endpoint=https://<address>:<port> --token=<token> --audio="%%L""
reg add HKEY_CLASSES_ROOT\*\shell\ChanifyImage\command /ve /t REG_SZ /d "cmd.exe /c chanify send --endpoint=https://<address>:<port> --token=<token> --image="%%L""

@REM 注册文件右键名称
reg add HKEY_CLASSES_ROOT\*\shell\ChanifyFile /ve /t REG_SZ /d 作为文件发送
reg add HKEY_CLASSES_ROOT\*\shell\ChanifyAudio /ve /t REG_SZ /d 作为音频发送
reg add HKEY_CLASSES_ROOT\*\shell\ChanifyImage /ve /t REG_SZ /d 作为图片发送
