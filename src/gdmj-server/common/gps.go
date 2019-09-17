package common

import (
	"math"
)

var (
	EARTHRADIUS = 6370996.81
)

func GetDistance(point1, point2 map[string]float64) float64 {
	point1["longitude"] = getLoop(point1["longitude"], -180, 180)
	point1["latitude"] = getRange(point1["latitude"], -74, 74)
	point2["longitude"] = getLoop(point2["longitude"], -180, 180)
	point2["latitude"] = getRange(point2["latitude"], -74, 74)

	x1 := toRadians(point1["longitude"])
	y1 := toRadians(point1["latitude"])
	x2 := toRadians(point2["longitude"])
	y2 := toRadians(point2["latitude"])

	return EARTHRADIUS * math.Acos(math.Sin(y1)*math.Sin(y2)+math.Cos(y1)*math.Cos(y2)*math.Cos(x2-x1))
}

func getLoop(v, a, b float64) float64 {
	if v > b {
		v -= b - a
	}
	if v < a {
		v += b - a
	}
	return v
}

func getRange(v, a, b float64) float64 {
	if a != 0 {
		v = math.Max(v, a)
	}

	if b != 0 {
		v = math.Min(v, b)
	}
	return v
}

func toRadians(angdeg float64) float64 {
	return math.Pi * angdeg / 180
}

func CheckLocation(location []float64) bool {
	if math.Abs(location[0]) > 180 || math.Abs(location[1]) > 90 {
		return false
	}
	return true
}
