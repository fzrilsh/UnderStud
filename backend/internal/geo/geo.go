package geo

import "math"

// DistanceM menghitung jarak dua titik (haversine) dalam meter.
func DistanceM(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusM = 6371000.0
	rad := math.Pi / 180
	dLat := (lat2 - lat1) * rad
	dLng := (lng2 - lng1) * rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusM * c
}

// WithinRadius melaporkan apakah (lat,lng) berada dalam radius meter dari pusat.
func WithinRadius(lat, lng, centerLat, centerLng float64, radiusM int) bool {
	return DistanceM(lat, lng, centerLat, centerLng) <= float64(radiusM)
}

// ValidCoord memvalidasi rentang koordinat. Menolak NaN/Inf (audit M6):
// perbandingan dengan NaN selalu false sehingga bisa lolos cek rentang.
func ValidCoord(lat, lng float64) bool {
	if math.IsNaN(lat) || math.IsInf(lat, 0) || math.IsNaN(lng) || math.IsInf(lng, 0) {
		return false
	}
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180
}
