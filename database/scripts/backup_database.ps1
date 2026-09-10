[CmdletBinding()]
param(
    [string]$ProjectRoot
)

$ErrorActionPreference = 'Stop'

if ([string]::IsNullOrWhiteSpace($ProjectRoot)) {
    $ProjectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..\..')).Path
} else {
    $ProjectRoot = (Resolve-Path -LiteralPath $ProjectRoot).Path
}

$envPath = Join-Path $ProjectRoot 'backend\.env'
if (-not (Test-Path -LiteralPath $envPath -PathType Leaf)) {
    throw "Missing local environment file: $envPath"
}

$settings = @{}
foreach ($line in Get-Content -LiteralPath $envPath) {
    if ($line -match '^\s*([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)\s*$') {
        $settings[$Matches[1]] = $Matches[2].Trim().Trim('"').Trim("'")
    }
}

$requiredKeys = @(
    'DATABASE_HOST',
    'DATABASE_PORT',
    'DATABASE_NAME',
    'DATABASE_USER',
    'DATABASE_PASSWORD'
)
foreach ($key in $requiredKeys) {
    if (-not $settings.ContainsKey($key) -or [string]::IsNullOrWhiteSpace($settings[$key])) {
        throw "Missing $key in backend/.env"
    }
}

$pgDump = (Get-Command pg_dump.exe -ErrorAction SilentlyContinue).Source
if ([string]::IsNullOrWhiteSpace($pgDump)) {
    $knownPaths = @(
        'C:\Program Files\PostgreSQL\18\bin\pg_dump.exe',
        'C:\Program Files\PostgreSQL\17\bin\pg_dump.exe',
        'C:\Program Files\PostgreSQL\16\bin\pg_dump.exe'
    )
    $pgDump = $knownPaths | Where-Object { Test-Path -LiteralPath $_ -PathType Leaf } | Select-Object -First 1
}
if ([string]::IsNullOrWhiteSpace($pgDump)) {
    throw 'pg_dump.exe was not found. Add PostgreSQL bin to PATH or install PostgreSQL client tools.'
}

$databaseDirectory = Join-Path $ProjectRoot 'database'
$dumpPath = Join-Path $databaseDirectory 'bwp-sonasea.dump'
$schemaPath = Join-Path $databaseDirectory 'full_app_schema.sql'
$dumpTemp = "$dumpPath.partial"
$schemaTemp = "$schemaPath.partial"

foreach ($path in @($dumpTemp, $schemaTemp)) {
    if (Test-Path -LiteralPath $path) {
        $item = Get-Item -LiteralPath $path -Force
        if ($item.PSIsContainer) {
            throw "Refusing directory at temporary output path: $path"
        }
        [IO.File]::Delete((Resolve-Path -LiteralPath $path).Path)
    }
}

$connectionArguments = @(
    '--host', $settings['DATABASE_HOST'],
    '--port', $settings['DATABASE_PORT'],
    '--username', $settings['DATABASE_USER'],
    '--dbname', $settings['DATABASE_NAME']
)
$previousPassword = $env:PGPASSWORD
$env:PGPASSWORD = $settings['DATABASE_PASSWORD']
$published = $false

try {
    & $pgDump @connectionArguments '--format=custom' '--no-owner' '--no-privileges' '--file' $dumpTemp
    if ($LASTEXITCODE -ne 0) {
        throw "pg_dump backup failed with exit code $LASTEXITCODE"
    }

    & $pgDump @connectionArguments '--schema-only' '--no-owner' '--no-privileges' '--no-comments' '--file' $schemaTemp
    if ($LASTEXITCODE -ne 0) {
        throw "pg_dump schema export failed with exit code $LASTEXITCODE"
    }

    foreach ($path in @($dumpPath, $schemaPath)) {
        if (Test-Path -LiteralPath $path) {
            $item = Get-Item -LiteralPath $path -Force
            if ($item.PSIsContainer) {
                throw "Refusing directory at output path: $path"
            }
            [IO.File]::Delete((Resolve-Path -LiteralPath $path).Path)
        }
    }

    [IO.File]::Move((Resolve-Path -LiteralPath $dumpTemp).Path, $dumpPath)
    [IO.File]::Move((Resolve-Path -LiteralPath $schemaTemp).Path, $schemaPath)
    $published = $true
}
finally {
    if (-not $published) {
        foreach ($path in @($dumpTemp, $schemaTemp)) {
            if (Test-Path -LiteralPath $path -PathType Leaf) {
                [IO.File]::Delete((Resolve-Path -LiteralPath $path).Path)
            }
        }
    }

    if ($null -eq $previousPassword) {
        [Environment]::SetEnvironmentVariable('PGPASSWORD', $null, 'Process')
    } else {
        $env:PGPASSWORD = $previousPassword
    }
}

Get-Item -LiteralPath $dumpPath, $schemaPath |
    Select-Object FullName, Length, LastWriteTime
