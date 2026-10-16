# Deploy UnderStud ke Server

Panduan menyalakan UnderStud di server yang sudah menjalankan **infra bersama**
(compose di `~/infra`, Cloudflare Tunnel bersama, dan Postgres + pgbouncer milik sendiri).

Image dibangun otomatis oleh GitHub Actions dan dipublikasikan ke **GHCR**
(`ghcr.io/fzrilsh/understud`), jadi server cukup `pull` tanpa `build`.
Frontend sudah **ter-embed** di dalam binary Go — hanya ada **satu** container.

Ringkas alurnya:

```
push ke main  ->  GitHub Actions build image  ->  push ke GHCR
                                                      |
                                        server: docker compose pull
                                                      |
                                    ~/infra/docker-compose.yml (service understud)
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

## Langkah 1: Bikin Database UnderStud

Jalankan di server. Ganti password dengan yang kuat.

```bash
# Nama superuser Postgres-mu ada di ~/infra/.env (POSTGRES_USER).
# Jangan asumsikan "postgres" — banyak setup pakai nama lain.
docker exec infra-postgres-1 env | grep POSTGRES_USER
# pakai nilai itu sebagai -U di bawah:

# Masuk ke Postgres
docker exec -it infra-postgres-1 psql -U <POSTGRES_USER> -d postgres

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

Di server, file deploy ada di `~/understud-deploy/` (hasil `git clone`/scp folder
`deploy/` dari repo). `.env` disimpan terpisah di `~/services/understud/`.

> Catatan: user server adalah `deploy`, jadi `~` = `/home/deploy`. Karena itu
> `env_file` di compose ditulis absolut: `/home/deploy/services/understud/.env`.

```bash
mkdir -p ~/services/understud
cp ~/understud-deploy/.env.prod.example ~/services/understud/.env
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

Buka `~/infra/docker-compose.yml` dan tempel service `understud` dari
`~/understud-deploy/compose.service.yml` ke dalam key `services:`.

Sesuaikan dua hal:

```yaml
  understud:
    networks: [internal]                            # <- nama network milikmu
    env_file: /home/deploy/services/understud/.env   # <- path .env kamu
```

**Siapkan folder data lampiran** (bind mount). Container jalan sebagai non-root
**UID 10001**, jadi folder host harus dimiliki UID itu — kalau tidak, lampiran
chat gagal disimpan (`permission denied`):

```bash
sudo mkdir -p /home/deploy/understud_uploads
sudo chown -R 10001:10001 /home/deploy/understud_uploads
```

> `mkdir` biasa (tanpa sudo) membuat folder milik user `deploy` (biasanya UID 1000),
> sehingga container UID 10001 **tidak bisa menulis**. `chown` di atas wajib.

> **Tidak ada `ports`.** cloudflared (container di network `internal` yang sama)
> meneruskan langsung ke `http://understud:8080` lewat nama service — sama
> seperti service lain di server ini. Tidak ada port yang terekspos ke host.

---

## Langkah 5: Jalankan Migrasi

Migrasi dijalankan **langsung ke Postgres** (bukan pgbouncer), karena goose butuh
sesi yang stabil.

```bash
cd ~/infra
DIRECT_DATABASE_URL='postgres://understud:PASSWORD@postgres:5432/understud?sslmode=disable' \
  bash ~/understud-deploy/migrate.sh
```

Script itu menjalankan container sekali pakai di network yang sama, menjalankan
`cli migrate`, lalu menghapus container itu. Kalau berhasil akan muncul
`OK — migrasi selesai.`

---

## Langkah 6: Nyalakan Service

```bash
cd ~/infra
docker compose pull understud
docker compose up -d understud
```

Cek:

```bash
docker compose ps understud
docker compose logs --tail=30 understud
```

Di log harus muncul `api listening` dan `migrasi otomatis dimatikan (MIGRATE_ON_START=false)`.

Tes dari server (lewat network docker, tanpa port host):

```bash
docker run --rm --network internal curlimages/curl:latest -s http://understud:8080/healthz
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
    service: http://understud:8080
```

> Karena cloudflared jalan sebagai container di network yang sama, pakai
> **nama service** (`understud:8080`) — bukan `localhost`/port host.

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
docker compose ps understud
docker compose logs --tail=30 understud
curl -I https://understud.fazrilsh.com
```

Terakhir, buka **https://understud.fazrilsh.com** di browser.

---

## Bikin Akun Staf

Setelah aplikasi jalan, bikin akun asli (jangan pakai akun demo di produksi):

```bash
cd ~/infra
docker compose exec understud /app/cli create-staff \
  --email guru@sekolah.sch.id --password 'password-kuat' \
  --name 'Ibu Rina' --role BK --kelas ''
```

Nilai `--role`: `BK` (Guru BK, semua kelas) atau `WALI_KELAS` (wajib isi
`--kelas`, mis. `--kelas 'X IPA 1'`).

Data demo (opsional, hanya untuk uji coba):

```bash
docker compose exec understud /app/cli seed-demo
```

Seed mengisi 8 kelas SMA (IPA/IPS) × 8 siswa + observasi, 8 wali kelas, 2 guru BK,
dan 12 chat anonim.

Akun demo (password semua: `password123`):

| Email | Peran |
|---|---|
| `bk@demo.local` | Guru BK |
| `bk2@demo.local` | Guru BK (kedua) |
| `wali@demo.local` | Wali Kelas — X IPA 1 |
| `wali.ipa2@demo.local` | Wali Kelas — X IPA 2 |
| `wali.ips1@demo.local` | Wali Kelas — X IPS 1 |
| `wali.xi.ipa1@demo.local` | Wali Kelas — XI IPA 1 |
| `wali.xi.ips1@demo.local` | Wali Kelas — XI IPS 1 |

---

## Cara Update

Setiap push ke `main` yang menyentuh `backend/`, `frontend/`, atau Dockerfile akan
membangun image baru di GitHub Actions. Setelah selesai (cek tab **Actions**):

```bash
cd ~/infra
docker compose pull understud
docker compose up -d --no-deps understud
docker image prune -f
```

Kalau ada migrasi baru (ada file baru di `backend/internal/db/migrations/`),
jalankan Langkah 5 dulu sebelum `up -d`.
