package ui

import (
	"avledger/internal/database"
	"avledger/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// showSettingsDialog opens the settings modal for configuring the logbook holder profile.
func showSettingsDialog(parent fyne.Window, db *database.DB, onSave func(models.Settings)) {
	// Load current values
	holderName, _ := db.GetSetting("holder_name")
	licenceNumber, _ := db.GetSetting("licence_number")

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Nome Cognome")
	nameEntry.SetText(holderName)

	licenceEntry := widget.NewEntry()
	licenceEntry.SetPlaceHolder("es. IE.145.006 297")
	licenceEntry.SetText(licenceNumber)

	form := widget.NewForm(
		widget.NewFormItem("Nome titolare logbook", nameEntry),
		widget.NewFormItem("N° Licenza / AML", licenceEntry),
	)

	note := widget.NewLabelWithStyle(
		"Questi dati verranno stampati nel footer del PDF esportato.",
		fyne.TextAlignLeading,
		fyne.TextStyle{Italic: true},
	)

	content := container.NewVBox(form, note)

	var d dialog.Dialog

	saveBtn := widget.NewButtonWithIcon("Salva impostazioni", theme.DocumentSaveIcon(), func() {
		_ = db.SetSetting("holder_name", nameEntry.Text)
		_ = db.SetSetting("licence_number", licenceEntry.Text)
		onSave(models.Settings{
			HolderName:    nameEntry.Text,
			LicenceNumber: licenceEntry.Text,
		})
		d.Hide()
		dialog.ShowInformation("Salvato", "Impostazioni salvate correttamente.", parent)
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButtonWithIcon("Annulla", theme.CancelIcon(), func() {
		d.Hide()
	})

	buttons := container.NewHBox(cancelBtn, saveBtn)
	fullContent := container.NewBorder(nil, buttons, nil, nil, content)

	d = dialog.NewCustom("Impostazioni", "✕", fullContent, parent)
	d.Resize(fyne.NewSize(420, 220))
	d.Show()
}
