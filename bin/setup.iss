; setup.iss
[Setup]
AppName=Tr
AppVersion=1.1.0
DefaultDirName={pf}\Tr
DefaultGroupName=Tr
UninstallDisplayIcon={app}\Tr.exe
Compression=lzma2
SolidCompression=yes
PrivilegesRequired=admin
ChangesEnvironment=yes
OutputBaseFilename=Tr_v1.1.0_Setup
SetupIconFile=../assets/Tr.ico

[Files]
Source: "Tr.exe"; DestDir: "{app}"

[Icons]
Name: "{group}\Tr"; Filename: "{app}\Tr.exe"
Name: "{group}\Uninstall Tr"; Filename: "{uninstallexe}"

; 注意：这里没有 [Run] 节，安装完成后不会自动运行程序

[Registry]
; 将安装目录添加到系统 PATH（仅当尚未添加时）
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Check: NeedsAddPath('{app}')

[Code]
function NeedsAddPath(Param: string): boolean;
var
  OrigPath: string;
begin
  if not RegQueryStringValue(HKLM, 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment', 'Path', OrigPath) then
  begin
    Result := True;
    exit;
  end;
  // 检查是否已存在，避免重复添加
  Result := Pos(';' + Param + ';', ';' + OrigPath + ';') = 0;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
var
  PathVal: string;
  AppDir: string;
  NewPath: string;
begin
  if CurUninstallStep = usPostUninstall then
  begin
    AppDir := ExpandConstant('{app}');
    if RegQueryStringValue(HKLM, 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment', 'Path', PathVal) then
    begin
      // 从PATH中精确移除安装目录
      StringChangeEx(PathVal, ';' + AppDir, '', True);
      StringChangeEx(PathVal, AppDir + ';', '', True);
      StringChangeEx(PathVal, AppDir, '', True);
      // 清理可能产生的多余分号
      while Pos(';;', PathVal) > 0 do StringChangeEx(PathVal, ';;', ';', True);
      // 清理首尾可能存在的分号
      if (Length(PathVal) > 0) and (PathVal[1] = ';') then Delete(PathVal, 1, 1);
      if (Length(PathVal) > 0) and (PathVal[Length(PathVal)] = ';') then Delete(PathVal, Length(PathVal), 1);
      
      RegWriteStringValue(HKLM, 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment', 'Path', PathVal);
    end;
  end;
end;