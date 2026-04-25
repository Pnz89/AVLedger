package ui

import (
	"encoding/json"
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

	w := a.NewWindow("AVLedger - The logbook that belongs to you")
	w.SetIcon(assets.ResourceLogoPng)
	w.Resize(fyne.NewSize(1280, 760))
	w.SetMaster()

	showProfileSelector(a, w, customTheme)

	w.ShowAndRun()
}

func getLogoImage(a fyne.App) *canvas.Image {
	var logoImg *canvas.Image
	if a.Settings().ThemeVariant() == theme.VariantDark {
		logoImg = canvas.NewImageFromResource(assets.ResourceWordmarkDarkPng)
	} else {
		logoImg = canvas.NewImageFromResource(assets.ResourceWordmarkLightPng)
	}
	logoImg.FillMode = canvas.ImageFillContain
	return logoImg
}

func showProfileSelector(a fyne.App, w fyne.Window, customTheme *CustomTheme) {
	var profiles []models.UserProfile
	prefStr := a.Preferences().String("profiles")
	if prefStr != "" {
		_ = json.Unmarshal([]byte(prefStr), &profiles)
	} else {
		// Migration for existing single-user setups
		legacyPath := a.Preferences().String("dbPath")
		if legacyPath == "" {
			// First run migration check
			cloudFolders := detectCloudSyncFolders()
			for _, f := range cloudFolders {
				checkPath := filepath.Join(f, "AVLedger", "avledger.db")
				if _, err := os.Stat(checkPath); err == nil {
					legacyPath = checkPath
					break
				}
			}
		}
		if legacyPath != "" {
			profiles = append(profiles, models.UserProfile{Name: "Default", DBPath: legacyPath})
			b, _ := json.Marshal(profiles)
			a.Preferences().SetString("profiles", string(b))
		}
	}

	logoImg := getLogoImage(a)
	logoImg.SetMinSize(fyne.NewSize(200, 80))

	title := widget.NewLabel("Select Your Profile")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	var list *widget.List
	list = widget.NewList(
		func() int { return len(profiles) },
		func() fyne.CanvasObject {
			btn := widget.NewButton("Select", nil)
			btn.Importance = widget.HighImportance
			return container.NewBorder(nil, nil, widget.NewIcon(theme.AccountIcon()), btn, widget.NewLabel("Name"))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			row := o.(*fyne.Container)
			nameLabel := row.Objects[0].(*widget.Label)
			nameLabel.SetText(profiles[i].Name)
			nameLabel.TextStyle = fyne.TextStyle{Bold: true}
			
			btn := row.Objects[2].(*widget.Button)
			btn.OnTapped = func() {
				showMainApp(a, w, customTheme, profiles[i], profiles)
			}
		},
	)

	newBtn := widget.NewButtonWithIcon("Create New Profile", theme.ContentAddIcon(), func() {
		showNewProfileDialog(a, w, &profiles, func(newProfile models.UserProfile) {
			b, _ := json.Marshal(profiles)
			a.Preferences().SetString("profiles", string(b))
			list.Refresh()
			showMainApp(a, w, customTheme, newProfile, profiles)
		})
	})
	newBtn.Importance = widget.HighImportance

	cardContent := container.NewBorder(
		container.NewVBox(container.NewCenter(logoImg), widget.NewSeparator(), title, widget.NewLabel("")),
		container.NewPadded(newBtn),
		nil, nil,
		container.NewPadded(list),
	)

	card := widget.NewCard("", "", cardContent)
	centered := container.NewCenter(container.New(layout.NewGridWrapLayout(fyne.NewSize(450, 500)), card))

	w.SetContent(centered)
}

func showNewProfileDialog(a fyne.App, w fyne.Window, profiles *[]models.UserProfile, onCreated func(models.UserProfile)) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("E.g. Mario Rossi")

	cloudFolders := detectCloudSyncFolders()
	options := []string{"Local Storage (Default)"}
	for _, p := range cloudFolders {
		options = append(options, filepath.Base(p))
	}

	storageSelect := widget.NewSelect(options, nil)
	storageSelect.SetSelected(options[0])

	items := []*widget.FormItem{
		widget.NewFormItem("Profile Name", nameEntry),
		widget.NewFormItem("Save Location", storageSelect),
	}

	dialog.ShowForm("Create New Profile", "Create", "Cancel", items, func(confirm bool) {
		if !confirm {
			return
		}
		name := strings.TrimSpace(nameEntry.Text)
		if name == "" {
			dialog.ShowError(fmt.Errorf("Profile name cannot be empty"), w)
			return
		}

		for _, p := range *profiles {
			if strings.EqualFold(p.Name, name) {
				dialog.ShowError(fmt.Errorf("Profile name already exists"), w)
				return
			}
		}

		var dir string
		if storageSelect.Selected == "Local Storage (Default)" {
			dataDir, err := os.UserConfigDir()
			if err != nil {
				dataDir, _ = os.UserHomeDir()
			}
			dir = filepath.Join(dataDir, "avledger", "profiles", name)
		} else {
			for _, p := range cloudFolders {
				if filepath.Base(p) == storageSelect.Selected {
					dir = filepath.Join(p, "AVLedger", name)
					break
				}
			}
		}

		os.MkdirAll(dir, 0755)
		dbPath := filepath.Join(dir, "avledger.db")

		newProfile := models.UserProfile{Name: name, DBPath: dbPath}
		*profiles = append(*profiles, newProfile)
		onCreated(newProfile)

	}, w)
}

func saveProfiles(a fyne.App, profiles []models.UserProfile) {
	b, _ := json.Marshal(profiles)
	a.Preferences().SetString("profiles", string(b))
}

func showMainApp(a fyne.App, w fyne.Window, customTheme *CustomTheme, profile models.UserProfile, allProfiles []models.UserProfile) {
	db, err := database.Open(profile.DBPath)
	if err != nil {
		dialog.ShowError(err, w)
		showProfileSelector(a, w, customTheme)
		return
	}

	entries, err := db.ListEntries()
	if err != nil {
		entries = []models.LogEntry{}
	}

	settings := loadSettings(db)

	entryList := entries

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search tasks...")

	acSelect := widget.NewSelect([]string{"(All)"}, nil)
	startDateEntry := widget.NewEntry()
	startDateEntry.SetPlaceHolder("DD/MM/YYYY")
	endDateEntry := widget.NewEntry()
	endDateEntry.SetPlaceHolder("DD/MM/YYYY")
	catSelect := widget.NewSelect([]string{"(All)"}, nil)
	ataSelect := widget.NewSelect([]string{"(All)"}, nil)
	jobSelect := widget.NewSelect([]string{"(All)"}, nil)

	acSelect.SetSelected("(All)")
	catSelect.SetSelected("(All)")
	ataSelect.SetSelected("(All)")
	jobSelect.SetSelected("(All)")

	updateSelectOptions := func() {
		if opts, _ := db.GetDistinctValues("aircraft_engine_type"); opts != nil {
			acSelect.Options = append([]string{"(All)"}, opts...)
		}
		if opts, _ := db.GetDistinctValues("category"); opts != nil {
			catSelect.Options = append([]string{"(All)"}, opts...)
		}
		if opts, _ := db.GetDistinctValues("ata"); opts != nil {
			var truncatedOpts []string
			seen := make(map[string]bool)
			for _, opt := range opts {
				val := opt
				if idx := strings.Index(val, " - "); idx != -1 {
					val = strings.TrimSpace(val[:idx])
				}
				if !seen[val] {
					seen[val] = true
					truncatedOpts = append(truncatedOpts, val)
				}
			}
			ataSelect.Options = append([]string{"(All)"}, truncatedOpts...)
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
		cat := catSelect.Selected
		if cat == "(All)" {
			cat = ""
		}
		ata := ataSelect.Selected
		if ata == "(All)" {
			ata = ""
		}
		job := jobSelect.Selected
		if job == "(All)" {
			job = ""
		}
		return models.FilterOptions{
			SearchQuery:        searchEntry.Text,
			AircraftEngineType: ac,
			StartDate:          startDateEntry.Text,
			EndDate:            endDateEntry.Text,
			Category:           cat,
			JobType:            job,
			ATA:                ata,
		}
	}

	var et *entryTable
	var tableContent fyne.CanvasObject

	et, tableContent = buildTable(&entryList, w,
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
	startDateEntry.OnChanged = func(s string) { refreshAll() }
	endDateEntry.OnChanged = func(s string) { refreshAll() }
	catSelect.OnChanged = func(s string) { refreshAll() }
	ataSelect.OnChanged = func(s string) { refreshAll() }
	jobSelect.OnChanged = func(s string) { refreshAll() }

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

	aircraftsBtn := widget.NewButtonWithIcon("  Aircrafts", theme.FileApplicationIcon(), func() {
		showAircraftsDialog(w, db)
	})

	logoImg := getLogoImage(a)
	logoImg.SetMinSize(fyne.NewSize(120, 32))

	themeBtn := widget.NewButtonWithIcon("", theme.ColorPaletteIcon(), nil)
	themeBtn.OnTapped = func() {
		current := customTheme.lastVariant
		if customTheme.ForcedVariant != nil {
			current = *customTheme.ForcedVariant
		}

		if current == theme.VariantLight {
			dark := theme.VariantDark
			customTheme.ForcedVariant = &dark
		} else {
			light := theme.VariantLight
			customTheme.ForcedVariant = &light
		}
		a.Settings().SetTheme(customTheme)
		
		variant := *customTheme.ForcedVariant
		if variant == theme.VariantDark {
			logoImg.Resource = assets.ResourceWordmarkDarkPng
		} else {
			logoImg.Resource = assets.ResourceWordmarkLightPng
		}
		logoImg.Refresh()
	}

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
			
			for i, p := range allProfiles {
				if p.Name == profile.Name {
					allProfiles[i].DBPath = db.Path
					saveProfiles(a, allProfiles)
					break
				}
			}

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

	manageDBBtn := widget.NewButtonWithIcon("Manage DB", theme.SettingsIcon(), nil)
	manageDBBtn.OnTapped = func() {
		pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(manageDBBtn)
		pos.Y += manageDBBtn.Size().Height
		widget.ShowPopUpMenuAtPosition(manageMenu, w.Canvas(), pos)
	}

	switchProfileBtn := widget.NewButtonWithIcon("Switch Profile", theme.AccountIcon(), func() {
		db.Close()
		showProfileSelector(a, w, customTheme)
	})
	switchProfileBtn.Importance = widget.WarningImportance

	rightActions := container.NewHBox(switchProfileBtn, manageDBBtn)

	dbBar := container.NewBorder(nil, nil,
		container.NewHBox(dbIcon, widget.NewLabel(fmt.Sprintf("Profile: %s | DB:", profile.Name))),
		rightActions,
		dbPathLabel,
	)

	versionText := canvas.NewText("0.7.0", theme.DisabledColor())
	versionText.TextSize = 12
	versionText.TextStyle = fyne.TextStyle{Bold: true}

	titleCol := container.NewVBox(
		logoImg,
		versionText,
	)

	titleRow := container.NewHBox(titleCol)

	toolbar := container.NewHBox(
		titleRow,
		widget.NewSeparator(),
		newBtn,
		exportBtn,
		widget.NewSeparator(),
		aircraftsBtn,
		assessorsBtn,
		settingsBtn,
		layout.NewSpacer(),
		themeBtn,
	)

	filtersRow := container.NewGridWithColumns(6,
		container.NewBorder(nil, nil, widget.NewLabel("Aircraft:"), nil, acSelect),
		container.NewBorder(nil, nil, widget.NewLabel("Cat:"), nil, catSelect),
		container.NewBorder(nil, nil, widget.NewLabel("ATA:"), nil, ataSelect),
		container.NewBorder(nil, nil, widget.NewLabel("Job:"), nil, jobSelect),
		container.NewBorder(nil, nil, widget.NewLabel("From:"), nil, startDateEntry),
		container.NewBorder(nil, nil, widget.NewLabel("To:"), nil, endDateEntry),
	)

	searchRow := container.NewBorder(nil, nil, container.NewHBox(widget.NewIcon(theme.SearchIcon()), widget.NewLabel("Search:")), nil, searchEntry)

	searchContainer := container.NewVBox(
		searchRow,
		filtersRow,
	)

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
}

func reloadEntries(db *database.DB, list *[]models.LogEntry, w fyne.Window, opts models.FilterOptions) {
	var updated []models.LogEntry
	var err error
	if opts.SearchQuery == "" && opts.AircraftEngineType == "" && opts.StartDate == "" && opts.EndDate == "" && opts.Category == "" && opts.JobType == "" && opts.ATA == "" {
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
		uc.Close()

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
