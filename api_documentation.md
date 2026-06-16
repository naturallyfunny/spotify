# REST API Documentation: Spotify AI Agent Backend

Alat ini adalah backend antarmuka berbasis Agent untuk Spotify Web API, didesain agar siap digunakan pada serverless architecture (AWS Lambda) secara *stateless*.

## Base URL
```text
(Sesuai domain API Gateway AWS Lambda Anda / localhost:8080)
```

## Autentikasi
Aplikasi tidak menggunakan Authentication Header standar seperti Bearer Token. Sebagai gantinya, **setiap endpoint mewajibkan Request Header `x-user-id`**. Aplikasi akan menggunakan `x-user-id` ini untuk mengambil `refresh_token` Spotify dari database PostgreSQL (Neon DB), kemudian menukarkannya dengan `access_token` Spotify JWT berumur pendek (1 jam) secara *on-the-fly*.

---

## 1. Endpoint: Control Player
Mengontrol playback Spotify pengguna, termasuk memutar musik, mengendalikan status putar/jeda, dan volume.

- **URL:** `/play`
- **Method:** `POST`
- **Content-Type:** `application/json`

### Request Parameters

Selain body JSON di bawah, Anda wajib menyertakan **Header: `x-user-id`**.

| Parameter | Tipe Data | Wajib | Deskripsi |
| :--- | :--- | :---: | :--- |
| `device_id` | `string` | Tidak | ID perangkat Spotify. Jika tidak diisi, sistem akan otomatis mengambil ID dari perangkat pertama yang sedang aktif/tersedia milik pengguna. |
| `play_mu` | `string` | Tidak | Spotify URI untuk media yang ingin diputar. Contoh: `spotify:track:4cOdK2wGLETKBW3PvgPWqT`. |
| `command` | `string` | Tidak | Mengatur perintah jeda atau putar. Nilai yang diterima: `"pause"` atau `"resume"`. |
| `volume` | `integer` | Tidak | Presentase volume pemutar musik (nilai `0` - `100`). |

> [!NOTE]
> Parameter `play_mu`, `command`, dan `volume` bersifat independen dan saling melengkapi. Anda dapat memasukkan salah satu, dua, atau ketiganya sekaligus. Jika lebih dari satu dimasukkan, backend akan memprosesnya berbarengan *(concurrently)*.

### Processing Logic
1. **Validasi & Auth:** Server memvalidasi keberadaan header `x-user-id`, mengambil `refresh_token` dari database, lalu menarik `access_token` melalui endpoint Spotify Token.
2. **Device Fallback:** Apabila parameter `device_id` kosong, server akan mengambil daftar device dari Spotify dan menggunakan device pada antrean index pertama.
3. **Concurrency:** Menggunakan Goroutines, server mengeksekusi request `play_mu`, `command`, dan `volume` secara asinkron tanpa memblokir baris kode lainnya. Mutex digunakan untuk me-record status respons Spotify tiap command ke dalam array output balasan.

### Contoh Response (Success)
Seluruh balasan dari API ini dibungkus menggunakan Base Response dengan field utama `success`, `code`, `message`, dan `data`.

```json
{
  "success": true,
  "code": 0,
  "message": "Play command processed successfully",
  "data": {
    "success": true,
    "output": [
      "Now playing: spotify:track:4cOdK2wGLETKBW3PvgPWqT",
      "Playback resumed",
      "Volume set to 50%"
    ]
  }
}
```

---

## 2. Endpoint: Get Music & Library Data
Mengambil data musik dari pustaka Spotify atau mencari data lagu berdasarkan rekomendasi AI agen. Endpoint ini didesain se-paralel mungkin untuk menghemat waktu komputasi API Agent.

- **URL:** `/get-music`
- **Method:** `GET` / `POST` (Mendukung GET dengan Body atau standar POST)
- **Content-Type:** `application/json`

### Request Parameters

Selain body JSON di bawah, Anda wajib menyertakan **Header: `x-user-id`**.

| Parameter | Tipe Data | Wajib | Deskripsi |
| :--- | :--- | :---: | :--- |
| `query` | `string` | Tidak | String pencarian untuk menemukan daftar **Tracks/Lagu**. |
| `playlist_query` | `string` | Tidak | String pencarian untuk menemukan daftar **Playlist**. |
| `playlist_track` | `string` | Tidak | ID playlist (bukan URI), digunakan untuk menarik seluruh lagu yang ada di dalam playlist tersebut. |
| `genre_recomendation` | `string[]` | Tidak | Array daftar genre ("pop", "acoustic", dll) untuk menghasilkan lag-lagu rekomendasi Spotify yang relevan. |

### Processing Logic
Sistem menggunakan `sync.WaitGroup` untuk langsung memulai **sampai dengan 7 HTTP Request asinkron (Goroutines) ke Spotify Server** dalam satu waktu:

**Proses Mandatory (Wajib berjalan bagaimanapun parameter yang dikirim):**
1. Get Device List
2. Get User Playlists
3. Get Genre Seeds

**Proses Opsional (Hanya berjalan apabila parameternya tidak kosongan):**
4. Search Track (apabila `query` ada isinya)
5. Search Playlist (apabila `playlist_query` ada isinya)
6. Get Tracks From Playlist (apabila `playlist_track` ada isinya)
7. Get Recommendations (apabila `genre_recomendation` array tidak kosong)

Sistem akan menunggu seluruh sub-request selesai dieksekusi kurang dari batas aman limitasi dan langsung mengkonsolidasikan data balasan HTTP menjadi satu Body JSON utuh. 

### Contoh Response (Success)
Seluruh balasan dari API ini dibungkus menggunakan Base Response dengan field utama `success`, `code`, `message`, dan `data`. 
*Catatan: Objek di dalam data akan dihilangkan (`omitempty`) secara otomatis jika Anda tidak meminta parameternya di request.*

```json
{
  "success": true,
  "code": 0,
  "message": "Get music processed successfully",
  "data": {
    "success": true,
    "devices": [
      {
        "id": "1abc2def3ghi4jkl5mno6pqr",
        "is_active": true,
        "is_restricted": false,
        "name": "Iphone Sebastian",
        "type": "Smartphone",
        "volume_percent": 80
      }
    ],
    "user_playlist": {
      "items": [
        {
          "id": "37i9dQZF1DXcBWIGoYBM5M",
          "name": "Today's Top Hits",
          "description": "...",
          "tracks": {
            "total": 50
          },
          "external_urls": {}
        }
      ],
      "total": 1
    },
    "genre_seed": [
      "acoustic", "afrobeat", "alt-rock", "pop"
    ],
    "output": [
      "Found 1 device(s)",
      "Found 1 playlist(s)",
      "Found 120 genre seed(s)",
      "Track search for 'Queen': 10 result(s)"
    ],
    "search_results": { ... } // Hasil jika 'query' dikirim
  }
}
```

---

## Error Code & Penanganan Troubleshooting
Untuk kedua rute ini, kegagalan request ditandai dengan field `"success": false` beserta `"code"` gRPC style:

* **Code 3 (InvalidArgument):** Muncul jika format JSON *body* berantakan, parameter tidak valid, atau metode salah.
* **Code 16 (Unauthenticated):** Menandakan header `x-user-id` kosong, user tidak ditemukan di database, token Spotify telah expire/dicabut oleh user, atau gagal ditukar.
* **Code 5 (NotFound):** Muncul apabila (misalnya di endpoint `/play`) tidak ada device Spotify satupun yang aktif.
* **Code 13 (Internal):** Kegagalan saat sistem mencoba menghubungi Spotify secara masif (misal: Rate Limit).
