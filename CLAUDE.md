# spotify

Module Go `go.avagenc.com/spotify` — library untuk integrasi Spotify Web API.
Digunakan oleh `go.avagenc.com/ava` sebagai dependency langsung (modular monolith), bukan microservice.

## Status

Sedang dalam migrasi dari microservice (HTTP server + Lambda) ke pure library.
File `main.go`, `handlers/`, dan dependency Lambda akan dihapus.
Target struktur: semua public API di root package `package spotify`.

## Struktur

```
spotify/        root package, semua public API di sini
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
