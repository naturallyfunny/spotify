# spotify

Module Go `go.naturallyfunny.dev/spotify` — reusable public library untuk integrasi Spotify Web API.
Dirancang sebagai interface-based library agar dapat dipakai lintas project, tidak terikat ke satu database atau satu aplikasi.

## Struktur

```
client.go       Client, TokenStore interface, tipe Track/Playlist/Device
search.go       SearchTracks, SearchPlaylists, UserPlaylists, PlaylistTracks
player.go       Devices, Play, Pause, Resume, SetVolume
postgres/
  store.go      implementasi TokenStore di atas pgxpool, Migrate()
  migrations/   SQL files, di-embed via //go:embed
```

## Cara Pakai

```go
auth := spotifyauth.New(
    spotifyauth.WithClientID(clientID),
    spotifyauth.WithClientSecret(clientSecret),
)
store := postgres.New(pool, dsn)
client := spotify.New(store, auth)

tracks, err := client.SearchTracks(ctx, userID, "Queen")
err = client.Play(ctx, userID, deviceID, trackURI)
```

## Migrations

Pakai `golang-migrate`. Naming: `000N_deskripsi.up.sql` / `000N_deskripsi.down.sql`.
Semua statement wajib pakai `IF NOT EXISTS` / `IF EXISTS`. Jangan pernah edit migration yang sudah di-commit.

## Previous Session

Migrasi penuh dari microservice ke reusable public library:

- Rename module dari `spotify-api` → `go.avagenc.com/spotify` → `go.naturallyfunny.dev/spotify`
- Hapus seluruh layer HTTP: `main.go`, `handlers/`, Lambda artifacts, `api_documentation.md`
- Buat migrations `000001_init`, `000002_drop_device_id`, `000003_drop_gmail` — `device_id` dan `gmail` dihapus karena tidak relevan dengan tanggung jawab modul ini
- Buat `postgres/` package dengan `TokenStore` interface dan `Store` yang mengimplementasinya
- Ganti implementasi HTTP manual ke Spotify dengan `zmb3/spotify/v2`
- `Client` menerima `TokenStore` dan `*spotifyauth.Authenticator` sebagai dependency injection — tidak membangun credential sendiri
- `clientFor` membuat zmb3 client per-request per-user via oauth2 refresh token flow
- `postgres.New` menerima DSN eksplisit agar `Migrate()` tidak rekonstruksi DSN dari `ConnString()` yang fragile

## Conventions

- `TokenStore` interface didefinisikan di `client.go` — consumer defined interface
- `postgres.New(pool, dsn)` — DSN disimpan di Store untuk kebutuhan Migrate()
- Tidak ada `pkg/` — flat structure
- Commit message pakai conventional commits: `feat:`, `fix:`, `chore(migrate):` dst
