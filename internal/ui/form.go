package ui

import (
	"strings"

	"avledger/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// categories available for selection
var categories = []string{"A1", "A2", "A3", "A4", "B1.1", "B1.2", "B1.3", "B1.4", "B2", "B3", "C", "Mech"}

// showEntryForm opens a modal dialog to create or edit a LogEntry.
// onSave is called with the filled entry when the user confirms.
// If editing, pass the existing entry; for new entries pass an empty LogEntry.
func showEntryForm(parent fyne.Window, existing models.LogEntry, onSave func(models.LogEntry)) {
	isNew := existing.ID == 0

	title := "New Task"
	if !isNew {
		title = "Edit Entry"
	}

	// ---- Fields ----
	dateEntry := widget.NewEntry()
	dateEntry.SetPlaceHolder("DD MMM YYYY")
	dateEntry.SetText(existing.Date)

	aircraftEntry := widget.NewEntry()
	aircraftEntry.SetPlaceHolder("e.g. B737 NG (CFM56)")
	aircraftEntry.SetText(existing.AircraftEngineType)

	regEntry := widget.NewEntry()
	regEntry.SetPlaceHolder("e.g. EI-DAZ")
	regEntry.SetText(existing.RegMarks)

	taskEntry := widget.NewMultiLineEntry()
	taskEntry.SetPlaceHolder("Description of work performed…")
	taskEntry.SetText(existing.TaskDetail)
	taskEntry.SetMinRowsVisible(3)

	categorySelect := widget.NewSelect(categories, nil)
	if existing.Category != "" {
		categorySelect.SetSelected(existing.Category)
	} else {
		categorySelect.SetSelected("B1.1")
	}

	jobTypeEntry := widget.NewEntry()
	jobTypeEntry.SetPlaceHolder("e.g. Line, Base, Mod")
	jobTypeEntry.SetText(existing.JobType)

	ataEntry := widget.NewEntry()
	ataEntry.SetPlaceHolder("e.g. 32")
	ataEntry.SetText(existing.ATA)

	woEntry := widget.NewEntry()
	woEntry.SetPlaceHolder("Work Order Number")
	woEntry.SetText(existing.WorkOrderNumber)

	verifiedEntry := widget.NewMultiLineEntry()
	verifiedEntry.SetPlaceHolder("Name + authorisation n° / AML")
	verifiedEntry.SetText(existing.VerifiedBy)
	verifiedEntry.SetMinRowsVisible(2)

	// ---- Layout ----
	form := widget.NewForm(
		widget.NewFormItem("Date *", dateEntry),
		widget.NewFormItem("Aircraft / Engine *", aircraftEntry),
		widget.NewFormItem("Registration *", regEntry),
		widget.NewFormItem("Task Detail *", taskEntry),
		widget.NewFormItem("Category", categorySelect),
		widget.NewFormItem("Job type", jobTypeEntry),
		widget.NewFormItem("ATA", ataEntry),
		widget.NewFormItem("Work Order N°", woEntry),
		widget.NewFormItem("Verified by", verifiedEntry),
	)

	requiredLabel := widget.NewLabelWithStyle(
		"* Required fields",
		fyne.TextAlignLeading,
		fyne.TextStyle{Italic: true},
	)

	content := container.NewVBox(form, requiredLabel)
	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(540, 420))

	// ---- Dialog ----
	var d dialog.Dialog

	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		// Validate required fields
		var missing []string
		if strings.TrimSpace(dateEntry.Text) == "" {
			missing = append(missing, "Date")
		}
		if strings.TrimSpace(aircraftEntry.Text) == "" {
			missing = append(missing, "Aircraft / Engine")
		}
		if strings.TrimSpace(regEntry.Text) == "" {
			missing = append(missing, "Registration")
		}
		if strings.TrimSpace(taskEntry.Text) == "" {
			missing = append(missing, "Task Detail")
		}
		if len(missing) > 0 {
			dialog.ShowInformation(
				"Missing fields",
				"Please fill in the following required fields:\n• "+strings.Join(missing, "\n• "),
				parent,
			)
			return
		}

		entry := models.LogEntry{
			ID:                 existing.ID,
			Date:               strings.TrimSpace(dateEntry.Text),
			AircraftEngineType: strings.TrimSpace(aircraftEntry.Text),
			RegMarks:           strings.TrimSpace(regEntry.Text),
			TaskDetail:         strings.TrimSpace(taskEntry.Text),
			Category:           categorySelect.Selected,
			JobType:            strings.TrimSpace(jobTypeEntry.Text),
			ATA:                strings.TrimSpace(ataEntry.Text),
			WorkOrderNumber:    strings.TrimSpace(woEntry.Text),
			VerifiedBy:         strings.TrimSpace(verifiedEntry.Text),
		}
		d.Hide()
		onSave(entry)
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		d.Hide()
	})

	buttons := container.NewHBox(cancelBtn, saveBtn)
	fullContent := container.NewBorder(nil, buttons, nil, nil, scroll)

	d = dialog.NewCustom(title, "✕", fullContent, parent)
	d.Resize(fyne.NewSize(580, 520))
	d.Show()
}
