package frontmatter_test

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/Yakitrak/obsidian-cli/pkg/frontmatter"
	"github.com/stretchr/testify/assert"
	logging "gopkg.in/op/go-logging.v1"
)

func init() {
	// Suppress yq logging noise in tests
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	format := logging.MustStringFormatter(`%{message}`)
	backendFormatted := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatted)
	backendLeveled.SetLevel(logging.ERROR, "")
	logging.SetBackend(backendLeveled)
	log.SetFlags(0)
}

func TestProcessYAML(t *testing.T) {
	t.Run("Simple field extraction", func(t *testing.T) {
		yaml := "title: Hello World\nauthor: Jane"

		result, err := frontmatter.ProcessYAML(yaml, ".title")

		assert.NoError(t, err)
		assert.Equal(t, "Hello World\n", result)
	})

	t.Run("Nested field extraction", func(t *testing.T) {
		yaml := "metadata:\n  category: programming\n  views: 1234"

		result, err := frontmatter.ProcessYAML(yaml, ".metadata.category")

		assert.NoError(t, err)
		assert.Equal(t, "programming\n", result)
	})

	t.Run("Array extraction", func(t *testing.T) {
		yaml := "tags:\n  - golang\n  - testing\n  - yq"

		result, err := frontmatter.ProcessYAML(yaml, ".tags")

		assert.NoError(t, err)
		assert.Contains(t, result, "golang")
		assert.Contains(t, result, "testing")
	})

	t.Run("Array element extraction", func(t *testing.T) {
		yaml := "tags:\n  - golang\n  - testing\n  - yq"

		result, err := frontmatter.ProcessYAML(yaml, ".tags[0]")

		assert.NoError(t, err)
		assert.Equal(t, "golang\n", result)
	})

	t.Run("Array length", func(t *testing.T) {
		yaml := "tags:\n  - golang\n  - testing\n  - yq"

		result, err := frontmatter.ProcessYAML(yaml, ".tags | length")

		assert.NoError(t, err)
		assert.Equal(t, "3\n", result)
	})

	t.Run("Root extraction with dot", func(t *testing.T) {
		yaml := "title: Test\nauthor: Jane"

		result, err := frontmatter.ProcessYAML(yaml, ".")

		assert.NoError(t, err)
		assert.Contains(t, result, "title: Test")
		assert.Contains(t, result, "author: Jane")
	})

	t.Run("Empty input treated as empty object", func(t *testing.T) {
		result, err := frontmatter.ProcessYAML("", ".title")

		assert.NoError(t, err)
		assert.Equal(t, "null\n", result)
	})

	t.Run("Whitespace-only input treated as empty object", func(t *testing.T) {
		result, err := frontmatter.ProcessYAML("   \n  \n", ".title")

		assert.NoError(t, err)
		assert.Equal(t, "null\n", result)
	})
}

func TestProcessYAML_Modifications(t *testing.T) {
	t.Run("Set field value", func(t *testing.T) {
		yaml := "title: Old Title\nauthor: Jane"

		result, err := frontmatter.ProcessYAML(yaml, ".title = \"New Title\"")

		assert.NoError(t, err)
		assert.Contains(t, result, "title: New Title")
		assert.Contains(t, result, "author: Jane")
	})

	t.Run("Add new field", func(t *testing.T) {
		yaml := "title: Test"

		result, err := frontmatter.ProcessYAML(yaml, ".draft = true")

		assert.NoError(t, err)
		assert.Contains(t, result, "title: Test")
		assert.Contains(t, result, "draft: true")
	})

	t.Run("Append to array", func(t *testing.T) {
		yaml := "tags:\n  - golang\n  - testing"

		result, err := frontmatter.ProcessYAML(yaml, ".tags += [\"new-tag\"]")

		assert.NoError(t, err)
		assert.Contains(t, result, "golang")
		assert.Contains(t, result, "testing")
		assert.Contains(t, result, "new-tag")
	})

	t.Run("Multiple operations with pipe", func(t *testing.T) {
		yaml := "title: Test\ndraft: false"

		result, err := frontmatter.ProcessYAML(yaml, ".draft = true | .published = \"2025-10-01\"")

		assert.NoError(t, err)
		assert.Contains(t, result, "draft: true")
		assert.Contains(t, result, "published: \"2025-10-01\"")
	})

	t.Run("Delete field", func(t *testing.T) {
		yaml := "title: Test\nauthor: Jane\ndraft: false"

		result, err := frontmatter.ProcessYAML(yaml, "del(.draft)")

		assert.NoError(t, err)
		assert.Contains(t, result, "title: Test")
		assert.Contains(t, result, "author: Jane")
		assert.NotContains(t, result, "draft")
	})
}

func TestProcessYAML_Errors(t *testing.T) {
	t.Run("Invalid expression", func(t *testing.T) {
		yaml := "title: Test"

		_, err := frontmatter.ProcessYAML(yaml, "|")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "yaml processing error")
	})

	t.Run("Invalid YAML", func(t *testing.T) {
		yaml := "title: Test\n  invalid: indentation"

		_, err := frontmatter.ProcessYAML(yaml, ".title")

		assert.Error(t, err)
	})
}

func TestReconstructFile(t *testing.T) {
	t.Run("Basic reconstruction", func(t *testing.T) {
		fm := "title: Test\nauthor: Jane"
		body := "\n# Content here\n"

		result := frontmatter.ReconstructFile(fm, body)

		assert.Equal(t, "---\ntitle: Test\nauthor: Jane\n---\n\n# Content here\n", result)
	})

	t.Run("Empty frontmatter", func(t *testing.T) {
		fm := ""
		body := "# Content\n"

		result := frontmatter.ReconstructFile(fm, body)

		assert.Equal(t, "---\n\n---\n# Content\n", result)
	})

	t.Run("Frontmatter with trailing newlines", func(t *testing.T) {
		fm := "title: Test\n\n\n"
		body := "Content"

		result := frontmatter.ReconstructFile(fm, body)

		assert.Equal(t, "---\ntitle: Test\n---\nContent", result)
	})

	t.Run("Empty body", func(t *testing.T) {
		fm := "title: Test"
		body := ""

		result := frontmatter.ReconstructFile(fm, body)

		assert.Equal(t, "---\ntitle: Test\n---\n", result)
	})
}

func TestProcessFrontmatter(t *testing.T) {
	extractor := frontmatter.NewYAMLFrontmatterExtractor()

	t.Run("Extract and process valid frontmatter", func(t *testing.T) {
		content := "---\ntitle: Test\nauthor: Jane\n---\n\nContent"

		result, err := frontmatter.ProcessFrontmatter(extractor, ".title", content)

		assert.NoError(t, err)
		assert.Equal(t, "Test\n", result)
	})

	t.Run("Process with modification", func(t *testing.T) {
		content := "---\ntitle: Test\ndraft: false\n---\n\nContent"

		result, err := frontmatter.ProcessFrontmatter(extractor, ".draft = true", content)

		assert.NoError(t, err)
		assert.Contains(t, result, "draft: true")
	})

	t.Run("Error on invalid frontmatter", func(t *testing.T) {
		content := "---\nscalar value\n---\n\nContent"

		_, err := frontmatter.ProcessFrontmatter(extractor, ".title", content)

		assert.Error(t, err)
		assert.Equal(t, frontmatter.ErrScalarFrontmatter, err)
	})

	t.Run("Error on missing closing delimiter", func(t *testing.T) {
		content := "---\ntitle: Test\n\nContent"

		_, err := frontmatter.ProcessFrontmatter(extractor, ".title", content)

		assert.Error(t, err)
		assert.Equal(t, frontmatter.ErrInvalidFrontmatter, err)
	})

	t.Run("Process document without frontmatter", func(t *testing.T) {
		content := "# Just content\n\nNo frontmatter"

		result, err := frontmatter.ProcessFrontmatter(extractor, ".title", content)

		assert.NoError(t, err)
		assert.Equal(t, "null\n", result)
	})
}

func TestRun(t *testing.T) {
	extractor := frontmatter.NewYAMLFrontmatterExtractor()

	t.Run("Print mode returns only frontmatter", func(t *testing.T) {
		content := "---\ntitle: Test\nauthor: Jane\n---\n\n# Body content"

		result, err := frontmatter.Run(extractor, ".title", content, false)

		assert.NoError(t, err)
		assert.Equal(t, "Test\n", result)
		assert.NotContains(t, result, "Body content")
	})

	t.Run("Full file mode returns reconstructed file", func(t *testing.T) {
		content := "---\ntitle: Test\nauthor: Jane\n---\n\n# Body content"

		result, err := frontmatter.Run(extractor, ".draft = true", content, true)

		assert.NoError(t, err)
		assert.Contains(t, result, "---")
		assert.Contains(t, result, "draft: true")
		assert.Contains(t, result, "# Body content")
	})

	t.Run("Full file mode preserves body", func(t *testing.T) {
		content := "---\ntitle: Old\n---\n\n# Heading\n\nParagraph"

		result, err := frontmatter.Run(extractor, ".title = \"New\"", content, true)

		assert.NoError(t, err)
		lines := strings.Split(result, "\n")
		assert.Equal(t, "---", lines[0])
		assert.Contains(t, result, "title: New")
		assert.Contains(t, result, "# Heading")
		assert.Contains(t, result, "Paragraph")
	})

	t.Run("Error propagation", func(t *testing.T) {
		content := "---\nscalar\n---\n\nContent"

		_, err := frontmatter.Run(extractor, ".title", content, false)

		assert.Error(t, err)
		assert.Equal(t, frontmatter.ErrScalarFrontmatter, err)
	})

	t.Run("Complex yq expression in full file mode", func(t *testing.T) {
		content := "---\ntitle: Test\ntags:\n  - golang\n---\n\nContent"

		result, err := frontmatter.Run(extractor, ".tags += [\"new-tag\"] | .draft = false", content, true)

		assert.NoError(t, err)
		assert.Contains(t, result, "new-tag")
		assert.Contains(t, result, "draft: false")
		assert.Contains(t, result, "Content")
	})

	t.Run("Query expression that would create scalar frontmatter", func(t *testing.T) {
		content := "# Just content\n\nNo frontmatter"

		// Using .title (a query) in full file mode on empty frontmatter
		// This would create "---\nnull\n---" which is invalid
		result, err := frontmatter.Run(extractor, ".title", content, true)

		// The function succeeds because it doesn't validate output
		// But the result would be invalid if written
		assert.NoError(t, err)

		// Verify the output would be invalid by extracting it
		validationResult, validationErr := extractor.Extract(result)
		assert.NoError(t, validationErr)
		assert.False(t, validationResult.IsValid)
		assert.Equal(t, frontmatter.ErrScalarFrontmatter, validationResult.ValidationError)
	})

	t.Run("Query on file with valid frontmatter returns scalar", func(t *testing.T) {
		content := "---\ntitle: Test\n---\n\nContent"

		// Using .title (a query) would return just the value wrapped in frontmatter
		result, err := frontmatter.Run(extractor, ".title", content, true)

		assert.NoError(t, err)

		// The result would be "---\nTest\n---\nContent"
		// which has scalar frontmatter
		validationResult, validationErr := extractor.Extract(result)
		assert.NoError(t, validationErr)
		assert.False(t, validationResult.IsValid)
		assert.Equal(t, frontmatter.ErrScalarFrontmatter, validationResult.ValidationError)
	})
}
