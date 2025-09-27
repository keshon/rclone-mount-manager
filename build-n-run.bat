@echo off

rem
rem BUILD
rem

rem Get the build date
for /f "tokens=*" %%a in ('powershell -command "Get-Date -UFormat '%%Y-%%m-%%dT%%H:%%M:%%SZ'"') do set BUILD_DATE=%%a

rem Build command
go build -o rclone-mount.exe -ldflags "-X app/internal/version.BuildDate=%BUILD_DATE%" cmd\rclone-mount\main.go && rclone-mount.exe