@echo off
REM Windows 批处理脚本：使用 garble 混淆编译

echo Building with garble obfuscation...

REM 检查 garble 是否安装
where garble >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo Error: garble is not installed
    echo Install it with: go install mvdan.cc/garble@latest
    exit /b 1
)

REM 使用 garble 编译
garble build -o app.exe main.go

if %ERRORLEVEL% EQU 0 (
    echo Build successful! Binary: app.exe
    echo Note: The binary is obfuscated, making reverse engineering harder.
) else (
    echo Build failed
    exit /b 1
)

