package fsm

func validIdent(s string) bool {
	// Allow: A-Z a-z 0-9 _ - . :
	// Disallow spaces and weird unicode to keep it deterministic.
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') ||
			(r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') ||
			r == '_' || r == '-' || r == '.' || r == ':' {
			continue
		}
		return false
	}
	return true
}
