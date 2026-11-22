@echo off
chcp 65001 >nul
cd /d "%~dp0backend"
start "go-admin-backend" cmd /k "go run ."
cd /d "%~dp0frontend"
start "go-admin-frontend" cmd /k "npm run dev"
echo.
echo Go Admin 已启动：
echo   后端 http://localhost:8080
echo   前端 http://localhost:5173
echo 两个窗口独立运行，关闭窗口即停止对应服务。
echo.