# Codekop Go MVC

Starter project Go + Gin untuk pengguna PHP yang ingin memahami cara membangun
MVC dan REST API dengan Go secara bertahap. Struktur controller, model, repository,
helper, route, migration, testing, dan response JSON dibuat sengaja terlihat jelas
agar mudah dibandingkan dengan pola kerja PHP/CodeIgniter.

Referensi pola aplikasi PHP yang digunakan adalah repository
[php-mvc-codekop](https://github.com/fauzan1892/php-mvc-codekop). Project ini bukan
salinan seluruh fitur aplikasi PHP tersebut, tetapi versi belajar yang mengambil
konsep dasarnya lalu menerapkannya dengan gaya Go: dependency injection sederhana,
`database/sql`, Gin, validasi eksplisit, dan error handling yang jelas.

## Untuk developer framework PHP yang belajar Go

Jangan mencari padanan satu-per-satu untuk semua magic framework. Di Go, wiring
aplikasi biasanya terlihat jelas di source code. Peta belajarnya:

| Kebiasaan di PHP MVC | Cara berpikir di Go MVC ini |
|---|---|
| `index.php` sebagai front controller | `cmd/server/main.go` membuat server dan dependency |
| `Route::get()` | `api.GET()` pada file `internal/routes/*_routes.go` |
| `$this->request` | `*gin.Context` |
| `$this->input->post()` / `request()->json()` | `helper.DecodeJSON(ctx, &input)` |
| `Response::json()` | `helper.Success()` atau `helper.JSON()` |
| Controller class | struct controller dengan method `func(ctx *gin.Context)` |
| Model/Active Record | repository + `models` + `helper.ORM` |
| PDO query | `database/sql` atau `helper.ORM.Query()` |
| Query Builder CI3/CI4 | `Select().From().Where().Get()` |
| `result_array()` / `getResultArray()` | `GetResultArray()` |
| `row_array()` / `getRowArray()` | `GetRowArray()` |
| `$this->db->trans_start()` | `helper.WithTransaction()` |
| Middleware class | `gin.HandlerFunc` |
| `.env` / config PHP | `internal/app/config` + environment |

Urutan belajar yang disarankan:

1. Jalankan `/api/health`, `/api/echo`, dan `/api/route/:id` tanpa database.
2. Baca `internal/routes/api_routes.go` untuk melihat route → controller.
3. Baca `internal/controllers/item_controller.go` untuk melihat request dan validation.
4. Baca `internal/repositories/item_repository.go` untuk melihat model dan query.
5. Coba `internal/helper/orm.go` dengan query builder CI3/CI4.
6. Pelajari `internal/helper/transaction.go` untuk operasi multi-query yang atomic.
7. Setelah paham, baru pindahkan modul bisnis seperti Blog, Finance, atau Penjualan.

Perbedaan penting: Go tidak memakai inheritance controller/model sebagai pusat
framework. Dependency dikirim melalui constructor, error dikembalikan sebagai nilai,
dan query/database ditulis eksplisit supaya alurnya mudah di-debug.

## Jalankan dalam 3 menit

```bash
cp .env.example .env
go mod download
make dev
```

`make dev` memakai hot reload Air Go dari `$(go env GOPATH)/bin/air`. Instal sekali
jika belum ada:

```bash
go install github.com/air-verse/air@latest
```

Pastikan binary yang dipakai adalah Air Go. Beberapa sistem juga memiliki command
`air` untuk bahasa R; binary tersebut bukan hot-reloader Go yang dipakai project ini.

Setelah itu setiap perubahan file `.go` akan rebuild dan restart server otomatis.
Tanpa Air, `make dev` tetap menjalankan server dengan `go run` biasa.

File `.env` lokal dan collection Postman tersedia di folder `postman/`. Import
`Codekop Go MVC.postman_collection.json`, lalu pilih environment `Codekop Go MVC - Local`.

Tanpa konfigurasi database, endpoint stateless tetap bisa dicoba:

```bash
curl http://localhost:8080/api/health
curl -X POST http://localhost:8080/api/echo \
  -H 'Content-Type: application/json' \
  -d '{"nama":"Budi","role":"admin"}'
curl 'http://localhost:8080/api/route/ABC-123?source=php'
```

Semua response JSON memakai HTTP status dan field `code` yang sama:

```json
{
  "code": 200,
  "success": true,
  "data": { "message": "ok" }
}
```

Contoh validasi:

```json
{
  "code": 422,
  "success": false,
  "message": "Validation failed",
  "errors": { "name": ["field is required"] }
}
```

CRUD sample (setelah tabel `sample_items` dibuat):

```bash
curl http://localhost:8080/api/items
curl -X POST http://localhost:8080/api/items -H 'Content-Type: application/json' -d '{"name":"Keyboard","total":250000}'
curl -X PATCH http://localhost:8080/api/items/1 -H 'Content-Type: application/json' -d '{"total":225000}'
curl -X DELETE http://localhost:8080/api/items/1
```

Paging list:

```bash
curl 'http://localhost:8080/api/items?page=2&per_page=10'
curl 'http://localhost:8080/api/db/transaksi?page=1&per_page=20'
```

`limit` tetap diterima sebagai alias `per_page`. Response list memiliki `data` dan
`pagination` lengkap: `current_page`, `per_page`, `total`, `last_page`, `from`,
`to`, `next_page`, dan `prev_page`.

## Database MariaDB/MySQL atau PostgreSQL

Isi `.env` berikut untuk MariaDB/MySQL:

```dotenv
DB_DRIVER=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_SOCKET=
DB_NAME=codekop
DB_USER=root
DB_PASSWORD=secret
```

Untuk PostgreSQL, ubah `DB_DRIVER=postgres`, port menjadi `5432`, dan isi
`DB_SSLMODE` sesuai server. Jalankan migration dari folder `migrations/mysql` atau
`migrations/postgres`, lalu:

```bash
curl 'http://localhost:8080/api/db/transaksi?limit=20'
```

`DB_NAME` kosong sengaja didukung agar belajar routing/API tidak terhalang setup
database. Endpoint transaksi hanya contoh query sederhana (`id`, `no_trx`, `total`,
`created_at`), bukan implementasi aturan pembayaran. Jika database diisi tetapi
tidak dapat dihubungi, server berhenti saat startup supaya masalah konfigurasi
terlihat lebih awal.

## Mode development dan production

Development:

```bash
APP_ENV=development go run ./cmd/server
```

Production:

```bash
APP_ENV=production go build -buildvcs=false -trimpath -ldflags='-s -w' -o bin/gomvc ./cmd/server
APP_ENV=production ./bin/gomvc
```

Production menyembunyikan detail error internal, mengaktifkan header keamanan
lebih ketat, memakai timeout HTTP, dan tetap menulis log terstruktur JSON.
Development juga menampilkan log debug Gin; error database/panic dicatat detailnya
ke log, tetapi response API tetap hanya mengirim pesan aman.

## Peta konsep PHP → Go

| PHP MVC | Go MVC ini |
|---|---|
| `index.php` / front controller | `cmd/server/main.go` |
| `app/Config/Config.php` | `internal/app/config` + `.env` |
| `Route::get('/api/health', ...)` | `api.GET("/health", handler)` |
| Controller method | handler method `func(c *gin.Context)` |
| `Request::json()` | `helper.DecodeJSON` |
| `Response::json()` | `helper.JSON` / `helper.Success` |
| PDO | `database/sql` |
| Model | `internal/models/*.go` |
| Repository/Query | `internal/repositories/*.go` |
| `System\\Crud` | `internal/helper/crud.go` |
| Middleware | `gin.HandlerFunc` |
| `APP_ENV=development` | `APP_ENV=development` |

Alur request-nya tetap familiar:

```text
HTTP request → Gin middleware → `routes/` → `controllers/` → `services/` (opsional) → `repositories/` → database
```

Untuk modul baru, buat controller di `internal/controllers`, model di
`internal/models`, repository di `internal/repositories`, lalu daftarkan route di
`internal/routes/routes.go`. `internal/helper/crud.go` sudah menyediakan operasi
generic `List`, `Find`, `Insert`, `Update`, dan `Delete`; `validate.go` berisi
validasi dasar, `json.go` berisi response/decode JSON, dan `transaction.go`
menyediakan `Begin → Rollback/Commit` melalui helper `WithTransaction`.

Contoh transaksi database:

```go
err := helper.WithTransaction(ctx, db, func(tx *sql.Tx) error {
    _, err := tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, fromID)
    if err != nil {
        return err // otomatis rollback
    }
    _, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance + ? WHERE id = ?", amount, toID)
    return err // nil = commit, error = rollback
})
```

## Query builder gaya CodeIgniter

Helper ORM sengaja memakai nama yang dekat dengan CodeIgniter:

```go
db := helper.NewORM(sqlDB, "mysql")

result, err := db.Select("id, name, total").
    From("sample_items").
    Where("id", 1).
    OrderBy("id", "desc").
    Limit(20).
    Get(ctx)

rows := result.GetResultArray()
row := result.GetRowArray()
```

Untuk SQL langsung:

```go
result, err := db.Query(ctx, "SELECT * FROM sample_items WHERE total >= ?", 10000)
rows := result.GetResult()
row := result.GetRow()
```

`GetResult`, `GetResultArray`, `GetRow`, dan `GetRowArray` adalah alias agar
perpindahan dari CodeIgniter lebih mudah. Nilai tetap memakai parameter SQL;
nama tabel/kolom divalidasi oleh builder.

### Pemetaan query CodeIgniter 3/4

| CodeIgniter | Go MVC |
|---|---|
| `$this->db->select('id, name')` | `db.Select("id, name")` |
| `$this->db->from('sample_items')` | `.From("sample_items")` |
| `$this->db->where('id', $id)` | `.Where("id", id)` |
| `$this->db->order_by('id', 'DESC')` | `.OrderBy("id", "desc")` |
| `$this->db->limit(20)` | `.Limit(20)` |
| `$builder->get()` | `.Get(ctx)` |
| `$query->result()` | `result.GetResult()` |
| `$query->result_array()` | `result.GetResultArray()` |
| `$query->row()` | `result.GetRow()` |
| `$query->row_array()` | `result.GetRowArray()` |
| `$model->find($id)` | `db.Table("table").Find(ctx, id)` |

`Where()` dapat dipanggil beberapa kali dan akan digabung dengan `AND`:

```go
result, err := db.Select("id, name, total").
    From("sample_items").
    Where("name", "Keyboard").
    Where("total", 250000).
    Get(ctx)
if err != nil {
    return err
}
items := result.GetResultArray()
```

### Query satu row berdasarkan ID

```go
query := db.Table("sample_items")
item, err := query.Find(ctx, 1)
if errors.Is(err, sql.ErrNoRows) {
    // balas 404 dari controller
}
```

`Find()` memakai `WHERE id = ?` untuk MySQL/MariaDB dan placeholder PostgreSQL
yang sesuai. Identifier tabel/kolom tetap divalidasi, sedangkan nilai masuk sebagai
parameter database.

### Insert, update, delete

```go
id, err := db.Insert(ctx, "sample_items", map[string]any{
    "name": "Mouse",
    "total": 150000,
})

updated, err := db.Update(ctx, "sample_items",
    map[string]any{"total": 125000},
    map[string]any{"id": id},
)

deleted, err := db.Delete(ctx, "sample_items", map[string]any{"id": id})
```

`Update()` dan `Delete()` wajib menerima kondisi `where`, sehingga helper tidak
menjalankan perubahan massal secara tidak sengaja.

### Raw query

Gunakan `Query()` jika query terlalu spesifik untuk builder. Query harus berasal
dari source code, bukan gabungan string dari input user:

```go
// MySQL/MariaDB memakai ?.
result, err := db.Query(ctx,
    "SELECT id, name FROM sample_items WHERE total >= ?",
    10000,
)
if err != nil {
    return err
}
rows := result.GetResultArray()
```

Untuk PostgreSQL gunakan `$1`, `$2`, dan set `DB_DRIVER=postgres`.

### Model, repository, controller

Pola modulnya mirip aplikasi PHP:

```text
routes/items_routes.go
    -> controllers.ItemController
        -> repositories.ItemRepository
            -> models.Item
            -> helper.Helper / helper.ORM
                -> database/sql
```

Controller menangani HTTP dan validasi. Repository menangani query. Model menjadi
bentuk data response. `helper.ORM` cocok untuk query builder fleksibel, sedangkan
`helper.Helper` cocok untuk CRUD generic dengan allow-list kolom.

## Test dan build

```bash
make fmt
make test
make vet
make build
```

Go menggunakan `go test` sebagai test runner bawaan dan tidak membutuhkan
framework test tambahan untuk memulai belajar. Penjelasan struktur dan aturan
perubahan ada di [`AGENTS.md`](AGENTS.md), test API ada di [`testing/`](testing/),
sedangkan SQL ada di [`migrations/README.md`](migrations/README.md).
