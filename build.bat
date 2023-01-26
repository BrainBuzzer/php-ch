@echo off
SET INNOSETUP=%CD%\installer.iss
SET ORIG=%CD%
REM SET GOPATH=%CD%\src
SET GOBIN=%CD%\bin
REM Support for older architectures
SET GOARCH=386

REM Cleanup existing build if it exists
if exist bin\php-ch.exe (
  del bin\php-ch.exe
)

REM Make the executable and add to the binary directory
echo ----------------------------
echo Building php-ch.exe
echo ----------------------------
go build -o bin/php-ch.exe

REM Group the file with the helper binaries
move php-ch.exe %GOBIN%

REM Codesign the executable
echo ----------------------------
echo Sign the nvm executable...
echo ----------------------------
buildtools\signtool.exe sign /debug /tr http://timestamp.sectigo.com /td sha256 /fd sha256 /a %GOBIN%\php-ch.exe

for /f %%i in ('"%GOBIN%\php-ch.exe" version') do set AppVersion=%%i
echo php-ch.exe v%AppVersion% built.

REM Create the distribution folder
SET DIST=%CD%\dist\%AppVersion%

REM Remove old build files if they exist.
if exist %DIST% (
  echo ----------------------------
  echo Clearing old build in %DIST%
  echo ----------------------------
  rd /s /q "%DIST%"
)

REM Create the distribution directory
mkdir "%DIST%"

REM Generate the installer (InnoSetup)
echo ----------------------------
echo Generating installer...
echo ----------------------------
buildtools\iscc "%INNOSETUP%" "/o%DIST%"

echo ----------------------------
echo Sign the installer
echo ----------------------------
buildtools\signtool.exe sign /debug /tr http://timestamp.sectigo.com /td sha256 /fd sha256 /a %DIST%\php-ch-setup.exe

REM Generate checksums
echo ----------------------------
echo Generating checksums...
echo ----------------------------
for %%f in (%DIST%\*.*) do (certutil -hashfile "%%f" MD5 | find /i /v "md5" | find /i /v "certutil" >> "%%f.checksum.txt")
echo complete

echo ----------------------------
echo Cleaning up...
echo ----------------------------
del %GOBIN%\php-ch.exe
echo complete
@REM del %GOBIN%\php-ch-setup.exe

echo php-ch v%AppVersion% build completed. Available in %DIST%
@echo on
