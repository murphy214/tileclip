package main

import (
	"github.com/paulmach/go.geojson"
	"github.com/murphy214/tileclip"
	"github.com/murphy214/mercantile"
	"fmt"
	"io/ioutil"
)

func main() {
	// reading alaka geojson file
    bs,err := ioutil.ReadFile("alaska.geojson")
    if err != nil {
        fmt.Println(err)
    }
	feature,err := geojson.UnmarshalFeature(bs)
    if err != nil {
		fmt.Println(err)
	}

	// clipping method about a tile 
	// used for clipping a specific part of a feature into a single tile
	tileid := mercantile.TileID{1,8,5}
	about_tile_feature := tileclip.ClipTile(feature,tileid)
	about_tile_feature.Properties = map[string]interface{}{"COLORKEY":"purple","TILEID":mercantile.Tilestr(tileid)}
	fmt.Printf("About Tile: %+v Feature: %+v\n",tileid,about_tile_feature)

	// this clipping methods clips about many tiles at the same time splitting up each geometry 
	// return a tilemap (map[m.TileID]*geojson.Feature)
	// keep_parents determines whether to keep previous zoom levels in the tileid:feature map
	keep_parents := false
	tilemap := tileclip.ClipFeature(feature,int(tileid.Z),keep_parents)
	
	// collecting all features from map
	feats := []*geojson.Feature{}
	for k,v := range tilemap {
		v.Properties = map[string]interface{}{"TILEID":mercantile.Tilestr(k),"COLORKEY":"white"}
		feats = append(feats,v)
		fmt.Printf("Tile: %+v Feature: %+v\n",k,v)
	}

	feats = append(feats,about_tile_feature)
	tileclip.MakeFeatures(feats, "a.geojson")
}