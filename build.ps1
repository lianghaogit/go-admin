# 打单文件包：前端构建 -> 复制到 backend/web/dist -> 编译内嵌前端与后端的 Go 二进制
# 用法：powershell -ExecutionPolicy Bypass -File .\build.ps1
$ErrorActionPreference = 'Stop'
$root = $PSScriptRoot

Push-Location (Join-Path $root 'frontend')
npm run build
Pop-Location

# 只复制真正需要的静态文件，排除 .gz/.br 预压缩副本（由静态服务器按需处理即可）
robocopy (Join-Path $root 'frontend\dist') (Join-Path $root 'backend\web\dist') /MIR /XF *.gz *.br /NFL /NDL /NJH /NJS
if ($LASTEXITCODE -ge 8) { throw "复制前端产物失败，robocopy 退出码 $LASTEXITCODE" }

Push-Location (Join-Path $root 'backend')
$env:CGO_ENABLED = '0'
go build -trimpath -ldflags '-s -w' -o go-admin.exe .
Pop-Location

$exe = Join-Path $root 'backend\go-admin.exe'
'{0}  体积 {1:N2} MB' -f $exe, ((Get-Item $exe).Length / 1MB)
