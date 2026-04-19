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
	nameEntry.SetPlaceHolder("First and Last Name")
	nameEntry.SetText(holderName)

	licenceEntry := widget.NewEntry()
	licenceEntry.SetPlaceHolder("e.g. IT.66.XXXX")
	licenceEntry.SetText(licenceNumber)

	form := widget.NewForm(
		widget.NewFormItem("Logbook holder name", nameEntry),
		widget.NewFormItem("Licence N° / AML", licenceEntry),
	)

	note := widget.NewLabelWithStyle(
		"This information will be printed in the footer of the exported PDF.",
		fyne.TextAlignLeading,
		fyne.TextStyle{Italic: true},
	)

	content := container.NewVBox(form, note)

	var w fyne.Window

	saveBtn := widget.NewButtonWithIcon("Save settings", theme.DocumentSaveIcon(), func() {
		_ = db.SetSetting("holder_name", nameEntry.Text)
		_ = db.SetSetting("licence_number", licenceEntry.Text)
		onSave(models.Settings{
			HolderName:    nameEntry.Text,
			LicenceNumber: licenceEntry.Text,
		})
		w.Close()
		dialog.ShowInformation("Saved", "Settings saved successfully.", parent)
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		w.Close()
	})

	buttons := container.NewHBox(cancelBtn, saveBtn)
	fullContent := container.NewBorder(nil, buttons, nil, nil, content)

	w = fyne.CurrentApp().NewWindow("Settings")
	w.SetContent(container.NewPadded(fullContent))
	w.Resize(fyne.NewSize(420, 250))
	w.Show()
}
