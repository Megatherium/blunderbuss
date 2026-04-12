package ui

type GridMode string

const (
	Grid4x1 GridMode = "4x1"
	Grid2x2 GridMode = "2x2"
	Grid1x4 GridMode = "1x4"
)

// LayoutDimensions holds all calculated layout dimensions for the UI.
type LayoutDimensions struct {
	Width  int
	Height int

	TermWidth  int
	TermHeight int

	SidebarWidth int

	GridMode GridMode

	TWidth int
	HWidth int
	MWidth int
	AWidth int

	THeight int
	HHeight int
	MHeight int
	AHeight int

	InnerTHeight int
	InnerHHeight int
	InnerMHeight int
	InnerAHeight int
}

const (
	marginWithSidebar    = 8
	marginWithoutSidebar = 6
	columnCount          = 4
	harnessWidthFactor   = 2
	minAgentWidth        = 10
	minZoomedColumnWidth = 12 // Minimum width for H/M/A columns in zoom mode (for labels)
	borderWidth          = 2
)

// Compute calculates all layout dimensions based on terminal size, sidebar visibility,
// and zoom mode. When zoom is enabled, the ticket column expands by shrinking H/M/A
// columns to minimum width.
func Compute(termW, termH int, showSidebar, zoom bool) LayoutDimensions {
	h, v := docStyle.GetFrameSize()

	width := termW - h
	height := termH - v - verticalMargins - footerHeight

	if width < minWindowWidth {
		width = minWindowWidth
	}
	if height < minWindowHeight {
		height = minWindowHeight
	}

	var gridMode GridMode
	if width > 110 {
		gridMode = Grid4x1
	} else if width >= 60 {
		gridMode = Grid2x2
	} else {
		gridMode = Grid1x4
	}

	if gridMode == Grid1x4 || gridMode == Grid2x2 {
		showSidebar = false
	}

	var usableWidth int
	if showSidebar {
		usableWidth = width - marginWithSidebar
	} else {
		usableWidth = width - marginWithoutSidebar
	}

	var sidebarWidth, tWidth, hWidth, mWidth, aWidth int
	if showSidebar {
		sidebarWidth = usableWidth / columnCount
		usableWidth -= sidebarWidth
	} else {
		sidebarWidth = 0
	}

	switch gridMode {
	case Grid4x1:
		baseX := usableWidth / columnCount
		if zoom {
			hWidth = minZoomedColumnWidth
			mWidth = minZoomedColumnWidth
			aWidth = minAgentWidth
			tWidth = usableWidth - (hWidth + mWidth + aWidth)
			if tWidth < baseX {
				tWidth = baseX
				hWidth = baseX / harnessWidthFactor
				if hWidth == 0 {
					hWidth = baseX
				}
				mWidth = baseX
				aWidth = usableWidth - (tWidth + hWidth + mWidth)
			}
		} else {
			tWidth = baseX
			hWidth = baseX / harnessWidthFactor
			if hWidth == 0 {
				hWidth = baseX
			}
			mWidth = baseX
			aWidth = usableWidth - (tWidth + hWidth + mWidth)
			if aWidth < minAgentWidth {
				aWidth = minAgentWidth
			}
		}
	case Grid2x2:
		usableWidth -= 2 // one gap between columns
		topHalf := usableWidth / 2
		if zoom {
			hWidth = minZoomedColumnWidth
			tWidth = usableWidth - hWidth
			mWidth = topHalf
			aWidth = usableWidth - topHalf
		} else {
			tWidth = topHalf
			hWidth = usableWidth - topHalf
			mWidth = topHalf
			aWidth = usableWidth - topHalf
		}
	case Grid1x4:
		tWidth = usableWidth
		hWidth = usableWidth
		mWidth = usableWidth
		aWidth = usableWidth
	}

	listHeight := height - filterHeight
	var tHeight, hHeight, mHeight, aHeight int

	switch gridMode {
	case Grid4x1:
		tHeight = listHeight
		hHeight = listHeight
		mHeight = listHeight
		aHeight = listHeight
	case Grid2x2:
		listHeight -= 1 // one gap between rows
		row1 := listHeight / 2
		row2 := listHeight - row1
		tHeight = row1
		hHeight = row1
		mHeight = row2
		aHeight = row2
	case Grid1x4:
		listHeight -= 3 // three gaps between rows
		row1 := listHeight / 4
		row2 := listHeight / 4
		row3 := listHeight / 4
		row4 := listHeight - (row1 + row2 + row3)
		tHeight = row1
		hHeight = row2
		mHeight = row3
		aHeight = row4
	}

	inner := func(h int) int {
		ih := h - borderHeight - 1
		if ih < 1 {
			return 1
		}
		return ih
	}

	return LayoutDimensions{
		Width:        width,
		Height:       height,
		TermWidth:    termW,
		TermHeight:   termH,
		SidebarWidth: sidebarWidth,
		GridMode:     gridMode,
		TWidth:       tWidth,
		HWidth:       hWidth,
		MWidth:       mWidth,
		AWidth:       aWidth,
		THeight:      tHeight,
		HHeight:      hHeight,
		MHeight:      mHeight,
		AHeight:      aHeight,
		InnerTHeight: inner(tHeight),
		InnerHHeight: inner(hHeight),
		InnerMHeight: inner(mHeight),
		InnerAHeight: inner(aHeight),
	}
}
