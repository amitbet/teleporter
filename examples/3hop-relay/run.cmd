REM run central hub
start "hub" ..\..\agent\cli\cli.exe configB.json
timeout /t 1
REM run first endpoint
start "node1 (socks5 server)" ..\..\agent\cli\cli.exe configA.json
timeout /t 1
REM run second endpoint
start "node2" ..\..\agent\cli\cli.exe configC.json
