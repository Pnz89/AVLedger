package ui

import (
	"fmt"

	"avledger/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

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
		boldLabel("#"),
		boldLabel("Data"),
		boldLabel("Aeromobile / Motore"),
		boldLabel("Reg"),
		boldLabel("Cat."),
		boldLabel("ATA"),
		boldLabel("WO N°"),
		boldLabel("Dettaglio Lavoro"),
		boldLabel("Verificato da"),
		boldLabel(""),
	)
	headerBg := container.NewPadded(header)

	// Data list
	list := widget.NewList(
		func() int { return len(*entries) },
		func() fyne.CanvasObject {
			row := container.New(
				newProportionalLayout(),
				widget.NewLabel(""),  // #
				widget.NewLabel(""),  // date
				widget.NewLabel(""),  // aircraft
				widget.NewLabel(""),  // reg
				widget.NewLabel(""),  // category
				widget.NewLabel(""),  // ata
				widget.NewLabel(""),  // wo
				widget.NewLabel(""),  // task
				widget.NewLabel(""),  // verified
				container.NewHBox(   // actions
					widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil),
					widget.NewButtonWithIcon("", theme.DeleteIcon(), nil),
				),
			)
			return row
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			e := (*entries)[id]
			row := obj.(*fyne.Container)
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
			}
			labels[0].SetText(fmt.Sprintf("%d", e.ID))
			labels[1].SetText(e.Date)
			labels[2].SetText(e.AircraftEngineType)
			labels[3].SetText(e.RegMarks)
			labels[4].SetText(e.Category)
			labels[5].SetText(e.ATA)
			labels[6].SetText(e.WorkOrderNumber)
			labels[7].SetText(e.TaskDetail)
			labels[8].SetText(e.VerifiedBy)

			actions := row.Objects[9].(*fyne.Container)
			editBtn := actions.Objects[0].(*widget.Button)
			delBtn := actions.Objects[1].(*widget.Button)

			editBtn.OnTapped = func() { onEdit(e) }
			delBtn.Importance = widget.DangerImportance
			delBtn.OnTapped = func() {
				dialog.ShowConfirm(
					"Conferma eliminazione",
					"Eliminare questa entry?\nL'operazione non è reversibile.",
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

	content := container.NewBorder(headerBg, nil, nil, nil, list)
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
var colProportions = []float32{0.04, 0.07, 0.11, 0.06, 0.05, 0.04, 0.08, 0.32, 0.16, 0.07}

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

func boldLabel(text string) *widget.Label {
	lbl := widget.NewLabel(text)
	lbl.TextStyle = fyne.TextStyle{Bold: true}
	return lbl
}

// Ensure layout package is used
var _ = layout.NewGridLayout(1)
