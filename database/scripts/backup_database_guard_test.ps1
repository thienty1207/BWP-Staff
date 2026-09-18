[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

$scriptPath = Join-Path $PSScriptRoot 'backup_database.ps1'
$temporaryRoot = Join-Path ([IO.Path]::GetTempPath()) ("hotel-staff-backup-guard-" + [guid]::NewGuid().ToString('N'))

try {
	New-Item -ItemType Directory -Path (Join-Path $temporaryRoot 'backend') -Force | Out-Null
	Set-Content -LiteralPath (Join-Path $temporaryRoot 'backend\.env') -NoNewline -Value 'DATABASE_URL=postgres://ci-user:synthetic-password@127.0.0.1:5432/not_hotel_staff?sslmode=disable'

	$captured = $null
	try {
		& $scriptPath -ProjectRoot $temporaryRoot
		throw 'Expected backup guard to reject a non-hotel_staff target.'
	} catch {
		$captured = $_
	}

	if ($captured.Exception.Message -ne 'Refusing backup: DATABASE_URL targets "not_hotel_staff"; expected "hotel_staff".') {
		throw "Unexpected backup guard message: $($captured.Exception.Message)"
	}
	if ($captured.Exception.Message.Contains('synthetic-password')) {
		throw 'Backup guard error exposed the database password.'
	}
	foreach ($outputName in @('hotel_staff.dump', 'full_app_schema.sql')) {
		if (Test-Path -LiteralPath (Join-Path $temporaryRoot "database\$outputName")) {
			throw "Backup guard created output file: $outputName"
		}
	}
} finally {
	if (Test-Path -LiteralPath $temporaryRoot) {
		Remove-Item -LiteralPath $temporaryRoot -Recurse -Force
	}
}
