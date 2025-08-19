param(
    [Parameter(Mandatory=$true)]
    [string]$InstallFolder
)

# === Remove PowerShell completion script ===
$completionScriptPath = Join-Path $InstallFolder "__peddi-tooling-completion.ps1"
if (Test-Path $completionScriptPath) {
    Remove-Item -Force $completionScriptPath
    Write-Host "Removed completion script: $completionScriptPath"
}

# === Remove the line from the user's PowerShell profile ===
$profilePath = $PROFILE
if (Test-Path $profilePath) {
    $profileContent = Get-Content $profilePath | Where-Object { $_ -notmatch "__peddi-tooling-completion.ps1" }
    Set-Content -Path $profilePath -Value $profileContent -Encoding UTF8
    Write-Host "Removed completion line from profile: $profilePath"
}

# === Optionally remove the install folder ===
if (Test-Path $InstallFolder) {
    Remove-Item -Recurse -Force $InstallFolder
    Write-Host "Removed install folder: $InstallFolder"
}
