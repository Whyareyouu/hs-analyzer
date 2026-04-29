@echo off
setlocal

call "%~dp0scripts\run-app.cmd" %*
exit /b %ERRORLEVEL%
