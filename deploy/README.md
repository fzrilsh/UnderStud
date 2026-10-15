# Deploy UnderStud ke Server (VPS)

Panduan menyalakan UnderStud di server pribadi, dengan domain
`understud.fazrilsh.com` di belakang Cloudflare.

Image dibangun otomatis oleh GitHub Actions dan dipublikasikan ke **GHCR**
(GitHub Container Registry) sebagai paket **publik**, jadi server cukup
`pull` tanpa perlu `build` sama sekali.

Ringkas: **2 langkah** (set Docker, jalankan script). Sisanya sudah diurus.

---

## Arsitektur Deployment

```
Internet
   │  HTTPS (Cloudflare)
   ▼
Cloudflare  ──►  reverse proxy / tunnel di server
                        │  http://127.0.0.1:3000
                        ▼
                 ┌─────────────────────────────┐
                 │  web (nginx, SPA + /api)    │  ← satu-satunya yang terbuka
                 └──────────────┬──────────────┘
                                │  jaringan internal docker
                 ┌──────────────▼──────────────┐
                 │  api (Go)                   │  ← tidak dipublikasikan
                 └──────────────┬──────────────┘
                 ┌──────────────▼──────────────┐
                 │  db (PostgreSQL 17)         │  ← tidak dipublikasikan
                 └─────────────────────────────┘
```

Yang penting: **hanya `web`** yang mendengarkan di server (dan hanya di
`127.0.0.1`). Database dan API tidak bisa dijangkau dari internet.

---

## Yang Dibutuhkan

- Server Linux (VPS) dengan akses root/sudo. Arsitektur **x86_64** (amd64).
- **Docker** dan **Docker Compose** terpasang.
- **Cloudflare** mengelola domain `fazrilsh.com`.
- (Kalau mau AI) URL + API key provider OpenAI-compatible.

---

## Langkah 1: Pasang Docker

Kalau Docker belum ada:

```bash
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER    # supaya tak perlu sudo lagi (login ulang)
```

Pastikan:

```bash
docker --version
docker-compose --version    # atau: docker compose version
```

> Catatan: kalau server hanya punya plugin `docker compose` (dengan spasi),
> script deploy tetap jalan. Keduanya didukung.

---

## Langkah 2: Ambil project & jalankan

```bash
# Ambil hanya file yang dibutuhkan untuk deploy
git clone https://github.com/fzrilsh/UnderStud.git understud
cd understud
git checkout main
```

Jalankan script deploy:

```bash
bash deploy/deploy.sh
```

Yang script lakukan otomatis:

1. Cek Docker & Compose.
2. Buat `.env.prod` dari template, lalu **mengisi `JWT_SECRET` dan
   `POSTGRES_PASSWORD` dengan nilai acak** (via `openssl`).
3. Tarik image `ghcr.io/fzrilsh/understud-api:latest` dan
   `understud-web:latest`.
4. Jalankan stack.
5. Tunggu sampai API sehat, lalu tampilkan ringkasan.

Kalau berhasil, aplikasi sudah jalan di `http://127.0.0.1:3000` di server.

**Isi data demo** (opsional, sekali saja):

```bash
bash deploy/deploy.sh --seed
```

**Bikin akun staf asli** (bukan akun demo):

```bash
docker-compose -f docker-compose.prod.yml --env-file .env.prod \
  exec api /app/cli create-staff \
  --email guru@sekolah.sch.id --password 'password-kuat' \
  --name 'Ibu Rina' --role BK --kelas ''
```

Nilai `--role`: `BK` (Guru BK, akses semua kelas) atau `WALI_KELAS`
(wajib isi `--kelas`, mis. `--kelas 'X TJKT 1'`).

---

## Langkah 3: Hubungkan Domain

Pilih salah satu cara. Keduanya sudah teruji di lapangan.

### Cara A: Cloudflare Tunnel (paling aman, tanpa buka port)

Server tidak perlu membuka port 80/443 sama sekali.

```bash
# Pasang cloudflared
curl -fsSL https://pkg.cloudflare.com/cloudflared/install.sh | sudo bash 2>/dev/null || \
  curl -L -o cloudflared https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 && \
  sudo mv cloudflared /usr/local/bin/ && sudo chmod +x /usr/local/bin/cloudflared

cloudflared tunnel login
cloudflared tunnel create understud
cloudflared tunnel route dns understud understud.fazrilsh.com
```

Buat `~/.cloudflared/config.yml`:

```yaml
tunnel: understud
credentials-file: /root/.cloudflared/<TUNNEL-ID>.json
ingress:
  - hostname: understud.fazrilsh.com
    service: http://127.0.0.1:3000
  - service: http_status:404
```

Jalankan sebagai layanan:

```bash
sudo cloudflared service install
sudo systemctl enable --now cloudflared
```

### Cara B: Cloudflare Proxy (A record)

1. Di Cloudflare DNS, tambah **A record** `understud` -> IP server, dengan
   **Proxy status: Proxied** (awan oranye).
2. Di server, pasang reverse proxy yang menangani TLS dari Cloudflare.

**Opsi B1, pakai Caddy (paling ringkas):**

```bash
sudo apt install -y caddy     # atau sesuai distro
```

`/etc/caddy/Caddyfile`:

```
understud.fazrilsh.com {
    reverse_proxy 127.0.0.1:3000
}
```

```bash
sudo systemctl reload caddy
```

> Catatan: kalau Cloudflare di-set **Full (Strict)**, Caddy perlu sertifikat
> asli (bisa pakai Cloudflare Origin Certificate). Kalau **Flexible**,
> Cloudflare sudah menangani HTTPS ke pengguna dan Caddy cukup HTTP.

**Opsi B2, pakai nginx host:**

```nginx
server {
    listen 80;
    server_name understud.fazrilsh.com;
    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        client_max_body_size 25m;
    }
}
```

### Setelah domain terhubung

1. Pastikan `CORS_ORIGINS` di `.env.prod` memuat `https://understud.fazrilsh.com`
   (sudah default).
2. Pastikan `COOKIE_SECURE=true` (sudah default).
3. Cek: buka `https://understud.fazrilsh.com` di browser.

> Penting: `COOKIE_SECURE=true` artinya cookie hanya dikirim lewat HTTPS.
> Kalau diakses lewat `http://<ip>:3000` langsung, login staf akan gagal
> (memang begitu seharusnya). Selalu akses lewat domain HTTPS.

---

## Cara Update ke Versi Terbaru

Setiap push ke `main` yang menyentuh `backend/` atau `frontend/` akan
membangun ulang image di GitHub Actions. Setelah build selesai (cek tab
**Actions** di GitHub):

```bash
bash deploy/deploy.sh --update
```

Perintah itu menarik image terbaru dan me-restart container tanpa
menyentuh database (data aman di volume).

Kalau mau versi spesifik (lebih aman untuk produksi), ubah di `.env.prod`:

```ini
IMAGE_TAG=sha-1a2b3c4     # lihat daftar tag di tab Packages GitHub
```

lalu jalankan `bash deploy/deploy.sh --update`.

---

## Perintah Sehari-hari

| Aksi | Perintah |
|---|---|
| Update ke versi terbaru | `bash deploy/deploy.sh --update` |
| Lihat log | `bash deploy/deploy.sh --logs` |
| Hentikan (data tetap aman) | `bash deploy/deploy.sh --down` |
| Isi data demo | `bash deploy/deploy.sh --seed` |
| Cek status container | `docker-compose -f docker-compose.prod.yml --env-file .env.prod ps` |
| Backup database | `docker-compose -f docker-compose.prod.yml --env-file .env.prod exec -T db pg_dump -U postgres understud > backup.sql` |

---

## Kalau Ada Masalah

**Image gagal ditarik (`denied` / `unauthorized`)**
Paket GHCR harus publik. Buka `https://github.com/users/fzrilsh/packages/container/understud-api/settings`
dan pastikan visibilitasnya **Public**. Kalau tetap gagal, login dulu:
`echo $TOKEN | docker login ghcr.io -u fzrilsh --password-stdin`.

**Login staf gagal / selalu logout**
`JWT_SECRET` kemungkinan kosong atau berubah tiap restart. Cek `.env.prod`
sudah berisi `JWT_SECRET` minimal 32 karakter dan tidak diubah-ubah.

**Login jalan tapi cookie tidak tersimpan**
`COOKIE_SECURE=true` tapi diakses lewat HTTP. Akses lewat domain HTTPS.

**Upload lampiran gagal**
Pastikan `client_max_body_size` di reverse proxy host minimal `25m`
(nginx default hanya 1 MB). Di dalam container web sudah di-set 25m.

**Rate limit terasa terlalu ketat**
Semua pengguna di balik satu IP publik (NAT sekolah / Cloudflare) berbagi
satu bucket. Longgarkan `RATE_LIMIT_IP_MULTIPLIER` (default 20) di `.env.prod`.

**API tidak sehat saat deploy**
`bash deploy/deploy.sh --logs` lalu lihat pesan error. Penyebab paling umum:
`JWT_SECRET` kurang dari 32 karakter, atau `GEO_FENCE_ENABLED=true` tanpa
`SCHOOL_LAT`/`SCHOOL_LNG`.

---

## File Terkait

| File | Isi |
|---|---|
| `docker-compose.prod.yml` | Compose produksi (tarik dari GHCR, port di-hardening) |
| `deploy/deploy.sh` | Script setup + update sekali jalan |
| `deploy/.env.prod.example` | Template konfigurasi produksi |
| `.github/workflows/publish.yml` | Build & push image ke GHCR otomatis |
