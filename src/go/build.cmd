@echo off
setlocal

:: Mimic 'exit on error' behavior by checking errorlevel after commands that might fail
set "VERSION=%1"
if "%VERSION%"=="" (
    echo Error: Version argument is missing.
    exit /b 1
)

:: Move to script's directory
pushd "%~dp0"

:: Functionality previously in 'clean' - Remove existing build directory
if exist build\windows rmdir /s /q build\windows

:: Mimic 'prepare' functionality - Version-related operations could be added here
:: For demonstration, it's just echoing the version
echo Preparing version %VERSION%...

:: Build process - Adjusted to assume Windows target and no cross-compilation
set "PROJECT_NAME=m2cp"
if not exist build\windows-amd64 mkdir build\windows-amd64
go build -o build\windows-amd64\%PROJECT_NAME%.exe

:: Restore initial directory
popd

echo Build completed successfully.
endlocal
