# Migrations

Pilih satu folder sesuai database:

- `mysql/001_init.sql` untuk MariaDB/MySQL.
- `postgres/001_init.sql` untuk PostgreSQL.

Migration membuat dua tabel sederhana untuk belajar:

- `penjualan`: dipakai endpoint read-only `/api/db/transaksi`.
- `sample_items`: dipakai CRUD `/api/items`.

Tidak ada aturan bisnis pembayaran di sini. Tabel sengaja kecil agar fokus belajar
alur request, query parameter, prepared statement, dan CRUD.

Contoh menjalankan:

```bash
mariadb -uroot -p codekop < migrations/mysql/001_init.sql
psql -U postgres -d codekop -f migrations/postgres/001_init.sql
```
