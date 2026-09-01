@echo off
setlocal

set "PYTHONUTF8=1"
set "PYTHONIOENCODING=utf-8"

set "PYTHON_CMD=python"
where py >nul 2>nul
if %errorlevel% equ 0 set "PYTHON_CMD=py -3"

%PYTHON_CMD% -X utf8 "%~dp0agent_cli.py"
