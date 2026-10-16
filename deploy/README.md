# Deploy Understud ke Server

Panduan menyalakan Understud di server yang sudah menjalankan **infra bersama**
(compose di `~/infra`, Cloudflare Tunnel bersama, dan Postgres + pgbouncer milik sendiri).

Image dibangun otomatis oleh GitHub Actions dan dipublikasikan ke **GHCR** sebagai
paket publik, jadi server cukup `pull` tanpa `build`.

Ringkas alurnya:

```
push ke main  ->  GitHub Actions build image  ->  push ke GHCR
                                                      |
                                        server: docker compose pull
                                                      |
                                    ~/infra/docker-compose.yml (service understud-*)
                                                      |
                                     cloudflared -> understud.fazrilsh.com
```

**Tidak ada auto-deploy via SSH.** Deploy ke server dijalankan manual (3 perintah).

---

## Isi Folder Ini

| File | Isi |
|---|---|
| `compose.service.yml` | Fragmen service untuk ditempel ke `~/infra/docker-compose.yml` |
| `.env.prod.example` | Template konfigurasi (diisi manual di server) |
| `migrate.sh` | Jalankan migrasi database (karena migrasi otomatis dimatikan) |
| `README.md` | Dokumen ini |

---

## Yang Dibutuhkan di Server

- Docker + Docker Compose (sudah ada, karena infra bersama jalan).
- Postgres + pgbouncer (sudah ada).
- Cloudflare Tunnel (sudah ada).
- Network docker internal yang dipakai service lain (mis. `internal`).

---

## Langkah 1: Bikin Database Understud

Jalankan di server. Ganti password dengan yang kuat.

```bash
# Masuk ke Postgres
docker exec -it <nama-container-postgres> psql -U postgres

CREATE DATABASE understud;
CREATE USER understud WITH PASSWORD 'PASSWORD_KUAT_DI_SINI';
GRANT ALL PRIVILEGES ON DATABASE understud TO understud;
\c understud
GRANT ALL ON SCHEMA public TO understud;
ALTER SCHEMA public OWNER TO understud;
\q
```

> Migrasi goose perlu membuat tabel, jadi user ini butuh hak di schema `public`.

---

## Langkah 2: Daftarkan di pgbouncer

Tambahkan entry baru di `pgbouncer.ini` dan `userlist.txt`.

**`pgbouncer.ini`** (bagian `[databases]`):

```ini
understud = host=postgres port=5432 dbname=understud
```

**`userlist.txt`**: hash md5 yang sama dengan role Postgres. Bikin hash-nya:

```bash
echo -n "PASSWORD_KUAT_DI_SINI" | md5sum | awk '{print "md5"$1}'
```

Lalu pastikan role di Postgres memakai hash yang sama:

```sql
ALTER ROLE understud WITH PASSWORD 'md5<hash-yang-sama>';
```

Tambahkan ke `userlist.txt`:

```txt
"understud" "md5<hash-yang-sama>"
```

Restart pgbouncer setelah selesai.

---

## Langkah 3: Isi Konfigurasi

```bash
mkdir -p ~/services/understud
cp <repo>/deploy/.env.prod.example ~/services/understud/.env
nano ~/services/understud/.env
chmod 600 ~/services/understud/.env
```

Yang **wajib** diisi:

| Variabel | Isi |
|---|---|
| `DATABASE_URL` | Lewat pgbouncer, dengan `default_query_exec_mode=simple_protocol` |
| `JWT_SECRET` | `openssl rand -hex 32` (minimal 32 karakter) |
| `CORS_ORIGINS` | `https://understud.fazrilsh.com` (sudah default) |
| `COOKIE_SECURE` | `true` (sudah default) |
| `TRUST_PROXY_HEADERS` | `true` (sudah default) |

Catatan penting:

- `MIGRATE_ON_START=false` (sudah default). Aplikasi **tidak** menjalankan migrasi
  sendiri, karena di balik pgbouncer mode transaction itu bisa gagal. Migrasi
  dijalankan terpisah di Langkah 5.
- Kalau password mengandung karakter khusus (`@ : / ? # & =`), **URL-encode** dulu.
- `AI_BASE_URL` jangan diisi `localhost` (api jalan di container lain). Kosongkan
  saja kalau belum ada provider; fitur ringkasan AI nonaktif dan aplikasi tetap jalan.

---

## Langkah 4: Tambah Service ke `~/infra/docker-compose.yml`

Buka `~/infra/docker-compose.yml` dan tempel dua service dari
`deploy/compose.service.yml` ke dalam key `services:`.

Sesuaikan tiga hal:

```yaml
  understud-api:
    networks: [internal]                            # <- nama network milikmu
    env_file: /home/deploy/services/understud/.env   # <- path .env kamu

  understud-web:
    ports:
      - "127.0.0.1:3010:80"                         # <- ganti kalau 3010 dipakai
```

Port `3010` di sini hanya didengarkan di `127.0.0.1` (tidak terbuka ke internet);
cloudflared yang meneruskan ke situ.

---

## Langkah 5: Jalankan Migrasi

Migrasi dijalankan **langsung ke Postgres** (bukan pgbouncer), karena goose butuh
sesi yang stabil.

```bash
cd ~/infra
bash <repo>/deploy/migrate.sh
```

Script butuh `DIRECT_DATABASE_URL` yang menunjuk langsung ke Postgres:

```bash
DIRECT_DATABASE_URL='postgres://understud:PASSWORD@postgres:5432/understud?sslmode=disable' \
  bash <repo>/deploy/migrate.sh
```

Perintah itu menjalankan container sekali pakai di network yang sama, menjalankan
`cli migrate`, lalu menghapus container itu. Kalau berhasil akan muncul
`OK — migrasi database selesai.`

---

## Langkah 6: Nyalakan Service

```bash
cd ~/infra
docker compose pull understud-api understud-web
docker compose up -d understud-api understud-web
```

Cek:

```bash
docker compose ps understud-api understud-web
docker compose logs --tail=30 understud-api
```

Di log harus muncul `api listening` dan `migrasi otomatis dimatikan (MIGRATE_ON_START=false)`.

Tes dari server:

```bash
curl -s http://127.0.0.1:3010/healthz
# harus: {"data":{"db":"up","status":"ok"}}
```

> Kalau `db` bukan `up`, cek `DATABASE_URL` (password, host pgbouncer, dan
> `default_query_exec_mode=simple_protocol`).

---

## Langkah 7: Daftarkan ke Cloudflare Tunnel

Buka `~/infra/cloudflared/config.yml` dengan `sudo nano`, lalu tambahkan ingress
**di atas** baris `http_status:404`:

```yaml
  - hostname: understud.fazrilsh.com
    service: http://understud-web:80
```

> Perhatikan: karena cloudflared jalan sebagai container di network yang sama,
> pakai **nama service** (`understud-web:80`), bukan `127.0.0.1:3010`.

Lalu daftarkan DNS route:

```bash
docker run -it --rm \
  -v ~/infra/cloudflared:/home/nonroot/.cloudflared \
  cloudflare/cloudflared:latest tunnel route dns backend-tunnel understud.fazrilsh.com
```

> Ganti `backend-tunnel` dengan nama tunnel milikmu (lihat `tunnel:` di `config.yml`).

Restart cloudflared:

```bash
cd ~/infra && docker compose restart cloudflared
```

---

## Langkah 8: Verifikasi

```bash
docker compose ps understud-api understud-web
docker compose logs --tail=30 understud-api
curl -I https://understud.fazrilsh.com
```

Terakhir, buka **https://understud.fazrilsh.com** di browser.

---

## Bikin Akun Staf

Setelah aplikasi jalan, bikin akun asli (jangan pakai akun demo di produksi):

```bash
cd ~/infra
docker compose exec understud-api /app/cli create-staff \
  --email guru@sekolah.sch.id --password 'password-kuat' \
  --name 'Ibu Rina' --role BK --kelas ''
```

Nilai `--role`: `BK` (Guru BK, semua kelas) atau `WALI_KELAS` (wajib isi
`--kelas`, mis. `--kelas 'X TJKT 1'`).

Data demo (opsional, hanya untuk uji coba):

```bash
docker compose exec understud-api /app/cli seed-demo
```

---

## Cara Update

Setiap push ke `main` yang menyentuh `backend/` atau `frontend/` akan membangun
image baru di GitHub Actions. Setelah selesai (cek tab **Actions**):

```bash
cd ~/infra
docker compose pull understud-api understud-web
docker compose up -d --no-deps understud-api understud-web
docker image prune -f
```

Kalau ada migrasi baru (ada file baru di `backend/internal/db/migrations/`),
jalankan Langkah 5 dulu sebelum `up -d`.

Untuk memakai versi tertentu (bukan `latest`), ubah tag di
`~/infra/docker-compose.yml`:

```yaml
image: ghcr.io/fzrilsh/understud-api:sha-935a557
```

---

## Lampiran: Supaya File Upload Bertahan

Secara default lampiran disimpan di dalam container `understud-api`. Kalau
container di-recreate, lampiran ikut hilang.

Supaya bertahan, tambahkan volume di service `understud-api`:

```yaml
  understud-api:
    volumes:
      - understud_uploads:/app/data
```

lalu daftarkan volume di bagian `volumes:` paling bawah compose:

```yaml
volumes:
  understud_uploads:
```

> Catatan: container jalan sebagai user `appuser` (UID 10001). Kalau memakai bind
> mount (folder host, bukan named volume), `chown 10001:10001` foldernya dulu.

---

## Kalau Ada Masalah

**`db` di healthz bukan `up`**
Cek `DATABASE_URL`: host pgbouncer, password, dan pastikan ada
`default_query_exec_mode=simple_protocol`. Cek juga entry pgbouncer sudah
ditambahkan (Langkah 2).

**Aplikasi gagal start dengan error migrasi**
`MIGRATE_ON_START` masih `true`. Set `false` di `.env`, karena di balik
pgbouncer migrasi otomatis memang tidak dijalankan.

**Login staf gagal terus / selalu logout**
`JWT_SECRET` kosong atau berubah tiap recreate. Pastikan terisi dan tetap.

**Cookie tidak tersimpan / login gagal padahal password benar**
`COOKIE_SECURE=true` tapi diakses lewat HTTP. Akses lewat domain HTTPS.

**Upload lampiran gagal**
Kalau ada proxy lain di depan, pastikan `client_max_body_size` minimal `25m`.
Di dalam container web sudah di-set 25m.

**Halaman domain balas 502**
Cek ingress cloudflared: `service:` harus menunjuk `http://understud-web:80`
(nama service), bukan `127.0.0.1:3010`.

**Image gagal ditarik (`denied`)**
Paket GHCR harus publik. Cek
`https://github.com/users/fzrilsh/packages/container/understud-api/settings`.
