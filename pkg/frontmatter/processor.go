package frontmatter

import (
	"fmt"
	"strings"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

type Extractor interface {
	Extract(content string) (*ExtractionResult, error)
}

func ProcessYAML(yamlString, expression string) (string, error) {
	yqlib.InitExpressionParser()

	decoder := yqlib.NewYamlDecoder(yqlib.ConfiguredYamlPreferences)
	encoder := yqlib.NewYamlEncoder(yqlib.ConfiguredYamlPreferences)

	input := yamlString
	if strings.TrimSpace(yamlString) == "" {
		input = "{}"
	}

	stringEval := yqlib.NewStringEvaluator()
	result, err := stringEval.Evaluate(expression, input, encoder, decoder)
	if err != nil {
		return "", fmt.Errorf("yaml processing error: %w", err)
	}

	return result, nil
}

func ReconstructFile(processedFrontmatter, body string) string {
	var output strings.Builder
	output.WriteString("---\n")
	output.WriteString(strings.TrimSpace(processedFrontmatter) + "\n")
	output.WriteString("---\n")
	output.WriteString(body)
	return output.String()
}

func ProcessFrontmatter(extractor Extractor, expression, mdContent string) (string, error) {
	result, err := extractor.Extract(mdContent)
	if err != nil {
		return "", err
	}

	if !result.IsValid {
		return "", result.ValidationError
	}

	processed, err := ProcessYAML(result.Frontmatter, expression)
	if err != nil {
		return "", err
	}

	return processed, nil
}

func Run(extractor Extractor, expression, mdContent string, fullFile bool) (string, error) {
	processedFrontmatter, err := ProcessFrontmatter(extractor, expression, mdContent)
	if err != nil {
		return "", err
	}

	if !fullFile {
		return processedFrontmatter, nil
	}

	result, err := extractor.Extract(mdContent)
	if err != nil {
		return "", err
	}

	return ReconstructFile(processedFrontmatter, result.Body), nil
}
