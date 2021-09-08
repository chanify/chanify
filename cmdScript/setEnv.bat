@echo off
%1 mshta vbscript:CreateObject("Shell.Application").ShellExecute("cmd.exe","/c %~s0 ::","","runas",1)(window.close)&&exit
cd /d "%~dp0"
mkdir C:\"Program Files"\chanify
xcopy /QY chanify.exe C:\"Program Files"\chanify\

setx path "%path%;C:\Program Files\chanify" /m

reg delete HKEY_CLASSES_ROOT\*\shell\Chanify\command /ve /f
reg delete HKEY_CLASSES_ROOT\*\shell\Chanify /ve /f
reg add HKEY_CLASSES_ROOT\*\shell\Chanify\command /ve /t REG_SZ /d "cmd.exe /c chanify send --endpoint=http://<address>:<port> --token=<token> --file="%%L""
reg add HKEY_CLASSES_ROOT\*\shell\Chanify /ve /t REG_SZ /d 发送到手机

