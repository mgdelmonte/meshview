package meshview

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fogleman/fauxgl"
)


// LoadMesh (MGD)
func LoadMesh(path string) (*MeshData, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".stl":
		return LoadSTL(path)
	case ".obj":
		return LoadOBJ(path)
	}
	return nil, fmt.Errorf("unrecognized mesh extension: %s", ext)
}

func boxForData(data []float32) fauxgl.Box {
	minx := data[0]
	maxx := data[0]
	miny := data[1]
	maxy := data[1]
	minz := data[2]
	maxz := data[2]
	for i := 0; i < len(data); i += 3 {
		x := data[i+0]
		y := data[i+1]
		z := data[i+2]
		if x < minx {
			minx = x
		}
		if x > maxx {
			maxx = x
		}
		if y < miny {
			miny = y
		}
		if y > maxy {
			maxy = y
		}
		if z < minz {
			minz = z
		}
		if z > maxz {
			maxz = z
		}
	}
	min := fauxgl.Vector{X:float64(minx), Y:float64(miny), Z:float64(minz)}
	max := fauxgl.Vector{X:float64(maxx), Y:float64(maxy), Z:float64(maxz)}
	return fauxgl.Box{Min:min, Max:max}
}
