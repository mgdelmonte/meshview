package meshview

import (
	//"fmt"
	"math"

	"github.com/fogleman/fauxgl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// Interactor (MGD) handles events
type Interactor interface {
	Matrix(window *glfw.Window) fauxgl.Matrix
	CursorPositionCallback(window *glfw.Window, x, y float64)
	MouseButtonCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey)
	KeyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey)
	ScrollCallback(window *glfw.Window, dx, dy float64)
}

// BindInteractor (MGD) binds it to callbacks
func BindInteractor(window *glfw.Window, interactor Interactor) {
	window.SetCursorPosCallback(glfw.CursorPosCallback(interactor.CursorPositionCallback))
	window.SetMouseButtonCallback(glfw.MouseButtonCallback(interactor.MouseButtonCallback))
	window.SetKeyCallback(glfw.KeyCallback(interactor.KeyCallback))
	window.SetScrollCallback(glfw.ScrollCallback(interactor.ScrollCallback))
}

// Turntable

// type Turntable struct {
// 	Sensitivity float64
// 	Dx, Dy      float64
// 	Px, Py      float64
// 	Scroll      float64
// 	Rotate      bool
// }

// func NewTurntable() Interactor {
// 	t := Turntable{}
// 	t.Sensitivity = 0.5
// 	return &t
// }

// func (t *Turntable) CursorPositionCallback(window *glfw.Window, x, y float64) {
// 	if t.Rotate {
// 		t.Dx += x - t.Px
// 		t.Dy += y - t.Py
// 		t.Px = x
// 		t.Py = y
// 		t.Dy = math.Max(t.Dy, -90/t.Sensitivity)
// 		t.Dy = math.Min(t.Dy, 90/t.Sensitivity)
// 	}
// }

// func (t *Turntable) MouseButtonCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
// 	if button == glfw.MouseButton1 {
// 		if action == glfw.Press {
// 			t.Rotate = true
// 			t.Px, t.Py = window.GetCursorPos()
// 		} else if action == glfw.Release {
// 			t.Rotate = false
// 		}
// 	}
// }

// func (t *Turntable) ScrollCallback(window *glfw.Window, dx, dy float64) {
// 	t.Scroll += dy
// }

// func (t *Turntable) Matrix() fauxgl.Matrix {
// 	s := math.Pow(0.98, t.Scroll)
// 	a1 := fauxgl.Radians(-t.Dx * t.Sensitivity)
// 	a2 := fauxgl.Radians(-t.Dy * t.Sensitivity)
// 	m := fauxgl.Identity()
// 	m = m.Scale(fauxgl.V(s, s, s))
// 	m = m.Rotate(fauxgl.V(math.Cos(a1), math.Sin(a1), 0), a2)
// 	m = m.Rotate(fauxgl.V(0, 0, 1), a1)
// 	return m
// }

// func (t *Turntable) Translation() fauxgl.Vector {
// 	return fauxgl.Vector{}
// }

// Arcball

// Arcball (MGD)
type Arcball struct {
	Sensitivity float64
	Start       fauxgl.Vector
	Current     fauxgl.Vector
	Rotation    fauxgl.Matrix
	Translation fauxgl.Vector
	Scroll      float64
	Rotate      bool
	Pan         bool
}

// NewArcball (MGD)
func NewArcball() Interactor {
	a := Arcball{}
	a.Sensitivity = 20
	a.Rotation = fauxgl.Identity()
	return &a
}

// CursorPositionCallback (MGD)
func (a *Arcball) CursorPositionCallback(window *glfw.Window, x, y float64) {
	if a.Rotate {
		a.Current = arcballVector(window)
	}
	if a.Pan {
		a.Current = screenPosition(window)
	}
}

// MouseButtonCallback (MGD)
func (a *Arcball) MouseButtonCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if button == glfw.MouseButton1 {
		if action == glfw.Press {
			if mods == 0 {
				v := arcballVector(window)
				a.Start = v
				a.Current = v
				a.Rotate = true
			} else {
				v := screenPosition(window)
				a.Start = v
				a.Current = v
				a.Pan = true
			}
		} else if action == glfw.Release {
			if a.Rotate {
				m := arcballRotate(a.Start, a.Current, a.Sensitivity)
				a.Rotation = m.Mul(a.Rotation)
				a.Rotate = false
			}
			if a.Pan {
				d := a.Current.Sub(a.Start)
				a.Translation = a.Translation.Add(d)
				a.Pan = false
			}
		}
	}
}

// KeyCallback (MGD)
func (a *Arcball) KeyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if (action == glfw.Press || action == glfw.Repeat) && mods == 0 {
		if key >= glfw.Key1 && key <= glfw.Key7 {
			a.Translation = fauxgl.Vector{}
			a.Scroll = 0
		}
		switch key {
		case glfw.Key1:
			a.Rotation = fauxgl.Identity()
		case glfw.Key2:
			a.Rotation = fauxgl.Identity().Rotate(fauxgl.V(0, 0, 1), math.Pi/2)
		case glfw.Key3:
			a.Rotation = fauxgl.Identity().Rotate(fauxgl.V(0, 0, 1), math.Pi)
		case glfw.Key4:
			a.Rotation = fauxgl.Identity().Rotate(fauxgl.V(0, 0, 1), -math.Pi/2)
		case glfw.Key5:
			a.Rotation = fauxgl.Identity().Rotate(fauxgl.V(1, 0, 0), math.Pi/2)
		case glfw.Key6:
			a.Rotation = fauxgl.Identity().Rotate(fauxgl.V(1, 0, 0), -math.Pi/2)
		case glfw.Key7:
			a.Rotation = fauxgl.Identity().Rotate(fauxgl.V(1, 1, 0).Normalize(), -math.Pi/4).Rotate(fauxgl.V(0, 0, 1), math.Pi/4)
		case glfw.KeyLeft:
			a.Rotation = a.Rotation.Rotate(fauxgl.V(0,0,1), -math.Pi/60)
		case glfw.KeyRight:
			a.Rotation = a.Rotation.Rotate(fauxgl.V(0,0,1), math.Pi/60)
		case glfw.KeyUp:
			if sliceIndex < sliceMax {
				sliceIndex++
				lastMatrix = fauxgl.Matrix{}
			}
		case glfw.KeyDown:
			if sliceIndex > 0 {
				sliceIndex--
				lastMatrix = fauxgl.Matrix{}
			}
		}
	}
}

// ScrollCallback (MGD)
func (a *Arcball) ScrollCallback(window *glfw.Window, dx, dy float64) {
	a.Scroll += dy
}

// Matrix (MGD)
func (a *Arcball) Matrix(window *glfw.Window) fauxgl.Matrix {
	w, h := window.GetFramebufferSize()
	aspect := float64(w) / float64(h)
	r := a.Rotation
	if a.Rotate {
		r = arcballRotate(a.Start, a.Current, a.Sensitivity).Mul(r)
	}
	t := a.Translation
	if a.Pan {
		t = t.Add(a.Current.Sub(a.Start))
	}
	s := math.Pow(0.98, a.Scroll)
	m := fauxgl.Identity()
	m = m.Scale(fauxgl.V(s, s, s))
	m = r.Mul(m)
	m = m.Translate(t)
	m = m.LookAt(fauxgl.V(0, -3, 0), fauxgl.V(0, 0, 0), fauxgl.V(0, 0, 1))
	//fmt.Println("skipping", aspect)
	m = m.Perspective(50, aspect, 0.1, 100)
	return m
}

func screenPosition(window *glfw.Window) fauxgl.Vector {
	x, y := window.GetCursorPos()
	w, h := window.GetSize()
	x = (x/float64(w))*2 - 1
	y = (y/float64(h))*2 - 1
	return fauxgl.Vector{X:x, Y:0, Z:-y}
}

func arcballVector(window *glfw.Window) fauxgl.Vector {
	x, y := window.GetCursorPos()
	w, h := window.GetSize()
	x = (x/float64(w))*2 - 1
	y = (y/float64(h))*2 - 1
	x /= 4
	y /= 4
	x = -x
	q := x*x + y*y
	if q <= 1 {
		z := math.Sqrt(1 - q)
		return fauxgl.Vector{X:x, Y:z, Z:y}
	} 
	return fauxgl.Vector{X:x, Y:0, Z:y}.Normalize()
}

func arcballRotate(a, b fauxgl.Vector, sensitivity float64) fauxgl.Matrix {
	const eps = 1e-9
	dot := b.Dot(a)
	if math.Abs(dot) < eps || math.Abs(dot-1) < eps {
		return fauxgl.Identity()
	} else if math.Abs(dot+1) < eps {
		return fauxgl.Rotate(a.Perpendicular(), math.Pi*sensitivity)
	} else {
		angle := math.Acos(dot)
		v := b.Cross(a).Normalize()
		return fauxgl.Rotate(v, angle*sensitivity)
	}
}
