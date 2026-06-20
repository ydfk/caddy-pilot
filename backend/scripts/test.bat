@echo off
setlocal
set ROOT=%~dp0..

call "%~dp0check-cgo.bat" || exit /b 1

pushd "%ROOT%"
go test ./...
set EXIT=%errorlevel%
popd

exit /b %EXIT%
