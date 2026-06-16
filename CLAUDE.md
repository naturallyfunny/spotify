# spotify

Module Go `go.avagenc.com/spotify` — library untuk integrasi Spotify Web API.
Digunakan oleh `go.avagenc.com/ava` sebagai dependency langsung (modular monolith), bukan microservice.

## Status

Layer microservice sudah dihapus (main.go, handlers/, Lambda artifacts).
Restrukturisasi package belum selesai — isi `spotify/` perlu dinaikkan ke root,
dan `db/` perlu dipindah ke `internal/db/`.

## Struktur

Saat ini:
```
spotify/        package spotify (library, player, client) — belum di root
db/             koneksi PostgreSQL — belum jadi internal
migrations/     SQL files
```

Target:
```
spotify.go      root package, public API (Client, konstruktor)
library.go
player.go
internal/db/    koneksi PostgreSQL, tidak diekspos keluar
migrations/     SQL files, di-embed ke binary via //go:embed
```

## Migrations

Pakai `golang-migrate`. Naming: `000N_deskripsi.up.sql` / `000N_deskripsi.down.sql`.
Semua statement wajib pakai `IF NOT EXISTS` / `IF EXISTS`. Jangan pernah edit migration yang sudah di-commit.

## Conventions

- Public API: method pada `spotify.Client`
- Tidak ada `pkg/` — flat structure
- Commit message pakai conventional commits: `feat:`, `fix:`, `chore(migrate):` dst
