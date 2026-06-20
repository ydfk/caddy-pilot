@echo off
setlocal
set ROOT=%~dp0..

pushd "%ROOT%"
docker compose -f docker-compose.dev.yml up --build
set EXIT=%errorlevel%
popd

exit /b %EXIT%
