@echo off
where gcc >nul 2>nul
if %errorlevel% equ 0 exit /b 0

echo [ERROR] gorm.io/driver/sqlite requires CGO and a C compiler.
echo Install MinGW-w64 and add gcc.exe to PATH, or run scripts\dev-docker.bat.
exit /b 1
