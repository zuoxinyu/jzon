package jzon

// Verify verifies this node by another JSON which has a particular format,
// the given JSON should define the format of each field by a regular
// expression and an level number. If there were some field can't pass
// the regexp, the level numbers would give errors respectively
func (jz *Jzon) Verify(format *Jzon) (ok bool, err error) {

	return
}
