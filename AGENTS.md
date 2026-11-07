# Panduan Belajar dan Kontribusi Codekop Go MVC

Dokumen ini adalah peta untuk memahami project, terutama bagi developer PHP yang
sedang pindah ke Go. Baca berurutan: konfigurasi, server, route, handler, CRUD,
baru kemudian database.

## Tujuan project

Project ini bukan clone penuh `codekop_v5`. Ini adalah versi belajar yang
memindahkan pola pentingnya ke Go + Gin:

```text
PHP front controller
  -> route
  -> controller
  -> model/PDO

Go server
  -> Gin middleware
  -> Gin route
  -> handler/controller
  -> repository + crud.Helper
  -> database/sql
```

Endpoint dan tabel sample sengaja netral. Jangan memasukkan aturan `status_bayar`,
`Midtrans`, RBAC, atau modul bisnis besar sebelum fondasi ini benar-benar dipahami.

## Struktur folder

```text
cmd/server/main.go          # bootstrap: .env, config, DB, HTTP server
internal/app/               # composition root; membuat router Gin
internal/app/config/         # APP_ENV, HTTP, database, .env loader
internal/app/database/       # pembukaan database/sql + driver
internal/routes/             # semua deklarasi URL dan HTTP method
  routes.go                  # aggregator route + dependency wiring
  api_routes.go              # health, echo, route parameter
  items_routes.go            # CRUD sample items
  transaksi_routes.go        # query transaksi
internal/controllers/        # controller HTTP, validasi request, response
internal/models/             # struct data/domain yang dikirim repository
internal/repositories/       # query database dan pemanggilan helper CRUD
internal/middleware/        # security header dan request logger
internal/helper/             # crud.go, validate.go, json.go
migrations/mysql/           # SQL MariaDB/MySQL
migrations/postgres/        # SQL PostgreSQL
testing/                    # smoke/integration test API dengan httptest
```

## Cara membaca satu request

Contoh `POST /api/items`:

1. `cmd/server/main.go` memuat `.env`, membuat `sql.DB`, lalu memanggil `app.New`.
2. `internal/routes/routes.go` mendaftarkan route `api.POST("/items", itemController.Store)`.
3. Gin menjalankan middleware recovery, logger, dan security header.
4. `controllers.ItemController.Store` membaca JSON, melakukan validasi input ringan, dan memberi
   status HTTP `422` jika data tidak valid.
5. `repositories.ItemRepository.Create` mengirim data ke `helper.Helper.Insert`.
6. `helper.Helper` meng-allow-list kolom, memakai placeholder SQL, lalu menjalankan
   prepared statement.
7. Controller mengembalikan envelope JSON `{ "success": true, "data": ... }`.

Handler tidak boleh membangun SQL mentah dari input user. Query berada di repository
atau helper database.

## Database transaction

Untuk beberapa query yang harus berhasil bersama-sama, gunakan
`helper.WithTransaction`. Callback menerima `*sql.Tx`; return `nil` berarti commit,
return error berarti rollback. Jangan memanggil `db.ExecContext` di tengah callback
untuk query yang seharusnya ikut transaksi—gunakan `tx.ExecContext`.

## Pemetaan istilah PHP ke Go

| PHP di `codekop_v5` | Project Go |
|---|---|
| `index.php` | `cmd/server/main.go` |
| `app/Config/Config.php` | `internal/app/config` + `.env` |
| `app/Routes/api.php` | route group `/api` di `internal/routes/routes.go` |
| `Controller` | `internal/controllers/*_controller.go` |
| `Request::json()` | `helper.DecodeJSON(c, &input)` |
| `Response::json()` | `helper.Success` / `helper.Error` |
| PDO | `*sql.DB` dari `database/sql` |
| Model | `internal/models/*.go` |
| Query/model PHP | `internal/repositories/*.go` |
| `System\\Crud` | `internal/helper/crud.go` |
| Middleware class | `gin.HandlerFunc` |
| `APP_ENV` | `APP_ENV` |

## Menambah modul baru

1. Buat `internal/controllers/blog_controller.go` untuk HTTP input/output.
2. Buat `internal/models/blog.go` untuk bentuk data yang keluar/masuk.
3. Buat `internal/repositories/blog_repository.go` untuk query atau pemanggilan `helper.Helper`.
4. Jika ada aturan bisnis lebih dari validasi sederhana, tambahkan `service.go` di
   antara handler dan repository.
5. Daftarkan route pada group `/api` di `internal/routes/routes.go`.
6. Buat migration di kedua folder database jika tabel baru diperlukan.
7. Tambahkan smoke test di package yang terkait.

Test API lintas package diletakkan di `testing/`; test detail helper boleh tetap
diletakkan di package sumbernya sebagai `*_test.go`.

Untuk CRUD sederhana, gunakan helper:

```go
crudHelper, err := helper.New(db, driver, "blogs", []string{"id", "title", "created_at"})
```

Nama tabel dan kolom adalah kode aplikasi, bukan input request. Nilai user tetap
masuk sebagai parameter SQL. Ini alasan helper memisahkan identifier dari values.

Query database sederhana memakai `helper.NewORM` dengan gaya CodeIgniter:
`Select().From().Where().OrderBy().Limit().Get()`. Hasil bisa dibaca melalui
`GetResultArray()`, `GetRowArray()`, `GetResult()`, atau `GetRow()`.

## Database dan dialect

- `DB_DRIVER=mysql` membuka MariaDB/MySQL melalui `go-sql-driver/mysql`.
- `DB_DRIVER=postgres` membuka PostgreSQL melalui `pgx/v5/stdlib` sebagai driver
  `database/sql`.
- `DB_SOCKET=/tmp/mysql.sock` memakai Unix socket MariaDB/MySQL; jika kosong,
  koneksi memakai `DB_HOST` dan `DB_PORT`.
- MySQL memakai placeholder `?` dan backtick identifier.
- PostgreSQL memakai placeholder `$1`, `$2`, dan double-quote identifier.

Perbedaan itu disembunyikan oleh `crud.Helper`, sehingga repository tetap mudah
dibaca.

## Development vs production

Development menyalakan mode debug Gin dan cocok untuk `go run`.
`make dev` memakai Air Go dari `$(go env GOPATH)/bin/air` bila terpasang; konfigurasi
watcher ada di `.air.toml`. Folder `tmp/` adalah output sementara dan tidak boleh
di-commit. Jangan memakai binary `air` milik bahasa R.
Production memakai mode release Gin, timeout HTTP, header keamanan lebih ketat, dan
log JSON. `ErrorLogger` menyimpan detail error internal di log tanpa mengirimnya ke
client. Jangan menaruh password database di source code; gunakan environment server
atau secret manager.

## Perintah wajib sebelum commit

```bash
gofmt -w cmd internal
go test ./...
go vet ./...
go build ./cmd/server
```

Jika dependency belum ada, jalankan `go mod download` di mesin yang memiliki akses
internet.
