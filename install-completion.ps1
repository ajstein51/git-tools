param(
    [Parameter(Mandatory=$true)]
    [string]$InstallFolder
)

# === 1. Update user PATH in the Registry AND for the current session ===
$newPathEntry = $InstallFolder.TrimEnd('\')
$userPathKey = "Registry::HKEY_CURRENT_USER\Environment"
$oldUserPath = (Get-ItemProperty -Path $userPathKey -Name Path -ErrorAction SilentlyContinue).Path

if (-not $oldUserPath) { $oldUserPath = "" }

# Check if the path is already present
$pathArray = $oldUserPath -split ';' | ForEach-Object { $_.Trim() }
if (-not ($pathArray -contains $newPathEntry)) {
    $newUserPath = if ([string]::IsNullOrEmpty($oldUserPath)) { $newPathEntry } else { "$oldUserPath;$newPathEntry" }
    Set-ItemProperty -Path $userPathKey -Name Path -Value $newUserPath
    Write-Host "User PATH registry key updated with: $newPathEntry"

    # IMPORTANT: Update the PATH for this script's current session
    $env:Path = $newUserPath
} else {
    Write-Host "PATH entry already exists."
    $env:Path = $oldUserPath # Ensure current session has the correct PATH
}


# === 2. Generate and Persist PowerShell Completion Script ===
$completionScriptPath = Join-Path $InstallFolder "__peddi-tooling-completion.ps1"

try {
    Write-Host "Generating completion script..."
    # Now that the PATH is set for this session, we can be more confident this works.
    & "$InstallFolder\peddi-tooling.exe" completion powershell | Out-File -FilePath $completionScriptPath -Encoding UTF8

    # Ensure the PowerShell profile file and its directory exist
    $profilePath = $PROFILE
    $profileDir = Split-Path $profilePath -Parent
    if (-not (Test-Path $profileDir)) {
        Write-Host "Profile directory not found. Creating: $profileDir"
        New-Item -ItemType Directory -Path $profileDir -Force | Out-Null
    }
    # Create the profile file if it doesn't exist
    if (-not (Test-Path $profilePath)) {
        Write-Host "Profile file not found. Creating: $profilePath"
        New-Item -ItemType File -Path $profilePath -Force | Out-Null
    }

    # Add the sourcing line to the profile if it's not already there
    $sourceCommand = ". `"$completionScriptPath`"" # Use quotes to handle spaces in path
    if (-not (Select-String -Path $profilePath -Pattern ([regex]::Escape($sourceCommand)) -Quiet)) {
        Add-Content -Path $profilePath -Value $sourceCommand
        Write-Host "PowerShell completion persisted in $profilePath"
    } else {
        Write-Host "PowerShell completion already configured in profile."
    }

} catch {
    Write-Warning "Failed to generate or persist completion script: $_"
}

# === 3. Broadcast Environment Change to the System ===
# This tells other apps (like Explorer) to reload environment variables.
Write-Host "Broadcasting environment variable changes to the system..."
try {
    $csCode = @"
    using System;
    using System.Runtime.InteropServices;
    public class Win32 {
        [DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
        public static extern IntPtr SendMessageTimeout(
            IntPtr hWnd, uint Msg, UIntPtr wParam, string lParam,
            uint fuFlags, uint uTimeout, out UIntPtr lpdwResult);
    }
"@
    Add-Type -TypeDefinition $csCode
    $HWND_BROADCAST = [IntPtr]0xffff;
    $WM_SETTINGCHANGE = 0x1a;
    $result = [UIntPtr]::Zero
    [Win32]::SendMessageTimeout($HWND_BROADCAST, $WM_SETTINGCHANGE, [UIntPtr]::Zero, "Environment", 2, 5000, [ref]$result) | Out-Null
    Write-Host "Broadcast sent."
} catch {
    Write-Warning "Failed to broadcast environment variable change. A restart or logoff may be required."
}

Write-Host "Installation complete. For completion to work, please CLOSE and REOPEN any PowerShell windows."