package ui

import (
	"fmt"
	"os"
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
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Run initialises and starts the AVLedger application.
func Run() {
	a := app.NewWithID("com.avledger.app")
	customTheme := &CustomTheme{}
	a.Settings().SetTheme(customTheme)
	a.SetIcon(assets.ResourceLogoPng)

	w := a.NewWindow("AVLedger — Maintenance Logbook")
	w.SetIcon(assets.ResourceLogoPng)
	w.Resize(fyne.NewSize(1280, 760))
	w.SetMaster()

	// ---- Open DB ----
	customDBPath := a.Preferences().StringWithFallback("dbPath", "")
	isFirstRun := customDBPath == ""
	var autoDiscoveredCloudDB string

	if isFirstRun {
		cloudFolders := detectCloudSyncFolders()
		for _, f := range cloudFolders {
			checkPath := filepath.Join(f, "AVLedger", "avledger.db")
			if _, err := os.Stat(checkPath); err == nil {
				autoDiscoveredCloudDB = checkPath
				break
			}
		}

		if autoDiscoveredCloudDB != "" {
			a.Preferences().SetString("dbPath", autoDiscoveredCloudDB)
			customDBPath = autoDiscoveredCloudDB
		}
	}

	db, err := database.Open(customDBPath)
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
			showEntryForm(w, db, e, func(updated models.LogEntry) {
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
		showEntryForm(w, db, models.LogEntry{}, func(e models.LogEntry) {
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

	assessorsBtn := widget.NewButtonWithIcon("  Assessors", theme.AccountIcon(), func() {
		showAssessorsDialog(w, db)
	})

	// ---- Theme toggle button ----
	themeBtn := widget.NewButtonWithIcon("", theme.ColorPaletteIcon(), nil)
	themeBtn.OnTapped = func() {
		// Determine what variant we are currently displaying
		current := customTheme.lastVariant
		if customTheme.ForcedVariant != nil {
			current = *customTheme.ForcedVariant
		}

		// Toggle it
		if current == theme.VariantDark {
			light := theme.VariantLight
			customTheme.ForcedVariant = &light
		} else {
			dark := theme.VariantDark
			customTheme.ForcedVariant = &dark
		}
		a.Settings().SetTheme(customTheme)
	}

	// ---- DB path bar ----
	dbIcon := widget.NewIcon(theme.StorageIcon())
	dbPathLabel := widget.NewLabel(db.Path)
	dbPathLabel.TextStyle = fyne.TextStyle{Monospace: true}
	dbPathLabel.Truncation = fyne.TextTruncateEllipsis

	changeDBItem := fyne.NewMenuItem("Change DB", func() {
		fileDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
			if err != nil || uc == nil {
				return
			}
			uc.Close()

			newPath := uc.URI().Path()
			if err := db.SwitchTo(newPath); err != nil {
				dialog.ShowError(err, w)
				return
			}

			a.Preferences().SetString("dbPath", db.Path)
			dbPathLabel.SetText(db.Path)

			updateSelectOptions()
			refreshAll()

			dialog.ShowInformation("Database Switched", "Successfully loaded the selected database.", w)
		}, w)

		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".db", ".sqlite", ".sqlite3"}))
		fileDialog.Show()
	})
	changeDBItem.Icon = theme.DocumentCreateIcon()

	openFolderItem := fyne.NewMenuItem("Open Folder", func() {
		openFolder(filepath.Dir(db.Path))
	})
	openFolderItem.Icon = theme.FolderOpenIcon()

	manageMenu := fyne.NewMenu("Manage DB", changeDBItem, openFolderItem)

	var manageDBBtn *widget.Button
	manageDBBtn = widget.NewButtonWithIcon("Manage DB", theme.SettingsIcon(), func() {
		pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(manageDBBtn)
		pos.Y += manageDBBtn.Size().Height
		widget.ShowPopUpMenuAtPosition(manageMenu, w.Canvas(), pos)
	})

	rightActions := container.NewHBox(manageDBBtn)

	dbBar := container.NewBorder(nil, nil,
		container.NewHBox(dbIcon, widget.NewLabel("Database:")),
		rightActions,
		dbPathLabel,
	)

	// ---- Header / title ----
	titleText := canvas.NewText("AVLedger", theme.PrimaryColor())
	titleText.TextSize = 22
	titleText.TextStyle = fyne.TextStyle{Bold: true}

	versionText := canvas.NewText("0.5.3", theme.DisabledColor())
	versionText.TextSize = 12
	versionText.TextStyle = fyne.TextStyle{Bold: true}

	titleRowTop := container.NewHBox(
		titleText,
		container.NewCenter(versionText),
	)

	subtitleText := canvas.NewText("Aircraft Maintenance Logbook", nil)
	subtitleText.TextSize = 11
	subtitleText.TextStyle = fyne.TextStyle{Italic: true}

	titleCol := container.NewVBox(titleRowTop, subtitleText)

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
		assessorsBtn,
		settingsBtn,
		layout.NewSpacer(),
		themeBtn,
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

	// ---- Notification if auto-discovered ----
	if isFirstRun && autoDiscoveredCloudDB != "" {
		dialog.ShowInformation("Cloud Logbook Found",
			fmt.Sprintf("Welcome!\nWe automatically found and connected to your existing logbook in the cloud folder:\n\n%s", filepath.Dir(autoDiscoveredCloudDB)), w)
	}

	// ---- Cloud Backup Prompter ----
	if isFirstRun && autoDiscoveredCloudDB == "" && !a.Preferences().Bool("cloudPrompted") {
		cloudFolders := detectCloudSyncFolders()
		var validPaths []string
		for _, p := range cloudFolders {
			if !strings.HasPrefix(db.Path, p) {
				validPaths = append(validPaths, p)
			}
		}

		if len(validPaths) > 0 {
			options := make([]string, len(validPaths))
			for i, p := range validPaths {
				options[i] = filepath.Base(p)
			}
			options = append(options, "Local Storage")

			selectWidget := widget.NewSelect(options, nil)
			selectWidget.SetSelected(options[0])

			content := container.NewVBox(
				widget.NewLabel("Where would you like to save your AVLedger database?\nWe detected cloud sync folders that allow automatic backups."),
				selectWidget,
			)

			dialog.ShowCustomConfirm("Select Database Location", "Confirm", "Cancel", content, func(yes bool) {
				if yes {
					a.Preferences().SetBool("cloudPrompted", true)
					
					if selectWidget.Selected == "Local Storage" {
						dialog.ShowInformation("Local Storage", "Database will be kept locally.", w)
						return
					}

					var targetFolder string
					for _, p := range validPaths {
						if filepath.Base(p) == selectWidget.Selected {
							targetFolder = p
							break
						}
					}
					if targetFolder != "" {
						dest := filepath.Join(targetFolder, "AVLedger")
						if err := db.MoveTo(dest); err != nil {
							dialog.ShowError(err, w)
							return
						}
						a.Preferences().SetString("dbPath", db.Path)
						dbPathLabel.SetText(db.Path)
						dialog.ShowInformation("Success", "Database moved securely and backup initialized.", w)
					}
				} else {
					// If they cancel, we can optionally also mark it as prompted,
					// or leave it so it prompts next time. 
					// We'll mark it prompted so it defaults to local storage and doesn't annoy them.
					a.Preferences().SetBool("cloudPrompted", true)
				}
			}, w)
		}
	}

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
		if err := pdf.Export(path, *entries, s, db); err != nil {
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

func detectCloudSyncFolders() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	targets := []string{
		filepath.Join(home, "Dropbox"),
		filepath.Join(home, "OneDrive"),
		filepath.Join(home, "Google Drive"),
		filepath.Join(home, "Nextcloud"),
		filepath.Join(home, "ownCloud"),
		filepath.Join(home, "pCloudDrive"),
	}

	// For Mac, also check Library/CloudStorage
	if runtime.GOOS == "darwin" {
		cloudStorage := filepath.Join(home, "Library", "CloudStorage")
		if entries, err := os.ReadDir(cloudStorage); err == nil {
			for _, e := range entries {
				if e.IsDir() {
					targets = append(targets, filepath.Join(cloudStorage, e.Name()))
				}
			}
		}
	}

	var found []string
	for _, t := range targets {
		if info, err := os.Stat(t); err == nil && info.IsDir() {
			found = append(found, t)
		}
	}

	return found
}
