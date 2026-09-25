package avatar

import "testing"

func TestFromEmailStable(t *testing.T) {
	if FromEmail("") != "#6366f1" {
		t.Fatal(FromEmail(""))
	}
	if FromEmail("  Admin@Example.com ") != FromEmail("admin@example.com") {
		t.Fatal("case")
	}
	got := FromEmail("admin@example.com")
	for _, c := range palette {
		if got == c {
			return
		}
	}
	t.Fatal(got)
}

func TestFloorMod(t *testing.T) {
	if floorMod(-1, 10) != 9 {
		t.Fatal(floorMod(-1, 10))
	}
	if javaHash("a") != 97 {
		t.Fatal(javaHash("a"))
	}
}
