package tea

// PointerShape is the name of a mouse pointer shape, set with
// [View.PointerShape].
//
// The constants below are the shape names from the CSS cursor property, which
// kitty, Ghostty, and foot accept. kitty doesn't support [PointerContextMenu],
// [PointerAllScroll], [PointerColResize], or [PointerRowResize]. Terminals
// ignore names they don't recognize, or fall back to their default shape.
//
// Because PointerShape is a string type, terminal-specific names can be used
// too. For example, xterm only accepts X11 cursor font names, such as
// "left_ptr" or "hand2":
//
//	v.PointerShape = "hand2"
type PointerShape string

// Pointer shapes, named after the CSS cursor property. See
// https://developer.mozilla.org/en-US/docs/Web/CSS/cursor for what each shape
// looks like.
const (
	// General.
	PointerDefault     PointerShape = "default"      // The platform's default pointer, usually an arrow.
	PointerContextMenu PointerShape = "context-menu" // A context menu is available.
	PointerHelp        PointerShape = "help"         // Help is available.
	PointerPointer     PointerShape = "pointer"      // A link or other clickable element, usually a hand.
	PointerProgress    PointerShape = "progress"     // Busy, but the user can still interact.
	PointerWait        PointerShape = "wait"         // Busy, and the user can't interact.

	// Selection.
	PointerCell         PointerShape = "cell"          // A cell or set of cells can be selected.
	PointerCrosshair    PointerShape = "crosshair"     // A precise position can be selected.
	PointerText         PointerShape = "text"          // Text can be selected, usually an I-beam.
	PointerVerticalText PointerShape = "vertical-text" // Vertical text can be selected.

	// Drag and drop.
	PointerAlias      PointerShape = "alias"       // An alias or shortcut will be created.
	PointerCopy       PointerShape = "copy"        // Something will be copied.
	PointerMove       PointerShape = "move"        // Something will be moved.
	PointerNoDrop     PointerShape = "no-drop"     // The item can't be dropped here.
	PointerNotAllowed PointerShape = "not-allowed" // The action isn't allowed.
	PointerGrab       PointerShape = "grab"        // Something can be grabbed and dragged.
	PointerGrabbing   PointerShape = "grabbing"    // Something is being dragged.

	// Scrolling and resizing.
	PointerAllScroll  PointerShape = "all-scroll"  // Something can be scrolled in any direction.
	PointerColResize  PointerShape = "col-resize"  // A column can be resized horizontally.
	PointerRowResize  PointerShape = "row-resize"  // A row can be resized vertically.
	PointerNResize    PointerShape = "n-resize"    // An edge can be moved north.
	PointerEResize    PointerShape = "e-resize"    // An edge can be moved east.
	PointerSResize    PointerShape = "s-resize"    // An edge can be moved south.
	PointerWResize    PointerShape = "w-resize"    // An edge can be moved west.
	PointerNEResize   PointerShape = "ne-resize"   // A corner can be moved north-east.
	PointerNWResize   PointerShape = "nw-resize"   // A corner can be moved north-west.
	PointerSEResize   PointerShape = "se-resize"   // A corner can be moved south-east.
	PointerSWResize   PointerShape = "sw-resize"   // A corner can be moved south-west.
	PointerEWResize   PointerShape = "ew-resize"   // Something can be resized east and west.
	PointerNSResize   PointerShape = "ns-resize"   // Something can be resized north and south.
	PointerNESWResize PointerShape = "nesw-resize" // Something can be resized north-east and south-west.
	PointerNWSEResize PointerShape = "nwse-resize" // Something can be resized north-west and south-east.

	// Zooming.
	PointerZoomIn  PointerShape = "zoom-in"  // Something can be zoomed in.
	PointerZoomOut PointerShape = "zoom-out" // Something can be zoomed out.
)
