# run central hub
gnome-terminal -- ../../agent/cli/cli configB.json
sleep 1
# run first endpoint (socks5)
gnome-terminal -- ../../agent/cli/cli configA.json
sleep 1
# run second endpoint (remote)
gnome-terminal -- ../../agent/cli/cli configC.json
