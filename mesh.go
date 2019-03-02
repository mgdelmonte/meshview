package meshview

import (
	"log"

	"github.com/fogleman/fauxgl"
	"github.com/fogleman/slicer"
	"github.com/go-gl/gl/v2.1/gl"
	//"github.com/go-gl/gl/v3.3-core/gl"
)

// Vectors2Vao converts a slice of vectors into a vao
func Vectors2Vao(vectors []fauxgl.Vector) Vao {
	buffer := make([]float32, len(vectors)*3)
	for i, v := range vectors {
		copy(buffer[i*3:], v.Points())
	}
	return NewVao(buffer)
}

// // Tessellate converts a slice of vectors into a tessellated polygon vao
// func Tessellate(vectors []fauxgl.Vector) Vao {
// 	buffer := make([]float32, len(vectors)*3)
// 	for i, v := range vectors {
// 		copy(buffer[i*3:], v.Points())
// 	}
// 	glu.NewMesh()
// 	return NewVao(buffer)
// }


// Vao is a buffered vertex array with length
type Vao struct {
	Buf uint32
	Len int32
}

// NewVao makes a Vao from a []float32
func NewVao(buffer []float32) Vao {
	// log.Println("making NewVao from buffer len ", len(buffer))
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(buffer)*4, gl.Ptr(buffer), gl.STATIC_DRAW)
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)
	return Vao{vao, int32(len(buffer))}
}

// Draw draws a vao as triangles
func (vao Vao) Draw() {
	gl.BindVertexArray(vao.Buf)
	gl.DrawArrays(gl.TRIANGLES, 0, vao.Len)
	gl.BindVertexArray(0)
}

// DrawPolygon draws a vao as a polygon
func (vao Vao) DrawPolygon() {
	gl.BindVertexArray(vao.Buf)
	gl.DrawArrays(gl.POLYGON, 0, vao.Len)
}

// DrawLines draws a vao as lines
func (vao Vao) DrawLines() {
	gl.BindVertexArray(vao.Buf)
	gl.DrawArrays(gl.LINES, 0, vao.Len)
}

// DrawLineStrip draws a vao as linestrip
func (vao Vao) DrawLineStrip() {
	gl.BindVertexArray(vao.Buf)
	gl.DrawArrays(gl.LINE_STRIP, 0, vao.Len)
}

// Triangles2Vao converts triangles to a Vao
func Triangles2Vao(triangles []*fauxgl.Triangle) Vao {
	// create buffer
	buffer := make([]float32, len(triangles)*9)
	for i, t := range triangles {
		copy(buffer[i*9:], t.Points())
	}
	return NewVao(buffer)
}

// Model contains the mesh plus vaos and view data
type Model struct {
	Mesh       *fauxgl.Mesh
	Slices     []slicer.Layer
	Transform  fauxgl.Matrix
	MeshVao    Vao
	SliceVaos  [][]Vao
	// BoxVao     Vao
}

// NewModel makes a Model from a Mesh
func NewModel(mesh *fauxgl.Mesh) *Model {
	r := Model{}
	r.Mesh = mesh
	box := mesh.BoundingBox()

	// compute transform to scale and center mesh
	scale := fauxgl.V(2, 2, 2).Div(box.Size()).MinComponent()
	r.Transform = fauxgl.Identity()
	r.Transform = r.Transform.Translate(box.Center().Negate())
	r.Transform = r.Transform.Scale(fauxgl.V(scale, scale, scale))

	// make 10 slices
	r.Slices = slicer.SliceMesh(r.Mesh, (box.Max.Z-box.Min.Z)/250)
	
	// cleanup
	for _, slice := range r.Slices {
		for _, path := range slice.Paths {
			for _, v := range path {
				if v.Z != slice.Z {
					log.Println("slice", slice.Z, "bad point", v)
				}
			}
			if path[0] != path[len(path)-1] {
				log.Println("slice", slice.Z, "has unclosed path", path[0], path[len(path)-1])
			}
		}
	}

	return &r
}

// Draw (MGD)
func (model *Model) Draw() {
	model.MeshVao.Draw()
	// TODO draw active slice and bounding box
}

// Destroy (MGD)
func (model *Model) Destroy() {
	gl.DeleteBuffers(1, &model.MeshVao.Buf)
}

// LoadModel loads a mesh and creates the model
func LoadModel(path string) (*Model, error) {
	mesh, err := fauxgl.LoadMesh(path)
	if err != nil {
		return nil, err
	}
	log.Println("loaded model")
	return NewModel(mesh), nil
}


// MeshData (MGD)
type MeshData struct {
	Buffer []float32
	Box    fauxgl.Box
	Triangles []*fauxgl.Triangle 
}

// Mesh (MGD)
type Mesh struct {
	Transform    fauxgl.Matrix
	VertexBuffer uint32
	VertexCount  int32
	Triangles    []*fauxgl.Triangle 
	SliceBuffer  uint32
	SliceCount   int32
}

// NewMesh (MGD)
func NewMesh(data *MeshData) *Mesh {
	// compute transform to scale and center mesh
	scale := fauxgl.V(2, 2, 2).Div(data.Box.Size()).MinComponent()
	transform := fauxgl.Identity()
	transform = transform.Translate(data.Box.Center().Negate())
	transform = transform.Scale(fauxgl.V(scale, scale, scale))

	// generate vbo
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(data.Buffer)*4, gl.Ptr(data.Buffer), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	// compute number of vertices
	count := int32(len(data.Buffer) / 3)

	return &Mesh{transform, vbo, count, data.Triangles, 0, 0}
}


// Slice generates a vbo for the slice at z
func (mesh *Mesh) Slice(z float64) {
	// copy triangles
	triangles := make([]*slicer.Triangle, len(mesh.Triangles))
	for i, t := range mesh.Triangles {
		triangles[i] = slicer.NewTriangle(t)
	}
	paths := slicer.GetPaths(triangles, z)
	buffer := []float32{}
	for _, path := range paths {
		for _, v := range path {
			buffer = append(buffer, float32(v.X))
			buffer = append(buffer, float32(v.Y))
			buffer = append(buffer, float32(v.Z))
		}
	}
	log.Println("slice buffer is ", len(buffer))
	// generate vbo
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(buffer)*4, gl.Ptr(buffer), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	mesh.SliceBuffer = vbo
	mesh.SliceCount = int32(len(buffer))
}

// Draw (MGD)
func (mesh *Mesh) Draw(positionAttrib uint32) {
	gl.BindBuffer(gl.ARRAY_BUFFER, mesh.VertexBuffer)
	gl.EnableVertexAttribArray(positionAttrib)
	gl.VertexAttribPointer(positionAttrib, 3, gl.FLOAT, false, 12, gl.PtrOffset(0))
	gl.DrawArrays(gl.TRIANGLES, 0, mesh.VertexCount)
	gl.DisableVertexAttribArray(positionAttrib)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}

// Destroy (MGD)
func (mesh *Mesh) Destroy() {
	gl.DeleteBuffers(1, &mesh.VertexBuffer)
}
