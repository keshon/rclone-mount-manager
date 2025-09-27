@echo off
setlocal

rem ============================================================
rem Build rclone-mount.exe with icon + DPI-aware manifest
rem ============================================================

cd cmd\rclone-mount

rem Generate .syso with icon + manifest
rsrc -ico ../../app.ico -manifest ../../app.manifest -o app.syso

cd ../..

rem Get the build date
for /f "tokens=*" %%a in ('powershell -command "Get-Date -UFormat '%%Y-%%m-%%dT%%H:%%M:%%SZ'"') do set BUILD_DATE=%%a

rem Build without console window
go build -ldflags="-H=windowsgui -X app/internal/version.BuildDate=%BUILD_DATE%" -o rclone-mount.exe ./cmd/rclone-mount

endlocal
