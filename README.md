# ✈️ AVLedger

**AVLedger** is a minimalist, open-source, EASA based logbook tool for **Aircraft Maintenance Engineers (AME)** — built with Go, designed with the [KISS principle](https://en.wikipedia.org/wiki/KISS_principle), and made for people who have better things to do than wait for software.

Log your maintenance tasks, export your logbook, move on.

---

## Why AVLedger?

AMTs don't have time for bloated interfaces or sluggish exports. AVLedger gets out of the way and does exactly what it says:

- 🔧 **Track maintenance tasks** — record every job done on every aircraft, clearly and quickly
- 📄 **Export to PDF in seconds** — because paperwork should never be the bottleneck
- 💾 **Own your data** — SQLite means your data is always one file away from safe. Zero subscriptions required. The software does not track any user data whatsoever, giving you total freedom and absolute control over how you manage your information.

---

## Features

| Feature | Description |
|---|---|
| 🔧 Task logging | Record maintenance work efficiently including ATA, Job Type, Workorders etc. |
| 🔍 Advanced filtering | Instantly narrow down your maintenance history by Aircraft, Registration, Category, or Job Type. |
| ⚡ Fast PDF export | Generate beautifully styled, ready-to-print PDF logbooks seamlessly (with smart text wrapping). |
| ☁️ Smart Cloud Backup | Automatically detects cloud sync folders (Nextcloud, Google Drive, OneDrive, Dropbox). Features **Auto-Discovery** to magically reconnect to an existing cloud database on fresh installations without user intervention. |
| 💽 Hot-swappable DB | Full manual control. Connect to or switch between any local/remote SQLite database files instantly from the UI without restarting. |
| 🎨 Refined UI | A modern, minimal Fyne-based interface featuring custom themes, zebra-striped tables, and clear visual hierarchy. |
| 🪶 Lightweight | Minimal footprint, standalone binaries, no heavy unnecessary dependencies. |
| 🔓 FOSS | Free and open-source, forever. |

---

## Philosophy

AVLedger is built around **KISS** — *Keep It Simple, Stupid*.

No locked-in cloud SaaS. No subscriptions. No unnecessary complexity. Just a tool that works smoothly on your machine, under your control. If it doesn't help an AME log a task faster or export cleaner, it doesn't belong in AVLedger.

---

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) 1.21 or later

### Build from source

```bash
git clone https://github.com/Pnz89/avledger.git
cd avledger
go build -o avledger .
```

### Run

```bash
./avledger
```

*(Note: Binaries can be built for Linux, macOS, FreeBSD, and Windows!)*

---

## Tech Stack

- **Language:** [Go](https://golang.org/)
- **UI Framework:** [Fyne](https://fyne.io/)
- **Database:** [SQLite](https://www.sqlite.org/) — single file, zero setup, easy backup
- **PDF generation:** [go-pdf/fpdf](https://github.com/go-pdf/fpdf)

---

## Contributing

Contributions are welcome. Please keep the KISS philosophy in mind — if a feature adds complexity without clear value, it probably doesn't belong here.

1. Fork the repo
2. Create your branch (`git checkout -b feature/your-feature`)
3. Commit your changes (`git commit -m 'Add some feature'`)
4. Push to the branch (`git push origin feature/your-feature`)
5. Open a Pull Request

---

## License

AVLedger is licensed under the **[GNU General Public License v2.0](./LICENSE)**.

You are free to use, study, modify, and distribute this software, provided that any derivative work is also distributed under the same license.

---

> *Built for the people who keep planes in the air — not for the people who like pretty dashboards.*
