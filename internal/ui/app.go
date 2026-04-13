package ui

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"avledger/internal/assets"
	"avledger/internal/database"
	"avledger/internal/models"
	"avledger/internal/pdf"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/data/binding"
)

// Run initialises and starts the AVLedger application.
func Run() {
	a := app.NewWithID("com.avledger.app")
	a.Settings().SetTheme(&CustomTheme{})
	a.SetIcon(assets.ResourceLogoPng)

	w := a.NewWindow("AVLedger — Maintenance Logbook")
	w.SetIcon(assets.ResourceLogoPng)
	w.Resize(fyne.NewSize(1280, 760))
	w.SetMaster()

	// ---- Open DB ----
	db, err := database.Open()
	if err != nil {
		dialog.ShowError(err, w)
		a.Quit()
		return
	}
	defer db.Close()

	// ---- Load initial data ----
	entries, err := db.ListEntries()
	if err != nil {
		entries = []models.LogEntry{}
	}

	settings := loadSettings(db)

	// ---- Mutable entry list shared with table ----
	entryList := entries

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search tasks...")

	acSelect := widget.NewSelect([]string{"(All)"}, nil)
	regSelect := widget.NewSelect([]string{"(All)"}, nil)
	catSelect := widget.NewSelect([]string{"(All)"}, nil)
	jobSelect := widget.NewSelect([]string{"(All)"}, nil)

	acSelect.SetSelected("(All)")
	regSelect.SetSelected("(All)")
	catSelect.SetSelected("(All)")
	jobSelect.SetSelected("(All)")

	updateSelectOptions := func() {
		if opts, _ := db.GetDistinctValues("aircraft_engine_type"); opts != nil {
			acSelect.Options = append([]string{"(All)"}, opts...)
		}
		if opts, _ := db.GetDistinctValues("reg_marks"); opts != nil {
			regSelect.Options = append([]string{"(All)"}, opts...)
		}
		if opts, _ := db.GetDistinctValues("category"); opts != nil {
			catSelect.Options = append([]string{"(All)"}, opts...)
		}
		if opts, _ := db.GetDistinctValues("job_type"); opts != nil {
			jobSelect.Options = append([]string{"(All)"}, opts...)
		}
	}
	updateSelectOptions()

	getFilterOptions := func() models.FilterOptions {
		ac := acSelect.Selected
		if ac == "(All)" {
			ac = ""
		}
		reg := regSelect.Selected
		if reg == "(All)" {
			reg = ""
		}
		cat := catSelect.Selected
		if cat == "(All)" {
			cat = ""
		}
		job := jobSelect.Selected
		if job == "(All)" {
			job = ""
		}
		return models.FilterOptions{
			SearchQuery:        searchEntry.Text,
			AircraftEngineType: ac,
			RegMarks:           reg,
			Category:           cat,
			JobType:            job,
		}
	}

	// ---- Build table ----
	// et declared before callbacks so closures can reference it
	var et *entryTable
	var tableContent fyne.CanvasObject

	et, tableContent = buildTable(&entryList, w,
		// onEdit
		func(e models.LogEntry) {
			showEntryForm(w, e, func(updated models.LogEntry) {
				if err := db.UpdateEntry(updated); err != nil {
					dialog.ShowError(err, w)
					return
				}
				updateSelectOptions()
				reloadEntries(db, &entryList, w, getFilterOptions())
				et.Refresh()
			})
		},
		// onDelete
		func(id int64) {
			if err := db.DeleteEntry(id); err != nil {
				dialog.ShowError(err, w)
				return
			}
			updateSelectOptions()
			reloadEntries(db, &entryList, w, getFilterOptions())
			et.Refresh()
		},
	)

	// ---- Count label & Time (status bar) ----
	countText := widget.NewLabel("")

	clockStr := binding.NewString()
	clockText := widget.NewLabelWithData(clockStr)
	clockText.TextStyle = fyne.TextStyle{Monospace: true, Italic: true}

	go func() {
		for {
			now := time.Now().UTC()
			datePart := strings.ToUpper(now.Format("02 Jan 2006"))
			timePart := now.Format("15:04:05 UTC")
			clockStr.Set(datePart + " " + timePart)
			time.Sleep(time.Second)
		}
	}()

	updateCount := func() {
		n := len(entryList)
		countText.SetText(fmt.Sprintf("%d entr%s in the logbook", n, pluralIt(n)))
	}
	updateCount()

	refreshAll := func() {
		reloadEntries(db, &entryList, w, getFilterOptions())
		et.Refresh()
		updateCount()
	}

	searchEntry.OnChanged = func(s string) { refreshAll() }
	acSelect.OnChanged = func(s string) { refreshAll() }
	regSelect.OnChanged = func(s string) { refreshAll() }
	catSelect.OnChanged = func(s string) { refreshAll() }
	jobSelect.OnChanged = func(s string) { refreshAll() }

	// ---- Toolbar buttons ----
	newBtn := widget.NewButtonWithIcon("  Task", theme.ContentAddIcon(), func() {
		showEntryForm(w, models.LogEntry{}, func(e models.LogEntry) {
			if _, err := db.CreateEntry(e); err != nil {
				dialog.ShowError(err, w)
				return
			}
			updateSelectOptions()
			refreshAll()
		})
	})
	newBtn.Importance = widget.HighImportance

	exportBtn := widget.NewButtonWithIcon("  Export PDF", theme.DocumentPrintIcon(), func() {
		exportToPDF(w, db, &entryList, settings)
	})

	settingsBtn := widget.NewButtonWithIcon("  Settings", theme.SettingsIcon(), func() {
		showSettingsDialog(w, db, func(s models.Settings) {
			settings = s
		})
	})

	// ---- DB path bar ----
	dbIcon := widget.NewIcon(theme.StorageIcon())
	dbPathLabel := widget.NewLabel(db.Path)
	dbPathLabel.TextStyle = fyne.TextStyle{Monospace: true}
	dbPathLabel.Truncation = fyne.TextTruncateEllipsis

	openFolderBtn := widget.NewButtonWithIcon("Open folder", theme.FolderOpenIcon(), func() {
		openFolder(filepath.Dir(db.Path))
	})

	dbBar := container.NewBorder(nil, nil,
		container.NewHBox(dbIcon, widget.NewLabel("Database:")),
		openFolderBtn,
		dbPathLabel,
	)

	// ---- Header / title ----
	titleText := canvas.NewText("AVLedger", theme.PrimaryColor())
	titleText.TextSize = 22
	titleText.TextStyle = fyne.TextStyle{Bold: true}

	subtitleText := canvas.NewText("Aircraft Maintenance Logbook", nil)
	subtitleText.TextSize = 11
	subtitleText.TextStyle = fyne.TextStyle{Italic: true}

	titleCol := container.NewVBox(titleText, subtitleText)

	logoImg := canvas.NewImageFromResource(assets.ResourceLogoPng)
	logoImg.FillMode = canvas.ImageFillContain
	logoImg.SetMinSize(fyne.NewSize(40, 40))

	titleRow := container.NewHBox(logoImg, titleCol)

	toolbar := container.NewHBox(
		titleRow,
		widget.NewSeparator(),
		newBtn,
		exportBtn,
		widget.NewSeparator(),
		settingsBtn,
	)

	filtersRow := container.NewGridWithColumns(4,
		container.NewBorder(nil, nil, widget.NewLabel("Aircraft:"), nil, acSelect),
		container.NewBorder(nil, nil, widget.NewLabel("Reg:"), nil, regSelect),
		container.NewBorder(nil, nil, widget.NewLabel("Cat:"), nil, catSelect),
		container.NewBorder(nil, nil, widget.NewLabel("Job:"), nil, jobSelect),
	)

	searchRow := container.NewBorder(nil, nil, container.NewHBox(widget.NewIcon(theme.SearchIcon()), widget.NewLabel("Search:")), nil, searchEntry)

	searchContainer := container.NewVBox(
		searchRow,
		filtersRow,
	)

	// ---- Assemble layout ----
	toolsCard := widget.NewCard("", "", container.NewVBox(
		container.NewPadded(toolbar),
		widget.NewSeparator(),
		container.NewPadded(searchContainer),
	))
	dbCard := widget.NewCard("", "", container.NewPadded(dbBar))

	topArea := container.NewPadded(container.NewVBox(toolsCard, dbCard))

	bottomBar := container.NewPadded(
		container.NewBorder(nil, nil, countText, clockText),
	)

	content := container.NewBorder(
		topArea,
		bottomBar,
		nil, nil,
		container.NewPadded(tableContent),
	)

	w.SetContent(content)
	w.ShowAndRun()
}

// ---- Helpers ----

func reloadEntries(db *database.DB, list *[]models.LogEntry, w fyne.Window, opts models.FilterOptions) {
	var updated []models.LogEntry
	var err error
	if opts.SearchQuery == "" && opts.AircraftEngineType == "" && opts.RegMarks == "" && opts.Category == "" && opts.JobType == "" {
		updated, err = db.ListEntries()
	} else {
		updated, err = db.SearchEntries(opts)
	}
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	*list = updated
}

func loadSettings(db *database.DB) models.Settings {
	name, _ := db.GetSetting("holder_name")
	lic, _ := db.GetSetting("licence_number")
	return models.Settings{HolderName: name, LicenceNumber: lic}
}

func exportToPDF(w fyne.Window, db *database.DB, entries *[]models.LogEntry, s models.Settings) {
	if len(*entries) == 0 {
		dialog.ShowInformation("No data",
			"The logbook is empty. Add at least one entry before exporting.", w)
		return
	}

	saveDialog := dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
		if err != nil || uc == nil {
			return
		}
		uc.Close() // fpdf writes directly to the file path

		path := uc.URI().Path()
		if err := pdf.Export(path, *entries, s); err != nil {
			dialog.ShowError(err, w)
			return
		}
		dialog.ShowInformation("Export complete",
			fmt.Sprintf("PDF saved in:\n%s", path), w)
		openFile(path)
	}, w)

	saveDialog.SetFileName(fmt.Sprintf("AVLedger_%s.pdf", time.Now().Format("2006-01-02")))
	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".pdf"}))
	saveDialog.Show()
}

func openFolder(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("explorer", path)
	}
	if cmd != nil {
		_ = cmd.Start()
	}
}

func openFile(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("start", path)
	}
	if cmd != nil {
		_ = cmd.Start()
	}
}

func pluralIt(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}
