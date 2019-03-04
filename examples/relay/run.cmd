REM run central hub
cd agentb
start cli.exe
timeout /t 1
REM run first endpoint
cd ../agenta
start cli.exe
timeout /t 1
REM run second endpoint
cd ../agentc
start cli.exe