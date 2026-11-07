# Testing

Folder ini berisi test dari sudut pandang pengguna API, bukan test detail
internal package.

```bash
go test ./...
```

`api_test.go` menggunakan `httptest`, jadi tidak membuka port dan tidak
memerlukan MariaDB/PostgreSQL. Untuk mencoba CRUD database sungguhan:

1. Jalankan salah satu migration di folder `migrations/`.
2. Isi `DB_NAME`, `DB_USER`, dan `DB_PASSWORD` pada `.env`.
3. Jalankan server dengan `make dev`.
4. Jalankan curl CRUD yang ada di `README.md`.
