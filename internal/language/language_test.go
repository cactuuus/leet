package language

import "testing"

// We gather a few known languages to use in tests, so we don't have to repeat the full struct literals.
// If the known map changes, these will automatically reflect the changes.
// If a language is removed, tests should be updated accordingly else they will fail.
var (
	golang = known["golang"]
	cpp = known["cpp"]
)

func TestGet(t *testing.T) {
	tests := []struct {
		name		   	string
		input    		string
		expectedLang 	Language
		expectedFound 	bool
	} {
		// fetch by slug (actual and case-insensitive)
		{"slug-golang", "golang",		golang, 	true},
		{"slug-cpp", 	"cpp", 			cpp, 		true},
		{"slug-GOLANG", "GOLANG", 		golang, 	true},
		{"slug-CPP", 	"CPP", 			cpp, 		true},
		// fetch by name (actual and case-insensitive)
		{"name-go", 	"go", 			golang, 	true},
		{"name-Go", 	"Go", 			golang, 	true},
		{"name-c++", 	"c++", 			cpp, 		true},
		{"name-C++", 	"C++", 			cpp, 		true},
		// tests non-acceptable inputs
		{"nonexistent", "nonexistent", 	Language{}, false},
		{"empty", 		"", 			Language{}, false},
		{"numeric", 	"123", 			Language{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, found := Get(tt.input)
			if found != tt.expectedFound {
				t.Fatalf("Get(%q) found = %v; want %v", tt.input, found, tt.expectedFound)
			}
			if l != tt.expectedLang {
				t.Errorf("Get(%q) = %v; want %v", tt.input, l, tt.expectedLang)
			}
		})
	}
}

func TestAll(t *testing.T) {
	langs := All()

	if len(langs) != len(known) {
		t.Errorf("All() returned %d languages; want %d", len(langs), len(known))
	}

	for _, l := range langs {
		got, ok := Get(l.Slug)
		if !ok {
			t.Fatalf("All() returned language not retrievable by Get(): %q", l.Slug)
		}
		if got != l {
			t.Errorf("Get(%q) = %+v, want %+v", l.Slug, got, l)
		}
	}
}
