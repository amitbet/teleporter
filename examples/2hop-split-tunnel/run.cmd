start "node2" ..\..\agent\cli\cli.exe configB.json
timeout /t 1
start "node1 (socks5 server)"  ..\..\agent\cli\cli.exe configA.json
