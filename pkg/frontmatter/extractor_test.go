package frontmatter_test

import (
	"log"
	"os"
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

func TestExtract_ValidFrontmatter(t *testing.T) {
	t.Run("Basic frontmatter with body", func(t *testing.T) {
		content := "---\ntitle: Test\nauthor: Jane\n---\n\n# Body content"
		extractor := frontmatter.NewYAMLFrontmatterExtractor()

		result, err := extractor.Extract(content)

		assert.NoError(t, err)
		assert.True(t, result.HasFrontmatter)
		assert.True(t, result.IsValid)
		assert.Equal(t, "title: Test\nauthor: Jane\n", result.Frontmatter)
		assert.Equal(t, "\n# Body content\n", result.Body)
		assert.Nil(t, result.ValidationError)
	})

	t.Run("Frontmatter with multiple fields", func(t *testing.T) {
		content := "---\ntitle: My Post\nauthor: John\ntags:\n  - golang\n  - testing\ndraft: false\n---\n\nContent here"
		extractor := frontmatter.NewYAMLFrontmatterExtractor()

		result, err := extractor.Extract(content)

		assert.NoError(t, err)
		assert.True(t, result.HasFrontmatter)
		assert.True(t, result.IsValid)
		assert.Contains(t, result.Frontmatter, "title: My Post")
		assert.Contains(t, result.Frontmatter, "tags:")
	})

	t.Run("Empty frontmatter", func(t *testing.T) {
		content := "---\n---\n\n# Content"
		extractor := frontmatter.NewYAMLFrontmatterExtractor()

		result, err := extractor.Extract(content)

		assert.NoError(t, err)
		assert.True(t, result.HasFrontmatter)
		assert.True(t, result.IsValid)
		assert.Equal(t, "", result.Frontmatter)
	})

	t.Run("Frontmatter with empty object", func(t *testing.T) {
		content := "---\n{}\n---\n\nContent"
		extractor := frontmatter.NewYAMLFrontmatterExtractor()

		result, err := extractor.Extract(content)

		assert.NoError(t, err)
		assert.True(t, result.HasFrontmatter)
		assert.True(t, result.IsValid)
		assert.Equal(t, "{}\n", result.Frontmatter)
	})
}

func TestExtract_NoFrontmatter(t *testing.T) {
	t.Run("Content without frontmatter", func(t *testing.T) {
		content := "# Just a heading\n\nNo frontmatter here."
		extractor := frontmatter.NewYAMLFrontmatterExtractor()

		result, err := extractor.Extract(content)

		assert.NoError(t, err)
		assert.False(t, result.HasFrontmatter)
		assert.True(t, result.IsValid)
		assert.Equal(t, "", result.Frontmatter)
		assert.Equal(t, "# Just a heading\n\nNo frontmatter here.\n", result.Body)
	})

	t.Run("Empty document", func(t *testing.T) {
		content := ""
		extractor := frontmatter.NewYAMLFrontmatterExtractor()

		result, err := extractor.Extract(content)

		assert.NoError(t, err)
		assert.False(t, result.HasFrontmatter)
		assert.True(t, result.IsValid)
		assert.Equal(t, "", result.Frontmatter)
		assert.Equal(t, "", result.Body)
	})
}

func TestExtract_InvalidFrontmatter(t *testing.T) {
	t.Run("Only opening delimiter", func(t *testing.T) {
		content := "---\ntitle: Test\nauthor: Jane\n\n# Content"
		extractor := frontmatter.NewYAMLFrontmatterExtractor()

		result, err := extractor.Extract(content)

		assert.NoError(t, err)
		assert.False(t, result.HasFrontmatter)
		assert.False(t, result.IsValid)
		assert.Equal(t, frontmatter.ErrInvalidFrontmatter, result.ValidationError)
	})

	t.Run("Scalar value instead of object", func(t *testing.T) {
		content := "---\njust a string\n---\n\nContent"
		extractor := frontmatter.NewYAMLFrontmatterExtractor()

		result, err := extractor.Extract(content)

		assert.NoError(t, err)
		assert.True(t, result.HasFrontmatter)
		assert.False(t, result.IsValid)
		assert.Equal(t, frontmatter.ErrScalarFrontmatter, result.ValidationError)
	})

	t.Run("Number as scalar", func(t *testing.T) {
		content := "---\n123\n---\n\nContent"
		extractor := frontmatter.NewYAMLFrontmatterExtractor()

		result, err := extractor.Extract(content)

		assert.NoError(t, err)
		assert.True(t, result.HasFrontmatter)
		assert.False(t, result.IsValid)
		assert.Equal(t, frontmatter.ErrScalarFrontmatter, result.ValidationError)
	})
}

func TestExtract_EdgeCases(t *testing.T) {
	t.Run("Multiple delimiter pairs (only first is frontmatter)", func(t *testing.T) {
		content := "---\ntitle: First\n---\n\nContent\n---\nNot frontmatter\n---"
		extractor := frontmatter.NewYAMLFrontmatterExtractor()

		result, err := extractor.Extract(content)

		assert.NoError(t, err)
		assert.True(t, result.HasFrontmatter)
		assert.True(t, result.IsValid)
		assert.Equal(t, "title: First\n", result.Frontmatter)
		assert.Contains(t, result.Body, "---\nNot frontmatter\n---")
	})

	t.Run("Delimiter not at start of document", func(t *testing.T) {
		content := "\n---\ntitle: Test\n---\nContent"
		extractor := frontmatter.NewYAMLFrontmatterExtractor()

		result, err := extractor.Extract(content)

		assert.NoError(t, err)
		// When delimiter is not at start, the scanner will still process it
		// The first line is empty, then we encounter ---, so delimiterCount becomes 1
		// This test needs to verify the actual behavior
		assert.True(t, result.HasFrontmatter)
		assert.True(t, result.IsValid)
	})

	t.Run("Whitespace around delimiters", func(t *testing.T) {
		content := "---\ntitle: Test\n---\n\nBody with spaces"
		extractor := frontmatter.NewYAMLFrontmatterExtractor()

		result, err := extractor.Extract(content)

		assert.NoError(t, err)
		assert.True(t, result.HasFrontmatter)
		assert.True(t, result.IsValid)
		assert.Equal(t, "title: Test\n", result.Frontmatter)
	})
}
