# uninstall-completion.ps1
# This script ONLY removes configuration.

param(
    [Parameter(Mandatory=$true)]
    [string]$InstallFolder
)

# === 1. Remove entry from user PATH ===
$pathToRemove = $InstallFolder.TrimEnd('\')
$userPathKey = "Registry::HKEY_CURRENT_USER\Environment"
$currentUserPath = (Get-ItemProperty -Path $userPathKey -Name Path -ErrorAction SilentlyContinue).Path

# Only proceed if the PATH variable actually exists
if ($currentUserPath) {
    # Split the path into an array of individual directories
    $pathArray = $currentUserPath -split ';'
    
    # Create a new array, keeping only the paths that ARE NOT our application's path
    $newPathArray = $pathArray | Where-Object { $_.Trim() -ne $pathToRemove }
    
    # Join the remaining paths back together with semicolons
    $newUserPath = $newPathArray -join ';'
    
    # Set the cleaned-up path back into the registry
    Set-ItemProperty -Path $userPathKey -Name Path -Value $newUserPath
    Write-Host "Removed '$pathToRemove' from user PATH."
}

# === 2. Remove completion from PowerShell profile ===
$profilePath = $PROFILE

# Only proceed if the user's profile file exists
if (Test-Path $profilePath) {
    # This pattern will find the line that sources our specific completion script
    $sourceCommandPattern = ".*__peddi-tooling-completion.ps1.*"
    $currentProfileContent = Get-Content $profilePath
    
    # Create new content by keeping only the lines that DO NOT match our pattern
    $newProfileContent = $currentProfileContent | Where-Object { $_ -notmatch $sourceCommandPattern }
    
    # Overwrite the profile with the cleaned-up content
    Set-Content -Path $profilePath -Value $newProfileContent
    Write-Host "Removed completion sourcing from PowerShell profile."
}

# === 3. Broadcast Environment Change to the System ===
# This tells other apps to reload environment variables.
Write-Host "Broadcasting environment variable changes..."
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
} catch { }

Write-Host "Uninstallation of PATH and profile settings complete."