# ego monitoring stack startup script
# Prerequisites: Docker
# Usage:
#   .\restart-monitoring.ps1              # start all monitoring services
#   .\restart-monitoring.ps1 -Down        # stop and remove monitoring services

param(
    [switch] $Down
)

$root = Split-Path -Parent $MyInvocation.MyCommand.Path

$services = @("prometheus", "loki", "promtail", "grafana")

function Write-Step($text) {
    Write-Host "`n=== $text ==="
}

if ($Down) {
    Write-Step "stopping monitoring stack"
    Push-Location $root
    docker compose stop @services 2>&1 | Out-Null
    docker compose rm -f @services 2>&1 | Out-Null
    Pop-Location
    Write-Host "  monitoring stack stopped and removed"
    return
}

Write-Step "Docker monitoring stack"

# Ensure log directory exists for promtail mount
$logDir = "$root/server/.tmp/logs/server"
if (-not (Test-Path $logDir)) {
    New-Item -ItemType Directory -Path $logDir -Force | Out-Null
}

Push-Location $root
docker compose up -d prometheus loki promtail grafana 2>&1
$result = $LASTEXITCODE
Pop-Location

if ($result -ne 0) {
    Write-Host "[ERROR] failed to start monitoring stack"
    exit 1
}

Write-Step "health check"
$allReady = $true

# Wait for Prometheus
Write-Host "  waiting for prometheus..."
$ok = $false
for ($i = 0; $i -lt 20; $i++) {
    try {
        $null = Invoke-WebRequest -Uri "http://localhost:9090/-/healthy" -UseBasicParsing -TimeoutSec 1
        $ok = $true
        break
    } catch {}
    Start-Sleep -Seconds 1
}
if ($ok) { Write-Host "  prometheus ready" -ForegroundColor Green }
else { Write-Host "  [WARN] prometheus not ready" -ForegroundColor Yellow; $allReady = $false }

# Wait for Loki
Write-Host "  waiting for loki..."
$ok = $false
for ($i = 0; $i -lt 20; $i++) {
    try {
        $null = Invoke-WebRequest -Uri "http://localhost:3100/ready" -UseBasicParsing -TimeoutSec 1
        $ok = $true
        break
    } catch {}
    Start-Sleep -Seconds 1
}
if ($ok) { Write-Host "  loki ready" -ForegroundColor Green }
else { Write-Host "  [WARN] loki not ready" -ForegroundColor Yellow; $allReady = $false }

# Wait for Grafana
Write-Host "  waiting for grafana..."
$ok = $false
for ($i = 0; $i -lt 20; $i++) {
    try {
        $null = Invoke-WebRequest -Uri "http://localhost:3200/api/health" -UseBasicParsing -TimeoutSec 1
        $ok = $true
        break
    } catch {}
    Start-Sleep -Seconds 1
}
if ($ok) { Write-Host "  grafana ready" -ForegroundColor Green }
else { Write-Host "  [WARN] grafana not ready" -ForegroundColor Yellow; $allReady = $false }

if (-not $allReady) {
    Write-Host "`n  retry: .\restart-monitoring.ps1" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "  monitoring stack running"
Write-Host "  Prometheus: http://localhost:9090"
Write-Host "  Grafana:    http://localhost:3200  (admin/admin)"
Write-Host "  Loki:       http://localhost:3100"
Write-Host ""
