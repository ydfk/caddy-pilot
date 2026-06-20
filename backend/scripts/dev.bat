@echo off
setlocal
set ROOT=%~dp0..

call "%~dp0check-cgo.bat" || exit /b 1

pushd "%ROOT%"
air -c .air.toml
set EXIT=%errorlevel%
popd

exit /b %EXIT%
