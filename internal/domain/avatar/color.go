package avatar

// Palette and hash match com.reactcms.auth.util.AvatarColor (Java String.hashCode).
var palette = []string{
	"#6366f1", "#0ea5e9", "#8b5cf6", "#14b8a6", "#f59e0b",
	"#ef4444", "#22c55e", "#ec4899", "#64748b", "#06b6d4",
}

// FromEmail returns a stable palette color for an email address.
func FromEmail(email string) string {
	var trimmed string
	trimmed = trimLower(email)
	if trimmed == "" {
		return palette[0]
	}
	var h int32
	h = javaHash(trimmed)
	var idx int32
	idx = floorMod(h, int32(len(palette)))
	return palette[idx]
}

func trimLower(s string) string {
	var i int
	var j int
	i, j = 0, len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\n' || s[j-1] == '\r') {
		j--
	}
	var b []byte
	b = []byte(s[i:j])
	for k, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[k] = c + ('a' - 'A')
		}
	}
	return string(b)
}

func javaHash(s string) int32 {
	var h int32
	for i := 0; i < len(s); i++ {
		h = 31*h + int32(s[i])
	}
	return h
}

func floorMod(a, m int32) int32 {
	var r int32
	r = a % m
	if r < 0 {
		r += m
	}
	return r
}
