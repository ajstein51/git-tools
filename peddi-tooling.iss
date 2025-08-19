[Setup]
AppName=Peddi Tooling
AppVersion=1.0
DefaultDirName={commonpf}\PeddiTooling
DefaultGroupName=Peddi Tooling
OutputBaseFilename=peddi-tooling-installer
Compression=lzma
SolidCompression=yes
PrivilegesRequired=admin
ArchitecturesInstallIn64BitMode=x64os


[Files]
Source: "C:\Users\AJ\source\repos\git-tools\peddi-tooling.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "C:\Users\AJ\source\repos\git-tools\install-completion.ps1"; DestDir: "{tmp}"; Flags: deleteafterinstall

[Icons]
Name: "{group}\Peddi Tooling"; Filename: "{app}\peddi-tooling.exe"

[Registry]
Root: HKCU; Subkey: "Environment"; ValueType: expandsz; ValueName: "Path"; \
    ValueData: "{olddata};{app}"; Flags: preservestringtype uninsdeletevalue


[Run]
Filename: "{win}\System32\WindowsPowerShell\v1.0\powershell.exe"; \
    Parameters: "-NoProfile -ExecutionPolicy Bypass -File ""{tmp}\install-completion.ps1"""; Flags: runhidden
