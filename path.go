package jzon

// Path searches a child node in an object or an array, if the
// node at the path doesn't exist, an error will be thrown out
func (jz *Jzon) Path(path string) (g *Jzon, err error) {
	// TODO:
	return
}

// SearchPath determines if there exists the node on the given path
func (jz *Jzon) SearchPath(path string) (exists bool) {
	// TODO:
	return
}

func parsePath(root *Jzon, path []byte) (closure func(jzon *Jzon) *Jzon, err error) {
	var curr *Jzon
	switch path[0] {
	case '$':
		// current: start
		// enter:   dollar
		// expect: '.' | '['
		curr = root
	case '.':
		// current: dollar | key
		// enter: key
		// expect: alpha
	case 'a','b':
		// current: dollar | key
		// enter: key
		// expect: '.' | '['
		key, path, err := parseKey(path)
		if err != nil {
			return nil, err
		}
		curr, err := curr.ValueOf(key)
		return parsePath(curr, path)

	case '[':
		// current: dollar | key
		// enter: key
		// expect: '.' | '['
	}

	return
}


