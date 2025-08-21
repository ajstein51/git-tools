[Setup]
AppName=Peddi Tooling
AppVersion=1.0
DefaultDirName={userappdata}\PeddiTooling
DefaultGroupName=Peddi Tooling
OutputBaseFilename=peddi-tooling-installer
Compression=lzma
SolidCompression=yes
PrivilegesRequired=lowest
ArchitecturesInstallIn64BitMode=x64os

[Files]
Source: "D:\source\repos\tooling\git-tools\peddi-tooling.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "D:\source\repos\tooling\git-tools\uninstall-completion.ps1"; DestDir: "{app}"; Flags: ignoreversion
Source: "D:\source\repos\tooling\git-tools\install-completion.ps1"; DestDir: "{tmp}"; Flags: deleteafterinstall

[Icons]
Name: "{group}\Peddi Tooling"; Filename: "{app}\peddi-tooling.exe"

[Registry]
; We do not touch system PATH, user PATH is updated in PowerShell script

[Run]
Filename: "{win}\System32\WindowsPowerShell\v1.0\powershell.exe"; \
Parameters: "-NoProfile -ExecutionPolicy Bypass -File ""{tmp}\install-completion.ps1"" -InstallFolder ""{app}"""; \
Flags: runhidden

[UninstallRun]
Filename: "{win}\System32\WindowsPowerShell\v1.0\powershell.exe"; \
Parameters: "-NoProfile -ExecutionPolicy Bypass -File ""{app}\uninstall-completion.ps1"" -InstallFolder ""{app}"""; \
Flags: runhidden

