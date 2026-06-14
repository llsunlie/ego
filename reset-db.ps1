$ErrorActionPreference = "Stop"

$DbService = if ($env:DB_SERVICE) { $env:DB_SERVICE } else { "postgres" }
$DbUser = if ($env:POSTGRES_USER) { $env:POSTGRES_USER } else { "ego" }
$DbName = if ($env:POSTGRES_DB) { $env:POSTGRES_DB } else { "ego" }
$EsService = if ($env:ES_SERVICE) { $env:ES_SERVICE } else { "elasticsearch" }
$EsUrl = if ($env:ELASTICSEARCH_URL) { $env:ELASTICSEARCH_URL } else { "http://localhost:9200" }

function Log($Message) {
    Write-Host "[$(Get-Date -Format HH:mm:ss)] $Message" -ForegroundColor Green
}

function Warn($Message) {
    Write-Host "[$(Get-Date -Format HH:mm:ss)] $Message" -ForegroundColor Yellow
}

function DockerComposeServiceRunning($Service) {
    docker compose ps $Service --status running *> $null
    return $LASTEXITCODE -eq 0
}

if (-not (DockerComposeServiceRunning $DbService)) {
    Warn "postgres is not running; starting docker compose service '$DbService'..."
    docker compose up -d $DbService
}

Log "waiting for postgres..."
do {
    docker compose exec -T $DbService pg_isready -U $DbUser -d $DbName *> $null
    if ($LASTEXITCODE -ne 0) {
        Start-Sleep -Milliseconds 500
    }
} while ($LASTEXITCODE -ne 0)

Log "truncating all public tables in database '$DbName'..."
$Sql = @'
DO $$
DECLARE
  table_list text;
BEGIN
  SELECT string_agg(format('%I.%I', schemaname, tablename), ', ')
  INTO table_list
  FROM pg_tables
  WHERE schemaname = 'public';

  IF table_list IS NULL THEN
    RAISE NOTICE 'no public tables found';
  ELSE
    EXECUTE 'TRUNCATE TABLE ' || table_list || ' RESTART IDENTITY CASCADE';
  END IF;
END $$;
'@
$Sql | docker compose exec -T $DbService psql -U $DbUser -d $DbName -v ON_ERROR_STOP=1
Log "done - all public tables cleared"

if (-not (Get-Command curl.exe -ErrorAction SilentlyContinue)) {
    Warn "curl.exe is not available; skipped Elasticsearch cleanup"
    exit 0
}

if (-not (DockerComposeServiceRunning $EsService)) {
    Warn "elasticsearch is not running; starting docker compose service '$EsService'..."
    docker compose up -d $EsService
}

Log "waiting for elasticsearch..."
do {
    curl.exe -fsS "$EsUrl/_cluster/health?wait_for_status=yellow&timeout=1s" *> $null
    if ($LASTEXITCODE -ne 0) {
        Start-Sleep -Milliseconds 500
    }
} while ($LASTEXITCODE -ne 0)

Log "deleting all non-system Elasticsearch indices at '$EsUrl'..."
$IndicesRaw = curl.exe -fsS "$EsUrl/_cat/indices?h=index"
$Indices = $IndicesRaw -split "`n" | ForEach-Object { $_.Trim() } | Where-Object { $_ -and -not $_.StartsWith(".") }
if (-not $Indices) {
    Warn "no non-system Elasticsearch indices found"
} else {
    foreach ($Index in $Indices) {
        curl.exe -fsS -X DELETE "$EsUrl/$Index" *> $null
        Log "  deleted ES index $Index"
    }
}

Log "done - Elasticsearch indices cleared"
