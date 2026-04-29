@echo off
setlocal

set "SCRIPT_DIR=%~dp0"
where pwsh.exe >nul 2>nul
if %ERRORLEVEL%==0 (
    pwsh.exe -ExecutionPolicy Bypass -File "%SCRIPT_DIR%run-app.ps1" %*
    exit /b %ERRORLEVEL%
)

powershell.exe -ExecutionPolicy Bypass -File "%SCRIPT_DIR%run-app.ps1" %*
exit /b %ERRORLEVEL%
