package controllers

import (
	"fmt"
	"melange/app/framework"
	"net/http"
)

func getTilesFromPlugins(p []framework.Plugin) map[string]map[string]framework.Tile {
	output := make(map[string]map[string]framework.Tile)
	for _, v := range p {
		output[v.Id] = v.Tiles
	}
	return output
}

type AllTiles struct {
	Path string
}

func (a *AllTiles) Handle(req *http.Request) framework.View {
	data, err := framework.AllPlugins(a.Path)
	if err != nil {
		fmt.Println("Error getting all plugins.", err)
		return framework.Error500
	}

	return &framework.JSONView{
		Content: getTilesFromPlugins(data),
	}
}

type CurrentTiles struct {
}

func (c *CurrentTiles) Handle(req *http.Request) framework.View {
	return framework.Error500
}
