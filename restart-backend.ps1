# Restart ego backend: kill old process, build, start new.
# Usage: .\restart-backend.ps1

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$grpcPort = if ($env:PORT) { $env:PORT } else { "9443" }
$webPort  = if ($env:WEB_PORT) { $env:WEB_PORT } else { "9080" }

# ── kill old backend ──────────────────────────────────────────────
Write-Host "=== killing old backend (ports $grpcPort $webPort) ==="
$pids = netstat -ano 2>$null | Select-String ":(9443|9080) " | Select-String "LISTENING" | ForEach-Object { (-split $_)[-1] } | Sort-Object -Unique
if ($pids) {
    foreach ($p in $pids) {
        taskkill /F /PID $p 2>$null
        Write-Host "  killed PID $p"
    }
}
Start-Sleep -Seconds 0.5

# ── build ─────────────────────────────────────────────────────────
Write-Host "=== building backend ==="
Push-Location "$root/server"
$prev = $ErrorActionPreference
$ErrorActionPreference = "Stop"
try {
    go build -o "$env:TEMP/ego-server.exe" ./cmd/ego/
    Write-Host "  build ok"
} catch {
    Write-Host "[ERROR] BUILD FAILED: $_"
    Pop-Location
    exit 1
} finally {
    $ErrorActionPreference = $prev
}
Pop-Location

# ── start ─────────────────────────────────────────────────────────
Write-Host "=== starting backend ==="
$proc = Start-Process -FilePath "$env:TEMP/ego-server.exe" `
    -WorkingDirectory "$root/server" `
    -PassThru `
    -NoNewWindow `
    -RedirectStandardError "$env:TEMP/ego-server-err.log" `
    -RedirectStandardOutput "$env:TEMP/ego-server-out.log"

Start-Sleep -Seconds 2

if ($proc.HasExited) {
    Write-Host "[ERROR] backend exited with code $($proc.ExitCode)" -ForegroundColor Red
    Write-Host "--- stderr ---" -ForegroundColor Red
    if (Test-Path "$env:TEMP/ego-server-err.log") {
        Get-Content "$env:TEMP/ego-server-err.log" | ForEach-Object { Write-Host $_ -ForegroundColor Red }
    }
    exit 1
}

for ($i = 0; $i -lt 20; $i++) {
    if (netstat -ano 2>$null | Select-String ":$grpcPort .*LISTENING") {
        Write-Host ""
        Write-Host "  backend ready   gRPC :${grpcPort}  gRPC-web :${webPort}" -ForegroundColor Green
        exit 0
    }
    Start-Sleep -Seconds 0.3
}

Write-Host "[ERROR] backend failed to start" -ForegroundColor Red
exit 1
