package pdf

import (
	"bytes"
	"fmt"
	"time"

	"avledger/internal/assets"
	"avledger/internal/models"

	"github.com/go-pdf/fpdf"
)

const (
	pageW  = 297.0 // A4 landscape width mm
	pageH  = 210.0 // A4 landscape height mm
	margin = 10.0
)

// column widths in mm — total must equal pageW - 2*margin = 277mm
var colWidths = []float64{20, 28, 22, 70, 20, 15, 14, 26, 62}

// Export generates an A4 landscape PDF from the provided entries and settings,
// writing the result to the given file path.
func Export(path string, entries []models.LogEntry, s models.Settings) error {
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
	const rowsPerPage = 14
	pages := chunkEntries(entries, rowsPerPage)
	if len(pages) == 0 {
		pages = [][]models.LogEntry{{}} // at least one blank page
	}

	for pageIdx, pageEntries := range pages {
		pdf.AddPage()
		drawPage(pdf, pageEntries, s, pageIdx+1, len(pages))
	}

	return pdf.OutputFileAndClose(path)
}

// drawPage renders a single page of the logbook.
func drawPage(pdf *fpdf.Fpdf, entries []models.LogEntry, s models.Settings, pageNum, totalPages int) {
	// ---- Table geometry ----
	const (
		headerH    = 12.0 // total header height (two sub-rows of 6mm each)
		rowH       = 9.5  // data row height
		rowsOnPage = 14
	)

	tableTop := margin + 22.0
	tableW := pageW - 2*margin // 277 mm
	tableH := headerH + float64(rowsOnPage)*rowH

	// ---- Logo (top-right) ----
	pdf.ImageOptions("logo", pageW-margin-18, margin, 18, 18, false, fpdf.ImageOptions{ImageType: "png", ReadDpi: true}, 0, "")

	// ---- Title ----
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetXY(margin, margin+11)
	pdf.CellFormat(tableW, 7, "MAINTENANCE EXPERIENCE RECORD", "", 0, "L", false, 0, "")

	// ===========================================================
	// Grid: draw all borders first with Line/Rect, then add text.
	// This guarantees every line is present and connected.
	// ===========================================================
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.3)

	// Outer border of the whole table
	pdf.Rect(margin, tableTop, tableW, tableH, "D")

	// Horizontal separator: header / data area
	pdf.Line(margin, tableTop+headerH, margin+tableW, tableTop+headerH)

	// Horizontal separator between the two header sub-rows
	pdf.Line(margin, tableTop+headerH/2, margin+tableW, tableTop+headerH/2)

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

	// ---- Header fill ----
	pdf.SetFillColor(220, 220, 220)
	pdf.Rect(margin, tableTop, tableW, headerH, "F")

	// ---- Header text (two sub-rows per column) ----
	subH := headerH / 2.0

	hLine1 := []string{
		"Date",
		"Aircraft /",
		"Reg",
		"Task Detail",
		"Category",
		"Job type",
		"ATA",
		"Work Order",
		"Verified by",
	}
	hLine2 := []string{
		"",
		"Engine Type",
		"Marks",
		"",
		"(A,B1,B2,C)",
		"",
		"",
		"Number",
		"(Signature + Auth / AML)",
	}

	pdf.SetFont("Helvetica", "B", 7.5)
	xCol = margin
	for i, txt := range hLine1 {
		pdf.SetXY(xCol, tableTop)
		pdf.CellFormat(colWidths[i], subH, txt, "", 0, "C", false, 0, "")
		pdf.SetXY(xCol, tableTop+subH)
		pdf.CellFormat(colWidths[i], subH, hLine2[i], "", 0, "C", false, 0, "")
		xCol += colWidths[i]
	}

	// ---- Data rows ----
	for i := 0; i < rowsOnPage; i++ {
		y := tableTop + headerH + float64(i)*rowH

		var e models.LogEntry
		hasData := i < len(entries)
		if hasData {
			e = entries[i]
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
			e.VerifiedBy,
		}

		xCol = margin
		for j, cell := range cells {
			pdf.SetXY(xCol, y)
			align := "C"
			if j == 3 || j == 8 {
				align = "L"
			}
			style := ""
			if hasData && (j == 3 || j == 8) {
				style = "I"
			}
			fitCellText(pdf, colWidths[j], rowH, cell, style, align)
			xCol += colWidths[j]
		}
	}

	// ---- Footer ----
	footerTop := tableTop + tableH + 3.5
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetXY(margin, footerTop)
	pdf.CellFormat(tableW, 5,
		"I hereby declare that the information given on this logbook page is true in every respect",
		"", 1, "L", false, 0, "")

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
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.3)
	sigLineX := margin + 22
	pdf.Line(sigLineX, sigY+4.5, sigLineX+sigLineW, sigY+4.5)

	// ---- Page number (right-anchored, dynamic width) ----
	pdf.SetFont("Helvetica", "", 7)
	pageLabel := fmt.Sprintf("AVLedger - Page %d of %d", pageNum, totalPages)
	labelW := pdf.GetStringWidth(pageLabel) + 2.0 // +2mm padding
	pdf.SetXY(pageW-margin-labelW, pageH-margin-4)
	pdf.CellFormat(labelW, 4, pageLabel, "", 0, "R", false, 0, "")
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

// fitCellText renders text inside a cell, shrinking the font size as needed
// (from 8pt down to 4pt) to ensure the text never overflows the cell width.
func fitCellText(pdf *fpdf.Fpdf, w, h float64, txt, style, align string) {
	const (
		startSize = 8.0
		minSize   = 4.0
	)
	// Leave a small horizontal padding so text doesn't touch the border.
	const hPad = 1.0

	size := startSize
	pdf.SetFont("Helvetica", style, size)
	for size > minSize && pdf.GetStringWidth(txt) > w-hPad {
		size -= 0.5
		pdf.SetFont("Helvetica", style, size)
	}
	pdf.CellFormat(w, h, txt, "", 0, align, false, 0, "")
}
