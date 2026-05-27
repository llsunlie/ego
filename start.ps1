# ego full-stack startup script (backend + frontend)
# Prerequisites: Docker, Go, Flutter, server/.env
# Usage:
#   .\start.ps1                    # start both backend + frontend
#   .\start.ps1 -BackendOnly       # start backend only
#   .\start.ps1 -FrontendOnly      # start frontend only
#   .\start.ps1 -SkipDockerCheck   # skip PostgreSQL Docker check
#   .\start.ps1 -SkipBackendBuild  # skip Go build

param(
    [switch] $BackendOnly,
    [switch] $FrontendOnly,
    [switch] $SkipDockerCheck,
    [switch] $SkipBackendBuild
)

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$grpcPort = if ($env:PORT) { $env:PORT } else { "9443" }
$webPort  = if ($env:WEB_PORT) { $env:WEB_PORT } else { "9080" }

$bkProc = $null

function Write-Step($text) {
    Write-Host "`n=== $text ==="
}

function Cleanup {
    Write-Host "`n=== shutting down ===" -ForegroundColor Yellow
    if ($bkProc -and -not $bkProc.HasExited) {
        Write-Host "  stopping backend (PID $($bkProc.Id))..."
        & taskkill /F /PID $bkProc.Id 2>$null | Out-Null
    }
    Write-Host "  done"
}

# 1. Docker PostgreSQL
if (-not $SkipDockerCheck -and -not $FrontendOnly) {
    Write-Step "Docker PostgreSQL"

    $containerName = "ego-postgres-1"
    $container = docker ps -a --filter "name=$containerName" --format "{{.Names}} {{.Status}}" 2>$null

    if (-not $container) {
        Write-Host "[ERROR] container '$containerName' not found. Run: docker compose up -d postgres"
        exit 1
    }

    if ($container -match "Exited") {
        Write-Host "  starting container..."
        docker start $containerName 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-Host "[ERROR] failed to start container"
            exit 1
        }
    }
    else {
        Write-Host "  container already running"
    }

    Write-Host "  waiting for PostgreSQL..."
    $pgReady = $false
    for ($i = 0; $i -lt 30; $i++) {
        $log = docker logs $containerName 2>&1 | Select-String "ready to accept connections"
        if ($log) {
            $pgReady = $true
            Write-Host "  PostgreSQL ready"
            break
        }
        Start-Sleep -Seconds 1
    }
    if (-not $pgReady) { Write-Host "[WARN] PostgreSQL may not be ready yet..." }
}

# 2. Frontend web build (before backend, so WEB_DIR is set for backend process)
if (-not $BackendOnly) {
    Write-Step "Flutter web build"

    Push-Location "$root/client"
    flutter pub get 2>&1 | Out-Null

    Write-Host "  building web (release)..."
    $buildOut = flutter build web 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERROR] flutter build web failed:`n$buildOut"
        Pop-Location
        exit 1
    }
    Write-Host "  build ok"

    $webDir = "$root/client/build/web"
    Pop-Location

    Write-Host "  backend will serve static files from client/build/web"
    $env:WEB_DIR = $webDir
}

# 3. Backend
if (-not $FrontendOnly) {
    Write-Step "killing old backend (ports $grpcPort $webPort)"
    $pids = netstat -ano 2>$null |
        Select-String ":(9443|9080) " |
        Select-String "LISTENING" |
        ForEach-Object { (-split $_)[-1] } |
        Sort-Object -Unique
    if ($pids) {
        foreach ($p in $pids) {
            & taskkill /F /PID $p 2>$null | Out-Null
            Write-Host "  killed PID $p"
        }
        Start-Sleep -Seconds 1
    }
    else { Write-Host "  no old process" }

    if (-not $SkipBackendBuild) {
        Write-Step "building backend"
        Push-Location "$root/server"
        $buildOut = go build -o "$env:TEMP/ego-server.exe" ./cmd/ego/ 2>&1
        $buildOk = $LASTEXITCODE -eq 0
        Pop-Location
        if (-not $buildOk) {
            Write-Host "[ERROR] BUILD FAILED:`n$buildOut"
            exit 1
        }
        Write-Host "  build ok"
    }
    else {
        Write-Step "skipping backend build"
        if (-not (Test-Path "$env:TEMP/ego-server.exe")) {
            Write-Host "[ERROR] binary not found at $env:TEMP/ego-server.exe. Re-run without -SkipBackendBuild."
            exit 1
        }
    }

    Write-Step "starting backend"
    $bkProc = Start-Process -FilePath "$env:TEMP/ego-server.exe" `
        -WorkingDirectory "$root/server" `
        -PassThru `
        -NoNewWindow `
        -RedirectStandardError "$env:TEMP/ego-server-err.log" `
        -RedirectStandardOutput "$env:TEMP/ego-server-out.log"

    Start-Sleep -Seconds 2

    if ($bkProc.HasExited) {
        Write-Host "[ERROR] backend exited (code $($bkProc.ExitCode))"
        Get-Content "$env:TEMP/ego-server-err.log" -ErrorAction SilentlyContinue | Write-Host
        exit 1
    }

    $bkReady = $false
    for ($i = 0; $i -lt 20; $i++) {
        if (netstat -ano 2>$null | Select-String ":$grpcPort .*LISTENING") {
            Write-Host "  backend ready   gRPC :${grpcPort}  gRPC-web :${webPort}" -ForegroundColor Green
            $bkReady = $true
            break
        }
        Start-Sleep -Seconds 0.3
    }
    if (-not $bkReady) {
        Write-Host "[ERROR] backend failed to start"
        exit 1
    }
}

# 4. Banner + wait
if (-not $FrontendOnly) {
    Write-Host ""
    Write-Host "                                                       "
    Write-Host "  ego dev server running                               "
    Write-Host ""
    if (-not $BackendOnly) {
        Write-Host "   Web UI:   http://localhost:${webPort}" -ForegroundColor Green
    }
    Write-Host "   gRPC:     localhost:${grpcPort}"
    Write-Host "   gRPC-web: localhost:${webPort}"
    Write-Host "   Adminer:  http://localhost:10081"
    Write-Host ""
    Write-Host "  Press Ctrl+C to stop..."
    try { $bkProc.WaitForExit() } catch {}
}

Cleanup
