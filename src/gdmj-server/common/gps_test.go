package common

import (
	"testing"
)

func TestGetDistance(t *testing.T) {
	var z float64
	t.Log(z, -z, 1/z, -1/z, z/z)
	point1 := make(map[string]float64, 2)
	point2 := make(map[string]float64, 2)
	point1["longitude"] = 29.9
	point1["latitude"] = 11.2
	point2["longitude"] = 30.12
	point2["latitude"] = 11.2
	length := GetDistance(point1, point2)
	t.Log(length)
}
