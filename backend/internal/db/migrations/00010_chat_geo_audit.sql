-- +goose Up
-- Audit lokasi (spec §11.4 poin 6): catat koordinat yang diterima saat siswa
-- mengirim laporan. Dipakai untuk audit bila geofence diaktifkan (mitigasi
-- spoofing, spec §11.5). Nullable: geofence OFF / klien tak kirim → NULL.
ALTER TABLE chats ADD COLUMN geo_lat DOUBLE PRECISION;
ALTER TABLE chats ADD COLUMN geo_lng DOUBLE PRECISION;
ALTER TABLE chats ADD COLUMN geo_accuracy DOUBLE PRECISION;

-- +goose Down
ALTER TABLE chats DROP COLUMN geo_accuracy;
ALTER TABLE chats DROP COLUMN geo_lng;
ALTER TABLE chats DROP COLUMN geo_lat;
