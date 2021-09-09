@echo off

@REM 管理员方式打开
%1 mshta vbscript:CreateObject("Shell.Application").ShellExecute("cmd.exe","/c %~s0 ::","","runas",1)(window.close)&&exit

@REM 删除自定义注册表项
reg delete HKEY_CLASSES_ROOT\*\shell\ChanifyFile\command /ve /f
reg delete HKEY_CLASSES_ROOT\*\shell\ChanifyAudio\command /ve /f
reg delete HKEY_CLASSES_ROOT\*\shell\ChanifyImage\command /ve /f
reg delete HKEY_CLASSES_ROOT\*\shell\ChanifyFile /ve /f
reg delete HKEY_CLASSES_ROOT\*\shell\ChanifyAudio /ve /f
reg delete HKEY_CLASSES_ROOT\*\shell\ChanifyImage /ve /f
