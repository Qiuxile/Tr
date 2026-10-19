; setup.iss -- Inno Setup script for Tr
; Build first:  go build -trimpath -ldflags "-s -w -X github.com/Qiuxile/Tr/internal/tr.Version=2.2.0" -o Output\Tr.exe .
; Then compile: iscc setup.iss

[Setup]
AppId={{6F1C2B7A-3D4E-4F8A-9B21-7C5D0E9A4F13}
AppName=Tr
AppVersion=2.2.0
AppPublisher=Qiuxile
AppPublisherURL=https://github.com/Qiuxile/Tr
AppSupportURL=https://github.com/Qiuxile/Tr/issues
AppUpdatesURL=https://github.com/Qiuxile/Tr/releases
VersionInfoVersion=2.2.0
VersionInfoDescription=Tr Terminal Translator Setup
VersionInfoCompany=Qiuxile
VersionInfoCopyright=Copyright (c) 2026 Surile
DefaultDirName={autopf}\Tr
DefaultGroupName=Tr
UninstallDisplayIcon={app}\Tr.exe
Compression=lzma2
SolidCompression=yes
PrivilegesRequired=admin
ChangesEnvironment=yes
OutputBaseFilename=Tr_v2.2.0_Setup
SetupIconFile=assets\Tr.ico
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
WizardStyle=modern

[Tasks]
Name: "desktopicon";   Description: "创建桌面快捷方式";             GroupDescription: "附加任务："; Flags: checkedonce
Name: "startmenuicon"; Description: "创建开始菜单快捷方式";         GroupDescription: "附加任务："; Flags: checkedonce
Name: "addtopath";     Description: "将安装目录加入系统 PATH 环境变量"; GroupDescription: "附加任务："; Flags: checkedonce

[Files]
Source: "Output\Tr.exe"; DestDir: "{app}"

[Icons]
; 开始菜单快捷方式：受任务控制
Name: "{group}\Tr"; Filename: "{app}\Tr.exe"; WorkingDir: "{app}"; Tasks: startmenuicon

; 卸载快捷方式：总是创建，方便卸载
Name: "{group}\Uninstall Tr"; Filename: "{uninstallexe}"

; 桌面快捷方式：受任务控制，默认已选
Name: "{autodesktop}\Tr"; Filename: "{app}\Tr.exe"; WorkingDir: "{app}"; Tasks: desktopicon

[Code]
const
  EnvKey = 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment';

var
  AddPathTaskSelected: Boolean;

function PathContainsDir(const PathVal, Dir: string): Boolean;
begin
  Result := Pos(';' + Uppercase(Dir) + ';',
                ';' + Uppercase(PathVal) + ';') > 0;
end;

procedure AddAppToPath;
var
  PathVal, AppDir, NewPath: string;
begin
  AppDir := ExpandConstant('{app}');

  if not RegQueryStringValue(HKLM, EnvKey, 'Path', PathVal) then
    PathVal := '';

  if PathContainsDir(PathVal, AppDir) then
    Exit;

  if Trim(PathVal) = '' then
    NewPath := AppDir
  else
    NewPath := PathVal + ';' + AppDir;

  RegWriteExpandStringValue(HKLM, EnvKey, 'Path', NewPath);
end;

procedure RemoveAppFromPath;
var
  PathVal, AppDir: string;
begin
  AppDir := ExpandConstant('{app}');

  if not RegQueryStringValue(HKLM, EnvKey, 'Path', PathVal) then
    Exit;

  PathVal := ';' + PathVal + ';';
  StringChangeEx(PathVal, ';' + AppDir + ';', ';', False);

  if (Length(PathVal) > 0) and (PathVal[1] = ';') then
    Delete(PathVal, 1, 1);
  if (Length(PathVal) > 0) and (PathVal[Length(PathVal)] = ';') then
    Delete(PathVal, Length(PathVal), 1);

  while Pos(';;', PathVal) > 0 do
    StringChangeEx(PathVal, ';;', ';', False);

  RegWriteExpandStringValue(HKLM, EnvKey, 'Path', PathVal);
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssInstall then
    AddPathTaskSelected := WizardIsTaskSelected('addtopath');

  if CurStep = ssPostInstall then
  begin
    if AddPathTaskSelected then
      AddAppToPath;
  end;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usPostUninstall then
    RemoveAppFromPath;
end;
