package gocvui

/*
#include <stdlib.h>
#include "modules/highgui_gocv.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"strings"
	"unsafe"

	"gocv.io/x/gocv"
)

// Lib version
const VERSION string = "2.7"

// Constants regarding component interactions
const ROW int = 0
const COLUMN int = 1
const DOWN int = 2
const CLICK int = 3
const OVER int = 4
const OUT int = 5
const UP int = 6
const IS_DOWN int = 7

// Constants regarding mouse buttons
const LEFT_BUTTON int = 0
const MIDDLE_BUTTON int = 1
const RIGHT_BUTTON int = 2

// Constants regarding components
const TRACKBAR_HIDE_SEGMENT_LABELS int = 1
const TRACKBAR_HIDE_STEP_SCALE int = 2
const TRACKBAR_DISCRETE int = 4
const TRACKBAR_HIDE_MIN_MAX_LABELS int = 8
const TRACKBAR_HIDE_VALUE_LABEL int = 16
const TRACKBAR_HIDE_LABELS int = 32

//! Mouse Events see cv::MouseCallback
// MouseEventTypes
const EVENT_MOUSEMOVE int = 0     //!< indicates that the mouse pointer has moved over the window.
const EVENT_LBUTTONDOWN int = 1   //!< indicates that the left mouse button is pressed.
const EVENT_RBUTTONDOWN int = 2   //!< indicates that the right mouse button is pressed.
const EVENT_MBUTTONDOWN int = 3   //!< indicates that the middle mouse button is pressed.
const EVENT_LBUTTONUP int = 4     //!< indicates that left mouse button is released.
const EVENT_RBUTTONUP int = 5     //!< indicates that right mouse button is released.
const EVENT_MBUTTONUP int = 6     //!< indicates that middle mouse button is released.
const EVENT_LBUTTONDBLCLK int = 7 //!< indicates that left mouse button is double clicked.
const EVENT_RBUTTONDBLCLK int = 8 //!< indicates that right mouse button is double clicked.
const EVENT_MBUTTONDBLCLK int = 9 //!< indicates that middle mouse button is double clicked.
const EVENT_MOUSEWHEEL int = 10   //!< positive and negative values mean forward and backward scrolling, respectively.
const EVENT_MOUSEHWHEEL int = 11  //!< positive and negative values mean right and left scrolling, respectively.

//! Mouse Event Flags see cv::MouseCallback
const EVENT_FLAG_LBUTTON int = 1   //!< indicates that the left mouse button is down.
const EVENT_FLAG_RBUTTON int = 2   //!< indicates that the right mouse button is down.
const EVENT_FLAG_MBUTTON int = 4   //!< indicates that the middle mouse button is down.
const EVENT_FLAG_CTRLKEY int = 8   //!< indicates that CTRL Key is pressed.
const EVENT_FLAG_SHIFTKEY int = 16 //!< indicates that SHIFT Key is pressed.
const EVENT_FLAG_ALTKEY int = 32   //!< indicates that ALT Key is pressed.

// Internal things
const CVUI_ANTIALISED int = int(gocv.LineAA) // cv2.LINE_AA
const CVUI_FILLED int = -1

var __internal Internal

// Access points to internal global namespaces.
func init() {
	__internal = NewInternal()
}

type Window struct {
	*gocv.Window
	Name         string
	OnMouse      func(theEvent int, theX, theY int, theFlags int, theContext *Context)
	OnMouseParam interface{}
}

func (w *Window) SetMouseCallback(callbackFunc func(theEvent int, theX, theY int, theFlags int, theContext *Context), context Context) {
	// w.OnMouse = callbackFunc
	// w.OnMouseParam = context

	cName := C.CString(__internal.CurrentContext)
	defer C.free(unsafe.Pointer(cName))

	cFunc := C.MouseCallback(callbackFunc)
	C.Set_Mouse_Callback(cName, cFunc, 0)
}

// Represent a 2D point.
type Point struct {
	X int
	Y int
}

func NewPoint(theX int, theY int) Point {
	return Point{theX, theY}
}

func (p *Point) Inside(theRect Rect) bool {
	return theRect.Contains(*p)
}

// Represent a rectangle.
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

func NewRect(theX, theY, theWidth, theHeight int) Rect {
	return Rect{theX, theY, theWidth, theHeight}
}

func (r *Rect) Contains(thePoint Point) bool {
	return thePoint.X >= r.X && thePoint.X <= (r.X+r.Width) && thePoint.Y >= r.Y && thePoint.Y <= (r.Y+r.Height)
}

func (r *Rect) Area() int {
	return r.Width * r.Height
}

// Represent the size of something, i.e. width and height.
// It is essentially a simplified version of Rect where x and y are zero.
type Size struct {
	X      int
	Y      int
	Width  int
	Height int
}

func NewSize(theWidth, theHeight int) Size {
	return Size{0, 0, theWidth, theHeight}
}

// Describe a block structure used by gocvui to handle `begin*()` and `end*()` calls.
type Block struct {
	Where   gocv.Mat
	Rect    Rect
	Fill    Rect
	Anchor  Point
	Padding int
	Type    int
}

func NewBlock() Block {
	block := Block{
		Where:   gocv.Mat{},
		Rect:    Rect{},
		Fill:    Rect{},
		Anchor:  Point{},
		Padding: 0,
		Type:    0,
	}

	block.Reset()
	return block
}

func (b *Block) Reset() {
	b.Rect = NewRect(0, 0, 0, 0)
	b.Fill = NewRect(0, 0, 0, 0)
	b.Anchor = NewPoint(0, 0)
	b.Padding = 0
}

// Describe a component label, including info about a shortcut.
// If a label contains "Re&start", then:
// - hasShortcut will be true
// - shortcut will be "s"
// - textBeforeShortcut will be "Re"
// - textAfterShortcut will be "tart"
type Label struct {
	HasShortcut        bool
	Shortcut           byte
	TextBeforeShortcut string
	TextAfterShortcut  string
}

func NewLabel() Label {
	return Label{
		HasShortcut:        false,
		Shortcut:           0,
		TextBeforeShortcut: "",
		TextAfterShortcut:  "",
	}
}

// Describe a mouse button
type MouseButton struct {
	JustReleased bool
	JustPressed  bool
	Pressed      bool
}

func NewMouseButton() *MouseButton {
	return &MouseButton{
		JustReleased: false, // if the mouse button was released, i.e. click event.
		JustPressed:  false, // if the mouse button was just pressed, i.e. true for a frame when a button is down.
		Pressed:      false, // if the mouse button is pressed or not.
	}
}

func (b *MouseButton) Reset() {
	b.JustReleased = false
	b.JustPressed = false
	b.Pressed = false
}

// Describe the information of the mouse cursor
type Mouse struct {
	Buttons   map[int]*MouseButton
	AnyButton *MouseButton
	Position  Point
}

func NewMouse() Mouse {
	return Mouse{
		Buttons: map[int]*MouseButton{ // status of each button. Use gocvui.{RIGHT,LEFT,MIDDLE}_BUTTON to access the buttons.
			LEFT_BUTTON:   NewMouseButton(),
			MIDDLE_BUTTON: NewMouseButton(),
			RIGHT_BUTTON:  NewMouseButton(),
		},
		AnyButton: NewMouseButton(), // represent the behavior of all mouse buttons combined
		Position:  NewPoint(0, 0),   // x and y coordinates of the mouse at the moment.
	}
}

// Describe a (window) context.
type Context struct {
	Window *gocv.Window
	Mouse  Mouse
}

func NewContext() Context {
	return Context{
		Window: nil,
		Mouse:  NewMouse(),
	}
}

// Describe the inner parts of the trackbar component.
type TrackbarParams struct {
	Min         float64
	Max         float64
	Step        float64
	Segments    int
	Options     int
	LabelFormat string
}

func NewTrackbarParams(theMin, theMax, theStep float64, theSegments int, theLabelFormat string, theOptions int) TrackbarParams {
	if theMax <= 0 {
		theMax = 25.
	}
	if theLabelFormat == "" {
		theLabelFormat = "%.0Lf"
	}

	return TrackbarParams{
		Min:         theMin,
		Max:         theMax,
		Step:        theStep,
		Segments:    theSegments,
		Options:     theOptions,
		LabelFormat: theLabelFormat,
	}
}

// This class contains all stuff that gocvui uses internally to render
// and control interaction with components.
type Internal struct {
	DefaultContext  string
	CurrentContext  string
	Contexts        map[string]Context // indexed by the window name.
	Buffer          []interface{}
	LastKeyPressed  int // TODO: collect it per window
	DelayWaitKey    int
	Screen          Block
	Stack           []Block // TODO: make it dynamic
	StackCount      int
	TrackbarMarginX int
	render          Render
}

func NewInternal() Internal {
	internal := Internal{
		DefaultContext:  "",
		CurrentContext:  "",
		Contexts:        map[string]Context{}, // indexed by the window name.
		Buffer:          nil,
		LastKeyPressed:  -1, // TODO: collect it per window
		DelayWaitKey:    -1,
		Screen:          NewBlock(),
		Stack:           nil, // TODO: make it dynamic
		StackCount:      -1,
		TrackbarMarginX: 14,
		render:          Render{},
	}

	stack := []Block{}
	for i := 0; i < 100; i++ {
		stack = append(stack, NewBlock())
	}
	internal.Stack = stack

	render := Render{
		internal: &internal,
	}

	internal.render = render
	return internal
}

func (in *Internal) IsMouseButton(theButton *MouseButton, theQuery int) bool {
	aRet := false

	if theQuery == CLICK || theQuery == UP {
		aRet = theButton.JustReleased
	} else if theQuery == DOWN {
		aRet = theButton.JustPressed
	} else if theQuery == IS_DOWN {
		aRet = theButton.Pressed
	}

	return aRet
}

// Return the last position of the mouse.
// param theWindowName name of the window whose mouse cursor will be used. If nothing is informed (default), the function will return the position of the mouse cursor for the default window (the one informed in `gocvui::Init()`).
// return a point containing the position of the mouse cursor in the speficied window.
func (in *Internal) MouseW(theWindowName string) Point {
	return in.GetContext(theWindowName).Mouse.Position
}

// Query the mouse for events, e.g. "is any button down now?". Available queries are:

// * `gocvui::DOWN`: any mouse button was pressed. `gocvui::mouse()` returns `true` for a single frame only.
// * `gocvui::UP`: any mouse button was released.  `gocvui::mouse()` returns `true` for a single frame only.
// * `gocvui::CLICK`: any mouse button was clicked (went down then up, no matter the amount of frames in between). `gocvui::mouse()` returns `true` for a single frame only.
// * `gocvui::IS_DOWN`: any mouse button is currently pressed. `gocvui::mouse()` returns `true` for as long as the button is down/pressed.

// It is easier to think of this function as the answer to a questions. For instance, asking if any mouse button went down:

// ```
// if (gocvui::Mouse(gocvui::DOWN)) {
// 	// Any mouse button just went down.
// }
// ```

// The window whose mouse will be queried depends on the context. If `gocvui::Mouse(query)` is being called after
// `gocvui::Context()`, the window informed in the context will be queried. If no context is available, the default
// window (informed in `gocvui::Init()`) will be used.

// Parameters
// ----------
// theQuery: int
// 	Integer describing the intended mouse query. Available queries are `gocvui::DOWN`, `gocvui::UP`, `gocvui::CLICK`, and `gocvui::IS_DOWN`.

// Mouse(const cv::String&)
// Mouse(const cv::String&, int)
// Mouse(const cv::String&, int, int)
// Mouse(int, int)
func (in *Internal) MouseQ(theQuery int) bool {
	return in.MouseWQ("", theQuery)
}

// Query the mouse for events in a particular window. This function behave exactly like `gocvui::mouse(int theQuery)`
// with the difference that queries are targeted at a particular window.

// \param theWindowName name of the window that will be queried.
// \param theQuery an integer describing the intended mouse query. Available queries are `gocvui::DOWN`, `gocvui::UP`, `gocvui::CLICK`, and `gocvui::IS_DOWN`.

// Mouse(const cv::String&)
// Mouse(const cv::String&, int, int)
// Mouse(int, int)
// Mouse(int)
func (in *Internal) MouseWQ(theWindowName string, theQuery int) bool {
	aButton := in.GetContext(theWindowName).Mouse.AnyButton
	aRet := in.IsMouseButton(aButton, theQuery)
	return aRet
}

// Query the mouse for events in a particular button in a particular window. This function behave exactly
// like `gocvui::Mouse(int theButton, int theQuery)`, with the difference that queries are targeted at
// a particular mouse button in a particular window instead.

// \param theWindowName name of the window that will be queried.
// \param theButton an integer describing the mouse button to be queried. Possible values are `gocvui::LEFT_BUTTON`, `gocvui::MIDDLE_BUTTON` and `gocvui::LEFT_BUTTON`.
// \param theQuery an integer describing the intended mouse query. Available queries are `gocvui::DOWN`, `gocvui::UP`, `gocvui::CLICK`, and `gocvui::IS_DOWN`.
func (in *Internal) MouseWBQ(theWindowName string, theButton int, theQuery int) bool {
	if theButton != RIGHT_BUTTON && theButton != MIDDLE_BUTTON && theButton != LEFT_BUTTON {
		__internal.Error(6, "Invalid mouse button. Are you using one of the available: cvui.{RIGHT,MIDDLE,LEFT}_BUTTON ?")
	}

	aButton := in.GetContext(theWindowName).Mouse.Buttons[theButton]
	aRet := in.IsMouseButton(aButton, theQuery)

	return aRet
}

func (in *Internal) Init(theWindowName string, theDelayWaitKey int) {
	in.DefaultContext = theWindowName
	in.CurrentContext = theWindowName
	in.DelayWaitKey = theDelayWaitKey
	in.LastKeyPressed = -1
}

func (in *Internal) BitsetHas(theBitset int, theValue int) bool {
	return (theBitset & theValue) != 0
}

func (in *Internal) Error(theId int, theMessage string) {
	fmt.Println("[CVUI] Fatal error (code ", theId, "): ", theMessage)
	gocv.WaitKey(100000)
	os.Exit(-1)
}

func (in *Internal) GetContext(theWindowName string) Context {
	if len(theWindowName) != 0 {
		// Get context in particular
		return in.Contexts[theWindowName]
	} else if len(in.CurrentContext) != 0 {
		// No window provided, return currently active context.
		return in.Contexts[in.CurrentContext]
	} else if len(in.DefaultContext) != 0 {
		// We have no active context, so let"s use the default one.
		return in.Contexts[in.DefaultContext]
	} else {
		// Apparently we have no window at all! <o>
		// This should not happen. Probably gocvui::Init() was never called.
		errMsg := "Unable to read context. Did you forget to call cvui.init()?"
		in.Error(5, errMsg)
		return Context{}
	}
}

func (in *Internal) UpdateLayoutFlow(theBlock *Block, theSize Size) {
	Max := func(a, b int) int {
		if a > b {
			return a
		} else {
			return b
		}
	}

	if theBlock.Type == ROW {
		aValue := theSize.Width + theBlock.Padding

		theBlock.Anchor.X += int(aValue)
		theBlock.Fill.Width += aValue
		theBlock.Fill.Height = Max(theSize.Height, theBlock.Fill.Height)

	} else if theBlock.Type == COLUMN {
		aValue := theSize.Height + theBlock.Padding

		theBlock.Anchor.Y += int(aValue)
		theBlock.Fill.Height += aValue
		theBlock.Fill.Width = Max(theSize.Width, theBlock.Fill.Width)
	}
}

func (in *Internal) BlockStackEmpty() bool {
	return in.StackCount == -1
}

func (in *Internal) TopBlock() *Block {
	if in.StackCount < 0 {
		in.Error(3, "You are using a function that should be enclosed by begin*() and end*(), but you probably forgot to call begin*().")
	}

	return &in.Stack[in.StackCount]
}

func (in *Internal) PushBlock() Block {
	in.StackCount += 1
	return in.Stack[in.StackCount]
}

func (in *Internal) PopBlock() Block {
	// Check if there is anything to be popped out from the stack.
	if in.StackCount < 0 {
		in.Error(1, "Mismatch in the number of begin*()/end*() calls. You are calling one more than the other.")
	}

	aIndex := in.StackCount
	in.StackCount -= 1

	return in.Stack[aIndex]
}

func (in *Internal) CreateLabel(theLabel string) Label {
	i := 0
	aBefore := []byte{}
	aAfter := []byte{}

	aLabel := NewLabel()
	aLabel.HasShortcut = false
	aLabel.Shortcut = 0
	aLabel.TextBeforeShortcut = ""
	aLabel.TextAfterShortcut = ""

	for i < len(theLabel) {
		c := theLabel[i]
		if c == '&' && i < len(theLabel)-1 {
			aLabel.HasShortcut = true
			aLabel.Shortcut = theLabel[i+1]
			i += 1
		} else if !aLabel.HasShortcut {
			aBefore = append(aBefore, c)
		} else {
			aAfter = append(aAfter, c)
		}
		i += 1
	}

	aLabel.TextBeforeShortcut = string(aBefore)
	aLabel.TextAfterShortcut = string(aAfter)

	return aLabel
}

func (in *Internal) Text(theBlock *Block, theX, theY int, theText string, theFontScale float64, theColor uint32, theUpdateLayout bool) {

	aSizeInfo := gocv.GetTextSize(theText, gocv.FontHersheySimplex, theFontScale, 1)

	aTextSize := NewSize(aSizeInfo.X, aSizeInfo.Y)
	aPos := NewPoint(theX, theY+int(aTextSize.Height))

	in.render.Text(theBlock, theText, aPos, theFontScale, theColor)

	if theUpdateLayout {
		// Add an extra pixel to the height to overcome OpenCV font size problems.
		aTextSize.Height += 1
		in.UpdateLayoutFlow(theBlock, aTextSize)
	}
}

func (in *Internal) Counter(theBlock *Block, theX, theY int, theValue []int, theStep int, theFormat string) int {
	aContentArea := NewRect(theX+22, theY, 48, 22)

	if in.ButtonWH(theBlock, theX, theY, 22, 22, "-", false) {
		theValue[0] -= theStep
	}

	aText := fmt.Sprintf(theFormat, theValue[0])
	in.render.Counter(theBlock, aContentArea, aText)

	if in.ButtonWH(theBlock, aContentArea.X+aContentArea.Width, theY, 22, 22, "+", false) {
		theValue[0] += theStep
	}

	// Update the layout flow
	aSize := NewSize(22*2+aContentArea.Width, aContentArea.Height)
	in.UpdateLayoutFlow(theBlock, aSize)

	return theValue[0]
}

func (in *Internal) Checkbox(theBlock *Block, theX, theY int, theLabel string, theState []bool, theColor uint32) bool {
	aMouse := in.GetContext("").Mouse
	aRect := NewRect(theX, theY, 15, 15)
	aSizeInfo := gocv.GetTextSize(theLabel, gocv.FontHersheySimplex, 0.4, 1)
	aTextSize := NewRect(0, 0, aSizeInfo.X, aSizeInfo.Y)
	aHitArea := NewRect(theX, theY, aRect.Width+aTextSize.Width+6, aRect.Height)
	aMouseIsOver := aHitArea.Contains(aMouse.Position)

	if aMouseIsOver {
		in.render.Checkbox(theBlock, OVER, aRect)

		if aMouse.AnyButton.JustReleased {
			theState[0] = !theState[0]
		}
	} else {
		in.render.Checkbox(theBlock, OUT, aRect)
	}

	in.render.CheckboxLabel(theBlock, aRect, theLabel, aTextSize, theColor)

	if theState[0] {
		in.render.CheckboxCheck(theBlock, aRect)
	}

	// Update the layout flow
	aSize := NewSize(aHitArea.Width, aHitArea.Height)
	in.UpdateLayoutFlow(theBlock, aSize)

	return theState[0]
}

func (in *Internal) Clamp01(theValue float64) float64 {
	if theValue > 1. {
		theValue = 1.
	}

	if theValue < 0. {
		theValue = 0.
	}

	return theValue
}

func (in *Internal) TrackbarForceValuesAsMultiplesOfSmallStep(theParams TrackbarParams, theValue []float64) {
	if !in.BitsetHas(theParams.Options, TRACKBAR_DISCRETE) || theParams.Step == 0. {
		return
	}

	k := float64(theValue[0]-theParams.Min) / theParams.Step
	k = math.Round(k)
	theValue[0] = theParams.Min + theParams.Step*k
}

func (in *Internal) TrackbarXPixelToValue(theParams TrackbarParams, theBounding Rect, thePixelX int) float64 {
	aRatio := float64(thePixelX) - float64(theBounding.X+in.TrackbarMarginX)/float64(theBounding.Width-2.*in.TrackbarMarginX)
	aRatio = in.Clamp01(aRatio)
	aValue := theParams.Min + aRatio*(theParams.Max-theParams.Min)
	return aValue
}

func (in *Internal) TrackbarValueToXPixel(theParams TrackbarParams, theBounding Rect, theValue float64) int {
	aRatio := float64(theValue-theParams.Min) / (theParams.Max - theParams.Min)
	aRatio = in.Clamp01(aRatio)
	aPixelsX := float64(theBounding.X+in.TrackbarMarginX) + aRatio*float64(theBounding.Width-2*in.TrackbarMarginX)
	return int(aPixelsX)
}

func (in *Internal) IArea(theX, theY, theWidth, theHeight int) int {
	aMouse := in.GetContext("").Mouse

	// By default, return that the mouse is out of the interaction area.
	aRet := OUT

	// Check if the mouse is over the interaction area.
	interArea := NewRect(theX, theY, theWidth, theHeight)
	aMouseIsOver := interArea.Contains(aMouse.Position)

	if aMouseIsOver {
		if aMouse.AnyButton.Pressed {
			aRet = DOWN
		} else {
			aRet = OVER
		}
	}

	// Tell if the button was clicked or not
	if aMouseIsOver && aMouse.AnyButton.JustReleased {
		aRet = CLICK
	}

	return aRet
}

func (in *Internal) ButtonWH(theBlock *Block, theX, theY, theWidth, theHeight int, theLabel string, theUpdateLayout bool) bool {
	// Calculate the space that the label will fill
	aSizeInfo := gocv.GetTextSize(theLabel, gocv.FontHersheySimplex, 0.4, 1)
	aTextSize := NewRect(0, 0, aSizeInfo.X, aSizeInfo.Y)

	// Make the button big enough to house the label
	aRect := NewRect(theX, theY, theWidth, theHeight)

	// Render the button according to mouse interaction, e.g. OVER, DOWN, OUT.
	aStatus := in.IArea(theX, theY, aRect.Width, aRect.Height)
	in.render.Button(theBlock, aStatus, aRect, theLabel)
	in.render.ButtonLabel(theBlock, aStatus, aRect, theLabel, aTextSize)

	// Update the layout flow according to button size
	// if we were told to update.
	if theUpdateLayout {
		aSize := NewSize(theWidth, theHeight)
		in.UpdateLayoutFlow(theBlock, aSize)
	}

	aWasShortcutPressed := false

	// Handle keyboard shortcuts
	if in.LastKeyPressed != -1 {
		aLabel := in.CreateLabel(theLabel)

		if aLabel.HasShortcut && strings.EqualFold(strings.ToLower(string(aLabel.Shortcut)), strings.ToLower(string(byte(in.LastKeyPressed)))) {
			aWasShortcutPressed = true
		}
	}

	// Return true if the button was clicked
	return aStatus == CLICK || aWasShortcutPressed
}

func (in *Internal) Button(theBlock *Block, theX, theY int, theLabel string) bool {
	// Calculate the space that the label will fill
	aSizeInfo := gocv.GetTextSize(theLabel, gocv.FontHersheySimplex, 0.4, 1)
	aTextSize := NewRect(0, 0, aSizeInfo.X, aSizeInfo.Y)

	// Create a button based on the size of the text
	return in.ButtonWH(theBlock, theX, theY, aTextSize.Width+30, aTextSize.Height+18, theLabel, true)
}

func (in *Internal) ButtonI(theBlock *Block, theX, theY int, theIdle, theOver, theDown *gocv.Mat, theUpdateLayout bool) bool {
	aIdleRows := theIdle.Rows()
	aIdleCols := theIdle.Cols()

	aRect := NewRect(theX, theY, aIdleCols, aIdleRows)
	aStatus := in.IArea(theX, theY, aRect.Width, aRect.Height)

	if aStatus == OUT {
		in.render.Image(theBlock, aRect, theIdle)
	} else if aStatus == OVER {
		in.render.Image(theBlock, aRect, theOver)
	} else if aStatus == DOWN {
		in.render.Image(theBlock, aRect, theDown)
	}

	// Update the layout flow according to button size
	// if we were told to update.
	if theUpdateLayout {
		aSize := NewSize(aRect.Width, aRect.Height)
		in.UpdateLayoutFlow(theBlock, aSize)
	}

	// Return true if the button was clicked
	return aStatus == CLICK
}

func (in *Internal) Image(theBlock *Block, theX, theY int, theImage *gocv.Mat) {
	aImageRows := theImage.Rows()
	aImageCols := theImage.Cols()

	aRect := NewRect(theX, theY, aImageCols, aImageRows)

	// TODO: check for render outside the frame area
	in.render.Image(theBlock, aRect, theImage)

	// Update the layout flow according to image size
	aSize := NewSize(aImageCols, aImageRows)
	in.UpdateLayoutFlow(theBlock, aSize)
}

// 	def trackbar(self, theBlock, theX, theY, theWidth, theValue, theParams):
// 		aMouse = in.getContext().mouse
// 		aContentArea = Rect(theX, theY, theWidth, 45)
// 		aMouseIsOver = aContentArea.contains(aMouse.position)
// 		aValue = theValue[0]

// 		in.render.trackbar(theBlock, OVER if aMouseIsOver else OUT, aContentArea, theValue[0], theParams)

// 		if aMouse.anyButton.pressed and aMouseIsOver:
// 			theValue[0] = in.trackbarXPixelToValue(theParams, aContentArea, aMouse.position.x)

// 			if in.bitsetHas(theParams.options, TRACKBAR_DISCRETE):
// 				in.trackbarForceValuesAsMultiplesOfSmallStep(theParams, theValue)

// 		// Update the layout flow
// 		// TODO: use aSize = aContentArea.size()?
// 		in.updateLayoutFlow(theBlock, aContentArea)

// 		return theValue[0] != aValue

// 	def window(self, theBlock, theX, theY, theWidth, theHeight, theTitle):
// 		aTitleBar = Rect(theX, theY, theWidth, 20)
// 		aContent = Rect(theX, theY + aTitleBar.height, theWidth, theHeight - aTitleBar.height)

// 		in.render.window(theBlock, aTitleBar, aContent, theTitle)

// 		// Update the layout flow
// 		aSize = Size(theWidth, theHeight)
// 		in.updateLayoutFlow(theBlock, aSize)

// 	def rect(self, theBlock, theX, theY, theWidth, theHeight, theBorderColor, theFillingColor):
// 		aAnchor = Point(theX, theY);
// 		aRect = Rect(theX, theY, theWidth, theHeight);

// 		aRect.x = aAnchor.x + aRect.width if aRect.width < 0 else aAnchor.x
// 		aRect.y = aAnchor.y + aRect.height if aRect.height < 0 else aAnchor.y
// 		aRect.width = abs(aRect.width)
// 		aRect.height = abs(aRect.height)

// 		in.render.rect(theBlock, aRect, theBorderColor, theFillingColor)

// 		// Update the layout flow
// 		aSize = Size(aRect.width, aRect.height)
// 		in.updateLayoutFlow(theBlock, aSize)

func (in *Internal) Sparkline(theBlock *Block, theValues []float64, theX, theY, theWidth, theHeight int, theColor uint32) {
	aRect := NewRect(theX, theY, theWidth, theHeight)
	aHowManyValues := len(theValues)

	if aHowManyValues >= 2 {
		aMin, aMax := in.FindMinMax(theValues)
		in.render.Sparkline(theBlock, theValues, aRect, aMin, aMax, theColor)
	} else {
		msg := "No data."
		if aHowManyValues != 0 {
			msg = "Insufficient data points."
		}

		in.Text(theBlock, theX, theY, msg, 0.4, 0xCECECE, false)
	}

	// Update the layout flow
	aSize := NewSize(theWidth, theHeight)
	in.UpdateLayoutFlow(theBlock, aSize)
}

func (in *Internal) HexToScalar(theColor uint32) color.RGBA {
	aAlpha := uint8((theColor >> 24) & 0xff)
	aRed := uint8((theColor >> 16) & 0xff)
	aGreen := uint8((theColor >> 8) & 0xff)
	aBlue := uint8(theColor & 0xff)

	return color.RGBA{R: aRed, G: aGreen, B: aBlue, A: aAlpha}
}

func (in *Internal) IsString(theObj interface{}) bool {
	_, ok := theObj.(string)
	return ok
}

func (in *Internal) Begin(theType int, theWhere *gocv.Mat, theX, theY, theWidth, theHeight int, thePadding int) {
	aBlock := in.PushBlock()

	aBlock.Where = *theWhere

	aBlock.Rect.X = theX
	aBlock.Rect.Y = theY
	aBlock.Rect.Width = theWidth
	aBlock.Rect.Height = theHeight

	aBlock.Fill = aBlock.Rect
	aBlock.Fill.Width = 0
	aBlock.Fill.Height = 0

	aBlock.Anchor.X = theX
	aBlock.Anchor.Y = theY

	aBlock.Padding = thePadding
	aBlock.Type = theType
}

func (in *Internal) End(theType int) {
	aBlock := in.PopBlock()

	if aBlock.Type != theType {
		in.Error(4, "Calling wrong type of end*(). E.g. endColumn() instead of endRow(). Check if your begin*() calls are matched with their appropriate end*() calls.")
	}

	// If we still have blocks in the stack, we must update
	// the current top with the dimensions that were filled by
	// the newly popped block.

	if !in.BlockStackEmpty() {
		aTop := in.TopBlock()
		aSize := NewSize(aBlock.Rect.Width, aBlock.Rect.Height)

		// If the block has rect.width < 0 or rect.heigth < 0, it means the
		// user don"t want to calculate the block"s width/height. It"s up to
		// us do to the math. In that case, we use the block"s fill rect to find
		// out the occupied space. If the block"s width/height is greater than
		// zero, then the user is very specific about the desired size. In that
		// case, we use the provided width/height, no matter what the fill rect
		// actually is.

		if aBlock.Rect.Width < 0 {
			aSize.Width = aBlock.Fill.Width
		}

		if aBlock.Rect.Height < 0 {
			aSize.Height = aBlock.Fill.Height
		}

		in.UpdateLayoutFlow(aTop, aSize)
	}
}

// Find the min and max values of a vector
func (in *Internal) FindMinMax(theValues []float64) (float64, float64) {
	aMin := theValues[0]
	aMax := theValues[0]

	for _, aValue := range theValues {
		if aValue < aMin {
			aMin = aValue
		}

		if aValue > aMax {
			aMax = aValue
		}
	}

	return aMin, aMax
}

// Class that contains all rendering methods.
type Render struct {
	internal *Internal
}

func (r *Render) Rectangle(theWhere *gocv.Mat, theShape Rect, theColor color.RGBA, theThickness int) {
	aStartPoint := NewPoint(theShape.X, theShape.Y)
	aEndPoint := NewPoint(theShape.X+theShape.Width, theShape.Y+theShape.Height)
	rectangle := image.Rect(int(aStartPoint.X), int(aStartPoint.Y), int(aEndPoint.X), int(aEndPoint.Y))
	gocv.Rectangle(theWhere, rectangle, theColor, theThickness)
}

func (r *Render) Text(theBlock *Block, theText string, thePos Point, theFontScale float64, theColor uint32) {
	aPosition := image.Point{int(thePos.X), int(thePos.Y)}
	gocv.PutText(&theBlock.Where, theText, aPosition, gocv.FontHersheySimplex, theFontScale, r.internal.HexToScalar(theColor), 1)
}

func (r *Render) Counter(theBlock *Block, theShape Rect, theValue string) {
	r.Rectangle(&theBlock.Where, theShape, color.RGBA{0x29, 0x29, 0x29, 0xff}, CVUI_FILLED) // fill
	r.Rectangle(&theBlock.Where, theShape, color.RGBA{0x45, 0x45, 0x45, 0xff}, 1)           // border
	aSizeInfo := gocv.GetTextSize(theValue, gocv.FontHersheySimplex, 0.4, 1)
	aTextSize := NewRect(0, 0, aSizeInfo.X, aSizeInfo.Y)

	aPos := NewPoint(theShape.X+theShape.Width/2-aTextSize.Width/2, theShape.Y+aTextSize.Height/2+theShape.Height/2)
	gocv.PutText(&theBlock.Where, theValue, image.Point{int(aPos.X), int(aPos.Y)}, gocv.FontHersheySimplex, 0.4, color.RGBA{0xCE, 0xCE, 0xCE, 0xff}, 1)
}

func (r *Render) Button(theBlock *Block, theState int, theShape Rect, theLabel string) {
	// Outline
	r.Rectangle(&theBlock.Where, theShape, color.RGBA{0x29, 0x29, 0x29, 0xff}, 1)

	// Border
	theShape.X += 1
	theShape.Y += 1
	theShape.Width -= 2
	theShape.Height -= 2
	r.Rectangle(&theBlock.Where, theShape, color.RGBA{0x4A, 0x4A, 0x4A, 0xff}, 1)

	// Inside
	theShape.X += 1
	theShape.Y += 1
	theShape.Width -= 2
	theShape.Height -= 2

	colors := color.RGBA{0x32, 0x32, 0x32, 0xff}
	if theState == OUT {
		colors = color.RGBA{0x42, 0x42, 0x42, 0xff}
	} else if theState == OVER {
		colors = color.RGBA{0x52, 0x52, 0x52, 0xff}
	}

	r.Rectangle(&theBlock.Where, theShape, colors, CVUI_FILLED)
}

func (r *Render) Image(theBlock *Block, theRect Rect, theImage *gocv.Mat) {
	theBlock.Where = theImage.Region(image.Rectangle{image.Point{int(theRect.X), int(theRect.Y)}, image.Point{int(theRect.X + theRect.Width), int(theRect.Y + theRect.Height)}})
}

func (r *Render) PutText(theBlock *Block, theState int, theColor color.RGBA, theText string, thePosition Point) int {
	var aFontScale float64
	if theState == DOWN {
		aFontScale = 0.39
	} else {
		aFontScale = 0.4
	}

	var aTextSize Rect

	if theText != "" {
		aPosition := image.Pt(int(thePosition.X), int(thePosition.Y))
		gocv.PutText(&theBlock.Where, theText, aPosition, gocv.FontHersheySimplex, aFontScale, theColor, 1)

		aSizeInfo := gocv.GetTextSize(theText, gocv.FontHersheySimplex, aFontScale, 1)
		aTextSize = NewRect(0, 0, aSizeInfo.X, aSizeInfo.Y)
	}
	return aTextSize.Width
}

// 	def putTextCentered(self, theBlock, thePosition, theText):
// 		aFontScale = 0.3

// 		aSizeInfo, aBaseline = cv2.getTextSize(theText, cv2.FONT_HERSHEY_SIMPLEX, aFontScale, 1)
// 		aTextSize = Rect(0, 0, aSizeInfo[0], aSizeInfo[1])
// 		aPositionDecentered = Point(thePosition.x - aTextSize.width / 2, thePosition.y)
// 		cv2.putText(theBlock.where, theText, (int(aPositionDecentered.x), int(aPositionDecentered.y)), cv2.FONT_HERSHEY_SIMPLEX, aFontScale, (0xCE, 0xCE, 0xCE), 1, CVUI_ANTIALISED)

// 		return aTextSize.width

func (r *Render) ButtonLabel(theBlock *Block, theState int, theRect Rect, theLabel string, theTextSize Rect) {
	aPos := NewPoint(theRect.X+theRect.Width/2-theTextSize.Width/2, theRect.Y+theRect.Height/2+theTextSize.Height/2)
	aColor := color.RGBA{0xCE, 0xCE, 0xCE, 0xff}

	aLabel := r.internal.CreateLabel(theLabel)

	if !aLabel.HasShortcut {
		r.PutText(theBlock, theState, aColor, theLabel, aPos)
	} else {
		aWidth := r.PutText(theBlock, theState, aColor, aLabel.TextBeforeShortcut, aPos)
		aStart := aPos.X + aWidth
		aPos.X += aWidth

		aShortcut := ""
		aShortcut += string(aLabel.Shortcut)

		aWidth = r.PutText(theBlock, theState, aColor, aShortcut, aPos)
		aEnd := aStart + aWidth
		aPos.X += aWidth

		r.PutText(theBlock, theState, aColor, aLabel.TextAfterShortcut, aPos)
		gocv.Line(&theBlock.Where, image.Point{int(aStart), int(aPos.Y + 3)}, image.Point{int(aEnd), int(aPos.Y + 3)}, aColor, 1)
	}
}

// 	def trackbarHandle(self, theBlock, theState, theShape, theValue, theParams, theWorkingArea):
// 		aBarTopLeft = Point(theWorkingArea.x, theWorkingArea.y + theWorkingArea.height / 2)
// 		aBarHeight = 7

// 		// Draw the rectangle representing the handle
// 		aPixelX = r.internal.trackbarValueToXPixel(theParams, theShape, theValue)
// 		aIndicatorWidth = 3
// 		aIndicatorHeight = 4
// 		aPoint1 = Point(aPixelX - aIndicatorWidth, aBarTopLeft.y - aIndicatorHeight)
// 		aPoint2 = Point(aPixelX + aIndicatorWidth, aBarTopLeft.y + aBarHeight + aIndicatorHeight)
// 		aRect = Rect(aPoint1.x, aPoint1.y, aPoint2.x - aPoint1.x, aPoint2.y - aPoint1.y)

// 		aFillColor = 0x525252 if theState == OVER else 0x424242

// 		in.rect(theBlock, aRect, 0x212121, 0x212121)
// 		aRect.x += 1
// 		aRect.y += 1
// 		aRect.width -= 2
// 		aRect.height -= 2
// 		in.rect(theBlock, aRect, 0x515151, aFillColor)

// 		aShowLabel = r.internal.bitsetHas(theParams.options, TRACKBAR_HIDE_VALUE_LABEL) == false

// 		// Draw the handle label
// 		if aShowLabel:
// 			aTextPos = Point(aPixelX, aPoint2.y + 11)
// 			aText = theParams.labelFormat % theValue
// 			in.putTextCentered(theBlock, aTextPos, aText)

// 	def trackbarPath(self, theBlock, theState, theShape, theValue, theParams, theWorkingArea):
// 		aBarHeight = 7
// 		aBarTopLeft = Point(theWorkingArea.x, theWorkingArea.y + theWorkingArea.height / 2)
// 		aRect = Rect(aBarTopLeft.x, aBarTopLeft.y, theWorkingArea.width, aBarHeight)

// 		aBorderColor = 0x4e4e4e if theState == OVER else 0x3e3e3e

// 		in.rect(theBlock, aRect, aBorderColor, 0x292929)
// 		cv2.line(theBlock.where, (int(aRect.x + 1), int(aRect.y + aBarHeight - 2)), (int(aRect.x + aRect.width - 2), int(aRect.y + aBarHeight - 2)), (0x0e, 0x0e, 0x0e))

// 	def trackbarSteps(self, theBlock, theState, theShape, theValue, theParams, theWorkingArea):
// 		aBarTopLeft = Point(theWorkingArea.x, theWorkingArea.y + theWorkingArea.height / 2)
// 		aColor = (0x51, 0x51, 0x51)

// 		aDiscrete = r.internal.bitsetHas(theParams.options, TRACKBAR_DISCRETE)
// 		aFixedStep = theParams.step if aDiscrete else (theParams.max - theParams.min) / 20

// 		// TODO: check min, max and step to prevent infinite loop.
// 		aValue = theParams.min
// 		while aValue <= theParams.max:
// 			aPixelX = int(r.internal.trackbarValueToXPixel(theParams, theShape, aValue))
// 			aPoint1 = (aPixelX, int(aBarTopLeft.y))
// 			aPoint2 = (aPixelX, int(aBarTopLeft.y - 3))
// 			cv2.line(theBlock.where, aPoint1, aPoint2, aColor)
// 			aValue += aFixedStep

// 	def trackbarSegmentLabel(self, theBlock, theShape, theParams, theValue, theWorkingArea, theShowLabel):
// 		aColor = (0x51, 0x51, 0x51)
// 		aBarTopLeft = Point(theWorkingArea.x, theWorkingArea.y + theWorkingArea.height / 2)

// 		aPixelX = int(r.internal.trackbarValueToXPixel(theParams, theShape, theValue))

// 		aPoint1 = (aPixelX, int(aBarTopLeft.y))
// 		aPoint2 = (aPixelX, int(aBarTopLeft.y - 8))
// 		cv2.line(theBlock.where, aPoint1, aPoint2, aColor)

// 		if theShowLabel:
// 			aText = theParams.labelFormat % theValue
// 			aTextPos = Point(aPixelX, aBarTopLeft.y - 11)
// 			in.putTextCentered(theBlock, aTextPos, aText)

// 	def trackbarSegments(self, theBlock, theState, theShape, theValue, theParams, theWorkingArea):
// 		aSegments = 1 if theParams.segments < 1 else theParams.segments
// 		aSegmentLength = float(theParams.max - theParams.min) / aSegments

// 		aHasMinMaxLabels = r.internal.bitsetHas(theParams.options, TRACKBAR_HIDE_MIN_MAX_LABELS) == false

// 		// Render the min value label
// 		in.trackbarSegmentLabel(theBlock, theShape, theParams, theParams.min, theWorkingArea, aHasMinMaxLabels)

// 		// Draw large steps and labels
// 		aHasSegmentLabels = r.internal.bitsetHas(theParams.options, TRACKBAR_HIDE_SEGMENT_LABELS) == false
// 		// TODO: check min, max and step to prevent infinite loop.
// 		aValue = theParams.min
// 		while aValue <= theParams.max:
// 			in.trackbarSegmentLabel(theBlock, theShape, theParams, aValue, theWorkingArea, aHasSegmentLabels)
// 			aValue += aSegmentLength

// 		// Render the max value label
// 		in.trackbarSegmentLabel(theBlock, theShape, theParams, theParams.max, theWorkingArea, aHasMinMaxLabels)

// 	def trackbar(self, theBlock, theState, theShape, theValue, theParams):
// 		aWorkingArea = Rect(theShape.x + r.internal.trackbarMarginX, theShape.y, theShape.width - 2 * r.internal.trackbarMarginX, theShape.height)

// 		in.trackbarPath(theBlock, theState, theShape, theValue, theParams, aWorkingArea)

// 		aHideAllLabels = r.internal.bitsetHas(theParams.options, TRACKBAR_HIDE_LABELS)
// 		aShowSteps = r.internal.bitsetHas(theParams.options, TRACKBAR_HIDE_STEP_SCALE) == false

// 		if aShowSteps and aHideAllLabels == false:
// 			in.trackbarSteps(theBlock, theState, theShape, theValue, theParams, aWorkingArea)

// 		if aHideAllLabels == false:
// 			in.trackbarSegments(theBlock, theState, theShape, theValue, theParams, aWorkingArea)

// 		in.trackbarHandle(theBlock, theState, theShape, theValue, theParams, aWorkingArea)

func (r *Render) Checkbox(theBlock *Block, theState int, theShape Rect) {
	// Outline
	var colors color.RGBA
	if theState == OUT {
		colors = color.RGBA{0x63, 0x63, 0x63, 0xff}
	} else {
		colors = color.RGBA{0x80, 0x80, 0x80, 0xff}
	}

	r.Rectangle(&theBlock.Where, theShape, colors, 1)

	// Border
	theShape.X += 1
	theShape.Y += 1
	theShape.Width -= 2
	theShape.Height -= 2
	r.Rectangle(&theBlock.Where, theShape, color.RGBA{0x17, 0x17, 0x17, 0xff}, 1)

	// Inside
	theShape.X += 1
	theShape.Y += 1
	theShape.Width -= 2
	theShape.Height -= 2
	r.Rectangle(&theBlock.Where, theShape, color.RGBA{0x29, 0x29, 0x29, 0xff}, CVUI_FILLED)
}

// 	def checkbox(self, theBlock, theState, theShape):
// 		// Outline
// 		in.rectangle(theBlock.where, theShape, (0x63, 0x63, 0x63) if theState == OUT else (0x80, 0x80, 0x80))

// 		// Border
// 		theShape.x += 1
// 		theShape.y+=1
// 		theShape.width -= 2
// 		theShape.height -= 2
// 		in.rectangle(theBlock.where, theShape, (0x17, 0x17, 0x17))

// 		// Inside
// 		theShape.x += 1
// 		theShape.y += 1
// 		theShape.width -= 2
// 		theShape.height -= 2
// 		in.rectangle(theBlock.where, theShape, (0x29, 0x29, 0x29), CVUI_FILLED)
func (r *Render) CheckboxLabel(theBlock *Block, theRect Rect, theLabel string, theTextSize Rect, theColor uint32) {
	aPos := NewPoint(theRect.X+theRect.Width+6, theRect.Y+theTextSize.Height+theRect.Height/2-theTextSize.Height/2-1)
	r.Text(theBlock, theLabel, aPos, 0.4, theColor)
}

func (r *Render) CheckboxCheck(theBlock *Block, theShape Rect) {
	theShape.X += 1
	theShape.Y += 1
	theShape.Width -= 2
	theShape.Height -= 2
	r.Rectangle(&theBlock.Where, theShape, color.RGBA{B: 0xFF, G: 0xBF, R: 0x75, A: 0xff}, CVUI_FILLED)
}

func (r *Render) Window(theBlock *Block, theTitleBar Rect, theContent Rect, theTitle string) {
	aTransparecy := false
	aAlpha := 0.3
	aOverlay := theBlock.Where.Clone()

	// Render borders in the title bar
	r.Rectangle(&theBlock.Where, theTitleBar, color.RGBA{0x4A, 0x4A, 0x4A, 0xff}, 1)

	// Render the inside of the title bar
	theTitleBar.X += 1
	theTitleBar.Y += 1
	theTitleBar.Width -= 2
	theTitleBar.Height -= 2
	r.Rectangle(&theBlock.Where, theTitleBar, color.RGBA{0x21, 0x21, 0x21, 0xff}, CVUI_FILLED)

	// Render title text.
	aPos := NewPoint(theTitleBar.X+5, theTitleBar.Y+12)
	gocv.PutText(&theBlock.Where, theTitle, image.Point{int(aPos.X), int(aPos.Y)}, gocv.FontHersheySimplex, 0.4, color.RGBA{0xCE, 0xCE, 0xCE, 0xff}, 1)

	// Render borders of the body
	r.Rectangle(&theBlock.Where, theContent, color.RGBA{0x4A, 0x4A, 0x4A, 0xff}, 1)

	// Render the body filling.
	theContent.X += 1
	theContent.Y += 1
	theContent.Width -= 2
	theContent.Height -= 2
	r.Rectangle(&aOverlay, theContent, color.RGBA{0x31, 0x31, 0x31, 0xff}, CVUI_FILLED)

	if aTransparecy {
		theBlock.Where = aOverlay.Clone()
		// np.copyto(aOverlay, theBlock.Where) // theBlock.where.copyTo(aOverlay);
		r.Rectangle(&aOverlay, theContent, color.RGBA{0x31, 0x31, 0x31, 0xff}, CVUI_FILLED)
		gocv.AddWeighted(aOverlay, aAlpha, theBlock.Where, 1.0-aAlpha, 0.0, &theBlock.Where)
	} else {
		r.Rectangle(&theBlock.Where, theContent, color.RGBA{0x31, 0x31, 0x31, 0xff}, CVUI_FILLED)
	}
}

// 	def rect(self, theBlock, thePos, theBorderColor, theFillingColor):
// 		aBorderColor = r.internal.hexToScalar(theBorderColor)
// 		aFillingColor = r.internal.hexToScalar(theFillingColor)

// 		aHasFilling = aFillingColor[3] != 0xff

// 		if aHasFilling:
// 			in.rectangle(theBlock.where, thePos, aFillingColor, CVUI_FILLED, CVUI_ANTIALISED)

// 		// Render the border
// 		in.rectangle(theBlock.where, thePos, aBorderColor)

func (r *Render) Sparkline(theBlock *Block, theValues []float64, theRect Rect, theMin, theMax float64, theColor uint32) {
	aSize := len(theValues)
	i := 0
	aScale := theMax - theMin
	aGap := float64(theRect.Width) / float64(aSize)
	aPosX := float64(theRect.X)

	for i <= aSize-2 {
		x := int(aPosX)
		y := int((theValues[i]-theMin)/aScale)*-(theRect.Height-5) + theRect.Y + theRect.Height - 5
		aPoint1 := NewPoint(x, y)

		x = int(aPosX + aGap)
		y = int((theValues[i+1]-theMin)/aScale)*-(theRect.Height-5) + theRect.Y + theRect.Height - 5
		aPoint2 := NewPoint(x, y)

		gocv.Line(&theBlock.Where, image.Point{aPoint1.X, aPoint1.Y}, image.Point{aPoint2.X, aPoint2.Y}, r.internal.HexToScalar(theColor), 1)
		aPosX += aGap

		i += 1
	}
}

func handleMouse(theEvent int, theX, theY int, theFlags int, theContext *Context) {
	aButtons := []int{LEFT_BUTTON, MIDDLE_BUTTON, RIGHT_BUTTON}
	aEventsDown := []int{EVENT_LBUTTONDOWN, EVENT_MBUTTONDOWN, EVENT_RBUTTONDOWN}
	aEventsUp := []int{EVENT_LBUTTONUP, EVENT_MBUTTONUP, EVENT_RBUTTONUP}

	for i := LEFT_BUTTON; i < RIGHT_BUTTON+1; i++ {
		aBtn := aButtons[i]

		if theEvent == aEventsDown[i] {
			theContext.Mouse.AnyButton.JustPressed = true
			theContext.Mouse.AnyButton.Pressed = true
			theContext.Mouse.Buttons[aBtn].JustPressed = true
			theContext.Mouse.Buttons[aBtn].Pressed = true
		} else if theEvent == aEventsUp[i] {
			theContext.Mouse.AnyButton.JustReleased = true
			theContext.Mouse.AnyButton.Pressed = false
			theContext.Mouse.Buttons[aBtn].JustReleased = true
			theContext.Mouse.Buttons[aBtn].Pressed = false
		}
	}

	theContext.Mouse.Position.X = theX
	theContext.Mouse.Position.Y = theY
}

// 	See Also
// 	----------
// 	watch()
// 	context()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// 	Track UI interactions of a particular window. This function must be invoked
// 	for any window that will receive gocvui components. gocvui automatically calls `gocvui.watch()`
// 	for any window informed in `gocvui.init()`, so generally you don"t have to watch them
// 	yourin. If you initialized gocvui and told it *not* to create windows automatically,
// 	you need to call `gocvui.watch()` on those windows yourin. `gocvui.watch()` can
// 	automatically create a window before watching it, if it does not exist.

// 	Parameters
// 	----------
// 	theWindowName: str
// 		name of the window whose UI interactions will be tracked.
// 	theCreateNamedWindow: bool
// 		if an OpenCV window named `theWindowName` should be created before it is watched. Windows are created using `cv2.namedWindow()`. If this parameter is `false`, ensure you have called `cv2.namedWindow(WINDOW_NAME)` to create the window, otherwise gocvui will not be able to track its UI interactions.

// 	See Also
// 	----------
// 	init()
// 	context()
func Watch(theWindowName string, theCreateNamedWindow bool) {
	var window Window
	if theCreateNamedWindow {
		window.Window = gocv.NewWindow(theWindowName) //Open windows
		window.Name = theWindowName
	}
	aContex := NewContext()
	aContex.Window = window.Window
	aContex.Mouse.Position.X = 0.
	aContex.Mouse.Position.Y = 0.

	aContex.Mouse.AnyButton.Reset()
	aContex.Mouse.Buttons[RIGHT_BUTTON].Reset()
	aContex.Mouse.Buttons[MIDDLE_BUTTON].Reset()
	aContex.Mouse.Buttons[LEFT_BUTTON].Reset()

	__internal.Contexts[theWindowName] = aContex
	window.SetMouseCallback(handleMouse, __internal.Contexts[theWindowName])
}

// def context(theWindowName):
// 	"""
// 	Inform gocvui that all subsequent component calls belong to a window in particular.
// 	When using gocvui with multiple OpenCV windows, you must call gocvui component calls
// 	between `gocvui.contex(NAME)` and `gocvui.update(NAME)`, where `NAME` is the name of
// 	the window. That way, gocvui knows which window you are using (`NAME` in this case),
// 	so it can track mouse events, for instance.

// 	E.g.

// 	```
// 	// Code for window "window1".
// 	gocvui.context("window1")
// 	gocvui.text(frame, ...)
// 	gocvui.button(frame, ...)
// 	gocvui.update("window1")

// 	// somewhere else, code for "window2"
// 	gocvui.context("window2")
// 	gocvui.printf(frame, ...)
// 	gocvui.printf(frame, ...)
// 	gocvui.update("window2")

// 	// Show everything in a window
// 	gocv.Window.IMShow(frame)
// 	```

// 	Pay attention to the pair `gocvui.context(NAME)` and `gocvui.update(NAME)`, which
// 	encloses the component calls for that window. You need such pair for each window
// 	of your application.

// 	After calling `gocvui.update()`, you can show the result in a window using `gocv.Window.IMShow()`.
// 	If you want to save some typing, you can use `gocvui.imshow()`, which calls `gocvui.update()`
// 	for you and then shows the frame in a window.

// 	E.g.:

// 	```
// 	// Code for window "window1".
// 	gocvui.context("window1")
// 	gocvui.text(frame, ...)
// 	gocvui.button(frame, ...)
// 	gocvui.imshow("window1")

// 	// somewhere else, code for "window2"
// 	gocvui.context("window2")
// 	gocvui.printf(frame, ...)
// 	gocvui.printf(frame, ...)
// 	gocvui.imshow("window2")
// 	```

// 	In that case, you don"t have to bother calling `gocvui.update()` yourself, since
// 	`gocvui.imshow()` will do it for you.

// 	Parameters
// 	----------
// 	theWindowName: str
// 		name of the window that will receive components from all subsequent gocvui calls.

// 	See Also
// 	----------
// 	init()
// 	watch()
// 	"""
// 	__internal.currentContext = theWindowName

// 	Display an image in the specified window and update the internal structures of gocvui.
// 	This function can be used as a replacement for `gocv.Window.IMShow()`. If you want to use
// 	`gocv.Window.IMShow() instead of `gocvui.imshow()`, you must ensure you call `gocvui.update()`
// 	*after* all component calls and *before* `gocv.Window.IMShow()`, so gocvui can update its
// 	internal structures.

// 	In general, it is easier to call `gocvui.imshow()` alone instead of calling
// 	`gocvui.update()" immediately followed by `gocv.Window.IMShow()`.

// 	Parameters
// 	----------
// 	theWindowName: str
// 		name of the window that will be shown.
// 	theFrame: np.ndarray
// 		image, i.e. `np.ndarray`, to be shown in the window.

// 	See Also
// 	----------
// 	update()
// 	context()
// 	watch()
func Imshow(theWindowName string, theFrame gocv.Mat) {
	ctx := __internal.Contexts[theWindowName]

	Update(theWindowName)
	ctx.Window.IMShow(theFrame)
}

// def lastKeyPressed():
// 	"""
// 	Return the last key that was pressed. This function will only
// 	work if a value greater than zero was passed to `gocvui.init()`
// 	as the delay waitkey parameter.

// 	See Also
// 	----------
// 	init()
// 	"""
// 	return __internal.lastKeyPressed

// def mouse(theWindowName = ""):
// 	"""
// 	Return the last position of the mouse.

// 	Parameters
// 	----------
// 	theWindowName: str
// 		name of the window whose mouse cursor will be used. If nothing is informed (default), the function will return the position of the mouse cursor for the default window (the one informed in `gocvui.init()`).

// 	Returns
// 	----------
// 	a point containing the position of the mouse cursor in the speficied window.
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def mouse(theQuery):
// 	"""
// 	Query the mouse for events, e.g. "is any button down now?". Available queries are:

// 	* `gocvui.DOWN`: any mouse button was pressed. `gocvui.mouse()` returns `True` for a single frame only.
// 	* `gocvui.UP`: any mouse button was released.  `gocvui.mouse()` returns `True` for a single frame only.
// 	* `gocvui.CLICK`: any mouse button was clicked (went down then up, no matter the amount of frames in between). `gocvui.mouse()` returns `True` for a single frame only.
// 	* `gocvui.IS_DOWN`: any mouse button is currently pressed. `gocvui.mouse()` returns `True` for as long as the button is down/pressed.

// 	It is easier to think of this function as the answer to a questions. For instance, asking if any mouse button went down:

// 	```
// 	if gocvui.mouse(gocvui.DOWN):
// 	// Any mouse button just went down.

// 	```

// 	The window whose mouse will be queried depends on the context. If `gocvui.mouse(query)` is being called after
// 	`gocvui.context()`, the window informed in the context will be queried. If no context is available, the default
// 	window (informed in `gocvui.init()`) will be used.

// 	Parameters
// 	----------
// 	theQuery: int
// 		an integer describing the intended mouse query. Available queries are `gocvui.DOWN`, `gocvui.UP`, `gocvui.CLICK`, and `gocvui.IS_DOWN`.

// 	See Also
// 	----------
// 	mouse(str)
// 	mouse(str, int)
// 	mouse(str, int, int)
// 	mouse(int, int)
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def mouse(theWindowName, theQuery):
// 	"""
// 	Query the mouse for events in a particular window. This function behave exactly like `gocvui.mouse(int theQuery)`
// 	with the difference that queries are targeted at a particular window.

// 	Parameters
// 	----------
// 	theWindowName: str
// 		name of the window that will be queried.
// 	theQuery: int
// 		an integer describing the intended mouse query. Available queries are `gocvui.DOWN`, `gocvui.UP`, `gocvui.CLICK`, and `gocvui.IS_DOWN`.

// 	See Also
// 	----------
// 	mouse(str)
// 	mouse(str, int, int)
// 	mouse(int, int)
// 	mouse(int)
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def mouse(theButton, theQuery):
// 	"""
// 	Query the mouse for events in a particular button. This function behave exactly like `gocvui.mouse(int theQuery)`,
// 	with the difference that queries are targeted at a particular mouse button instead.

// 	Parameters
// 	----------
// 	theButton: int
// 		an integer describing the mouse button to be queried. Possible values are `gocvui.LEFT_BUTTON`, `gocvui.MIDDLE_BUTTON` and `gocvui.LEFT_BUTTON`.
// 	theQuery: int
// 		an integer describing the intended mouse query. Available queries are `gocvui.DOWN`, `gocvui.UP`, `gocvui.CLICK`, and `gocvui.IS_DOWN`.

// 	See Also
// 	----------
// 	mouse(str)
// 	mouse(str, int, int)
// 	mouse(int)
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def mouse(theWindowName, theButton, theQuery):
// 	"""
// 	Query the mouse for events in a particular button in a particular window. This function behave exactly
// 	like `gocvui.mouse(int theButton, int theQuery)`, with the difference that queries are targeted at
// 	a particular mouse button in a particular window instead.

// 	Parameters
// 	----------
// 	theWindowName: str
// 		name of the window that will be queried.
// 	theButton: int
// 		an integer describing the mouse button to be queried. Possible values are `gocvui.LEFT_BUTTON`, `gocvui.MIDDLE_BUTTON` and `gocvui.LEFT_BUTTON`.
// 	theQuery: int
// 		an integer describing the intended mouse query. Available queries are `gocvui.DOWN`, `gocvui.UP`, `gocvui.CLICK`, and `gocvui.IS_DOWN`.
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def button(theWhere, theX, theY, theLabel):
// 	"""
// 	Display a button. The size of the button will be automatically adjusted to
// 	properly house the label content.

// 	Parameters
// 	----------
// 	theWhere: np.ndarray
// 		image/frame where the component should be rendered.
// 	theX: int
// 		position X where the component should be placed.
// 	theY: int
// 		position Y where the component should be placed.
// 	theLabel: str
// 		text displayed inside the button.

// 	Returns
// 	----------
// 	`true` everytime the user clicks the button.
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def button(theWhere, theX, theY, theWidth, theHeight, theLabel):
// 	"""
// 	Display a button. The button size will be defined by the width and height parameters,
// 	no matter the content of the label.

// 	Parameters
// 	----------
// 	theWhere: np.ndarray
// 		image/frame where the component should be rendered.
// 	theX: int
// 		position X where the component should be placed.
// 	theY: int
// 		position Y where the component should be placed.
// 	theWidth: int
// 		width of the button.
// 	theHeight: int
// 		height of the button.
// 	theLabel: str
// 		text displayed inside the button.

// 	Returns
// 	----------
// 	`true` everytime the user clicks the button.
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def button(theWhere, theX, theY, theIdle, theOver, theDown):
// 	"""
// 	Display a button whose graphics are images (np.ndarray). The button accepts three images to describe its states,
// 	which are idle (no mouse interaction), over (mouse is over the button) and down (mouse clicked the button).
// 	The button size will be defined by the width and height of the images.

// 	Parameters
// 	----------
// 	theWhere: np.ndarray
// 		image/frame where the component should be rendered.
// 	theX: int
// 		position X where the component should be placed.
// 	theY: int
// 		position Y where the component should be placed.
// 	theIdle: np.ndarray
// 		an image that will be rendered when the button is not interacting with the mouse cursor.
// 	theOver: np.ndarray
// 		an image that will be rendered when the mouse cursor is over the button.
// 	theDown: np.ndarray
// 		an image that will be rendered when the mouse cursor clicked the button (or is clicking).

// 	Returns
// 	----------
// 	`true` everytime the user clicks the button.

// 	See Also
// 	----------
// 	button()
// 	image()
// 	iarea()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def image(theWhere, theX, theY, theImage):
// 	"""
// 	Display an image (np.ndarray).

// 	Parameters
// 	----------
// 	theWhere: np.ndarray
// 		image/frame where the provded image should be rendered.
// 	theX: int
// 		position X where the image should be placed.
// 	theY: int
// 		position Y where the image should be placed.
// 	theImage: np.ndarray
// 		image to be rendered in the specified destination.

// 	See Also
// 	----------
// 	button()
// 	iarea()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def checkbox(theWhere, theX, theY, theLabel, theState, theColor = 0xCECECE):
// 	"""
// 	Display a checkbox. You can use the state parameter to monitor if the
// 	checkbox is checked or not.

// 	Parameters
// 	----------
// 	theWhere: np.ndarray
// 		image/frame where the component should be rendered.
// 	theX: int
// 		position X where the component should be placed.
// 	theY: int
// 		position Y where the component should be placed.
// 	theLabel: str
// 		text displayed besides the clickable checkbox square.
// 	theState: [bool]
// 		array or list of booleans whose first position, i.e. theState[0], will be used to store the current state of the checkbox: `True` means the checkbox is checked.
// 	theColor: uint
// 		color of the label in the format `0xRRGGBB`, e.g. `0xff0000` for red.

// 	Returns
// 	----------
// 	a boolean value that indicates the current state of the checkbox, `true` if it is checked.
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def text(theWhere, theX, theY, theText, theFontScale = 0.4, theColor = 0xCECECE):
// 	"""
// 	Display a piece of text.

// 	Parameters
// 	----------
// 	theWhere: np.ndarray
// 		image/frame where the component should be rendered.
// 	theX: int
// 		position X where the component should be placed.
// 	theY: int
// 		position Y where the component should be placed.
// 	theText: str
// 		the text content.
// 	theFontScale: float
// 		size of the text.
// 	theColor: uint
// 		color of the text in the format `0xRRGGBB`, e.g. `0xff0000` for red.

// 	See Also
// 	----------
// 	printf()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def printf(theWhere, theX, theY, theFontScale, theColor, theFmt):
// 	"""
// 	Display a piece of text that can be formated using `C stdio"s printf()` style. For instance
// 	if you want to display text mixed with numbers, you can use:

// 	```
// 	printf(frame, 10, 15, 0.4, 0xff0000, "Text: %d and %f", 7, 3.1415)
// 	```

// 	Parameters
// 	----------
// 	theWhere: np.ndarray
// 		image/frame where the component should be rendered.
// 	theX: int
// 		position X where the component should be placed.
// 	theY: int
// 		position Y where the component should be placed.
// 	theFontScale: float
// 		size of the text.
// 	theColor: uint
// 		color of the text in the format `0xRRGGBB`, e.g. `0xff0000` for red.
// 	theFmt: str
// 		formating string as it would be supplied for `stdio"s printf()`, e.g. `"Text: %d and %f", 7, 3.1415`.

// 	See Also
// 	----------
// 	text()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def printf(theWhere, theX, theY, theFmt):
// 	"""
// 	Display a piece of text that can be formated using `C stdio"s printf()` style. For instance
// 	if you want to display text mixed with numbers, you can use:

// 	```
// 	printf(frame, 10, 15, 0.4, 0xff0000, "Text: %d and %f", 7, 3.1415)
// 	```

// 	The size and color of the text will be based on gocvui"s default values.

// 	Parameters
// 	----------
// 	theWhere: np.ndarray
// 		image/frame where the component should be rendered.
// 	theX: int
// 		position X where the component should be placed.
// 	theY: int
// 		position Y where the component should be placed.
// 	theFmt: str
// 		formating string as it would be supplied for `stdio"s printf()`, e.g. `"Text: %d and %f", 7, 3.1415`.

// 	See Also
// 	----------
// 	text()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def counter(theWhere, theX, theY, theValue, theStep = 1, theFormat = "%d"):
// 	"""
// 	Display a counter for integer values that the user can increase/descrease
// 	by clicking the up and down arrows.

// 	Parameters
// 	----------
// 	theWhere: np.ndarray
// 		image/frame where the component should be rendered.
// 	theX: int
// 		position X where the component should be placed.
// 	theY: int
// 		position Y where the component should be placed.
// 	theValue: [number]
// 		array or list of numbers whose first position, i.e. theValue[0], will be used to store the current value of the counter.
// 	theStep: number
// 		amount that should be increased/decreased when the user interacts with the counter buttons
// 	theFormat: str
// 		how the value of the counter should be presented, as it was printed by `stdio"s printf()`. E.g. `"%d"` means the value will be displayed as an integer, `"%0d"` integer with one leading zero, etc.

// 	Returns
// 	----------
// 	number that corresponds to the current value of the counter.
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def trackbar(theWhere, theX, theY, theWidth, theValue, theMin, theMax, theSegments = 1, theLabelFormat = "%.1Lf", theOptions = 0, theDiscreteStep = 1):
// 	"""
// 	Display a trackbar for numeric values that the user can increase/decrease
// 	by clicking and/or dragging the marker right or left. This component can use
// 	different types of data as its value, so it is imperative provide the right
// 	label format, e.g. "%d" for ints, otherwise you might end up with weird errors.

// 	Example:

// 	```
// 	// using float
// 	trackbar(where, x, y, width, &floatValue, 0.0, 50.0)

// 	// using float
// 	trackbar(where, x, y, width, &floatValue, 0.0f, 50.0f)

// 	// using char
// 	trackbar(where, x, y, width, &charValue, (char)1, (char)10)
// 	```

// 	Parameters
// 	----------
// 	theWhere: np.ndarray
// 		image/frame where the component should be rendered.
// 	theX: int
// 		position X where the component should be placed.
// 	theY: int
// 		position Y where the component should be placed.
// 	theWidth: int
// 		width of the trackbar.
// 	theValue: [number]
// 		array or list of numbers whose first position, i.e. theValue[0], will be used to store the current value of the trackbar. It will be modified when the user interacts with the trackbar. Any numeric type can be used, e.g. int, float, long double, etc.
// 	theMin: number
// 		minimum value allowed for the trackbar.
// 	theMax: number
// 		maximum value allowed for the trackbar.
// 	theSegments: int
// 		number of segments the trackbar will have (default is 1). Segments can be seen as groups of numbers in the scale of the trackbar. For example, 1 segment means a single groups of values (no extra labels along the scale), 2 segments mean the trackbar values will be divided in two groups and a label will be placed at the middle of the scale.
// 	theLabelFormat: str
// 		formating string that will be used to render the labels. If you are using a trackbar with integers values, for instance, you can use `%d` to render labels.
// 	theOptions: uint
// 		options to customize the behavior/appearance of the trackbar, expressed as a bitset. Available options are defined as `gocvui.TRACKBAR_` constants and they can be combined using the bitwise `|` operand. Available options are: `TRACKBAR_HIDE_SEGMENT_LABELS` (do not render segment labels, but do render min/max labels), `TRACKBAR_HIDE_STEP_SCALE` (do not render the small lines indicating values in the scale), `TRACKBAR_DISCRETE` (changes of the trackbar value are multiples of theDiscreteStep param), `TRACKBAR_HIDE_MIN_MAX_LABELS` (do not render min/max labels), `TRACKBAR_HIDE_VALUE_LABEL` (do not render the current value of the trackbar below the moving marker), `TRACKBAR_HIDE_LABELS` (do not render labels at all).
// 	theDiscreteStep: number
// 		amount that the trackbar marker will increase/decrease when the marker is dragged right/left (if option TRACKBAR_DISCRETE is ON)

// 	Returns
// 	----------
// 	`true` when the value of the trackbar changed.

// 	See Also
// 	----------
// 	counter()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def window(theWhere, theX, theY, theWidth, theHeight, theTitle):
// 	"""
// 	Display a window (a block with a title and a body).

// 	Parameters
// 	----------
// 	theWhere: np.ndarray
// 		image/frame where the component should be rendered.
// 	theX: int
// 		position X where the component should be placed.
// 	theY: int
// 		position Y where the component should be placed.
// 	theWidth: int
// 		width of the window.
// 	theHeight: int
// 		height of the window.
// 	theTitle: str
// 		text displayed as the title of the window.

// 	See Also
// 	----------
// 	rect()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def rect(theWhere, theX, theY, theWidth, theHeight, theBorderColor, theFillingColor = 0xff000000):
// 	"""
// 	Display a filled rectangle.

// 	Parameters
// 	----------
// 	theWhere: np.ndarray
// 		image/frame where the component should be rendered.
// 	theX: int
// 		position X where the component should be placed.
// 	theY: int
// 		position Y where the component should be placed.
// 	theWidth: int
// 		width of the rectangle.
// 	theHeight: int
// 		height of the rectangle.
// 	theBorderColor: uint
// 		color of rectangle"s border in the format `0xRRGGBB`, e.g. `0xff0000` for red.
// 	theFillingColor: uint
// 		color of rectangle"s filling in the format `0xAARRGGBB`, e.g. `0x00ff0000` for red, `0xff000000` for transparent filling.

// 	See Also
// 	----------
// 	image()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def sparkline(theWhere, theValues, theX, theY, theWidth, theHeight, theColor = 0x00FF00):
// 	"""
// 	Display the values of a vector as a sparkline.

// 	Parameters
// 	----------
// 	theWhere: np.ndarray
// 		image/frame where the component should be rendered.
// 	theValues: number[]
// 		array or list containing the numeric values to be used in the sparkline.
// 	theX: int
// 		position X where the component should be placed.
// 	theY: int
// 		position Y where the component should be placed.
// 	theWidth: int
// 		width of the sparkline.
// 	theHeight: int
// 		height of the sparkline.
// 	theColor: uint
// 		color of sparkline in the format `0xRRGGBB`, e.g. `0xff0000` for red.

// 	See Also
// 	----------
// 	trackbar()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def iarea(theX, theY, theWidth, theHeight):
// 	"""
// 	Create an interaction area that reports activity with the mouse cursor.
// 	The tracked interactions are returned by the function and they are:

// 	`OUT` when the cursor is not over the iarea.
// 	`OVER` when the cursor is over the iarea.
// 	`DOWN` when the cursor is pressed over the iarea, but not released yet.
// 	`CLICK` when the cursor clicked (pressed and released) within the iarea.

// 	This function creates no visual output on the screen. It is intended to
// 	be used as an auxiliary tool to create interactions.

// 	Parameters
// 	----------
// 	theX: int
// 		position X where the interactive area should be placed.
// 	theY: int
// 		position Y where the interactive area should be placed.
// 	theWidth: int
// 		width of the interactive area.
// 	theHeight: int
// 		height of the interactive area.

// 	Returns
// 	----------
// 	integer value representing the current state of interaction with the mouse cursor. It can be `OUT` (cursor is not over the area), `OVER` (cursor is over the area), `DOWN` (cursor is pressed over the area, but not released yet) and `CLICK` (cursor clicked, i.e. pressed and released, within the area).

// 	See Also
// 	----------
// 	button()
// 	image()
// 	"""
// 	return __internal.iarea(theX, theY, theWidth, theHeight)

// def beginRow(theWhere, theX, theY, theWidth = -1, theHeight = -1, thePadding = 0):
// 	"""
// 	Start a new row.

// 	One of the most annoying tasks when building UI is to calculate
// 	where each component should be placed on the screen. gocvui has
// 	a set of methods that abstract the process of positioning
// 	components, so you don"t have to think about assigning a
// 	X and Y coordinate. Instead you just add components and gocvui
// 	will place them as you go.

// 	You use `beginRow()` to start a group of elements. After `beginRow()`
// 	has been called, all subsequent component calls don"t have to specify
// 	the frame where the component should be rendered nor its position.
// 	The position of the component will be automatically calculated by gocvui
// 	based on the components within the group. All components are placed
// 	side by side, from left to right.

// 	E.g.

// 	```
// 	beginRow(frame, x, y, width, height)
// 	text("test")
// 	button("btn")
// 	endRow()
// 	```

// 	Rows and columns can be nested, so you can create columns/rows within
// 	columns/rows as much as you want. It"s important to notice that any
// 	component within `beginRow()` and `endRow()` *do not* specify the position
// 	where the component is rendered, which is also True for `beginRow()`.
// 	As a consequence, **be sure you are calling `beginRow(width, height)`
// 	when the call is nested instead of `beginRow(x, y, width, height)`**,
// 	otherwise gocvui will throw an error.

// 	E.g.

// 	```
// 	beginRow(frame, x, y, width, height)
// 	text("test")
// 	button("btn")

// 	beginColumn()      // no frame nor x,y parameters here!
// 	text("column1")
// 	text("column2")
// 	endColumn()
// 	endRow()
// 	```

// 	Don"t forget to call `endRow()` to finish the row, otherwise gocvui will throw an error.

// 	Parameters
// 	----------
// 	theWhere: np.ndarray
// 		image/frame where the components within this block should be rendered.
// 	theX: int
// 		position X where the row should be placed.
// 	theY: int
// 		position Y where the row should be placed.
// 	theWidth: int
// 		width of the row. If a negative value is specified, the width of the row will be automatically calculated based on the content of the block.
// 	theHeight: int
// 		height of the row. If a negative value is specified, the height of the row will be automatically calculated based on the content of the block.
// 	thePadding: int
// 		space, in pixels, among the components of the block.

// 	See Also
// 	----------
// 	beginColumn()
// 	endRow()
// 	endColumn()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def endRow():
// 	"""
// 	End a row. You must call this function only if you have previously called
// 	its counter part, the `beginRow()` function.

// 	See Also
// 	----------
// 	beginRow()
// 	beginColumn()
// 	endColumn()
// 	"""
// 	__internal.end(ROW)

// def beginColumn(theWhere, theX, theY, theWidth = -1, theHeight = -1, thePadding = 0):
// 	"""
// 	Start a new column.

// 	One of the most annoying tasks when building UI is to calculate
// 	where each component should be placed on the screen. gocvui has
// 	a set of methods that abstract the process of positioning
// 	components, so you don"t have to think about assigning a
// 	X and Y coordinate. Instead you just add components and gocvui
// 	will place them as you go.

// 	You use `beginColumn()` to start a group of elements. After `beginColumn()`
// 	has been called, all subsequent component calls don"t have to specify
// 	the frame where the component should be rendered nor its position.
// 	The position of the component will be automatically calculated by gocvui
// 	based on the components within the group. All components are placed
// 	below each other, from the top of the screen towards the bottom.

// 	E.g.

// 	```
// 	beginColumn(frame, x, y, width, height)
// 	text("test")
// 	button("btn")
// 	endColumn()
// 	```

// 	Rows and columns can be nested, so you can create columns/rows within
// 	columns/rows as much as you want. It"s important to notice that any
// 	component within `beginColumn()` and `endColumn()` *do not* specify the position
// 	where the component is rendered, which is also True for `beginColumn()`.
// 	As a consequence, **be sure you are calling `beginColumn(width, height)`
// 	when the call is nested instead of `beginColumn(x, y, width, height)`**,
// 	otherwise gocvui will throw an error.

// 	E.g.

// 	```
// 	beginColumn(frame, x, y, width, height)
// 	text("test")
// 	button("btn")

// 	beginRow()      // no frame nor x,y parameters here!
// 	text("column1")
// 	text("column2")
// 	endRow()
// 	endColumn()
// 	```

// 	Don"t forget to call `endColumn()` to finish the column, otherwise gocvui will throw an error.

// 	Parameters
// 	----------
// 	theWhere: np.ndarray
// 		image/frame where the components within this block should be rendered.
// 	theX: int
// 		position X where the row should be placed.
// 	theY: int
// 		position Y where the row should be placed.
// 	theWidth: int
// 		width of the column. If a negative value is specified, the width of the column will be automatically calculated based on the content of the block.
// 	theHeight: int
// 		height of the column. If a negative value is specified, the height of the column will be automatically calculated based on the content of the block.
// 	thePadding: int
// 		space, in pixels, among the components of the block.

// 	See Also
// 	----------
// 	beginRow()
// 	endColumn()
// 	endRow()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def endColumn():
// 	"""
// 	End a column. You must call this function only if you have previously called
// 	its counter part, i.e. `beginColumn()`.

// 	See Also
// 	----------
// 	beginColumn()
// 	beginRow()
// 	endRow()
// 	"""
// 	__internal.end(COLUMN)

// def beginRow(theWidth = -1, theHeight = -1, thePadding = 0):
// 	"""
// 	Start a row. This function behaves in the same way as `beginRow(frame, x, y, width, height)`,
// 	however it is suposed to be used within `begin*()/end*()` blocks since they require components
// 	not to inform frame nor x,y coordinates.

// 	IMPORTANT: this function can only be used within a `begin*()/end*()` block, otherwise it does nothing.

// 	Parameters
// 	----------
// 	theWidth: int
// 		width of the row. If a negative value is specified, the width of the row will be automatically calculated based on the content of the block.
// 	theHeight: int
// 		height of the row. If a negative value is specified, the height of the row will be automatically calculated based on the content of the block.
// 	thePadding: int
// 		space, in pixels, among the components of the block.

// 	See Also
// 	----------
// 	beginColumn()
// 	endRow()
// 	endColumn()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def beginColumn(theWidth = -1, theHeight = -1, thePadding = 0):
// 	"""
// 	Start a column. This function behaves in the same way as `beginColumn(frame, x, y, width, height)`,
// 	however it is suposed to be used within `begin*()/end*()` blocks since they require components
// 	not to inform frame nor x,y coordinates.

// 	IMPORTANT: this function can only be used within a `begin*()/end*()` block, otherwise it does nothing.

// 	Parameters
// 	----------
// 	theWidth: int
// 		width of the column. If a negative value is specified, the width of the column will be automatically calculated based on the content of the block.
// 	theHeight: int
// 		height of the column. If a negative value is specified, the height of the column will be automatically calculated based on the content of the block.
// 	thePadding: int
// 		space, in pixels, among the components of the block.

// 	See Also
// 	----------
// 	beginColumn()
// 	endRow()
// 	endColumn()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def space(theValue = 5):
// 	"""
// 	Add an arbitrary amount of space between components within a `begin*()` and `end*()` block.
// 	The function is aware of context, so if it is used within a `beginColumn()` and
// 	`endColumn()` block, the space will be vertical. If it is used within a `beginRow()`
// 	and `endRow()` block, space will be horizontal.

// 	IMPORTANT: this function can only be used within a `begin*()/end*()` block, otherwise it does nothing.

// 	Parameters
// 	----------
// 	theValue: int
// 		the amount of space to be added.

// 	See Also
// 	----------
// 	beginColumn()
// 	beginRow()
// 	endRow()
// 	endColumn()
// 	"""
// 	aBlock = __internal.topBlock()
// 	aSize = Size(theValue, theValue)

// 	__internal.updateLayoutFlow(aBlock, aSize)

// def text(theText, theFontScale = 0.4, theColor = 0xCECECE):
// 	"""
// 	Display a piece of text within a `begin*()` and `end*()` block.

// 	IMPORTANT: this function can only be used within a `begin*()/end*()` block, otherwise it does nothing.

// 	Parameters
// 	----------
// 	theText: str
// 		text content.
// 	theFontScale: float
// 		size of the text.
// 	theColor: uint
// 		color of the text in the format `0xRRGGBB`, e.g. `0xff0000` for red.

// 	See Also
// 	----------
// 	printf()
// 	beginColumn()
// 	beginRow()
// 	endRow()
// 	endColumn()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def button(theWidth, theHeight, theLabel):
// 	"""
// 	Display a button within a `begin*()` and `end*()` block.
// 	The button size will be defined by the width and height parameters,
// 	no matter the content of the label.

// 	IMPORTANT: this function can only be used within a `begin*()/end*()` block, otherwise it does nothing.

// 	Parameters
// 	----------
// 	theWidth: int
// 		width of the button.
// 	theHeight: int
// 		height of the button.
// 	theLabel: str
// 		text displayed inside the button. You can set shortcuts by pre-pending them with "&"

// 	Returns
// 	----------
// 	`true` everytime the user clicks the button.

// 	See Also
// 	----------
// 	beginColumn()
// 	beginRow()
// 	endRow()
// 	endColumn()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def button(theLabel):
// 	"""
// 	Display a button within a `begin*()` and `end*()` block. The size of the button will be
// 	automatically adjusted to properly house the label content.

// 	IMPORTANT: this function can only be used within a `begin*()/end*()` block, otherwise it does nothing.

// 	Parameters
// 	----------
// 	theLabel: str
// 		text displayed inside the button. You can set shortcuts by pre-pending them with "&"

// 	Returns
// 	----------
// 	`true` everytime the user clicks the button.

// 	See Also
// 	----------
// 	beginColumn()
// 	beginRow()
// 	endRow()
// 	endColumn()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def button(theIdle, theOver, theDown):
// 	"""
// 	Display a button whose graphics are images (np.ndarray).

// 	IMPORTANT: this function can only be used within a `begin*()/end*()` block, otherwise it does nothing.

// 	The button accepts three images to describe its states,
// 	which are idle (no mouse interaction), over (mouse is over the button) and down (mouse clicked the button).
// 	The button size will be defined by the width and height of the images.

// 	Parameters
// 	----------
// 	theIdle: np.ndarray
// 		image that will be rendered when the button is not interacting with the mouse cursor.
// 	theOver: np.ndarray
// 		image that will be rendered when the mouse cursor is over the button.
// 	theDown: np.ndarray
// 		image that will be rendered when the mouse cursor clicked the button (or is clicking).

// 	Returns
// 	----------
// 	`true` everytime the user clicks the button.

// 	See Also
// 	----------
// 	button()
// 	image()
// 	iarea()
// 	beginColumn()
// 	beginRow()
// 	endRow()
// 	endColumn()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def image(theImage):
// 	"""
// 	Display an image (np.ndarray) within a `begin*()` and `end*()` block

// 	IMPORTANT: this function can only be used within a `begin*()/end*()` block, otherwise it does nothing.

// 	Parameters
// 	----------
// 	theImage: np.ndarray
// 		image to be rendered in the specified destination.

// 	See Also
// 	----------
// 	button()
// 	iarea()
// 	beginColumn()
// 	beginRow()
// 	endRow()
// 	endColumn()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def checkbox(theLabel, theState, theColor = 0xCECECE):
// 	"""
// 	Display a checkbox within a `begin*()` and `end*()` block. You can use the state parameter
// 	to monitor if the checkbox is checked or not.

// 	IMPORTANT: this function can only be used within a `begin*()/end*()` block, otherwise it does nothing.

// 	Parameters
// 	----------
// 	theLabel: str
// 		text displayed besides the clickable checkbox square.
// 	theState: [bool]
// 		array or list of booleans whose first position, i.e. theState[0], will be used to store the current state of the checkbox: `True` means the checkbox is checked.
// 	theColor: uint
// 		color of the label in the format `0xRRGGBB`, e.g. `0xff0000` for red.

// 	Returns
// 	----------
// 	a boolean value that indicates the current state of the checkbox, `true` if it is checked.

// 	See Also
// 	----------
// 	beginColumn()
// 	beginRow()
// 	endRow()
// 	endColumn()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def printf(theFontScale, theColor, theFmt):
// 	"""
// 	Display a piece of text within a `begin*()` and `end*()` block.

// 	IMPORTANT: this function can only be used within a `begin*()/end*()` block, otherwise it does nothing.

// 	The text can be formated using `C stdio"s printf()` style. For instance if you want to display text mixed
// 	with numbers, you can use:

// 	```
// 	printf(0.4, 0xff0000, "Text: %d and %f", 7, 3.1415)
// 	```

// 	Parameters
// 	----------
// 	theFontScale: float
// 		size of the text.
// 	theColor: uint
// 		color of the text in the format `0xRRGGBB`, e.g. `0xff0000` for red.
// 	theFmt: str
// 		formating string as it would be supplied for `C stdio"s printf()`, e.g. `"Text: %d and %f", 7, 3.1415`.

// 	See Also
// 	----------
// 	text()
// 	beginColumn()
// 	beginRow()
// 	endRow()
// 	endColumn()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def printf(theFmt):
// 	"""
// 	Display a piece of text that can be formated using `C stdio"s printf()` style.

// 	IMPORTANT: this function can only be used within a `begin*()/end*()` block, otherwise it does nothing.

// 	For instance if you want to display text mixed with numbers, you can use:

// 	```
// 	printf(frame, 10, 15, 0.4, 0xff0000, "Text: %d and %f", 7, 3.1415)
// 	```

// 	The size and color of the text will be based on gocvui"s default values.

// 	Parameters
// 	----------
// 	theFmt: str
// 		formating string as it would be supplied for `stdio"s printf()`, e.g. `"Text: %d and %f", 7, 3.1415`.

// 	See Also
// 	----------
// 	text()
// 	beginColumn()
// 	beginRow()
// 	endRow()
// 	endColumn()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def counter(theValue, theStep = 1, theFormat = "%d"):
// 	"""
// 	Display a counter for integer values that the user can increase/descrease
// 	by clicking the up and down arrows.

// 	IMPORTANT: this function can only be used within a `begin*()/end*()` block, otherwise it does nothing.

// 	Parameters
// 	----------
// 	theValue: [number]
// 		array or list of numbers whose first position, i.e. theValue[0], will be used to store the current value of the counter.
// 	theStep: number
// 		amount that should be increased/decreased when the user interacts with the counter buttons.
// 	theFormat: str
// 		how the value of the counter should be presented, as it was printed by `C stdio"s printf()`. E.g. `"%d"` means the value will be displayed as an integer, `"%0d"` integer with one leading zero, etc.

// 	Returns
// 	----------
// 	number that corresponds to the current value of the counter.

// 	See Also
// 	----------
// 	printf()
// 	beginColumn()
// 	beginRow()
// 	endRow()
// 	endColumn()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def trackbar(theWidth, theValue, theMin, theMax, theSegments = 1, theLabelFormat = "%.1Lf", theOptions = 0, theDiscreteStep = 1):
// 	"""
// 	Display a trackbar for numeric values that the user can increase/decrease
// 	by clicking and/or dragging the marker right or left.

// 	IMPORTANT: this function can only be used within a `begin*()/end*()` block, otherwise it does nothing.

// 	This component uses templates so it is imperative that you make it very explicit
// 	the type of `theValue`, `theMin`, `theMax` and `theStep`, otherwise you might end up with
// 	weird compilation errors.

// 	Example:

// 	```
// 	// using float
// 	trackbar(width, &floatValue, 0.0, 50.0)

// 	// using float
// 	trackbar(width, &floatValue, 0.0f, 50.0f)

// 	// using char
// 	trackbar(width, &charValue, (char)1, (char)10)
// 	```

// 	Parameters
// 	----------
// 	theWidth: int
// 		the width of the trackbar.
// 	theValue: [number]
// 		array or list of numbers whose first position, i.e. theValue[0], will be used to store the current value of the trackbar. It will be modified when the user interacts with the trackbar. Any numeric type can be used, e.g. int, float, long double, etc.
// 	theMin: number
// 		minimum value allowed for the trackbar.
// 	theMax: number
// 		maximum value allowed for the trackbar.
// 	theSegments: int
// 		number of segments the trackbar will have (default is 1). Segments can be seen as groups of numbers in the scale of the trackbar. For example, 1 segment means a single groups of values (no extra labels along the scale), 2 segments mean the trackbar values will be divided in two groups and a label will be placed at the middle of the scale.
// 	theLabelFormat: str
// 		formating string that will be used to render the labels, e.g. `%.2Lf`. No matter the type of the `theValue` param, internally trackbar stores it as a `long float`, so the formating string will *always* receive a `long float` value to format. If you are using a trackbar with integers values, for instance, you can supress decimals using a formating string as `%.0Lf` to format your labels.
// 	theOptions: uint
// 		options to customize the behavior/appearance of the trackbar, expressed as a bitset. Available options are defined as `TRACKBAR_` constants and they can be combined using the bitwise `|` operand. Available options are: `TRACKBAR_HIDE_SEGMENT_LABELS` (do not render segment labels, but do render min/max labels), `TRACKBAR_HIDE_STEP_SCALE` (do not render the small lines indicating values in the scale), `TRACKBAR_DISCRETE` (changes of the trackbar value are multiples of informed step param), `TRACKBAR_HIDE_MIN_MAX_LABELS` (do not render min/max labels), `TRACKBAR_HIDE_VALUE_LABEL` (do not render the current value of the trackbar below the moving marker), `TRACKBAR_HIDE_LABELS` (do not render labels at all).
// 	theDiscreteStep: number
// 		amount that the trackbar marker will increase/decrease when the marker is dragged right/left (if option TRACKBAR_DISCRETE is ON)

// 	Returns
// 	----------
// 	`true` when the value of the trackbar changed.

// 	See Also
// 	----------
// 	counter()
// 	beginColumn()
// 	beginRow()
// 	endRow()
// 	endColumn()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def window(theWidth, theHeight, theTitle):
// 	"""
// 	Display a window (a block with a title and a body) within a `begin*()` and `end*()` block.

// 	IMPORTANT: this function can only be used within a `begin*()/end*()` block, otherwise it does nothing.

// 	Parameters
// 	----------
// 	theWidth: int
// 		width of the window.
// 	theHeight: int
// 		height of the window.
// 	theTitle: str
// 		text displayed as the title of the window.

// 	See Also
// 	----------
// 	rect()
// 	beginColumn()
// 	beginRow()
// 	endRow()
// 	endColumn()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def rect(theWidth, theHeight, theBorderColor, theFillingColor = 0xff000000):
// 	"""
// 	Display a rectangle within a `begin*()` and `end*()` block.

// 	IMPORTANT: this function can only be used within a `begin*()/end*()` block, otherwise it does nothing.

// 	Parameters
// 	----------
// 	theWidth: int
// 		width of the rectangle.
// 	theHeight: int
// 		height of the rectangle.
// 	theBorderColor: uint
// 		color of rectangle"s border in the format `0xRRGGBB`, e.g. `0xff0000` for red.
// 	theFillingColor: uint
// 		color of rectangle"s filling in the format `0xAARRGGBB`, e.g. `0x00ff0000` for red, `0xff000000` for transparent filling.

// 	See Also
// 	----------
// 	window()
// 	beginColumn()
// 	beginRow()
// 	endRow()
// 	endColumn()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// def sparkline(theValues, theWidth, theHeight, theColor = 0x00FF00):
// 	"""
// 	Display the values of a vector as a sparkline within a `begin*()` and `end*()` block.

// 	IMPORTANT: this function can only be used within a `begin*()/end*()` block, otherwise it does nothing.

// 	Parameters
// 	----------
// 	theValues: number[]
// 		array or list of numeric values that will be rendered as a sparkline.
// 	theWidth: int
// 		width of the sparkline.
// 	theHeight: int
// 		height of the sparkline.
// 	theColor: uint
// 		color of sparkline in the format `0xRRGGBB`, e.g. `0xff0000` for red.

// 	See Also
// 	----------
// 	beginColumn()
// 	beginRow()
// 	endRow()
// 	endColumn()
// 	"""
// 	print("This is wrapper function to help code autocompletion.")

// 	Update the library internal things. You need to call this function **AFTER** you are done adding/manipulating
// 	UI elements in order for them to react to mouse interactions.

// 	Parameters
// 	----------
// 	theWindowName: str
// 		name of the window whose components are being updated. If no window name is provided, gocvui uses the default window.

// 	See Also
// 	----------
// 	init()
// 	watch()
// 	context()
func Update(theArgs ...interface{}) {
	theWindowName := ""
	if len(theArgs) != 0 {
		var ok bool
		if theWindowName, ok = theArgs[0].(string); !ok {
			__internal.Error(7, "theWindowName error")
		}
	}

	aContext := __internal.GetContext(theWindowName)
	aContext.Mouse.AnyButton.JustReleased = false
	aContext.Mouse.AnyButton.JustPressed = false

	for i := LEFT_BUTTON; i < RIGHT_BUTTON+1; i++ {
		aContext.Mouse.Buttons[i].JustReleased = false
		aContext.Mouse.Buttons[i].JustPressed = false
	}

	__internal.Screen.Reset()

	// If we were told to keep track of the keyboard shortcuts, we
	// proceed to handle opencv event queue.
	if __internal.DelayWaitKey > 0 {
		__internal.LastKeyPressed = gocv.WaitKey(__internal.DelayWaitKey)
	}

	if !__internal.BlockStackEmpty() {
		__internal.Error(2, "Calling update() before finishing all begin*()/end*() calls. Did you forget to call a begin*() or an end*()? Check if every begin*() has an appropriate end*() call before you call update().")
	}
}

func Init(theArgs ...interface{}) {

	if __internal.IsString(theArgs[0]) {
		// Signature: init(theWindowName, theDelayWaitKey = -1, theCreateNamedWindow = True)
		aWindowName := theArgs[0].(string)
		aDelayWaitKey := -1
		if len(theArgs) >= 2 {
			aDelayWaitKey = theArgs[1].(int)
		}

		aCreateNamedWindow := true
		if len(theArgs) >= 3 {
			aCreateNamedWindow = theArgs[2].(bool)
		}

		__internal.Init(aWindowName, aDelayWaitKey)
		Watch(aWindowName, aCreateNamedWindow)
	} else {
		// Signature: init(theWindowNames[], theHowManyWindows, theDelayWaitKey = -1, theCreateNamedWindows = True)
		aWindowNames := theArgs[0].([]string)
		aHowManyWindows := theArgs[1].(int)
		aDelayWaitKey := -1
		if len(theArgs) >= 3 {
			aDelayWaitKey = theArgs[2].(int)
		}

		aCreateNamedWindows := true
		if len(theArgs) >= 4 {
			aCreateNamedWindows = theArgs[3].(bool)
		}

		__internal.Init(aWindowNames[0], aDelayWaitKey)
		for i := 0; i < aHowManyWindows; i++ {
			Watch(aWindowNames[i], aCreateNamedWindows)
		}

	}
}

func Text(theArgs ...interface{}) {
	var aBlock *Block
	var aX, aY int
	var aText string
	aFontScale := 0.4
	aColor := uint32(0xCECECE)

	if _, ok := theArgs[0].(gocv.Mat); ok {
		// Signature: text(theWhere, theX, theY, theText, theFontScale = 0.4, theColor = 0xCECECE)
		aWhere := theArgs[0].(gocv.Mat)
		aX = theArgs[1].(int)
		aY = theArgs[2].(int)
		aText = theArgs[3].(string)

		if len(theArgs) >= 5 {
			aFontScale = theArgs[4].(float64)
		}

		if len(theArgs) >= 6 {
			aColor = theArgs[5].(uint32)
		}

		__internal.Screen.Where = aWhere
		aBlock = &__internal.Screen
	} else {
		// Signature: text(theText, theFontScale = 0.4, theColor = 0xCECECE)
		aBlock = __internal.TopBlock()
		aX = aBlock.Anchor.X
		aY = aBlock.Anchor.Y
		aText = theArgs[0].(string)

		if len(theArgs) >= 2 {
			aFontScale = theArgs[1].(float64)
		}

		if len(theArgs) >= 3 {
			aColor = theArgs[2].(uint32)
		}
	}
	__internal.Text(aBlock, aX, aY, aText, aFontScale, aColor, true)
}

// def printf(*theArgs):
// 	if isinstance(theArgs[0], np.ndarray):
// 		// Signature: printf(theWhere, theX, theY, ...)
// 		aWhere = theArgs[0]
// 		aX = theArgs[1]
// 		aY = theArgs[2]

// 		__internal.screen.where = aWhere
// 		aBlock = __internal.screen

// 		aArgs = theArgs[3:]
// 	else:
// 		// Row/column function
// 		aBlock = __internal.topBlock()
// 		aX = aBlock.anchor.x
// 		aY = aBlock.anchor.y
// 		aArgs = theArgs

// 	if __internal.isString(aArgs[0]):
// 		// Signature: printf(theWhere, theX, theY, theFmt, ...)
// 		aFontScale = 0.4
// 		aColor = 0xCECECE
// 		aFmt = aArgs[0]
// 		aFmtArgs = aArgs[1:]
// 	else:
// 		// Signature: printf(theWhere, theX, theY, theFontScale, theColor, theFmt, ...)
// 		aFontScale = aArgs[0]
// 		aColor = aArgs[1]
// 		aFmt = aArgs[2]
// 		aFmtArgs = aArgs[3:]

// 	aText = aFmt % aFmtArgs
// 	__internal.text(aBlock, aX, aY, aText, aFontScale, aColor, True)

// def counter(*theArgs):
// 	if isinstance(theArgs[0], np.ndarray):
// 		// Signature: counter(theWhere, theX, theY, theValue, theStep = 1, theFormat = "")
// 		aWhere = theArgs[0]
// 		aX = theArgs[1]
// 		aY = theArgs[2]
// 		aValue = theArgs[3]
// 		aStep = theArgs[4] if len(theArgs) >= 5 else 1
// 		aFormat = theArgs[5] if len(theArgs) >= 6 else ""

// 		__internal.screen.where = aWhere
// 		aBlock = __internal.screen
// 	else:
// 		// Signature: counter(theValue, theStep = 1, theFormat = "%d")
// 		aBlock = __internal.topBlock()
// 		aX = aBlock.anchor.x
// 		aY = aBlock.anchor.y
// 		aValue = theArgs[0]
// 		aStep = theArgs[1] if len(theArgs) >= 2 else 1
// 		aFormat = theArgs[2] if len(theArgs) >= 3 else ""

// 	if not aFormat:
// 		aIsInt = isinstance(aValue[0], int) == True and isinstance(aStep, int)
// 		aFormat = "%d" if aIsInt else "%.1f"

// 	return __internal.counter(aBlock, aX, aY, aValue, aStep, aFormat)

func Checkbox(theArgs ...interface{}) bool {
	var aBlock *Block
	var aX, aY int
	var aLabel string
	var aState []bool
	aColor := uint32(0xCECECE)
	if _, ok := theArgs[0].(gocv.Mat); ok {
		// Signature: checkbox(theWhere, theX, theY, theLabel, theState, theColor = 0xCECECE)
		aWhere := theArgs[0].(gocv.Mat)
		aX = theArgs[1].(int)
		aY = theArgs[2].(int)
		aLabel = theArgs[3].(string)
		aState = theArgs[4].([]bool)
		if len(theArgs) >= 6 {
			aColor = theArgs[5].(uint32)
		}
		__internal.Screen.Where = aWhere
		aBlock = &__internal.Screen
	} else {
		// Signature: checkbox(theLabel, theState, theColor = 0xCECECE)
		aBlock = __internal.TopBlock()
		aX = aBlock.Anchor.X
		aY = aBlock.Anchor.Y
		aLabel = theArgs[0].(string)
		aState = theArgs[1].([]bool)
		if len(theArgs) >= 3 {
			aColor = theArgs[2].(uint32)
		}
	}

	return __internal.Checkbox(aBlock, aX, aY, aLabel, aState, aColor)
}

// def checkbox(*theArgs):
// 	if isinstance(theArgs[0], np.ndarray):
// 		// Signature: checkbox(theWhere, theX, theY, theLabel, theState, theColor = 0xCECECE)
// 		aWhere = theArgs[0]
// 		aX = theArgs[1]
// 		aY = theArgs[2]
// 		aLabel = theArgs[3]
// 		aState = theArgs[4]
// 		aColor = theArgs[5] if len(theArgs) >= 6 else 0xCECECE

// 		__internal.screen.where = aWhere
// 		aBlock = __internal.screen
// 	else:
// 		// Signature: checkbox(theLabel, theState, theColor = 0xCECECE)
// 		aBlock = __internal.topBlock()
// 		aX = aBlock.anchor.x
// 		aY = aBlock.anchor.y
// 		aLabel = theArgs[0]
// 		aState = theArgs[1]
// 		aColor = theArgs[2] if len(theArgs) >= 3 else 0xCECECE

// 	return __internal.checkbox(aBlock, aX, aY, aLabel, aState, aColor)

// def mouse(*theArgs):
// 	if len(theArgs) == 3:
// 		// Signature: mouse(theWindowName, theButton, theQuery)
// 		aWindowName = theArgs[0]
// 		aButton = theArgs[1]
// 		aQuery = theArgs[2]
// 		return __internal.mouseWBQ(aWindowName, aButton, aQuery)
// 	elif len(theArgs) == 2:
// 		// Signatures: mouse(theWindowName, theQuery) or mouse(theButton, theQuery)
// 		if __internal.isString(theArgs[0]):
// 			// Signature: mouse(theWindowName, theQuery)
// 			aWindowName = theArgs[0]
// 			aQuery = theArgs[1]
// 			return __internal.mouseWQ(aWindowName, aQuery)
// 		else:
// 			// Signature: mouse(theButton, theQuery)
// 			aButton = theArgs[0]
// 			aQuery = theArgs[1]
// 			return __internal.mouseBQ(aButton, aQuery)
// 	elif len(theArgs) == 1 and isinstance(theArgs[0], int):
// 		// Signature: mouse(theQuery)
// 		aQuery = theArgs[0]
// 		return __internal.mouseQ(aQuery)
// 	else:
// 		// Signature: mouse(theWindowName = "")
// 		aWindowName = theArgs[0] if len(theArgs) == 1 else ""
// 		return __internal.mouseW(aWindowName)

func Button(theArgs ...interface{}) (bool, error) {
	var aWhere gocv.Mat
	var aX, aY int
	var aBlock *Block
	var aArgs []interface{}
	var aLabel string

	_, ok1 := theArgs[0].(gocv.Mat)
	_, ok2 := theArgs[1].(gocv.Mat)
	if ok1 && !ok2 {
		// Signature: button(Mat, theX, theY, ...)
		aWhere = theArgs[0].(gocv.Mat)
		aX = theArgs[1].(int)
		aY = theArgs[2].(int)

		__internal.Screen.Where = aWhere
		aBlock = &__internal.Screen

		aArgs = theArgs[3:]
	} else {
		// Row/column function
		aBlock = __internal.TopBlock()
		aX = aBlock.Anchor.X
		aY = aBlock.Anchor.Y
		aArgs = theArgs
	}

	if len(aArgs) == 1 {
		// Signature: button(theLabel)
		aLabel = aArgs[0].(string)
		return __internal.Button(aBlock, aX, aY, aLabel), nil
	} else if len(aArgs) == 3 {
		if _, ok := aArgs[0].(int); ok {
			// Signature: button(theWidth, theHeight, theLabel)
			aWidth := aArgs[0].(int)
			aHeight := aArgs[1].(int)
			aLabel = aArgs[2].(string)
			return __internal.ButtonWH(aBlock, aX, aY, aWidth, aHeight, aLabel, true), nil
		} else {
			// Signature: button(theIdle, theOver, theDown)
			aIdle := aArgs[0].(gocv.Mat)
			aOver := aArgs[1].(gocv.Mat)
			aDown := aArgs[2].(gocv.Mat)
			return __internal.ButtonI(aBlock, aX, aY, &aIdle, &aOver, &aDown, true), nil
		}
	} else {
		// TODO: check this case here
		print("Problem?")
		return false, errors.New("args error")
	}
}

func Image(theArgs ...interface{}) {
	if _, ok := theArgs[0].(gocv.Mat); ok && len(theArgs) > 1 {

		// Signature: image(Mat, ...)
		aWhere := theArgs[0].(gocv.Mat)
		aX := theArgs[1].(int)
		aY := theArgs[2].(int)
		aImage := theArgs[3].(gocv.Mat)

		__internal.Screen.Where = aWhere
		__internal.Image(&__internal.Screen, aX, aY, &aImage)
	} else {
		// Row/column function, signature is image(...)
		aImage := theArgs[0].(gocv.Mat)
		aBlock := __internal.TopBlock()

		__internal.Image(aBlock, aBlock.Anchor.X, aBlock.Anchor.Y, &aImage)
	}
}

// def trackbar(*theArgs):
// 	// TODO: re-factor this two similar blocks by slicing theArgs into aArgs
// 	if isinstance(theArgs[0], np.ndarray):
// 		// Signature: trackbar(theWhere, theX, theY, theWidth, theValue, theMin, theMax, theSegments = 1, theLabelFormat = "%.1Lf", theOptions = 0, theDiscreteStep = 1)
// 		aWhere = theArgs[0]
// 		aX = theArgs[1]
// 		aY = theArgs[2]
// 		aWidth = theArgs[3]
// 		aValue = theArgs[4]
// 		aMin = theArgs[5]
// 		aMax = theArgs[6]
// 		aSegments = theArgs[7] if len(theArgs) >= 8 else 1
// 		aLabelFormat = theArgs[8] if len(theArgs) >= 9 else "%.1Lf"
// 		aOptions = theArgs[9] if len(theArgs) >= 10 else 0
// 		aDiscreteStep = theArgs[10] if len(theArgs) >= 11 else 1

// 		__internal.screen.where = aWhere
// 		aBlock = __internal.screen
// 	else:
// 		// Signature: trackbar(theWidth, theValue, theMin, theMax, theSegments = 1, theLabelFormat = "%.1Lf", theOptions = 0, theDiscreteStep = 1)
// 		aBlock = __internal.topBlock()
// 		aX = aBlock.anchor.x
// 		aY = aBlock.anchor.y
// 		aWidth = theArgs[0]
// 		aValue = theArgs[1]
// 		aMin = theArgs[2]
// 		aMax = theArgs[3]
// 		aSegments = theArgs[4] if len(theArgs) >= 5 else 1
// 		aLabelFormat = theArgs[5] if len(theArgs) >= 6 else "%.1Lf"
// 		aOptions = theArgs[6] if len(theArgs) >= 7 else 0
// 		aDiscreteStep = theArgs[7] if len(theArgs) >= 8 else 1

// 	// TODO: adjust aLabelFormat based on type of aValue
// 	aParams = TrackbarParams(aMin, aMax, aDiscreteStep, aSegments, aLabelFormat, aOptions)
// 	aResult = __internal.trackbar(aBlock, aX, aY, aWidth, aValue, aParams)

// 	return aResult

// def window(*theArgs):
// 	if isinstance(theArgs[0], np.ndarray):
// 		// Signature: window(theWhere, theX, theY, theWidth, theHeight, theTitle)
// 		aWhere = theArgs[0]
// 		aX = theArgs[1]
// 		aY = theArgs[2]
// 		aWidth = theArgs[3]
// 		aHeight = theArgs[4]
// 		aTitle = theArgs[5]

// 		__internal.screen.where = aWhere
// 		__internal.window(__internal.screen, aX, aY, aWidth, aHeight, aTitle)
// 	else:
// 		// Row/column function, signature: window(theWidth, theHeight, theTitle)
// 		aWidth = theArgs[0]
// 		aHeight = theArgs[1]
// 		aTitle = theArgs[2]

// 		aBlock = __internal.topBlock()
// 		__internal.window(aBlock, aBlock.anchor.x, aBlock.anchor.y, aWidth, aHeight, aTitle)

// def rect(*theArgs):
// 	if isinstance(theArgs[0], np.ndarray):
// 		// Signature: rect(theWhere, theX, theY, theWidth, theHeight, theBorderColor, theFillingColor = 0xff000000)
// 		aWhere = theArgs[0]
// 		aX = theArgs[1]
// 		aY = theArgs[2]
// 		aWidth = theArgs[3]
// 		aHeight = theArgs[4]
// 		aBorderColor = theArgs[5]
// 		aFillingColor = theArgs[6] if len(theArgs) >= 7 else 0xff000000

// 		__internal.screen.where = aWhere
// 		aBlock = __internal.screen
// 	else:
// 		// Signature: rect(theWidth, theHeight, theBorderColor, theFillingColor = 0xff000000)
// 		aBlock = __internal.topBlock()
// 		aX = aBlock.anchor.x
// 		aY = aBlock.anchor.y
// 		aWidth = theArgs[0]
// 		aHeight = theArgs[1]
// 		aBorderColor = theArgs[2]
// 		aFillingColor = theArgs[3] if len(theArgs) >= 4 else 0xff000000

// 	__internal.rect(aBlock, aX, aY, aWidth, aHeight, aBorderColor, aFillingColor)

func Sparkline(theArgs ...interface{}) {
	var aBlock *Block
	var aValues []float64
	var aX, aY, aWidth, aHeight int
	aColor := uint32(0x00FF00)

	if _, ok := theArgs[0].(gocv.Mat); ok {
		// Signature: sparkline(theWhere, theValues, theX, theY, theWidth, theHeight, theColor = 0x00FF00)
		aWhere := theArgs[0].(gocv.Mat)
		aValues = theArgs[1].([]float64)
		aX = theArgs[2].(int)
		aY = theArgs[3].(int)
		aWidth = theArgs[4].(int)
		aHeight = theArgs[5].(int)

		if len(theArgs) >= 7 {
			aColor = theArgs[6].(uint32)
		}
		__internal.Screen.Where = aWhere
		aBlock = &__internal.Screen
	} else {
		// Signature: sparkline(theValues, theWidth, theHeight, theColor = 0x00FF00)
		aBlock := __internal.TopBlock()
		aValues = theArgs[0].([]float64)
		aX = aBlock.Anchor.X
		aY = aBlock.Anchor.Y
		aWidth = theArgs[1].(int)
		aHeight = theArgs[2].(int)

		if len(theArgs) >= 4 {
			aColor = theArgs[3].(uint32)
		}
	}

	__internal.Sparkline(aBlock, aValues, aX, aY, aWidth, aHeight, aColor)
}

func BeginRow(theArgs ...interface{}) {
	if _, ok := theArgs[0].(gocv.Mat); ok && len(theArgs) > 0 {
		// Signature: beginRow(theWhere, theX, theY, theWidth = -1, theHeight = -1, thePadding = 0):
		aWhere := theArgs[0].(gocv.Mat)
		aX := theArgs[1].(int)
		aY := theArgs[2].(int)

		aWidth := -1
		if len(theArgs) >= 4 {
			aWidth = theArgs[3].(int)
		}

		aHeight := -1
		if len(theArgs) >= 5 {
			aHeight = theArgs[4].(int)
		}

		aPadding := 0
		if len(theArgs) >= 6 {
			aPadding = theArgs[5].(int)
		}

		__internal.Begin(ROW, &aWhere, aX, aY, aWidth, aHeight, aPadding)
	} else {
		// Signature: beginRow(theWidth = -1, theHeight = -1, thePadding = 0)
		aWidth := -1
		if len(theArgs) >= 1 {
			aWidth = theArgs[0].(int)
		}

		aHeight := -1
		if len(theArgs) >= 2 {
			aHeight = theArgs[1].(int)
		}

		aPadding := 0
		if len(theArgs) >= 3 {
			aPadding = theArgs[2].(int)
		}

		aBlock := __internal.TopBlock()
		__internal.Begin(ROW, &aBlock.Where, aBlock.Anchor.X, aBlock.Anchor.Y, aWidth, aHeight, aPadding)
	}
}

func BeginColumn(theArgs ...interface{}) {

	if _, ok := theArgs[0].(gocv.Mat); ok && len(theArgs) > 0 {
		// Signature: beginColumn(theWhere, theX, theY, theWidth = -1, theHeight = -1, thePadding = 0):
		aWhere := theArgs[0].(gocv.Mat)
		aX := theArgs[1].(int)
		aY := theArgs[2].(int)
		aWidth := -1
		if len(theArgs) >= 4 {
			aWidth = theArgs[3].(int)
		}

		aHeight := -1
		if len(theArgs) >= 5 {
			aHeight = theArgs[4].(int)
		}

		aPadding := 0
		if len(theArgs) >= 6 {
			aPadding = theArgs[5].(int)
		}

		__internal.Begin(COLUMN, &aWhere, aX, aY, aWidth, aHeight, aPadding)
	} else {
		// Signature: beginColumn(theWidth = -1, theHeight = -1, thePadding = 0):
		aWidth := -1
		if len(theArgs) >= 1 {
			aWidth = theArgs[0].(int)
		}

		aHeight := -1
		if len(theArgs) >= 2 {
			aHeight = theArgs[1].(int)
		}

		aPadding := 0
		if len(theArgs) >= 3 {
			aPadding = theArgs[2].(int)
		}

		aBlock := __internal.TopBlock()
		__internal.Begin(COLUMN, &aBlock.Where, aBlock.Anchor.X, aBlock.Anchor.Y, aWidth, aHeight, aPadding)
	}
}
