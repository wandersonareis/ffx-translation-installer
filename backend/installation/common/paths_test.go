package common

import "testing"

func TestNormalizeRelPath(t *testing.T) {
	valid := map[string]string{
		"data/mods/a.txt":        "data/mods/a.txt",
		`data\mods\a.txt`:        "data/mods/a.txt",
		"data/mods/sub/b.bin":    "data/mods/sub/b.bin",
		"data/mods/../../ok.txt": "ok.txt",
		"data//mods///file.txt":  "data/mods/file.txt",
	}
	for input, want := range valid {
		got, err := NormalizeRelPath(input)
		if err != nil {
			t.Fatalf("NormalizeRelPath(%q) erro: %v", input, err)
		}
		if got != want {
			t.Fatalf("NormalizeRelPath(%q) = %q, esperado %q", input, got, want)
		}
	}

	invalid := []string{
		"",
		"   ",
		"../evil.txt",
		`..\evil.txt`,
		"data/../../evil.txt",
		"/etc/passwd",
		`\windows\system32\evil.dll`,
		`C:\evil.txt`,
		`c:/evil.txt`,
		"//server/share/evil.txt",
		`\\server\share\evil.txt`,
		"data/mods/",
		".",
		"..",
		"data/mods/a?.txt",
		"data/mods/a*.txt",
		"data/mods/a:b.txt",
		"data/mods/CON",
		"data/mods/NUL.txt",
		"data/mods/COM1.bin",
		"data/mods/a.txt.",
	}
	for _, input := range invalid {
		if got, err := NormalizeRelPath(input); err == nil {
			t.Fatalf("NormalizeRelPath(%q) deveria falhar, retornou %q", input, got)
		}
	}
}
