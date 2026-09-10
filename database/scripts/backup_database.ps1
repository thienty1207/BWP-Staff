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

if (-not $settings.ContainsKey('DATABASE_URL') -or [string]::IsNullOrWhiteSpace($settings['DATABASE_URL'])) {
    throw 'Missing DATABASE_URL in backend/.env'
}

try {
    $databaseUri = [Uri]$settings['DATABASE_URL']
} catch {
    throw 'DATABASE_URL in backend/.env is invalid'
}
if ($databaseUri.Scheme -notin @('postgres', 'postgresql') -or [string]::IsNullOrWhiteSpace($databaseUri.Host)) {
    throw 'DATABASE_URL must use the postgres or postgresql scheme'
}

$userInfo = $databaseUri.UserInfo.Split(':', 2)
if ($userInfo.Count -ne 2 -or [string]::IsNullOrWhiteSpace($userInfo[0])) {
    throw 'DATABASE_URL must contain a database username and password'
}
$databaseUser = [Uri]::UnescapeDataString($userInfo[0])
$databasePassword = [Uri]::UnescapeDataString($userInfo[1])
$databaseName = [Uri]::UnescapeDataString($databaseUri.AbsolutePath.TrimStart('/'))
if ([string]::IsNullOrWhiteSpace($databaseName)) {
    throw 'DATABASE_URL must contain a database name'
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
    '--host', $databaseUri.Host,
    '--port', $databaseUri.Port,
    '--username', $databaseUser,
    '--dbname', $databaseName
)
$previousPassword = $env:PGPASSWORD
$env:PGPASSWORD = $databasePassword
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
