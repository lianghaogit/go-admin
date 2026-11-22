# ============================================================
# go-admin PostgreSQL 备份脚本（Windows PowerShell）
# 用法：
#   pwsh -File backend/scripts/pg_backup.ps1
#   可选参数：-Container my-postgres -Database go_admin -OutDir data\backup
# 产物：data\backup\go_admin_<yyyyMMdd_HHmmss>.sql（含 --clean --if-exists，可直接回滚）
# 依赖：本机可执行 docker CLI，且容器内存在 pg_dump（postgres 官方镜像自带）
# ============================================================
param(
    [string]$Container = "my-postgres",
    [string]$Database = "go_admin",
    [string]$User = "postgres",
    [string]$OutDir = ""
)

$ErrorActionPreference = "Stop"

if ($OutDir -eq "") {
    $OutDir = Join-Path (Split-Path -Parent $PSScriptRoot) "data\backup"
}
if (-not (Test-Path -LiteralPath $OutDir)) {
    New-Item -ItemType Directory -Path $OutDir | Out-Null
}

$stamp = Get-Date -Format "yyyyMMdd_HHmmss"
$outFile = Join-Path $OutDir ("{0}_{1}.sql" -f $Database, $stamp)

Write-Host "导出 $Database 到 $outFile ..."
# 用 Start-Process 直接重定向 stdout 到文件，避免 PowerShell 管道改写换行/编码
$errFile = "$outFile.err"
$args = @("exec", $Container, "pg_dump", "-U", $User, "-d", $Database, "--clean", "--if-exists", "--no-owner")
$proc = Start-Process -FilePath "docker" -ArgumentList $args -NoNewWindow -Wait -PassThru `
    -RedirectStandardOutput $outFile -RedirectStandardError $errFile
if ($proc.ExitCode -ne 0) {
    Write-Host "备份失败，查看 $errFile"
    exit $proc.ExitCode
}

$size = (Get-Item -LiteralPath $outFile).Length
if ($size -le 0) {
    Write-Host "备份文件为空，备份失败"
    exit 1
}
Remove-Item -LiteralPath "$outFile.err" -ErrorAction SilentlyContinue
Write-Host ("备份完成：{0}（{1:N0} 字节）" -f $outFile, $size)
