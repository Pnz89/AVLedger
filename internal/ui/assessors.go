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

// showAssessorsDialog opens a window to manage the list of assessors.
func showAssessorsDialog(parent fyne.Window, db *database.DB) {
	var d dialog.Dialog
	var list *widget.List
	var currentAssessors []models.Assessor

	reload := func() {
		assessors, err := db.ListAssessors()
		if err != nil {
			dialog.ShowError(err, parent)
			return
		}
		currentAssessors = assessors
		if list != nil {
			list.Refresh()
		}
	}
	reload()

	list = widget.NewList(
		func() int { return len(currentAssessors) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.AccountIcon()),
				widget.NewLabel("Name placeholder..."),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			box := o.(*fyne.Container)
			label := box.Objects[1].(*widget.Label)
			label.SetText(currentAssessors[i].Name)
		},
	)

	list.OnSelected = func(i widget.ListItemID) {
		selected := currentAssessors[i]
		showAssessorForm(parent, db, selected, func() {
			reload()
			list.UnselectAll()
		})
	}

	addBtn := widget.NewButtonWithIcon("Add Assessor", theme.ContentAddIcon(), func() {
		showAssessorForm(parent, db, models.Assessor{}, func() {
			reload()
		})
	})
	addBtn.Importance = widget.HighImportance

	closeBtn := widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() {
		d.Hide()
	})

	buttons := container.NewHBox(closeBtn, widget.NewSeparator(), addBtn)
	
	content := container.NewBorder(
		widget.NewLabelWithStyle("Select an assessor to edit, or add a new one:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		buttons, nil, nil,
		list,
	)

	d = dialog.NewCustom("Manage Assessors", "✕", content, parent)
	d.Resize(fyne.NewSize(450, 350))
	d.Show()
}

func showAssessorForm(parent fyne.Window, db *database.DB, existing models.Assessor, onComplete func()) {
	isNew := existing.ID == 0
	title := "New Assessor"
	if !isNew {
		title = "Edit Assessor"
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetText(existing.Name)
	nameEntry.SetPlaceHolder("Name and surname")

	licenseEntry := widget.NewEntry()
	licenseEntry.SetText(existing.LicenseNumber)
	licenseEntry.SetPlaceHolder("License number")

	approvalEntry := widget.NewEntry()
	approvalEntry.SetText(existing.CompanyApproval)
	approvalEntry.SetPlaceHolder("Company approval")

	form := widget.NewForm(
		widget.NewFormItem("Name / Surname *", nameEntry),
		widget.NewFormItem("License Number", licenseEntry),
		widget.NewFormItem("Company Approval", approvalEntry),
	)

	var d dialog.Dialog

	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		if nameEntry.Text == "" {
			dialog.ShowInformation("Missing field", "Name is required.", parent)
			return
		}

		existing.Name = nameEntry.Text
		existing.LicenseNumber = licenseEntry.Text
		existing.CompanyApproval = approvalEntry.Text

		if isNew {
			if _, err := db.CreateAssessor(existing); err != nil {
				dialog.ShowError(err, parent)
				return
			}
		} else {
			if err := db.UpdateAssessor(existing); err != nil {
				dialog.ShowError(err, parent)
				return
			}
		}
		d.Hide()
		if onComplete != nil {
			onComplete()
		}
	})
	saveBtn.Importance = widget.HighImportance

	delBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		dialog.ShowConfirm("Confirm", "Are you sure you want to delete this assessor?", func(yes bool) {
			if yes {
				if err := db.DeleteAssessor(existing.ID); err != nil {
					dialog.ShowError(err, parent)
					return
				}
				d.Hide()
				if onComplete != nil {
					onComplete()
				}
			}
		}, parent)
	})
	if isNew {
		delBtn.Hide()
	}

	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		d.Hide()
	})

	buttons := container.NewHBox(delBtn, widget.NewSeparator(), cancelBtn, saveBtn)
	content := container.NewBorder(nil, buttons, nil, nil, container.NewVBox(form))

	d = dialog.NewCustom(title, "✕", content, parent)
	d.Resize(fyne.NewSize(450, 250))
	d.Show()
}
