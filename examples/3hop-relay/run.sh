# run central hub
../../agent/cli/cli configB.json &
sleep 1
# run first endpoint (socks5)
../../agent/cli/cli configA.json &
sleep 1
# run second endpoint (remote)
../../agent/cli/cli configC.json &
