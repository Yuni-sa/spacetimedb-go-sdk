package main

import "math"

// Vector2 with utility methods
type Vector2 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (v Vector2) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vector2) Normalize() Vector2 {
	length := v.Length()
	if length == 0 {
		return Vector2{0, 0}
	}
	return Vector2{v.X / length, v.Y / length}
}
