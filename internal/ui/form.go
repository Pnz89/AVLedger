package ui

import (
	"strings"
	"time"

	"avledger/internal/database"
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
func showEntryForm(parent fyne.Window, db *database.DB, existing models.LogEntry, onSave func(models.LogEntry)) {
	isNew := existing.ID == 0

	title := "New Task"
	if !isNew {
		title = "Edit Entry"
	}

	// ---- Fields ----
	dateEntry := widget.NewEntry()
	dateEntry.SetPlaceHolder("DD/MM/YYYY")
	if isNew && existing.Date == "" {
		dateEntry.SetText(time.Now().Format("02/01/2006"))
	} else {
		dateEntry.SetText(existing.Date)
	}

	aircraftEntry := widget.NewEntry()
	aircraftEntry.SetPlaceHolder("e.g. B737 NG (CFM56)")
	aircraftEntry.SetText(existing.AircraftEngineType)

	regEntry := widget.NewEntry()
	regEntry.SetPlaceHolder("e.g. I-DEMF")
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

	durationEntry := widget.NewEntry()
	durationEntry.SetPlaceHolder("e.g. 2.5")
	durationEntry.SetText(existing.Duration)

	assessors, _ := db.ListAssessors()
	var assessorNames []string
	for _, a := range assessors {
		assessorNames = append(assessorNames, a.Name)
	}

	verifiedEntry := widget.NewSelectEntry(assessorNames)
	verifiedEntry.SetPlaceHolder("Select or type Assessor name")
	verifiedEntry.SetText(existing.VerifiedBy)

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
		widget.NewFormItem("Duration (hrs)", durationEntry),
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

	// ---- Window ----
	var w fyne.Window

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
				w,
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
			Duration:           strings.TrimSpace(durationEntry.Text),
			VerifiedBy:         strings.TrimSpace(verifiedEntry.Text),
		}
		w.Close()
		onSave(entry)
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		w.Close()
	})

	buttons := container.NewHBox(cancelBtn, saveBtn)
	fullContent := container.NewBorder(nil, buttons, nil, nil, scroll)

	w = fyne.CurrentApp().NewWindow(title)
	w.SetContent(container.NewPadded(fullContent))
	w.Resize(fyne.NewSize(580, 520))
	w.Show()
}
