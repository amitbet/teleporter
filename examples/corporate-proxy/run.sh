# run proxy
.\proxy\proxy &
sleep 1
# run first endpoint
..\..\agent\cli\cli configB.json &
sleep 1
# run second endpoint (socks5)
..\..\agent\cli\cli configA.json &
