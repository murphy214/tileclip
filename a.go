package main

import (
	"github.com/paulmach/go.geojson"
	m "github.com/murphy214/mercantile"
	"math"
	"fmt"
	"io/ioutil"
)

type ClipGeom struct {
	Geom      [][]float64
	NewGeom   [][][]float64
	K1        float64
	K2        float64
	Axis      int
	IsPolygon bool
	SlicePos  int
}

type Slice struct {
	Pos   int
	Slice [][]float64
	Axis  int
}

func NewSlice(axis int) *Slice {
	return &Slice{Slice:[][]float64{}, Axis: axis}
}

func (slice *Slice) IntersectX(ax, ay, bx, by, x float64) float64 {
	t := (x - ax) / (bx - ax)
	if (bx-ax) == 0 {
		slice.Slice = append(slice.Slice,[]float64{x,ay})	
		slice.Pos += 1
		return 0.0
	} else {
		fmt.Println(t,ax,ay,bx,by,x,"here")
		slice.Slice = append(slice.Slice,[]float64{x,ay + (by-ay)*t})
		slice.Pos += 1
		fmt.Println(slice.Slice)
	}
	return t
}

func (slice *Slice) IntersectY(ax, ay, bx, by, y float64) float64 {
	t := (y - ay) / (by - ay)
	if (by-ay) == 0 {
		slice.Slice = append(slice.Slice,[]float64{ax,y})	
		slice.Pos += 1
		return 0.0
	} else {
		slice.Slice = append(slice.Slice,[]float64{ax + (bx-ax)*t,y})
		slice.Pos += 1
	}
	return t
}

func (slice *Slice) Intersect(ax, ay, bx, by, val float64) float64 {
	if slice.Axis == 0 {
		return slice.IntersectX(ax, ay, bx, by, val)
	} else if slice.Axis == 1 {
		return slice.IntersectY(ax, ay, bx, by, val)
	}
	return 0.0
}

func (slice *Slice) AddPoint(x, y float64) {
	slice.Slice = append(slice.Slice,[]float64{x,y})
	slice.Pos += 1
}

func (input *ClipGeom) clipLine() {
	slice := NewSlice(input.Axis)
	//lenn := 0
	//var segLen, t int
	var ax, ay, bx, by, a, b float64
	k1, k2 := input.K1, input.K2
	for i := 0; i < len(input.Geom)-1; i += 1 {
		ax = input.Geom[i][0]
		ay = input.Geom[i][1]
		bx = input.Geom[i+1][0]
		by = input.Geom[i+1][1]
		if input.Axis == 0 {
			a = ax
			b = bx
		} else if input.Axis == 1 {
			a = ay
			b = by
		}
		exited := false

		if a <= k1 {
			// ---|-->  | (line enters the clip region from the left)
			if b >= k1 {
				slice.Intersect(ax, ay, bx, by, k1)
				//if (trackMetrics) slice.start = len + segLen * t;
			}
		} else if a >= k2 {
			// |  <--|--- (line enters the clip region from the right)
			if b <= k2 {
				slice.Intersect(ax, ay, bx, by, k2)
			}
		} else {
			slice.AddPoint(ax, ay)
		}
		if b < k1 && a >= k1 {
			// <--|---  | or <--|-----|--- (line exits the clip region on the left)
			slice.Intersect(ax, ay, bx, by, k1)
			exited = true
		}
		if b > k2 && a <= k2 {
			// |  ---|--> or ---|-----|--> (line exits the clip region on the right)
			slice.Intersect(ax, ay, bx, by, k2)
			exited = true
		}

		if !input.IsPolygon && exited {
			input.NewGeom = append(input.NewGeom, slice.Slice)
			slice.Slice = [][]float64{}
			slice.Pos = 0
		}

	}

	// add the last point
	last := len(input.Geom) - 1
	ax = input.Geom[last][0]
	ay = input.Geom[last][1]
	if input.Axis == 0 {
		a = ax
	} else if input.Axis == 1 {
		a = ay
	}
	if a >= k1 && a <= k2 {
		slice.AddPoint(ax, ay)
	}

	// close the polygon if its endpoints are not the same after clipping
	last = len(slice.Slice) - 1
	if input.IsPolygon && last >= 3 && (slice.Slice[last][0] != slice.Slice[0][0] || slice.Slice[last][1] != slice.Slice[0][1]) {
		slice.AddPoint(slice.Slice[0][0],slice.Slice[0][1])
	}

	if slice.Pos > 0 {
		input.NewGeom = append(input.NewGeom, slice.Slice)
	}

}

// clips a single line
func clipLine(geom [][]float64, k1, k2 float64, axis int, IsPolygon bool) [][][]float64 {
	clipthing := &ClipGeom{Geom: geom, K1: k1, K2: k2, Axis: axis, IsPolygon: IsPolygon}
	clipthing.clipLine()
	return clipthing.NewGeom
}

// clipping lines
func clipLines(geom [][][]float64, k1, k2 float64, axis int, IsPolygon bool) [][][]float64 {
	if len(geom) == 0 {
		return [][][]float64{}
	} 
	clipthing := &ClipGeom{Geom: geom[0], K1: k1, K2: k2, Axis: axis, IsPolygon: IsPolygon}
	for pos := range geom {
		clipthing.Geom = geom[pos]
		clipthing.clipLine()

	}
	return clipthing.NewGeom
}

func clipMultiPolygon(geom [][][][]float64,k1,k2 float64,axis int, IsPolygon bool) [][][][]float64 {
	newgeom := [][][][]float64{}
	for i := range geom {
		tmp := clipLines(geom[i],k1,k2,axis,IsPolygon)	
		if len(tmp) > 0 {
			newgeom = append(newgeom,tmp)
		}
	}
	return newgeom
}

var Power7 = math.Pow(10,-7)

func DeltaFirstLast(firstpt,lastpt []float64) bool {
	dx,dy := math.Abs(firstpt[0] - lastpt[0]),math.Abs(firstpt[1]-lastpt[1])
	return dx > Power7 || dy > Power7
}

// checks and adds a polygon value
func LintPolygon(polygon [][][]float64) [][][]float64 {
	for pos,line := range polygon {
		fpt,lpt := line[0],line[len(line)-1]
		if DeltaFirstLast(fpt, lpt) {
			polygon[pos] = append(line,fpt)
		} else {
			line[len(line)-1] = fpt
			polygon[pos] = line
		}
	}
	return polygon
}

// checks and adds a point value for a mutli polygon
func LintMultiPolygon(multipolygon [][][][]float64) [][][][]float64 {
	for pos := range multipolygon {
		multipolygon[pos] = LintPolygon(multipolygon[pos])
	}
	return multipolygon
}

// clips an arbitary geometry from a geojson
func clip(geom geojson.Geometry,k1,k2 float64,axis int) geojson.Geometry {
	var newgeom geojson.Geometry
	switch geom.Type {
	case "LineString":
		newgeom.Type = "LineString"
		lines := clipLine(geom.LineString,k1,k2,axis,false)
		if len(lines) == 1 {
			newgeom.LineString = lines[0]
		} else {
			newgeom.LineString = [][]float64{}
			newgeom.Type = "MultiLineString"
			newgeom.MultiLineString = lines
		}
	case "MultiLineString":
		newgeom.Type = "MultiLineString"
		lines := clipLines(geom.MultiLineString,k1,k2,axis,false)
		if len(lines) == 1 {
			newgeom.LineString = lines[0]
			newgeom.Type = "LineString"
			newgeom.MultiLineString = [][][]float64{}
		} else {
			newgeom.MultiLineString = lines
		}
	case "Polygon":
		newgeom.Type = "Polygon"

		geom.Polygon = LintPolygon(geom.Polygon)
		lines := clipLines(geom.Polygon,k1,k2,axis,true)
		if len(lines) > 0 {
			newgeom.Polygon = lines
		}
	case "MultiPolygon":
		newgeom.Type = "MultiPolygon"

		geom.MultiPolygon = LintMultiPolygon(geom.MultiPolygon)
		multilines := clipMultiPolygon(geom.MultiPolygon, k1, k2, axis, true)
		if len(multilines) == 1 {
			newgeom.Polygon = multilines[0]
			newgeom.Type = "Polygon"
			newgeom.MultiPolygon = [][][][]float64{}
		} else if len(multilines) > 0 {
			newgeom.MultiPolygon = multilines
		}
	}
	return newgeom
}

func IsEmpty(geom geojson.Geometry) bool {
	switch geom.Type {
	case "Point":
		return true
	case "MultiPoint":
		return len(geom.MultiPoint) == 0
	case "LineString":
		return len(geom.LineString) == 0
	case "MultiLineString":
		return len(geom.MultiLineString) == 0
	case "Polygon":
		return len(geom.Polygon) == 0
	case "MultiPolygon":
		return len(geom.MultiPolygon) == 0
	}
	return false
}


// clips a tile 
func ClipTile(geom geojson.Geometry,tileid m.TileID) geojson.Geometry {
	bds := m.Bounds(tileid)
	geom = clip(geom,bds.W,bds.E,0)
	geom = clip(geom,bds.S,bds.N,1)
	return geom
}

// clips down a tile level
func ClipDown(geom geojson.Geometry,tileid m.TileID) map[m.TileID]geojson.Geometry {
	bds := m.Bounds(tileid)	
	cs := m.Children(tileid)
	cbds := m.Bounds(cs[0])

	mx,my := cbds.E,cbds.S
	
	l,r := clip(geom,bds.W,mx,0),clip(geom,mx,bds.E,0)
	ld,lu := clip(l,bds.S,my,1),clip(l,my,bds.N,1)
	rd,ru := clip(r,bds.S,my,1),clip(r,my,bds.N,1)
	fmt.Println(bds.S,my,bds.N)
	lut,rut,rdt,ldt := cs[0],cs[1],cs[2],cs[3]

	mymap := map[m.TileID]geojson.Geometry{
		lut:lu,
		rut:ru,
		rdt:rd,
		ldt:ld,
	}

	for k,v := range mymap {
		if IsEmpty(v) {
			delete(mymap,k)
		}
	}
	

	return mymap
}




func readfeatures() []*geojson.Feature {
	bs,_ := ioutil.ReadFile("county.geojson")
	fc,_ := geojson.UnmarshalFeatureCollection(bs)
	return fc.Features
}

func makefeatures(feats []*geojson.Feature,filename string) {
	fc := geojson.NewFeatureCollection()
	fc.Features = feats
	s,err := fc.MarshalJSON()
	fmt.Println(err)
	ioutil.WriteFile(filename,s,0677)
}



func newfeature(geom geojson.Geometry,props map[string]interface{}) *geojson.Feature {
	feat2 := geojson.NewFeature(&geom)
	feat2.Properties = map[string]interface{}{}
	for k,v := range props {
		feat2.Properties[k] = v
	}
	feat2.Properties["COLORKEY"] = "white"
	return feat2
}

func newfeatures(mymap map[m.TileID]geojson.Geometry,props map[string]interface{}) []*geojson.Feature {
	feats := make([]*geojson.Feature,len(mymap))
	i := 0
	for k,geom := range mymap {
		feat2 := &geojson.Feature{Geometry:&geojson.Geometry{}}
		feat2.Geometry.Type = geom.Type
		switch feat2.Geometry.Type {
		case "Point":
			feat2.Geometry.Point = geom.Point
		case "MultiPoint":
			feat2.Geometry.MultiPoint = geom.MultiPoint
		case "LineString":
			feat2.Geometry.LineString = geom.LineString
		case "MultiLineString":
			feat2.Geometry.MultiLineString = geom.MultiLineString
		case "Polygon":
			feat2.Geometry.Polygon = geom.Polygon
		case "MultiPolygon":
			feat2.Geometry.MultiPolygon = geom.MultiPolygon
		}
		fmt.Println(geom.Polygon)
		feat2.Properties = map[string]interface{}{}
		for k,v := range props {
			feat2.Properties[k] = v
		}
		feat2.Properties["TILEID"] = m.Tilestr(k)
		feat2.Properties["COLORKEY"] = "white"
		feats[i] = feat2
		i++
	}
	return feats
}




func main() {
	fs := readfeatures()
	feature := fs[100]
	feature.Properties["COLORKEY"] = "purple"
	geom := *feature.Geometry
	fpt := geom.Polygon[0][0]
	tile := m.Parent(m.Tile(fpt[0],fpt[1],12))
	newgeom := ClipTile(geom,tile)
	imap := ClipDown(newgeom,tile)

	feats := newfeatures(imap,feature.Properties)
	feats = append(feats,feature)
	fmt.Println(feats)
	makefeatures(feats,"a.geojson")

}