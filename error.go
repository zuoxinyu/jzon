package jzon

// Error implements the chained-call mechanism
// TODO:
type Error struct {
	val *Jzon
	err string
}
