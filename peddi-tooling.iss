; Script generated for peddi-tooling CLI
[Setup]
AppName=Peddi Tooling
AppVersion=1.0
DefaultDirName={pf}\PeddiTooling
DefaultGroupName=Peddi Tooling
OutputBaseFilename=peddi-tooling-installer
Compression=lzma
SolidCompression=yes

[Files]
Source: "C:\path\to\peddi-tooling.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\Peddi Tooling"; Filename: "{app}\peddi-tooling.exe"

[Run]
; Optionally add folder to user PATH
Filename: "powershell.exe"; Parameters: "-Command `"setx PATH `$([Environment]::GetEnvironmentVariable('PATH', 'User') + ';{app}')`""; Flags: runhidden
