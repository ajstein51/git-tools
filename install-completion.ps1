param(
    [Parameter(Mandatory=$true)]
    [string]$InstallFolder
)

# === Update user PATH ===
$oldUserPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
if (-not $oldUserPath) { $oldUserPath = "" }

$newPathEntry = $InstallFolder.TrimEnd('\')

if (-not ($oldUserPath -split ';' | ForEach-Object { $_.Trim() } | Where-Object { $_ -ieq $newPathEntry })) {
    $newUserPath = if ($oldUserPath) { "$oldUserPath;$newPathEntry" } else { $newPathEntry }
    [System.Environment]::SetEnvironmentVariable("Path", $newUserPath, "User")
}

Write-Host "User PATH updated with: $newPathEntry"

# === Generate PowerShell completion using Cobra CLI ===
$completionScriptPath = Join-Path $InstallFolder "__peddi-tooling-completion.ps1"

try {
    # Generate the completion script for PowerShell
    & "$InstallFolder\peddi-tooling.exe" completion powershell | Out-String | Set-Content -Path $completionScriptPath -Encoding UTF8

    # Load completion for current session
    . $completionScriptPath
    Write-Host "PowerShell tab completion loaded for peddi-tooling (current session)"

    # Persist completion in profile
    $profilePath = $PROFILE
    if (-not (Select-String -Path $profilePath -Pattern "__peddi-tooling-completion.ps1" -Quiet)) {
        Add-Content -Path $profilePath -Value ". '$completionScriptPath'"
        Write-Host "PowerShell completion persisted in $profilePath"
    }
} catch {
    Write-Warning "Failed to generate or load completion script: $_"
}

Write-Host "Installation complete. Restart PowerShell to ensure PATH and completion are fully active."
