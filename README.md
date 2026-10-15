# Dokan Dorpon

*Dokan Dorpon* ("the shop's mirror") is a grocery demand and stock manager.

A web app for a small grocery shop. It keeps track of stock, sales, expiry dates and profit, and it
suggests how much to buy based on **festivals and seasons** in Bangladesh.

- Backend: **Go** (standard library + `modernc.org/sqlite`)
- Database: **SQLite** (one file: `grocery.db`)
- Frontend: plain **HTML + CSS + JavaScript** (`fetch()`), no frameworks and no CDN, so it **works offline**

> University project. The full explanation of every file, function, SQL query and built-in function is in
> [VIVA_GUIDE.md](VIVA_GUIDE.md).

---

## Features

| Tab | What it does |
|---|---|
| **Dashboard** | Today's income / profit / items sold, active festivals, low stock and expiry alerts, 12-month chart |
| **Products** | List, add and edit products, set a discount (0–90%), see the final price after discount |
| **Sell** | Sell a product with a quantity, live total, "not enough stock" error, last 15 sales |
| **Stock in** | Record a purchase: quantity, buy price, optional expiry date |
| **Reports** | Daily / monthly / yearly report per product: sold qty, income, cost, profit, totals, best product |
| **Suggestions** | How much to buy (festival + season aware), discount suggestions with an "Apply" button |

Key rules:

- Stock changes in **only 3 ways**: a purchase (+), a sale (−), removing an expired batch (−). A new product starts with 0 stock.
- Every purchase is a **batch**. When we sell, the batch that **expires first is used first** (FEFO).
- Each sale saves **that day's buy price and sell price**, so old profit stays correct even if prices change later.
- Demand multipliers for Ramadan, Eid-ul-Fitr, Pohela Boishakh, Qurbani Eid, Durga Puja and the
  Summer / Monsoon / Winter seasons are written in the code.

---

## Folder structure

```
dokan-dorpon/
├── go.mod            module grocery, go 1.22
├── go.sum
├── main.go           opens the database, registers routes, starts the server on :8080
├── helpers.go        small helpers for dates, rounding and JSON
├── database.go       creates the tables + sample data on the first run
├── products.go       list / add / edit products, set discount
├── sales.go          sell, recent sales, stock in, remove expired batch
├── demand.go         festivals + seasons, multiplier, buy suggestions, occasions
├── suggestions.go    alerts (low stock, expiry) + discount suggestions
├── reports.go        daily / monthly / yearly report, chart data
├── static/
│   ├── index.html
│   ├── style.css
│   └── app.js
├── README.md
└── VIVA_GUIDE.md
```

---

## Install and run

### 1. Install Go

Download Go from https://go.dev/dl/ (**Go 1.22 or newer**). Open a new terminal and check:

```
go version
```

### 2. Go to the project folder

```
cd dokan-dorpon
```

(the folder that has `go.mod` in it)

### 3. Download the SQLite driver (once, needs internet)

**If you cloned the repo** (`go.mod` is already there):

```
go mod download
```

**If you are creating the project from zero:**

```
go mod init grocery
go mod edit -go=1.22
go get modernc.org/sqlite@v1.35.0
```

> **Why the fixed version `v1.35.0`?** The newest `modernc.org/sqlite` needs Go 1.26, and
> `go get modernc.org/sqlite` without a version changes the `go 1.22` line in `go.mod` to `go 1.26.0`.
> `v1.35.0` works with Go 1.22, and both of us get exactly the same version.
> So after cloning, **do not run `go get modernc.org/sqlite` without a version.**

### 4. Run

```
go run .
```

The terminal shows: `Server running at http://localhost:8080`

> The first build can take 1–2 minutes because the SQLite driver is big. Later builds are fast.

### 5. Open the browser

http://localhost:8080

Stop the server with **Ctrl + C** in the terminal.

---

## Sample data

On the first run, `grocery.db` is created. If it has no products, the app adds **18 products and one year
(365 days) of sales**, so the reports and the chart show real spikes (spices before Qurbani Eid, ice cream in
summer, dates during Ramadan). Some items are set up on purpose for the demo:

- **Bread**: its batch is already expired → it shows in the expiry alert with a "Remove from stock" button
- **Milk**: expires in 2 days and stock is below min stock → low stock alert + 30% discount suggestion
- **Rooh Afza**: only sells during Ramadan → "No sale in the last 30 days" → 20% discount suggestion
- **Ice Cream Cup**: far too much stock → 10% discount suggestion

### Reset the sample data

Stop the server (Ctrl + C), delete `grocery.db`, then run `go run .` again.

```
# PowerShell
Remove-Item grocery.db

# Git Bash
rm grocery.db
```

---

## API

| Method | URL | What it does |
|---|---|---|
| GET | `/api/products` | All products |
| POST | `/api/products` | Add a product |
| POST | `/api/products/update` | Edit a product |
| POST | `/api/products/discount` | Set discount (0–90) |
| POST | `/api/sales` | Sell |
| GET | `/api/sales/recent` | Last 15 sales |
| POST | `/api/purchases` | Stock in |
| POST | `/api/purchases/remove-expired` | Remove an expired batch |
| GET | `/api/report?period=2026-10-15` | Report (`2026-10-15` daily, `2026-10` monthly, `2026` yearly) |
| GET | `/api/chart?year=2026&product_id=3` | 12 months of chart data (`product_id` is optional) |
| GET | `/api/occasions` | Today's festivals/seasons and festivals in the next 60 days |
| GET | `/api/suggestions/buy` | How much to buy |
| GET | `/api/suggestions/discount` | Discount suggestions |
| GET | `/api/alerts` | Low stock + expiry alerts |

On an error the server sends a plain text message (for example `Not enough stock. Only 5 left`) and the page
shows it. Request/response examples and curl tests are in
[VIVA_GUIDE.md](VIVA_GUIDE.md#14-api-reference-and-curl-tests).

---

## Our coding rules

We follow these rules to keep the code simple and easy to explain:

- Only the Go standard library + one outside package, `modernc.org/sqlite` (pure Go, so no gcc/CGO on Windows)
- `net/http` with Go 1.22 routing patterns (`"GET /api/products"`), no Gin/Echo/GORM
- Everything in `package main`, flat files, no sub-packages
- No goroutines, generics or interfaces written by us; short functions
- Every error is checked (`if err != nil { ... }`)
- Slices start as `[]T{}`, so the JSON is `[]` and not `null`
- Dates are stored as TEXT `"YYYY-MM-DD"`
- `db.SetMaxOpenConns(1)`, and `rows` are closed before the next query
- A transaction (`Begin` / `defer Rollback` / `Commit`) whenever more than one table changes
- All values go into SQL through `?` placeholders (protects against SQL injection)
- No CDN in the frontend, the chart is plain HTML/CSS bars

---

## Progress

We are building the project step by step. `go build` must pass after every step.

- [x] Docs: `README.md`, `VIVA_GUIDE.md`
- [ ] Setup: `go.mod`, `main.go` (static file server only), `helpers.go`, `.gitignore` (`grocery.db`, `*.exe`)
- [ ] `database.go`: tables + foreign keys
- [ ] `products.go` + routes
- [ ] `sales.go` (sell, recent sales, stock in, remove expired) + routes
- [ ] `demand.go` (festivals, seasons, occasions, buy suggestions) + sample data in `database.go`
- [ ] `suggestions.go` (alerts, discount suggestions)
- [ ] `reports.go` (report, chart)
- [ ] `static/index.html` + `static/style.css` (khata ledger design)
- [ ] `static/app.js` (all tabs)
- [ ] Full test: every API with curl + every tab in the browser

---

## Troubleshooting

| Problem | Fix |
|---|---|
| `go is not recognized` | Close and reopen the terminal / VS Code after installing Go |
| `address already in use` (port 8080) | Another server is still running; stop it with Ctrl + C |
| Windows Firewall popup | Click "Allow". Even if you cancel, `localhost` still works |
| `database is locked` | If the folder is inside OneDrive, OneDrive may lock `grocery.db` while syncing. Move the project out of OneDrive (e.g. `C:\projects\dokan-dorpon`) |
| Page not found / 404 | Run `go run .` from the project folder, because the server looks for `static/` there |
| No internet at the viva | Run `go build` once with internet before the viva; after that it works offline |

---

## Team

| | Part |
|---|---|
| **Member 1** | Products, Reports, Dashboard, database setup |
| **Member 2** | Sell, Stock in, Expiry, Festival demand, Suggestions |

File-by-file split: [VIVA_GUIDE.md](VIVA_GUIDE.md#16-who-does-what).
