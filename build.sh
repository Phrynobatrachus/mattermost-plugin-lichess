GOOS=linux GOARCH=amd64 go build -o lichess-plugin.exe server/plugin.go
cd webapp && ./node_modules/.bin/webpack --mode=production && cd -
tar -czvf lichess-plugin.tar.gz lichess-plugin.exe plugin.json webapp/dist/*