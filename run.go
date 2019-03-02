package meshview

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/fogleman/fauxgl"
	"github.com/go-gl/gl/v2.1/gl"
	//"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

var vertexShader = `
#version 120
uniform mat4 matrix;
attribute vec4 position;
varying vec3 ec_pos;
void main() {
	gl_Position = matrix * position;
	ec_pos = vec3(gl_Position);
}
`

var fragmentShader = `
#version 120
varying vec3 ec_pos;
const vec3 light_direction = normalize(vec3(1, -1.5, 1));
const vec3 object_color = vec3(0x5b / 255.0, 0xac / 255.0, 0xe3 / 255.0);
void main() {
	vec3 ec_normal = normalize(cross(dFdx(ec_pos), dFdy(ec_pos)));
	float diffuse = max(0, dot(ec_normal, light_direction)) * 0.9 + 0.15;
	vec3 color = object_color * diffuse;
	gl_FragColor = vec4(color, 1);
}
`

func init() {
	runtime.LockOSThread()
}

func loadModel(path string, ch chan *Model) {
	go func() {
		start := time.Now()
		model, err := LoadModel(path)
		if err != nil {
			log.Println("load error")
			return // TODO: display an error
		}
		log.Printf("loaded %d triangles in %.3f seconds\n", len(model.Mesh.Triangles), time.Since(start).Seconds())
		ch <- model
	}()
}

var sliceIndex = 0
var sliceMax = 0
var lastMatrix = fauxgl.Matrix{}

// Run (MGD)
func Run(path string) {
	start := time.Now()

	// load model in the background
	ch := make(chan *Model)
	loadModel(path, ch)

	// initialize glfw
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	// create the window
	glfw.WindowHint(glfw.Samples, 4)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	window, err := glfw.CreateWindow(1920, 1080, path, nil, nil)
	if err != nil {
		panic(err)
	}
	window.Maximize()
	window.MakeContextCurrent()

	fmt.Printf("window shown at %.3f seconds\n", time.Since(start).Seconds())

	// initialize gl
	if err := gl.Init(); err != nil {
		panic(err)
	}

	// MGD
	//gl.Enable(gl.BLEND)
	//gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	
	glfw.SwapInterval(1)

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)
	gl.ClearColor(float32(0xd4)/255, float32(0xd9)/255, float32(0xde)/255, 1)

	// compile shaders
	program, err := compileProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}
	gl.UseProgram(program)

	matrixUniform := uniformLocation(program, "matrix")
	//positionAttrib := attribLocation(program, "position")

	var model *Model

	// create interactor
	interactor := NewArcball()
	BindInteractor(window, interactor)

	// Get supported line width range and step size
	var lineWidthSizes [2]float32
	gl.GetFloatv(gl.LINE_WIDTH_RANGE, &lineWidthSizes[0])
	var lineWidthStep float32
	gl.GetFloatv(gl.LINE_WIDTH_GRANULARITY, &lineWidthStep)
	log.Println("lws", lineWidthSizes, "lwstep", lineWidthStep)
	gl.LineWidth(1)
	gl.Disable(gl.LINE_STIPPLE)
	gl.Enable(gl.LINE_SMOOTH)
	gl.Hint(gl.LINE_SMOOTH_HINT,  gl.NICEST)

	// render function
	// MGD test not redrawing if no change
	render := func() {
		// WAS gl.Clear(gl.DEPTH_BUFFER_BIT | gl.COLOR_BUFFER_BIT)
		if model != nil {
			matrix := getMatrix(window, interactor, model)
			// MGD
			if matrix != lastMatrix {
				lastMatrix = matrix
				gl.Clear(gl.DEPTH_BUFFER_BIT | gl.COLOR_BUFFER_BIT)
				setMatrix(matrixUniform, matrix.Translate(fauxgl.V(-0.5,0,0)))
				model.MeshVao.Draw()
				// // box the model
				// a := float32(model.Mesh.BoundingBox().Min.MinComponent())
				// b := float32(model.Mesh.BoundingBox().Max.MaxComponent())
				// gl.Begin(gl.LINE_STRIP)
				// gl.Vertex3f(a,a,a)
				// gl.Vertex3f(a,a,b)
				// gl.Vertex3f(b,a,b)
				// gl.Vertex3f(b,a,a)
				// gl.Vertex3f(a,a,a)
				// gl.Vertex3f(a,b,a)
				// gl.Vertex3f(a,b,b)
				// gl.Vertex3f(b,b,b)
				// gl.Vertex3f(b,b,a)
				// gl.Vertex3f(a,b,a)
				// gl.End()

				setMatrix(matrixUniform, matrix.Translate(fauxgl.V(0.5, 0, 0)))
				slice := model.Slices[sliceIndex]
				for _, path := range slice.Paths {
					gl.Begin(gl.LINE_STRIP)
					for _, v := range path {
						gl.Vertex3f(float32(v.X), float32(v.Y), float32(v.Z))
					}
					gl.End()
				}

				// vaos := model.SliceVaos[sliceIndex]
				// for _, vao := range vaos {
				// 	vao.DrawLineStrip()
				// }

				// for _, vaos := range model.SliceVaos {
				// 	for _, vao := range vaos {
				// 		vao.DrawLines()
				// 	}
				// 	break
				// }
				// MGD
				// if mesh.SliceBuffer > 0 {
				// 	gl.BindBuffer(gl.ARRAY_BUFFER, mesh.SliceBuffer)
				// 	gl.EnableVertexAttribArray(positionAttrib)
				// 	gl.VertexAttribPointer(positionAttrib, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
				// 	setMatrix(matrixUniform, matrix.Translate(fauxgl.V(0.5, 0, 0)))
				// 	gl.DrawArrays(gl.LINES, 0, mesh.SliceCount)
				// 	gl.DisableVertexAttribArray(positionAttrib)
				// 	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
				// }
				window.SwapBuffers()
			}
			// MGD
			//layer := slicer.SliceMesh(mesh, 0.1)
			// WAS
			// setMatrix(matrixUniform, matrix)
			// mesh.Draw(positionAttrib)

		}
		//window.SwapBuffers()
	}

	// render during resize
	window.SetFramebufferSizeCallback(func(window *glfw.Window, w, h int) {
		//width, height := window.GetFramebufferSize()
		//log.Println("resizing", width, height)
		render()
	})

	// handle drop events
	window.SetDropCallback(func(window *glfw.Window, filenames []string) {
		loadModel(filenames[0], ch)
		window.SetTitle(filenames[0])
	})

	// main loop
	for !window.ShouldClose() {
		select {
		case newModel := <-ch:
			if model != nil {
				model.Destroy()
			}
			model = newModel
			model.MeshVao = Triangles2Vao(model.Mesh.Triangles)
			sliceIndex = 0
			sliceMax = len(model.Slices)-1

			for _, slice := range model.Slices {
				vaos := []Vao{}
				for _, p := range slice.Paths {
					vaos = append(vaos, Vectors2Vao(p))
				}
				model.SliceVaos = append(model.SliceVaos, vaos)
			}
			
			//log.Printf("first frame at %.3f seconds\n", time.Since(start).Seconds())
			//mesh.Slice((data.Box.Min.Z+data.Box.Max.Z)*0.1)
			//fmt.Printf("sliced at %.3f seconds\n", time.Since(start).Seconds())
		default:
		}
		render()
		glfw.PollEvents()
	}
}

func getMatrix(window *glfw.Window, interactor Interactor, model *Model) fauxgl.Matrix {
	return interactor.Matrix(window).Mul(model.Transform)
}
