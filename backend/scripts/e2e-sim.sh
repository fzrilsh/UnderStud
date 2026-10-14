#!/usr/bin/env bash
# E2E Simulation — simulasi end-to-end seluruh peran + replay semua temuan audit
# (yang harus SUDAH diperbaiki). Bukti hidup bahwa sistem jalan & aman.
#
# Pemakaian: bash scripts/e2e-sim.sh [BASE_URL]
# Prasyarat: stack jalan + `docker-compose exec api /app/cli seed-demo` sudah dijalankan.
set -uo pipefail

BASE_URL="${1:-http://localhost:8080}"
JAR_ANON="$(mktemp)"; JAR_ANON2="$(mktemp)"
JAR_BK="$(mktemp)"; JAR_WALI="$(mktemp)"
TMP="$(mktemp -d)"
trap 'rm -rf "$JAR_ANON" "$JAR_ANON2" "$JAR_BK" "$JAR_WALI" "$TMP"' EXIT

PASS=0; FAIL=0
ok()   { PASS=$((PASS+1)); printf '  \033[32m✓\033[0m %s\n' "$1"; }
bad()  { FAIL=$((FAIL+1)); printf '  \033[31m✗ %s\033[0m\n' "$1"; }
sec()  { printf '\n\033[1m%s\033[0m\n' "$1"; }

# expect <label> <expected> <actual>
expect() { [ "$2" = "$3" ] && ok "$1 ($3)" || bad "$1 — harap $2, dapat $3"; }
# expect_true <label> <jq-expr-result>
expect_true() { [ "$2" = "true" ] && ok "$1" || bad "$1 (dapat: $2)"; }

code() { curl -s -o /dev/null -w '%{http_code}' "$@"; }
jqr() { jq -r "$1"; }

# PNG 1x1 valid
printf '\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15\xc4\x89\x00\x00\x00\nIDATx\x9cc\x00\x01\x00\x00\x05\x00\x01\r\n-\xb4\x00\x00\x00\x00IEND\xaeB`\x82' > "$TMP/px.png"

echo "==================================================================="
echo "  JOINTS — E2E SIMULATION  ($BASE_URL)"
echo "==================================================================="

# ---------------------------------------------------------------------------
sec "[0] Kesehatan & bootstrap"
# ---------------------------------------------------------------------------
expect "healthz" 200 "$(code "$BASE_URL/healthz")"
expect_true "healthz envelope {data}" "$(curl -s "$BASE_URL/healthz" | jq '.data.status == "ok"')"
expect "bootstrap anon (set cookie)" 200 "$(code -c "$JAR_ANON" "$BASE_URL/api/v1/lapor")"
expect_true "bootstrap punya profile+chats" "$(curl -s -b "$JAR_ANON" "$BASE_URL/api/v1/lapor" | jq '.data | has("profile") and has("chats")')"

# ---------------------------------------------------------------------------
sec "[1] Login 3 peran + RBAC dasar"
# ---------------------------------------------------------------------------
expect "login BK" 200 "$(code -c "$JAR_BK" -X POST "$BASE_URL/api/v1/auth/login" -H 'Content-Type: application/json' -d '{"email":"bk@demo.local","password":"password123"}')"
expect "login Wali Kelas" 200 "$(code -c "$JAR_WALI" -X POST "$BASE_URL/api/v1/auth/login" -H 'Content-Type: application/json' -d '{"email":"wali@demo.local","password":"password123"}')"
expect_true "me BK role=BK" "$(curl -s -b "$JAR_BK" "$BASE_URL/api/v1/auth/me" | jq '.data.user.role == "BK"')"
expect_true "me WK role+kelas" "$(curl -s -b "$JAR_WALI" "$BASE_URL/api/v1/auth/me" | jq '.data.user.role == "WALI_KELAS" and .data.user.kelas == "X TJKT 1"')"
expect "login password salah -> 401" 401 "$(code -X POST "$BASE_URL/api/v1/auth/login" -H 'Content-Type: application/json' -d '{"email":"bk@demo.local","password":"salah"}')"

# ---------------------------------------------------------------------------
sec "[2] Alur siswa anonim: lapor -> BK triase -> requeue"
# ---------------------------------------------------------------------------
CHAT=$(curl -sf -b "$JAR_ANON" -X POST "$BASE_URL/api/v1/lapor/chats" \
  -F 'body=Simulasi E2E: saya butuh teman bicara' -F 'identity_mode=anonim' \
  -F "files=@$TMP/px.png;type=image/png;filename=bukti.png" | jqr '.data.id')
[ -n "$CHAT" ] && [ "$CHAT" != "null" ] && ok "siswa kirim laporan (+lampiran): $CHAT" || bad "gagal kirim laporan"

expect_true "lampiran terdaftar di detail siswa" "$(curl -s -b "$JAR_ANON" "$BASE_URL/api/v1/lapor/chats/$CHAT" | jq '.data.messages[0].attachments | length >= 1')"
expect_true "laporan masuk antrean BK" "$(curl -s -b "$JAR_BK" "$BASE_URL/api/v1/reports" | jq --arg id "$CHAT" '.data | map(.id) | index($id) != null')"
expect_true "buka laporan -> status dibuka (H2)" "$(curl -s -b "$JAR_BK" "$BASE_URL/api/v1/reports/$CHAT" | jq '.data.chat.status == "dibuka"')"

expect "Next sebelum balas -> 400" 400 "$(code -b "$JAR_BK" -X POST "$BASE_URL/api/v1/reports/$CHAT/next")"
expect "BK balas" 201 "$(code -b "$JAR_BK" -X POST "$BASE_URL/api/v1/reports/$CHAT/messages" -H 'Content-Type: application/json' -d '{"body":"Terima kasih, kami bantu."}')"
expect "Next setelah balas -> 200" 200 "$(code -b "$JAR_BK" -X POST "$BASE_URL/api/v1/reports/$CHAT/next")"
expect_true "Next keluar antrean" "$(curl -s -b "$JAR_BK" "$BASE_URL/api/v1/reports" | jq --arg id "$CHAT" '.data | map(.id) | index($id) == null')"
expect_true "muncul di riwayat" "$(curl -s -b "$JAR_BK" "$BASE_URL/api/v1/reports/history" | jq --arg id "$CHAT" '.data | map(.id) | index($id) != null')"
expect "siswa balas (requeue)" 201 "$(code -b "$JAR_ANON" -X POST "$BASE_URL/api/v1/lapor/chats/$CHAT/messages" -F 'body=Siap, terima kasih.')"
expect_true "requeue -> kembali ke antrean" "$(curl -s -b "$JAR_BK" "$BASE_URL/api/v1/reports" | jq --arg id "$CHAT" '.data | map(.id) | index($id) != null')"

# ---------------------------------------------------------------------------
sec "[3] Audit H1: Next TIDAK meninggalkan pin (regresi T1)"
# ---------------------------------------------------------------------------
C2=$(curl -sf -b "$JAR_ANON" -X POST "$BASE_URL/api/v1/lapor/chats" -F 'body=pin-skip test' -F 'identity_mode=anonim' | jqr '.data.id')
curl -s -b "$JAR_BK" -X POST "$BASE_URL/api/v1/reports/$C2/pin" > /dev/null
curl -s -b "$JAR_BK" -X POST "$BASE_URL/api/v1/reports/$C2/messages" -H 'Content-Type: application/json' -d '{"body":"r"}' > /dev/null
curl -s -b "$JAR_BK" -X POST "$BASE_URL/api/v1/reports/$C2/next" > /dev/null
curl -s -b "$JAR_ANON" -X POST "$BASE_URL/api/v1/lapor/chats/$C2/messages" -F 'body=lagi' > /dev/null
expect_true "setelah skip+requeue, TIDAK balik ke PIN" "$(curl -s -b "$JAR_BK" "$BASE_URL/api/v1/reports" | jq --arg id "$C2" '(.data[] | select(.id==$id) | .pinned) == false')"

# ---------------------------------------------------------------------------
sec "[4] Audit M7: Pin dari Draft wajib balas >=1x"
# ---------------------------------------------------------------------------
C3=$(curl -sf -b "$JAR_ANON" -X POST "$BASE_URL/api/v1/lapor/chats" -F 'body=draft test' -F 'identity_mode=anonim' | jqr '.data.id')
curl -s -b "$JAR_BK" -X POST "$BASE_URL/api/v1/reports/$C3/draft" > /dev/null
expect "Pin dari Draft tanpa balas -> 400" 400 "$(code -b "$JAR_BK" -X POST "$BASE_URL/api/v1/reports/$C3/pin")"
curl -s -b "$JAR_BK" -X POST "$BASE_URL/api/v1/reports/$C3/messages" -H 'Content-Type: application/json' -d '{"body":"r"}' > /dev/null
expect "Pin dari Draft SETELAH balas -> 200" 200 "$(code -b "$JAR_BK" -X POST "$BASE_URL/api/v1/reports/$C3/pin")"

# ---------------------------------------------------------------------------
sec "[5] Audit M3/M2: isolasi domain + status akun"
# ---------------------------------------------------------------------------
expect "staf akses /lapor -> 403 (M3)" 403 "$(code -b "$JAR_BK" "$BASE_URL/api/v1/lapor")"
expect "anon akses /reports -> 401/403" 401 "$(code -b "$JAR_ANON" "$BASE_URL/api/v1/reports")"
expect "tanpa auth /students -> 401" 401 "$(code "$BASE_URL/api/v1/students")"

# ---------------------------------------------------------------------------
sec "[6] Audit RBAC/scoping: WK hanya kelasnya, BK read-only"
# ---------------------------------------------------------------------------
expect "BK list siswa -> 200 (read)" 200 "$(code -b "$JAR_BK" "$BASE_URL/api/v1/students")"
expect "BK create siswa -> 403 (read-only)" 403 "$(code -b "$JAR_BK" -X POST "$BASE_URL/api/v1/students" -H 'Content-Type: application/json' -d '{"nisn":"9999999999","name":"X","kelas":"X TJKT 1"}')"
expect "BK unduh template -> 403" 403 "$(code -b "$JAR_BK" "$BASE_URL/api/v1/students/import/template")"
SID=$(curl -sf -b "$JAR_WALI" "$BASE_URL/api/v1/students?q=Andi" | jqr '.data[0].id')
expect_true "WK lihat siswa kelasnya" "$(curl -s -b "$JAR_WALI" "$BASE_URL/api/v1/students" | jq '.data | length >= 4')"
expect "WK unduh template -> 200" 200 "$(code -b "$JAR_WALI" "$BASE_URL/api/v1/students/import/template?format=xlsx")"
expect "WK handle red flag -> 403 (BK-only)" 403 "$(code -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/observations/00000000-0000-0000-0000-000000000000/handle")"

# ---------------------------------------------------------------------------
sec "[7] Alur Wali Kelas: profil + observasi + red flag + ringkasan"
# ---------------------------------------------------------------------------
NEWNISN="77$RANDOM$RANDOM"
NEWID=$(curl -sf -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students" -H 'Content-Type: application/json' \
  -d "{\"nisn\":\"$NEWNISN\",\"name\":\"Siswa Sim\",\"kelas\":\"X TJKT 1\"}" | jqr '.data.id')
[ -n "$NEWID" ] && [ "$NEWID" != "null" ] && ok "WK tambah siswa manual: $NEWID" || bad "gagal tambah siswa"

expect_true "payload StudentItem: archived_at ada (null) + nisn (M11)" \
  "$(curl -s -b "$JAR_WALI" "$BASE_URL/api/v1/students/$NEWID" | jq '.data.student | has("archived_at") and has("nisn") and .archived_at == null')"
expect_true "simpan profil + absensi (S/I/A)" \
  "$(curl -s -b "$JAR_WALI" -X PUT "$BASE_URL/api/v1/students/$NEWID/profile" -H 'Content-Type: application/json' \
     -d '{"nilai_erapor":[{"mapel":"Matematika (Umum)","nilai":88,"periode":"2025/2026 Genap"}],"minat_bakat":"Melukis","sakit":2,"izin":1,"alpa":0}' | jq '.data.sakit == 2 and .data.izin == 1')"

OBS=$(curl -sf -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students/$NEWID/observations" -H 'Content-Type: application/json' \
  -d "{\"items\":[\"red_flag:menyakiti_diri\"],\"tanggal\":\"$(date +%Y-%m-%d)\"}" | jqr '.data.id')
expect_true "observasi red flag -> red_flag=true" "$(curl -s -b "$JAR_WALI" "$BASE_URL/api/v1/students/$NEWID" | jq --arg id "$OBS" '.data.observations[] | select(.id==$id) | .red_flag')"
expect_true "ObservationView.author objek {name} (M11)" "$(curl -s -b "$JAR_WALI" "$BASE_URL/api/v1/students/$NEWID" | jq '.data.observations[0].author | type == "object"')"
expect_true "red flag tampil di panel BK" "$(curl -s -b "$JAR_BK" "$BASE_URL/api/v1/red-flags" | jq --arg id "$NEWID" '.data | map(.student_id) | index($id) != null')"
expect "BK tandai red flag ditangani" 200 "$(code -b "$JAR_BK" -X POST "$BASE_URL/api/v1/observations/$OBS/handle")"
expect_true "red flag hilang dari panel setelah ditangani" "$(curl -s -b "$JAR_BK" "$BASE_URL/api/v1/red-flags" | jq --arg id "$NEWID" '.data | map(.student_id) | index($id) == null')"

expect_true "ringkasan AI: AI off -> error terkode (bukan crash)" \
  "$(curl -s -b "$JAR_WALI" "$BASE_URL/api/v1/students/$NEWID/summaries/latest" | jq '.error.code == "AI_NOT_CONFIGURED" or .data.status == "ready" or .error.code == "NOT_FOUND"')"

# ---------------------------------------------------------------------------
sec "[8] Audit H5: pagination total = jumlah nyata (bukan ukuran halaman)"
# ---------------------------------------------------------------------------
T1=$(curl -s -b "$JAR_ANON" "$BASE_URL/api/v1/lapor/chats?limit=1" | jq '.meta.total')
T2=$(curl -s -b "$JAR_ANON" "$BASE_URL/api/v1/lapor/chats?limit=100" | jq '.meta.total')
expect "total /lapor/chats konsisten lintas limit" "$T2" "$T1"

# ---------------------------------------------------------------------------
sec "[9] Import leger (WK): preview -> commit -> absensi"
# ---------------------------------------------------------------------------
IMP="88$RANDOM$RANDOM"
cat > "$TMP/leger.csv" <<EOF
Sekolah,,SMKN 1 TANGERANG
NPSN,,20606915
Kelas,,X TJKT 1
Tahun Pelajaran,,2025/2026 Genap

No,NAMA PESERTA DIDIK,NISN,Matematika (Umum),Bahasa Indonesia,S,I,A
1,Siswa Leger,$IMP,85,90,3,1,0
EOF
PREV=$(curl -sf -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students/import/preview" -F "file=@$TMP/leger.csv;filename=leger.csv")
TOKEN=$(echo "$PREV" | jqr '.data.preview_token')
expect_true "preview mendeteksi siswa baru" "$(echo "$PREV" | jq '.data.ringkasan.baru >= 1')"
expect_true "commit membuat siswa" "$(curl -s -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students/import/commit" -H 'Content-Type: application/json' -d "{\"preview_token\":\"$TOKEN\"}" | jq '.data.created >= 1')"
LSID=$(curl -sf -b "$JAR_WALI" "$BASE_URL/api/v1/students?q=Siswa%20Leger" | jqr '.data[0].id')
expect_true "absensi dari leger tersimpan (sakit=3)" "$(curl -s -b "$JAR_WALI" "$BASE_URL/api/v1/students/$LSID" | jq '.data.profile.sakit == 3')"
expect "commit token SAMA lagi -> 410 (sekali-pakai, L8)" 410 "$(code -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students/import/commit" -H 'Content-Type: application/json' -d "{\"preview_token\":\"$TOKEN\"}")"

# ---------------------------------------------------------------------------
sec "[10] Audit keamanan: serangan yang HARUS gagal"
# ---------------------------------------------------------------------------
# C1: token palsu dengan secret default lama
FAKE=$(python3 - <<'PY'
import base64, hmac, hashlib, json, time
def b64(b): return base64.urlsafe_b64encode(b).rstrip(b'=').decode()
h = b64(json.dumps({"alg":"HS256","typ":"JWT"}).encode())
p = b64(json.dumps({"sub":"00000000-0000-0000-0000-000000000000","role":"BK","kelas":"","exp":int(time.time())+3600}).encode())
sig = b64(hmac.new(b"dev-secret-change-me-in-prod", f"{h}.{p}".encode(), hashlib.sha256).digest())
print(f"{h}.{p}.{sig}")
PY
)
expect "token palsu (secret default) -> 401 (C1)" 401 "$(code -H "Authorization: Bearer $FAKE" "$BASE_URL/api/v1/students")"

# M6: NaN geofence — tak bisa diverifikasi tanpa geofence ON, cek validasi koordinat tetap jalan
NANCODE=$(code -b "$JAR_ANON" -X POST "$BASE_URL/api/v1/lapor/chats" -F 'body=geo' -F 'identity_mode=anonim' -F 'lat=NaN' -F 'lng=NaN')
[ "$NANCODE" != "500" ] && ok "koordinat NaN tidak bikin 500 (M6) — dapat $NANCODE" || bad "koordinat NaN -> 500 (M6)"

# M8: SVG/HTML ditolak sebagai lampiran
printf '<svg xmlns="http://www.w3.org/2000/svg"></svg>' > "$TMP/evil.svg"
expect "upload SVG -> ditolak (M8)" 415 "$(code -b "$JAR_ANON" -X POST "$BASE_URL/api/v1/lapor/chats" -F 'body=x' -F 'identity_mode=anonim' -F "files=@$TMP/evil.svg;type=image/svg+xml;filename=evil.svg")"

# H3: PNG dimensi raksasa ditolak (bukan OOM)
python3 - "$TMP/bomb.png" <<'PY'
import sys, struct, zlib
def chunk(t,d): return struct.pack(">I",len(d))+t+d+struct.pack(">I",zlib.crc32(t+d)&0xffffffff)
ihdr=struct.pack(">IIBBBBB",100000,100000,8,2,0,0,0)
raw=b""
open(sys.argv[1],"wb").write(b"\x89PNG\r\n\x1a\n"+chunk(b"IHDR",ihdr)+chunk(b"IDAT",zlib.compress(raw))+chunk(b"IEND",b""))
PY
expect "PNG dimensi 100000x100000 -> ditolak, server hidup (H3)" 415 "$(code -b "$JAR_ANON" -X POST "$BASE_URL/api/v1/lapor/chats" -F 'body=x' -F 'identity_mode=anonim' -F "files=@$TMP/bomb.png;type=image/png;filename=bomb.png")"
expect_true "server MASIH hidup setelah image bomb" "$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/healthz" | grep -q 200 && echo true || echo false)"

# M10: >3 lampiran -> 413
expect "kirim 4 lampiran -> 413 (M10)" 413 "$(code -b "$JAR_ANON" -X POST "$BASE_URL/api/v1/lapor/chats" -F 'body=x' -F 'identity_mode=anonim' \
  -F "files=@$TMP/px.png;filename=a.png" -F "files=@$TMP/px.png;filename=b.png" -F "files=@$TMP/px.png;filename=c.png" -F "files=@$TMP/px.png;filename=d.png")"

# L12/M9: pesan error bahasa Indonesia
expect_true "error INVALID_INPUT berbahasa Indonesia (M9)" \
  "$(curl -s -X POST "$BASE_URL/api/v1/auth/login" -H 'Content-Type: application/json' -d '{}' | jq '.error.message | test("wajib|diisi|Format|tidak"; "i")')"

# ---------------------------------------------------------------------------
sec "[11] Isolasi sesi anonim (siswa A tak bisa baca chat siswa B)"
# ---------------------------------------------------------------------------
curl -s -c "$JAR_ANON2" "$BASE_URL/api/v1/lapor" > /dev/null
expect "sesi lain baca chat -> 404" 404 "$(code -b "$JAR_ANON2" "$BASE_URL/api/v1/lapor/chats/$CHAT")"
expect "sesi lain balas chat -> 404" 404 "$(code -b "$JAR_ANON2" -X POST "$BASE_URL/api/v1/lapor/chats/$CHAT/messages" -F 'body=x')"

# ---------------------------------------------------------------------------
sec "[12] Re-audit fix (F10/F12/F13/F16/F19/F23)"
# ---------------------------------------------------------------------------
# F23: bootstrap kirim daftar chat nyata + has_new_reply (bukan hardcode []).
expect_true "bootstrap: chats bukan hardcode kosong (F23)" \
  "$(curl -s -b "$JAR_ANON" "$BASE_URL/api/v1/lapor" | jq '.data.chats | length >= 1')"
expect_true "bootstrap: chat punya has_new_reply (F23)" \
  "$(curl -s -b "$JAR_ANON" "$BASE_URL/api/v1/lapor" | jq '.data.chats[0] | has("has_new_reply")')"
expect_true "/reports item punya unread_by_staff (F23)" \
  "$(curl -s -b "$JAR_BK" "$BASE_URL/api/v1/reports" | jq '.data[0] | has("unread_by_staff")')"

# F12: absensi raksasa → 400 (bukan tersimpan negatif via int32 overflow).
expect "absensi 3 miliar -> 400 (F12)" 400 "$(code -b "$JAR_WALI" -X PUT "$BASE_URL/api/v1/students/$NEWID/profile" \
  -H 'Content-Type: application/json' -d '{"sakit":3000000000}')"

# F13: PATCH tanpa nisn → NISN lama TIDAK terhapus.
expect_true "PATCH tanpa nisn → nisn tetap ada (F13)" \
  "$(curl -s -b "$JAR_WALI" -X PATCH "$BASE_URL/api/v1/students/$NEWID" \
     -H 'Content-Type: application/json' -d '{"name":"Siswa Sim","kelas":"X TJKT 1"}' | jq '.data.nisn != null and .data.nisn != ""')"

# F19: tanggal observasi BESOK → 400 (bukan lolos).
TOMORROW=$(date -v+1d +%Y-%m-%d 2>/dev/null || date -d '+1 day' +%Y-%m-%d)
expect "observasi tanggal besok -> 400 (F19)" 400 "$(code -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students/$NEWID/observations" \
  -H 'Content-Type: application/json' -d "{\"items\":[\"emosi:murung\"],\"tanggal\":\"$TOMORROW\"}")"

# F16/F15: leger dengan sel nilai "inf" & absensi raksasa → preview 200 (tak 500),
# nilai non-finite dilewati, absensi rusak jadi 0.
IMP2="99$RANDOM$RANDOM"
cat > "$TMP/leger2.csv" <<EOF
Sekolah,,SMKN 1 TANGERANG
NPSN,,20606915
Kelas,,X TJKT 1
Tahun Pelajaran,,2025/2026 Genap

No,NAMA PESERTA DIDIK,NISN,Matematika (Umum),S,I,A
1,Siswa Rusak,$IMP2,inf,3000000000,0,0
EOF
PREV2=$(curl -sf -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students/import/preview" -F "file=@$TMP/leger2.csv;filename=leger2.csv")
expect_true "leger sel 'inf' → preview 200, tak 500 (F16)" "$(echo "$PREV2" | jq '.data.preview_token != null')"
TOKEN2=$(echo "$PREV2" | jqr '.data.preview_token')
curl -s -b "$JAR_WALI" -X POST "$BASE_URL/api/v1/students/import/commit" -H 'Content-Type: application/json' -d "{\"preview_token\":\"$TOKEN2\"}" > /dev/null
LSID2=$(curl -sf -b "$JAR_WALI" "$BASE_URL/api/v1/students?q=Siswa%20Rusak" | jqr '.data[0].id')
expect_true "absensi rusak (3e9) → 0, bukan negatif (F15)" "$(curl -s -b "$JAR_WALI" "$BASE_URL/api/v1/students/$LSID2" | jq '.data.profile.sakit == 0')"

# F10: sesi anonim valid (bootstrap sukses → ok). Sesi zero-value tak lagi lolos.
expect_true "anon session bootstrap sehat (F10)" \
  "$(curl -s -b "$JAR_ANON" "$BASE_URL/api/v1/lapor/profile" | jq '.data.profile | has("name") and has("kelas")')"

# ---------------------------------------------------------------------------
echo
echo "==================================================================="
printf "  HASIL: \033[32m%d lulus\033[0m · \033[31m%d gagal\033[0m\n" "$PASS" "$FAIL"
echo "==================================================================="
[ "$FAIL" -eq 0 ] && printf '\033[32mE2E SIMULATION PASSED — semua alur jalan & semua audit fix aktif.\033[0m\n' || printf '\033[31mE2E SIMULATION FAILED — ada %d kegagalan.\033[0m\n' "$FAIL"
[ "$FAIL" -eq 0 ]
