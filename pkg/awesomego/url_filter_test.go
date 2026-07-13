package awesomego

import "testing"

func TestExtractRepoURLs(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
	}{
		{name: "golang go", url: "https://github.com/golang/go", wantOwner: "golang", wantRepo: "go"},
		{name: "golang repository", url: "https://github.com/golang/tools", wantOwner: "golang", wantRepo: "tools"},
		{name: "direct repository with query", url: "https://github.com/acme/project?tab=readme", wantOwner: "acme", wantRepo: "project"},
		{name: "marketplace route", url: "https://github.com/marketplace/actions", wantOwner: "", wantRepo: ""},
		{name: "trending route", url: "https://github.com/trending/go", wantOwner: "", wantRepo: ""},
		{name: "wiki page", url: "https://github.com/acme/project/wiki", wantOwner: "", wantRepo: ""},
		{name: "issues page", url: "https://github.com/acme/project/issues", wantOwner: "", wantRepo: ""},
		{name: "package subdirectory", url: "https://github.com/acme/project/tree/main/pkg/library", wantOwner: "acme", wantRepo: "project"},
		{name: "versioned document", url: "https://github.com/acme/project/blob/v2/README.md", wantOwner: "acme", wantRepo: "project"},
		{name: "non github host", url: "https://example.com/acme/project", wantOwner: "", wantRepo: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo := extractRepoURLs(tt.url)
			if owner != tt.wantOwner || repo != tt.wantRepo {
				t.Fatalf("extractRepoURLs(%q) = %q/%q, want %q/%q", tt.url, owner, repo, tt.wantOwner, tt.wantRepo)
			}
		})
	}
}
