# spotify

Module Go `go.naturallyfunny.dev/spotify` — reusable public library untuk integrasi Spotify Web API.
Dirancang sebagai interface-based library agar dapat dipakai lintas project, tidak terikat ke satu database atau satu aplikasi.

## Status

Layer microservice sudah dihapus (main.go, handlers/, Lambda artifacts).
Restrukturisasi package belum selesai:
- isi `spotify/` perlu dinaikkan ke root
- `db/` perlu diubah menjadi `postgres/` dengan implementasi `TokenStore`
- `migrations/` dipindah ke dalam `postgres/`
- `store.go` (interface `TokenStore`) perlu dibuat di root

## Struktur

Saat ini:
```
spotify/        package spotify (library, player, client) — belum di root
db/             koneksi PostgreSQL — belum direstrukturisasi
migrations/     SQL files — belum dipindah ke postgres/
```

Target:
```
spotify.go      root package, Client dan NewClient(store TokenStore)
library.go
player.go
store.go        interface TokenStore { GetRefreshToken(...) }
postgres/
  postgres.go   implementasi TokenStore pakai pgxpool
  migrations/   SQL files, di-embed via //go:embed
```

## Migrations

Pakai `golang-migrate`. Naming: `000N_deskripsi.up.sql` / `000N_deskripsi.down.sql`.
Semua statement wajib pakai `IF NOT EXISTS` / `IF EXISTS`. Jangan pernah edit migration yang sudah di-commit.

## Conventions

- Public API: method pada `spotify.Client`
- Storage: injeksi via `TokenStore` interface, bukan hardcode ke postgres
- Tidak ada `pkg/` — flat structure
- Commit message pakai conventional commits: `feat:`, `fix:`, `chore(migrate):` dst
