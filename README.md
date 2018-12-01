# tileclip - A Simple Tile Clipping Library

### What is it?

Clipping is pretty context dependent, meaning, depending on how your clipping certain geometries, several different approachs may work more effecient. This repository contains two functions for clipping under with two pretty different contexts.

### Minimal Example

Belove is an example of the two different clipping methods on an Alaska geojson polygon feature.

```go
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
  bs,_ := ioutil.ReadFile("alaska.geojson")
	feature,_ := geojson.UnmarshalFeature(bs)

	// used for clipping a specific part of a feature into a single tile
	tileid := mercantile.TileID{1,8,5}
	about_tile_feature := tileclip.ClipTile(feature,tileid)
	about_tile_feature.Properties = map[string]interface{}{"COLORKEY":"purple","TILEID":mercantile.Tilestr(tileid)}
	fmt.Printf("About Tile: %+v Feature: %+v\n",tileid,about_tile_feature)

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
```

#### Output

![](https://user-images.githubusercontent.com/10904982/49332813-0e57f880-f582-11e8-9b21-b7e9afed7c70.png)


### General Usage

#### ClipTile 

ClipTile clips a single feature about a single tile returning a single feature. 

```golang
func ClipTile(feature *geojson.Feature,mercantile m.TileID) *geojson.Feature
```

#### ClipFeature 

ClipFeature clips a feature at a defined endzoom and then collects all tiles with that given endzoom created. Keep_parents is a bool that determines wheter or not to delete the parent or upper tile k/v sets for the returned tilemap. 

```golang
func ClipTile(feature *geojson.Feature,endzoom int,keep_parents bool) map[m.TileID]*geojson.Feature
```


```
```
