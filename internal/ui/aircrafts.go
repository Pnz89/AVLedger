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

// showAircraftsDialog opens a window to manage the list of aircrafts.
func showAircraftsDialog(parent fyne.Window, db *database.DB) {
	var w fyne.Window
	var list *widget.List
	var currentAircrafts []models.Aircraft

	reload := func() {
		aircrafts, err := db.ListAircrafts()
		if err != nil {
			dialog.ShowError(err, parent)
			return
		}
		currentAircrafts = aircrafts
		if list != nil {
			list.Refresh()
		}
	}
	reload()

	list = widget.NewList(
		func() int { return len(currentAircrafts) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.FileApplicationIcon()),
				widget.NewLabel("Registration placeholder..."),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			box := o.(*fyne.Container)
			label := box.Objects[1].(*widget.Label)
			label.SetText(currentAircrafts[i].Registration + " - " + currentAircrafts[i].Aircraft)
		},
	)

	list.OnSelected = func(i widget.ListItemID) {
		selected := currentAircrafts[i]
		showAircraftForm(parent, db, selected, func() {
			reload()
			list.UnselectAll()
		})
	}

	addBtn := widget.NewButtonWithIcon("Add Aircraft", theme.ContentAddIcon(), func() {
		showAircraftForm(parent, db, models.Aircraft{}, func() {
			reload()
		})
	})
	addBtn.Importance = widget.HighImportance

	closeBtn := widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() {
		w.Close()
	})

	buttons := container.NewHBox(closeBtn, widget.NewSeparator(), addBtn)

	content := container.NewBorder(
		widget.NewLabelWithStyle("Select an aircraft to edit, or add a new one:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		buttons, nil, nil,
		list,
	)

	w = fyne.CurrentApp().NewWindow("Manage Aircrafts")
	w.SetContent(container.NewPadded(content))
	w.Resize(fyne.NewSize(450, 350))
	w.Show()
}

func showAircraftForm(parent fyne.Window, db *database.DB, existing models.Aircraft, onComplete func()) {
	isNew := existing.ID == 0
	title := "New Aircraft"
	if !isNew {
		title = "Edit Aircraft"
	}

	regEntry := widget.NewEntry()
	regEntry.SetText(existing.Registration)
	regEntry.SetPlaceHolder("e.g. EI-DAZ")

	acEntry := widget.NewEntry()
	acEntry.SetText(existing.Aircraft)
	acEntry.SetPlaceHolder("e.g. B737 NG")

	engineEntry := widget.NewEntry()
	engineEntry.SetText(existing.Engine)
	engineEntry.SetPlaceHolder("e.g. CFM56")

	form := widget.NewForm(
		widget.NewFormItem("Registration *", regEntry),
		widget.NewFormItem("Aircraft *", acEntry),
		widget.NewFormItem("Engine", engineEntry),
	)

	var w fyne.Window

	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		if regEntry.Text == "" || acEntry.Text == "" {
			dialog.ShowInformation("Missing fields", "Registration and Aircraft are required.", w)
			return
		}

		existing.Registration = regEntry.Text
		existing.Aircraft = acEntry.Text
		existing.Engine = engineEntry.Text

		if isNew {
			if _, err := db.CreateAircraft(existing); err != nil {
				dialog.ShowError(err, w)
				return
			}
		} else {
			if err := db.UpdateAircraft(existing); err != nil {
				dialog.ShowError(err, w)
				return
			}
		}
		w.Close()
		if onComplete != nil {
			onComplete()
		}
	})
	saveBtn.Importance = widget.HighImportance

	delBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		dialog.ShowConfirm("Confirm", "Are you sure you want to delete this aircraft?", func(yes bool) {
			if yes {
				if err := db.DeleteAircraft(existing.ID); err != nil {
					dialog.ShowError(err, w)
					return
				}
				w.Close()
				if onComplete != nil {
					onComplete()
				}
			}
		}, w)
	})
	if isNew {
		delBtn.Hide()
	}

	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		w.Close()
	})

	buttons := container.NewHBox(delBtn, widget.NewSeparator(), cancelBtn, saveBtn)
	content := container.NewBorder(nil, buttons, nil, nil, container.NewVBox(form))

	w = fyne.CurrentApp().NewWindow(title)
	w.SetContent(container.NewPadded(content))
	w.Resize(fyne.NewSize(450, 250))
	w.Show()
}
