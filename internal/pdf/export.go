package pdf

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"avledger/internal/assets"
	"avledger/internal/database"
	"avledger/internal/models"

	"github.com/go-pdf/fpdf"
)

const (
	pageW  = 297.0 // A4 landscape width mm
	pageH  = 210.0 // A4 landscape height mm
	margin = 10.0
)

// column widths in mm — total must equal pageW - 2*margin = 277mm
var colWidths = []float64{18, 26, 20, 57, 18, 14, 12, 22, 16, 52, 22}

// Export generates an A4 landscape PDF from the provided entries and settings,
// writing the result to the given file path.
func Export(path string, entries []models.LogEntry, s models.Settings, db *database.DB) error {
	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: "L",
		UnitStr:        "mm",
		SizeStr:        "A4",
		FontDirStr:     "",
	})

	opt := fpdf.ImageOptions{ImageType: "png", ReadDpi: true}
	pdf.RegisterImageOptionsReader("logo", opt, bytes.NewReader(assets.ResourceLogoPng.StaticContent))

	pdf.SetMargins(margin, margin, margin)
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetTitle("AVLedger — Maintenance Experience Record", false)
	pdf.SetAuthor(s.HolderName, false)

	// Split entries into pages of rowsPerPage rows each
	const rowsPerPage = 12
	pages := chunkEntries(entries, rowsPerPage)
	if len(pages) == 0 {
		pages = [][]models.LogEntry{{}} // at least one blank page
	}

	for pageIdx, pageEntries := range pages {
		pdf.AddPage()
		drawPage(pdf, pageEntries, s, pageIdx+1, len(pages), db)
	}

	return pdf.OutputFileAndClose(path)
}

// drawPage renders a single page of the logbook.
func drawPage(pdf *fpdf.Fpdf, entries []models.LogEntry, s models.Settings, pageNum, totalPages int, db *database.DB) {
	// ---- Table geometry ----
	const (
		headerH    = 12.0 // total header height (two sub-rows of 6mm each)
		rowH       = 11.0 // data row height
		rowsOnPage = 12
	)

	tableTop := margin + 22.0
	tableW := pageW - 2*margin // 277 mm
	tableH := headerH + float64(rowsOnPage)*rowH

	// ---- Logo (top-right) ----
	pdf.ImageOptions("logo", pageW-margin-18, margin, 18, 18, false, fpdf.ImageOptions{ImageType: "png", ReadDpi: true}, 0, "")

	// ---- Title ----
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetTextColor(30, 41, 59) // Slate 800
	pdf.SetXY(margin, margin+11)
	pdf.CellFormat(tableW, 7, "MAINTENANCE EXPERIENCE RECORD", "", 0, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)

	// ===========================================================
	// Grid: draw background fills first, then borders
	// ===========================================================

	// ---- Background Fills ----
	// Header fill
	pdf.SetFillColor(190, 200, 210) // Darker for B&W print
	pdf.Rect(margin, tableTop, tableW, headerH, "F")

	// Zebra striping for data rows
	pdf.SetFillColor(235, 240, 245) // Darker zebra striping
	for i := 0; i < rowsOnPage; i++ {
		if i%2 != 0 {
			y := tableTop + headerH + float64(i)*rowH
			pdf.Rect(margin, y, tableW, rowH, "F")
		}
	}

	// ---- Borders ----
	pdf.SetDrawColor(80, 90, 100) // Darker borders
	pdf.SetLineWidth(0.3)

	// Outer border of the whole table
	pdf.Rect(margin, tableTop, tableW, tableH, "D")

	// Horizontal separator: header / data area
	pdf.Line(margin, tableTop+headerH, margin+tableW, tableTop+headerH)

	// (Horizontal separator between the two header sub-rows removed as requested)

	// Horizontal separators between data rows
	for i := 1; i < rowsOnPage; i++ {
		y := tableTop + headerH + float64(i)*rowH
		pdf.Line(margin, y, margin+tableW, y)
	}

	// Vertical separators (column dividers — full table height)
	xCol := margin
	for i := 0; i < len(colWidths)-1; i++ {
		xCol += colWidths[i]
		pdf.Line(xCol, tableTop, xCol, tableTop+tableH)
	}

	// ---- Header text (two sub-rows per column) ----
	subH := headerH / 2.0

	hLine1 := []string{
		"DATE",
		"AIRCRAFT /",
		"REG",
		"TASK DETAIL",
		"CATEGORY",
		"JOB TYPE",
		"ATA",
		"WORK ORDER",
		"DURATION",
		"VERIFIED BY",
		"STAMP /",
	}
	hLine2 := []string{
		"",
		"ENGINE TYPE",
		"MARKS",
		"",
		"(A,B1,B2,C)",
		"",
		"",
		"NUMBER",
		"(HOURS)",
		"(AUTH / AML)",
		"SIGNATURE",
	}

	pdf.SetTextColor(20, 20, 20) // Very dark text for headers
	pdf.SetFont("Helvetica", "B", 7.5)
	xCol = margin
	for i, txt := range hLine1 {
		pdf.SetXY(xCol, tableTop)
		pdf.CellFormat(colWidths[i], subH, txt, "", 0, "C", false, 0, "")
		pdf.SetXY(xCol, tableTop+subH)
		pdf.CellFormat(colWidths[i], subH, hLine2[i], "", 0, "C", false, 0, "")
		xCol += colWidths[i]
	}
	pdf.SetTextColor(0, 0, 0) // Reset to black

	// ---- Data rows ----
	for i := 0; i < rowsOnPage; i++ {
		y := tableTop + headerH + float64(i)*rowH

		var e models.LogEntry
		hasData := i < len(entries)
		if hasData {
			e = entries[i]
		}

		verifiedBy := e.VerifiedBy
		if db != nil && verifiedBy != "" {
			if a, err := db.GetAssessorByName(verifiedBy); err == nil {
				var parts []string
				parts = append(parts, a.Name)
				if a.LicenseNumber != "" {
					parts = append(parts, a.LicenseNumber)
				}
				if a.CompanyApproval != "" {
					parts = append(parts, a.CompanyApproval)
				}
				verifiedBy = strings.Join(parts, " - ")
			}
		}

		cells := []string{
			e.Date,
			e.AircraftEngineType,
			e.RegMarks,
			e.TaskDetail,
			e.Category,
			e.JobType,
			e.ATA,
			e.WorkOrderNumber,
			e.Duration,
			verifiedBy,
			"", // STAMP column
		}

		xCol = margin
		for j, cell := range cells {
			pdf.SetXY(xCol, y)
			align := "C"
			if j == 3 || j == 9 {
				align = "L"
			}
			style := ""
			if hasData && (j == 3 || j == 9) {
				style = "I"
			}
			fitCellText(pdf, colWidths[j], rowH, cell, style, align)
			xCol += colWidths[j]
		}
	}

	// ---- Footer ----
	footerTop := tableTop + tableH + 3.5
	pdf.SetFont("Helvetica", "I", 8)
	pdf.SetTextColor(50, 50, 50)
	pdf.SetXY(margin, footerTop)
	pdf.CellFormat(tableW, 5,
		"I hereby declare that the information given on this logbook page is true in every respect",
		"", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)

	nameY := footerTop + 6.5
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetXY(margin, nameY)
	pdf.CellFormat(18, 5, "Name:", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8)
	pdf.CellFormat(70, 5, s.HolderName, "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 8)
	pdf.CellFormat(20, 5, "Licence:", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8)
	pdf.CellFormat(55, 5, s.LicenceNumber, "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 8)
	pdf.CellFormat(15, 5, "Date:", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8)
	pdf.CellFormat(35, 5, time.Now().Format("02/01/2006"), "", 0, "L", false, 0, "")

	// ---- Signature field ----
	sigY := nameY + 8.0
	const sigLineW = 60.0
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetXY(margin, sigY)
	pdf.CellFormat(22, 5, "Signature:", "", 0, "L", false, 0, "")
	// Draw the underline for pen signing
	pdf.SetDrawColor(80, 90, 100)
	pdf.SetLineWidth(0.3)
	sigLineX := margin + 22
	pdf.Line(sigLineX, sigY+4.5, sigLineX+sigLineW, sigY+4.5)

	// ---- Page number (right-anchored, dynamic width) ----
	pdf.SetFont("Helvetica", "", 7)
	pdf.SetTextColor(80, 90, 100)
	pageLabel := fmt.Sprintf("AVLedger - Page %d of %d", pageNum, totalPages)
	labelW := pdf.GetStringWidth(pageLabel) + 2.0 // +2mm padding
	pdf.SetXY(pageW-margin-labelW, pageH-margin-4)
	pdf.CellFormat(labelW, 4, pageLabel, "", 0, "R", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
}

// chunkEntries splits entries into slices of at most size elements.
func chunkEntries(entries []models.LogEntry, size int) [][]models.LogEntry {
	var chunks [][]models.LogEntry
	for len(entries) > 0 {
		n := size
		if len(entries) < n {
			n = len(entries)
		}
		chunks = append(chunks, entries[:n])
		entries = entries[n:]
	}
	return chunks
}

// wrapText splits txt into lines that each fit within maxW mm at the current
// font size. It respects word boundaries and hard-wraps individual words that
// are wider than maxW.
func wrapText(pdf *fpdf.Fpdf, txt string, maxW float64) []string {
	words := splitWords(txt)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	current := ""

	for _, word := range words {
		if word == "" {
			continue
		}
		candidate := word
		if current != "" {
			candidate = current + " " + word
		}

		if pdf.GetStringWidth(candidate) <= maxW {
			current = candidate
		} else {
			// Current word alone is wider than the cell — hard-break it.
			if current == "" {
				// Flush the oversized word as-is (font shrinking will handle it).
				lines = append(lines, word)
			} else {
				lines = append(lines, current)
				current = word
			}
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// splitWords splits a string by spaces, preserving non-empty tokens.
func splitWords(s string) []string {
	var words []string
	start := -1
	for i, ch := range s {
		if ch == ' ' {
			if start >= 0 {
				words = append(words, s[start:i])
				start = -1
			}
		} else {
			if start < 0 {
				start = i
			}
		}
	}
	if start >= 0 {
		words = append(words, s[start:])
	}
	return words
}

// fitCellText renders text inside a cell.
// Strategy:
//  1. At the current font size (starting at 8pt), attempt to word-wrap the
//     text into lines that fit within the cell width.
//  2. If the wrapped lines fit within the cell height, render them and return.
//  3. Otherwise reduce the font size by 0.5pt and try again, down to 4pt.
//  4. At the minimum size render whatever fits.
func fitCellText(pdf *fpdf.Fpdf, w, h float64, txt, style, align string) {
	const (
		startSize = 8.0
		minSize   = 4.0
		hPad      = 1.0 // horizontal padding so text doesn't touch borders
		vPad      = 0.5 // vertical padding (top + bottom total)
	)

	if txt == "" {
		pdf.SetFont("Helvetica", style, startSize)
		pdf.CellFormat(w, h, "", "", 0, align, false, 0, "")
		return
	}

	// Remember the starting X,Y so we can render at the correct position.
	startX := pdf.GetX()
	startY := pdf.GetY()

	size := startSize
	for {
		pdf.SetFont("Helvetica", style, size)

		// Line height: fpdf uses font size in points; 1pt ≈ 0.352778 mm.
		lineH := size * 0.352778 * 1.2 // 1.2 leading factor

		// How many wrapped lines fit vertically?
		maxLines := int((h - vPad) / lineH)
		if maxLines < 1 {
			maxLines = 1
		}

		lines := wrapText(pdf, txt, w-hPad)

		// Check both vertical fit (line count) AND horizontal fit (each line width).
		// A single very-long word with no spaces produces 1 line that is still too wide.
		allFitH := true
		for _, line := range lines {
			if pdf.GetStringWidth(line) > w-hPad {
				allFitH = false
				break
			}
		}

		if (len(lines) <= maxLines && allFitH) || size <= minSize {
			// It fits (or we've hit the minimum size) — render.
			// Clip to maxLines so we never overflow.
			if len(lines) > maxLines {
				lines = lines[:maxLines]
			}

			// Total block height for vertical centering.
			blockH := float64(len(lines)) * lineH
			topY := startY + (h-blockH)/2.0
			if topY < startY {
				topY = startY
			}

			for li, line := range lines {
				lineY := topY + float64(li)*lineH
				pdf.SetXY(startX, lineY)
				pdf.CellFormat(w, lineH, line, "", 0, align, false, 0, "")
			}

			// Restore X/Y to after the cell so the caller's layout isn't broken.
			pdf.SetXY(startX+w, startY)
			return
		}

		// Text doesn't fit at this size — try a smaller font.
		size -= 0.5
	}
}
