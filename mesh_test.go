package meshview

import (
	"runtime"
	"testing"
	"github.com/fogleman/fauxgl"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

func init() {
	runtime.LockOSThread()
}

func setup(t *testing.T) {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.Samples, 4)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	window, err := glfw.CreateWindow(200, 200, "test", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		panic(err)
	}	
}

func teardown(t *testing.T) {
	glfw.Terminate()
}

func TestGl(t *testing.T) {
	setup(t)
	defer teardown(t)
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	if vao == 0 {
		t.Errorf("bad vao")
	}
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	if vbo == 0 {
		t.Errorf("bad buf")
	}
}

func TestNewVao(t *testing.T) {
	setup(t)
	defer teardown(t)
	b := []float32{1,2,3}
	v := NewVao(b)
	if v.Buf == 0 {
		t.Errorf("bad buf")
	}
	if v.Len != int32(len(b)) {
		t.Errorf("bad len")
	}
}

func TestTriangles2Vao(t *testing.T) {
	setup(t)
	defer teardown(t)
	triangles := []*fauxgl.Triangle{}
	tr := fauxgl.NewTriangleForPoints(fauxgl.V(0,0,0), fauxgl.V(1,1,1), fauxgl.V(2,2,2))
	triangles = append(triangles, tr)
	triangles = append(triangles, tr)
	triangles = append(triangles, tr)
	v := Triangles2Vao(triangles)
	if v.Buf == 0 {
		t.Errorf("bad buf")
	}
	if v.Len != int32(len(triangles)*9) {
		t.Errorf("bad len %d", v.Len)
	}
}
