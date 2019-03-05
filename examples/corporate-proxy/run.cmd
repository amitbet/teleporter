REM run proxy
start "proxy" .\proxy\proxy.exe 
timeout /t 1
REM run first endpoint
start "node2" ..\..\agent\cli\cli.exe configB.json
timeout /t 1
REM run second endpoint
start "node1 (socks5 server)"  ..\..\agent\cli\cli.exe configA.json
