@echo off
setlocal

rem ============================================================
rem Build rclone-mount.exe with icon + DPI-aware manifest
rem ============================================================

cd cmd\rclone-mount

rem Generate .syso with icon + manifest
rsrc -ico ../../app.ico -manifest ../../app.manifest -o app.syso

cd ../..

rem Get Go version
for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i

rem Get the build date (UTC)
for /f "tokens=*" %%a in ('powershell -command "Get-Date -UFormat '%%Y-%%m-%%dT%%H:%%M:%%SZ'"') do set BUILD_DATE=%%a

rem Build without console window
go build -ldflags="-H=windowsgui -X app/internal/version.BuildDate=%BUILD_DATE% -X app/internal/version.GoVersion=%GO_VERSION%" -o rclone-mount.exe ./cmd/rclone-mount

endlocal
