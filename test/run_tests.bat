@echo off
echo ========================================
echo    Running Queue Server Tests
echo ========================================

REM Переходим в корневую директорию проекта
cd ..

echo.
echo [1/3] Running unit tests...
echo.

echo Running model tests...
go test ./test/unit/ -run TestModel -v
if %errorlevel% neq 0 exit /b %errorlevel%

echo.
echo Running repository tests...
go test ./test/unit/ -run TestRepository -v
if %errorlevel% neq 0 exit /b %errorlevel%

echo.
echo Running controller tests...
go test ./test/unit/ -run TestController -v
if %errorlevel% neq 0 exit /b %errorlevel%

echo.
echo Running worker pool tests...
go test ./test/unit/ -run TestWorkerPool -v
if %errorlevel% neq 0 exit /b %errorlevel%

echo.
echo [2/3] Running all unit tests together...
go test ./test/unit/ -v
if %errorlevel% neq 0 exit /b %errorlevel%

echo.
echo [3/3] Running integration tests...
go test ./test/integration/ -v -tags=integration
if %errorlevel% neq 0 exit /b %errorlevel%

echo.
echo ========================================
echo    All tests completed successfully!
echo ========================================
pause