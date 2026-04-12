# ✈️ AVLedger

**AVLedger** is a minimalist, open-source logbook tool for **Aircraft Maintenance Technicians (AMT)** — built with Go, designed with the [KISS principle](https://en.wikipedia.org/wiki/KISS_principle), and made for people who have better things to do than wait for software.

Log your maintenance tasks, export your logbook, move on.

---

## Why AVLedger?

AMTs don't have time for bloated interfaces or sluggish exports. AVLedger gets out of the way and does exactly what it says:

- 🔧 **Track maintenance tasks** — record every job done on every aircraft, clearly and quickly
- 📄 **Export to PDF in seconds** — because paperwork should never be the bottleneck
- 💾 **Backup your database instantly** — SQLite means your data is always one file away from safe

---

## Features

| Feature | Description |
|---|---|
| 🔧 Task logging | Record maintenance work performed on aircraft with full traceability |
| 📋 Logbook view | Browse and filter your maintenance history at a glance |
| ⚡ Fast PDF export | Generate ready-to-print logbooks without waiting |
| 💾 One-click DB backup | Instant SQLite database backup, no configuration needed |
| 🪶 Lightweight | Minimal footprint, no unnecessary dependencies |
| 🔓 FOSS | Free and open-source, forever |

---

## Philosophy

AVLedger is built around **KISS** — *Keep It Simple, Stupid*.

No cloud. No subscriptions. No unnecessary complexity. Just a tool that works, on your machine, under your control. If it doesn't help an AMT log a task faster, it doesn't belong in AVLedger.

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

---

## Tech Stack

- **Language:** [Go](https://golang.org/)
- **Database:** [SQLite](https://www.sqlite.org/) — single file, zero setup, easy backup
- **PDF generation:** built-in Go library

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
