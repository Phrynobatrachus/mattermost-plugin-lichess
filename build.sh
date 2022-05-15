GOOS=linux GOARCH=amd64 go build -o lichess-plugin-amd64 server/*.go
cd webapp && ./node_modules/.bin/webpack --mode=production && cd -
tar -czvf lichess-plugin.tar.gz lichess-plugin-amd64 plugin.json webapp/dist/*