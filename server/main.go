package main

import mmplugin "github.com/mattermost/mattermost-server/v6/plugin"

func main() {
	mmplugin.ClientMain(&Plugin{})
}
