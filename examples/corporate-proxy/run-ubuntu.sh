# run proxy
gnome-terminal -- .\proxy\proxy
sleep 1
# run first endpoint
gnome-terminal -- ..\..\agent\cli\cli configB.json
sleep 1
# run second endpoint (socks5)
gnome-terminal -- ..\..\agent\cli\cli configA.json
