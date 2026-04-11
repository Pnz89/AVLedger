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
var categories = []string{"A", "B1", "B2", "C"}

// showEntryForm opens a modal dialog to create or edit a LogEntry.
// onSave is called with the filled entry when the user confirms.
// If editing, pass the existing entry; for new entries pass an empty LogEntry.
func showEntryForm(parent fyne.Window, existing models.LogEntry, onSave func(models.LogEntry)) {
	isNew := existing.ID == 0

	title := "Nuovo Task"
	if !isNew {
		title = "Modifica Entry"
	}

	// ---- Fields ----
	dateEntry := widget.NewEntry()
	dateEntry.SetPlaceHolder("DD/MM/YYYY")
	dateEntry.SetText(existing.Date)

	aircraftEntry := widget.NewEntry()
	aircraftEntry.SetPlaceHolder("es. B737 NG (CFM56)")
	aircraftEntry.SetText(existing.AircraftEngineType)

	regEntry := widget.NewEntry()
	regEntry.SetPlaceHolder("es. EI-DAZ")
	regEntry.SetText(existing.RegMarks)

	taskEntry := widget.NewMultiLineEntry()
	taskEntry.SetPlaceHolder("Descrizione del lavoro eseguito…")
	taskEntry.SetText(existing.TaskDetail)
	taskEntry.SetMinRowsVisible(3)

	categorySelect := widget.NewSelect(categories, nil)
	if existing.Category != "" {
		categorySelect.SetSelected(existing.Category)
	} else {
		categorySelect.SetSelected("B1")
	}

	ataEntry := widget.NewEntry()
	ataEntry.SetPlaceHolder("es. 32")
	ataEntry.SetText(existing.ATA)

	woEntry := widget.NewEntry()
	woEntry.SetPlaceHolder("Numero Work Order")
	woEntry.SetText(existing.WorkOrderNumber)

	verifiedEntry := widget.NewMultiLineEntry()
	verifiedEntry.SetPlaceHolder("Nome + n° autorizzazione / AML")
	verifiedEntry.SetText(existing.VerifiedBy)
	verifiedEntry.SetMinRowsVisible(2)

	// ---- Layout ----
	form := widget.NewForm(
		widget.NewFormItem("Data *", dateEntry),
		widget.NewFormItem("Aeromobile / Motore *", aircraftEntry),
		widget.NewFormItem("Marche (Reg) *", regEntry),
		widget.NewFormItem("Dettaglio Lavoro *", taskEntry),
		widget.NewFormItem("Categoria", categorySelect),
		widget.NewFormItem("ATA", ataEntry),
		widget.NewFormItem("Work Order N°", woEntry),
		widget.NewFormItem("Verificato da", verifiedEntry),
	)

	requiredLabel := widget.NewLabelWithStyle(
		"* Campi obbligatori",
		fyne.TextAlignLeading,
		fyne.TextStyle{Italic: true},
	)

	content := container.NewVBox(form, requiredLabel)
	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(540, 420))

	// ---- Dialog ----
	var d dialog.Dialog

	saveBtn := widget.NewButtonWithIcon("Salva", theme.DocumentSaveIcon(), func() {
		// Validate required fields
		var missing []string
		if strings.TrimSpace(dateEntry.Text) == "" {
			missing = append(missing, "Data")
		}
		if strings.TrimSpace(aircraftEntry.Text) == "" {
			missing = append(missing, "Aeromobile / Motore")
		}
		if strings.TrimSpace(regEntry.Text) == "" {
			missing = append(missing, "Marche (Reg)")
		}
		if strings.TrimSpace(taskEntry.Text) == "" {
			missing = append(missing, "Dettaglio Lavoro")
		}
		if len(missing) > 0 {
			dialog.ShowInformation(
				"Campi mancanti",
				"Compila i seguenti campi obbligatori:\n• "+strings.Join(missing, "\n• "),
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
			ATA:                strings.TrimSpace(ataEntry.Text),
			WorkOrderNumber:    strings.TrimSpace(woEntry.Text),
			VerifiedBy:         strings.TrimSpace(verifiedEntry.Text),
		}
		d.Hide()
		onSave(entry)
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButtonWithIcon("Annulla", theme.CancelIcon(), func() {
		d.Hide()
	})

	buttons := container.NewHBox(cancelBtn, saveBtn)
	fullContent := container.NewBorder(nil, buttons, nil, nil, scroll)

	d = dialog.NewCustom(title, "✕", fullContent, parent)
	d.Resize(fyne.NewSize(580, 520))
	d.Show()
}
