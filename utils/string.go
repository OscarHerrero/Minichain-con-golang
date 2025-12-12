package utils

// TruncateString trunca un string de forma segura
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// TruncateWithDots trunca un string y agrega "..." al final
func TruncateWithDots(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
