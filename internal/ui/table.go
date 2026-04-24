package ui

import (
	"image/color"
	"strings"

	"avledger/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// resizeWatcher is an invisible widget that calls onResize whenever its
// Resize method is invoked by the parent layout. This is used to detect
// window-resize events and force the list to re-layout its rows.
type resizeWatcher struct {
	widget.BaseWidget
	onResize func()
}

func newResizeWatcher(onResize func()) *resizeWatcher {
	rw := &resizeWatcher{onResize: onResize}
	rw.ExtendBaseWidget(rw)
	return rw
}

func (rw *resizeWatcher) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(newInvisibleRect())
}

func (rw *resizeWatcher) Resize(size fyne.Size) {
	rw.BaseWidget.Resize(size)
	if rw.onResize != nil {
		rw.onResize()
	}
}

// invisibleRect is a zero-opacity canvas rectangle used as a placeholder.
type invisibleRect struct{ widget.BaseWidget }

func newInvisibleRect() *invisibleRect {
	r := &invisibleRect{}
	r.ExtendBaseWidget(r)
	return r
}
func (r *invisibleRect) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(&zeroObject{})
}

// zeroObject is a no-op CanvasObject with zero min size.
type zeroObject struct{}

func (z *zeroObject) Size() fyne.Size                  { return fyne.Size{} }
func (z *zeroObject) Resize(_ fyne.Size)                {}
func (z *zeroObject) Position() fyne.Position           { return fyne.Position{} }
func (z *zeroObject) Move(_ fyne.Position)              {}
func (z *zeroObject) MinSize() fyne.Size                { return fyne.Size{} }
func (z *zeroObject) Visible() bool                     { return false }
func (z *zeroObject) Show()                             {}
func (z *zeroObject) Hide()                             {}
func (z *zeroObject) Refresh()                          {}

// entryTable is a custom widget that displays log entries in a tabular layout.
type entryTable struct {
	entries  *[]models.LogEntry
	parent   fyne.Window
	onEdit   func(e models.LogEntry)
	onDelete func(id int64)
	list     *widget.List
}

// buildTable creates a scrollable list-based table for log entries.
func buildTable(
	entries *[]models.LogEntry,
	parent fyne.Window,
	onEdit func(e models.LogEntry),
	onDelete func(id int64),
) (*entryTable, fyne.CanvasObject) {

	et := &entryTable{
		entries:  entries,
		parent:   parent,
		onEdit:   onEdit,
		onDelete: onDelete,
	}

	// Header row
	header := container.New(
		newProportionalLayout(),
		boldTruncLabel("DATE"),
		boldTruncLabel("AIRCRAFT / ENGINE"),
		boldTruncLabel("REG"),
		boldTruncLabel(" CAT."),
		boldTruncLabel("JOB TYPE"),
		boldTruncLabel("ATA"),
		boldTruncLabel("WO N°"),
		boldTruncLabel("DURATION"),
		boldTruncLabel("TASK DETAIL"),
		boldTruncLabel("VERIFIED BY"),
	)
	headerBgRect := canvas.NewRectangle(theme.PrimaryColor())
	headerBg := container.NewStack(headerBgRect, container.NewPadded(header))

	// Data list
	list := widget.NewList(
		func() int { return len(*entries) },
		func() fyne.CanvasObject {
			rowBg := canvas.NewRectangle(color.Transparent)
			row := container.New(
				newProportionalLayout(),
				newTruncLabel(""),  // date
				newTruncLabel(""),  // aircraft
				newTruncLabel(""),  // reg
				newTruncLabel(""),  // category
				newTruncLabel(""),  // job type
				newTruncLabel(""),  // ata
				newTruncLabel(""),  // wo
				newTruncLabel(""),  // task
				newTruncLabel(""),  // duration
				newTruncLabel(""),  // verified
				container.NewHBox(  // actions
					widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil),
					widget.NewButtonWithIcon("", theme.DeleteIcon(), nil),
				),
			)
			return container.NewStack(rowBg, row)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			e := (*entries)[id]
			stack := obj.(*fyne.Container)
			rowBg := stack.Objects[0].(*canvas.Rectangle)
			row := stack.Objects[1].(*fyne.Container)

			if id%2 == 0 {
				rowBg.FillColor = theme.HoverColor()
			} else {
				rowBg.FillColor = color.Transparent
			}
			rowBg.Refresh()

			labels := []*widget.Label{
				row.Objects[0].(*widget.Label),
				row.Objects[1].(*widget.Label),
				row.Objects[2].(*widget.Label),
				row.Objects[3].(*widget.Label),
				row.Objects[4].(*widget.Label),
				row.Objects[5].(*widget.Label),
				row.Objects[6].(*widget.Label),
				row.Objects[7].(*widget.Label),
				row.Objects[8].(*widget.Label),
				row.Objects[9].(*widget.Label),
			}
			labels[0].SetText(e.Date)
			labels[1].SetText(e.AircraftEngineType)
			labels[2].SetText(e.RegMarks)
			labels[3].SetText(e.Category)
			labels[4].SetText(e.JobType)
			ataDisplay := e.ATA
			if idx := strings.Index(ataDisplay, " - "); idx != -1 {
				ataDisplay = strings.TrimSpace(ataDisplay[:idx])
			}
			labels[5].SetText(ataDisplay)
			labels[6].SetText(e.WorkOrderNumber)
			labels[7].SetText(e.Duration)
			labels[8].SetText(e.TaskDetail)
			labels[9].SetText(e.VerifiedBy)

			for _, lbl := range labels {
				lbl.Refresh()
			}

			actions := row.Objects[10].(*fyne.Container)
			editBtn := actions.Objects[0].(*widget.Button)
			delBtn := actions.Objects[1].(*widget.Button)

			editBtn.OnTapped = func() { onEdit(e) }
			delBtn.Importance = widget.DangerImportance
			delBtn.OnTapped = func() {
				dialog.ShowConfirm(
					"Confirm deletion",
					"Are you sure you want to delete this entry?\nThis operation is irreversible.",
					func(ok bool) {
						if ok {
							onDelete(e.ID)
						}
					},
					parent,
				)
			}
		},
	)

	et.list = list

	// resizeWatcher detects window resize and forces the list to re-layout
	// its rows so that truncated labels stay within their column boundaries.
	watcher := newResizeWatcher(func() {
		list.Refresh()
	})

	// Stack the watcher invisibly behind the list so it shares the same size.
	listWithWatcher := container.NewStack(watcher, list)

	content := container.NewBorder(headerBg, nil, nil, nil, listWithWatcher)
	return et, content
}

// Refresh reloads the list data.
func (et *entryTable) Refresh() {
	if et.list != nil {
		et.list.Refresh()
	}
}

// ---- Layout helpers ----

// proportions for each column (must sum to ~1.0)
var colProportions = []float32{0.07, 0.11, 0.06, 0.05, 0.06, 0.04, 0.08, 0.06, 0.24, 0.16, 0.07}

type proportionalLayout struct{}

func newProportionalLayout() fyne.Layout {
	return &proportionalLayout{}
}

func (p *proportionalLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	x := float32(0)
	for i, obj := range objects {
		if i >= len(colProportions) {
			break
		}
		w := size.Width * colProportions[i]
		obj.Resize(fyne.NewSize(w, size.Height))
		obj.Move(fyne.NewPos(x, 0))
		x += w
	}
}

func (p *proportionalLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(900, 30)
}

// newTruncLabel creates a label with ellipsis truncation enabled.
func newTruncLabel(text string) *widget.Label {
	lbl := widget.NewLabel(text)
	lbl.Truncation = fyne.TextTruncateEllipsis
	return lbl
}

// boldTruncLabel creates a bold label with ellipsis truncation enabled.
func boldTruncLabel(text string) *widget.Label {
	lbl := widget.NewLabel(text)
	lbl.TextStyle = fyne.TextStyle{Bold: true}
	lbl.Truncation = fyne.TextTruncateEllipsis
	return lbl
}

// boldLabel creates a bold label (kept for compatibility).
func boldLabel(text string) *widget.Label {
	lbl := widget.NewLabel(text)
	lbl.TextStyle = fyne.TextStyle{Bold: true}
	return lbl
}

// Ensure layout package is used
var _ = layout.NewGridLayout(1)
