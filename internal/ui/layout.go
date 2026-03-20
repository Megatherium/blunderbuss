package ui

// LayoutDimensions holds all calculated layout dimensions for the UI.
type LayoutDimensions struct {
	Width  int
	Height int

	TermWidth  int
	TermHeight int

	SidebarWidth int
	TWidth       int
	HWidth       int
	MWidth       int
	AWidth       int

	InnerListHeight int
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

	var usableWidth int
	if showSidebar {
		usableWidth = width - marginWithSidebar
	} else {
		usableWidth = width - marginWithoutSidebar
	}

	baseX := usableWidth / columnCount

	var sidebarWidth, tWidth, hWidth, mWidth, aWidth int
	if showSidebar {
		sidebarWidth = usableWidth / columnCount
		if zoom {
			// Zoom mode: shrink H/M/A to minimum, give rest to ticket column
			hWidth = minZoomedColumnWidth
			mWidth = minZoomedColumnWidth
			aWidth = minAgentWidth
			tWidth = usableWidth - (sidebarWidth + hWidth + mWidth + aWidth)
			if tWidth < baseX {
				// Safety: if ticket column would be too small, fall back to normal
				tWidth = baseX
				hWidth = baseX / harnessWidthFactor
				mWidth = baseX
				aWidth = usableWidth - (sidebarWidth + tWidth + hWidth + mWidth)
			}
		} else {
			// Normal mode: proportional distribution
			tWidth = baseX
			hWidth = baseX / harnessWidthFactor
			mWidth = baseX
			aWidth = usableWidth - (sidebarWidth + tWidth + hWidth + mWidth)
			if aWidth < minAgentWidth {
				aWidth = minAgentWidth
			}
		}
	} else {
		sidebarWidth = 0
		if zoom {
			// Zoom mode without sidebar
			hWidth = minZoomedColumnWidth
			mWidth = minZoomedColumnWidth
			aWidth = minAgentWidth
			tWidth = usableWidth - (hWidth + mWidth + aWidth)
			if tWidth < baseX {
				// Safety: fall back to normal
				tWidth = baseX
				hWidth = baseX
				mWidth = baseX
				aWidth = usableWidth - (tWidth + hWidth + mWidth)
			}
		} else {
			// Normal mode without sidebar
			tWidth = baseX
			hWidth = baseX
			mWidth = baseX
			aWidth = usableWidth - (tWidth + hWidth + mWidth)
		}
	}

	listHeight := height - filterHeight
	innerListHeight := listHeight - borderHeight - 1
	if innerListHeight < 1 {
		innerListHeight = 1
	}

	return LayoutDimensions{
		Width:           width,
		Height:          height,
		TermWidth:       termW,
		TermHeight:      termH,
		SidebarWidth:    sidebarWidth,
		TWidth:          tWidth,
		HWidth:          hWidth,
		MWidth:          mWidth,
		AWidth:          aWidth,
		InnerListHeight: innerListHeight,
	}
}
