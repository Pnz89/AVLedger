package ui

import (
	"strings"

	"avledger/internal/database"
	"avledger/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
)

// categories available for selection
var categories = []string{"A1", "A2", "A3", "A4", "B1.1", "B1.2", "B1.3", "B1.4", "B2", "B3", "C", "Mech"}

// ataChapters represents standard ATA 100 chapters
var ataChapters = []string{
	"05 - TIME LIMITS/MAINTENANCE CHECKS",
	"06 - DIMENSIONS AND AREAS",
	"07 - LIFTING AND SHORING",
	"08 - LEVELING AND WEIGHING",
	"09 - TOWING AND TAXIING",
	"10 - PARKING, MOORING, STORAGE AND RETURN TO SERVICE",
	"11 - PLACARDS AND MARKINGS",
	"12 - SERVICING - ROUTINE MAINTENANCE",
	"18 - VIBRATION AND NOISE ANALYSIS (HELICOPTER ONLY)",
	"20 - STANDARD PRACTICES - AIRFRAME",
	"21 - AIR CONDITIONING",
	"22 - AUTO FLIGHT",
	"23 - COMMUNICATIONS",
	"24 - ELECTRICAL POWER",
	"25 - EQUIPMENT/FURNISHINGS",
	"26 - FIRE PROTECTION",
	"27 - FLIGHT CONTROLS",
	"28 - FUEL",
	"29 - HYDRAULIC POWER",
	"30 - ICE AND RAIN PROTECTION",
	"31 - INDICATING/RECORDING SYSTEMS",
	"32 - LANDING GEAR",
	"33 - LIGHTS",
	"34 - NAVIGATION",
	"35 - OXYGEN",
	"36 - PNEUMATIC",
	"37 - VACUUM",
	"38 - WATER/WASTE",
	"39 - ELECTRICAL - ELECTRONIC PANELS AND MULTIPURPOSE COMPONENTS",
	"41 - WATER BALLAST",
	"42 - INTEGRATED MODULAR AVIONICS",
	"44 - CABIN SYSTEMS",
	"45 - DIAGNOSTIC AND MAINTENANCE SYSTEM",
	"46 - INFORMATION SYSTEMS",
	"47 - NITROGEN GENERATION SYSTEM",
	"49 - AIRBORNE AUXILIARY POWER",
	"51 - STANDARD PRACTICES AND STRUCTURES - GENERAL",
	"52 - DOORS",
	"53 - FUSELAGE",
	"54 - NACELLES/PYLONS",
	"55 - STABILIZERS",
	"56 - WINDOWS",
	"57 - WINGS",
	"60 - STANDARD PRACTICES - PROPELLER/ROTOR",
	"61 - PROPELLERS/PROPULSORS",
	"62 - MAIN ROTOR(S)",
	"63 - MAIN ROTOR DRIVE(S)",
	"64 - TAIL ROTOR",
	"65 - TAIL ROTOR DRIVE",
	"66 - ROTOR BLADE AND TAIL PYLON FOLDING",
	"67 - ROTORS FLIGHT CONTROL",
	"70 - STANDARD PRACTICES - ENGINE",
	"71 - POWER PLANT - GENERAL",
	"72 - ENGINE",
	"73 - ENGINE - FUEL AND CONTROL",
	"74 - IGNITION",
	"75 - BLEED AIR",
	"76 - ENGINE CONTROLS",
	"77 - ENGINE INDICATING",
	"78 - EXHAUST",
	"79 - OIL",
	"80 - STARTING",
	"81 - TURBINES (RECIPROCATING ENGINES)",
	"82 - WATER INJECTION",
	"83 - ACCESSORY GEAR BOXES",
	"84 - PROPULSION AUGMENTATION",
	"85 - FUEL CELL SYSTEMS",
	"91 - CHARTS",
	"92 - ELECTRICAL SYSTEM INSTALLATION",
	"95 - CREW INFORMATION",
}

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
	if existing.Date != "" {
		dateEntry.SetText(existing.Date)
	}

	aircraftEntry := widget.NewEntry()
	aircraftEntry.SetPlaceHolder("Please select a registration")
	aircraftEntry.SetText(existing.AircraftEngineType)
	aircraftEntry.Disable()

	aircrafts, _ := db.ListAircrafts()
	var aircraftRegs []string
	aircraftMap := make(map[string]models.Aircraft)
	for _, a := range aircrafts {
		aircraftRegs = append(aircraftRegs, a.Registration)
		aircraftMap[a.Registration] = a
	}

	regEntry := widget.NewSelect(aircraftRegs, func(s string) {
		if a, ok := aircraftMap[s]; ok {
			combined := a.Aircraft
			if a.Engine != "" {
				combined += " (" + a.Engine + ")"
			}
			aircraftEntry.SetText(combined)
		}
	})
	regEntry.PlaceHolder = "Select Registration..."
	if existing.RegMarks != "" {
		regEntry.SetSelected(existing.RegMarks)
	}

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

	ataEntry := xwidget.NewCompletionEntry(ataChapters)
	ataEntry.SetPlaceHolder("e.g. 32 - LANDING GEAR")
	ataEntry.SetText(existing.ATA)
	ataEntry.OnChanged = func(s string) {
		if s == "" {
			ataEntry.SetOptions(ataChapters)
			ataEntry.HideCompletion()
			return
		}
		s = strings.ToUpper(s)
		var filtered []string
		for _, opt := range ataChapters {
			if strings.Contains(strings.ToUpper(opt), s) {
				filtered = append(filtered, opt)
			}
		}
		ataEntry.SetOptions(filtered)
		if len(filtered) > 0 {
			ataEntry.ShowCompletion()
		} else {
			ataEntry.HideCompletion()
		}
	}

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

	verifiedEntry := widget.NewSelect(assessorNames, nil)
	verifiedEntry.PlaceHolder = "Select Assessor"
	if existing.VerifiedBy != "" {
		verifiedEntry.SetSelected(existing.VerifiedBy)
	}

	// ---- Layout ----
	form := widget.NewForm(
		widget.NewFormItem("Date *", dateEntry),
		widget.NewFormItem("Registration *", regEntry),
		widget.NewFormItem("Aircraft / Engine *", aircraftEntry),
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
		if regEntry.Selected == "" {
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
			RegMarks:           regEntry.Selected,
			TaskDetail:         strings.TrimSpace(taskEntry.Text),
			Category:           categorySelect.Selected,
			JobType:            strings.TrimSpace(jobTypeEntry.Text),
			ATA:                strings.TrimSpace(ataEntry.Text),
			WorkOrderNumber:    strings.TrimSpace(woEntry.Text),
			Duration:           strings.TrimSpace(durationEntry.Text),
			VerifiedBy:         verifiedEntry.Selected,
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
