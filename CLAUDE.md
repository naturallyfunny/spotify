# spotify

Module Go `go.naturallyfunny.dev/spotify` — reusable public library untuk integrasi Spotify Web API.
Dirancang sebagai interface-based library agar dapat dipakai lintas project, tidak terikat ke satu database atau satu aplikasi.

## Struktur

```
client.go       Client, TokenStore interface, tipe Track/Playlist/Device
search.go       SearchTracks, SearchPlaylists, UserPlaylists, PlaylistTracks, Recommendations
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

## Conventions

- `TokenStore` interface didefinisikan di `client.go` — consumer defined interface
- `postgres.New(pool, dsn)` — DSN disimpan di Store untuk kebutuhan Migrate()
- Tidak ada `pkg/` — flat structure
- Commit message pakai conventional commits: `feat:`, `fix:`, `chore(migrate):` dst
