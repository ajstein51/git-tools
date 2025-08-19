# List all paths you want to include
$pathsToAdd = @(
    "C:\Program Files\PowerShell\7"
    "C:\Windows\system32"
    "C:\Windows"
    "C:\Windows\System32\Wbem"
    "C:\Windows\System32\WindowsPowerShell\v1.0\"
    "C:\Windows\System32\OpenSSH\"
    "C:\Program Files (x86)\NVIDIA Corporation\PhysX\Common"
    "C:\Program Files\Microsoft SQL Server\130\Tools\Binn\"
    "C:\Program Files\Microsoft SQL Server\150\Tools\Binn\"
    "C:\Program Files\WiX Toolset v5.0\bin\"
    "C:\Program Files\Microsoft SQL Server\Client SDK\ODBC\170\Tools\Binn\"
    "C:\Program Files\dotnet\"
    "C:\Program Files (x86)\Microsoft SQL Server\160\DTS\Binn\"
    "C:\Program Files\TortoiseSVN\bin"
    "C:\Program Files\Git\cmd"
    "C:\Program Files\WiX Toolset v6.0\bin\"
    "C:\Program Files\PowerShell\7\"
    "C:\Program Files (x86)\Windows Kits\10\Windows Performance Toolkit\"
    "C:\Users\AJ\AppData\Local\Microsoft\WindowsApps"
    "C:\Users\AJ\.dotnet\tools"
    "C:\Users\AJ\AppData\Local\Programs\Microsoft VS Code\bin"
)

# Get current system PATH
$oldPath = [System.Environment]::GetEnvironmentVariable("Path","Machine")

# Combine paths, avoiding duplicates
$newPath = ($oldPath.Split(';') + $pathsToAdd | Sort-Object -Unique) -join ';'

# Set new system PATH
[System.Environment]::SetEnvironmentVariable("Path", $newPath, "Machine")

Write-Host "System PATH updated. You may need to restart your PowerShell or log off/on to see changes."
