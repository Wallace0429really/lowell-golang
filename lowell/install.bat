@echo off
:: LOWELL BETA 1.01 NUCLEAR INSTALLER
setlocal EnableDelayedExpansion
echo ==========================================
echo   LOWELL Stable V1.0 INSTALLER (NUCLEAR)
echo ==========================================
echo.

:: CHECK ADMINISTRATOR
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [ERROR] RIGHT‑CLICK AND "RUN AS ADMINISTRATOR"!
    pause
    exit /b
)

:: 获取本bat所在目录（不带末尾反斜杠）
set "LOWELL_DIR=%~dp0"
set "LOWELL_DIR=%LOWELL_DIR:~0,-1%"
echo [INFO] Install directory: %LOWELL_DIR%
echo.

:: 校验必须文件 lowell.exe + HELP.md
if not exist "%LOWELL_DIR%\lowell.exe" (
    echo [ERROR] lowell.exe MISSING! Run 'go build' first.
    pause
    exit /b
)
if not exist "%LOWELL_DIR%\HELP.md" (
    echo [WARNING] HELP.md NOT FOUND: help command will not work!
    echo.
)

:: Powershell 添加到系统PATH，处理带空格路径
echo [ACTION] Updating System PATH via PowerShell...
powershell -Command "$dir = '%LOWELL_DIR%';$p=[System.Environment]::GetEnvironmentVariable('Path','Machine');if($p -notmatch [regex]::Escape($dir)){$newpath=$p+';'+$dir;[System.Environment]::SetEnvironmentVariable('Path',$newpath,'Machine');Write‑Host '[SUCCESS] Added directory to System PATH!'}else{Write‑Host '[OK] Directory is already present in System PATH.'}"

echo.
echo ==========================================
echo   INSTALL FINISHED!
echo   ❗ CLOSE ALL existing CMD/PowerShell windows
echo   Open a NEW terminal, then run:
echo        lowell repl
echo ==========================================
pause
endlocal
