package claude

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillMD_Exists(t *testing.T) {
	path := filepath.Join("ccpr-review", "SKILL.md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("SKILL.md not found: %v", err)
	}
}

func TestSkillMD_ValidFrontmatter(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("ccpr-review", "SKILL.md"))
	if err != nil {
		t.Fatalf("failed to read SKILL.md: %v", err)
	}

	content := string(data)

	// Must start with frontmatter delimiter
	if !strings.HasPrefix(content, "---\n") {
		t.Fatal("SKILL.md must start with ---")
	}

	// Must have closing frontmatter delimiter
	rest := content[4:] // skip opening "---\n"
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		t.Fatal("SKILL.md missing closing --- for frontmatter")
	}

	frontmatter := rest[:idx]

	// Required fields
	requiredFields := []string{"name:", "description:"}
	for _, field := range requiredFields {
		if !strings.Contains(frontmatter, field) {
			t.Errorf("frontmatter missing required field %q", field)
		}
	}

	// Body must not be empty
	body := strings.TrimSpace(rest[idx+5:])
	if body == "" {
		t.Error("SKILL.md body is empty")
	}
}
