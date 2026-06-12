package runner

import (
	"os"

	"gopkg.in/yaml.v3"
)

type LanguageDef struct {
	ID         string   `yaml:"id"`
	Name       string   `yaml:"name"`
	SourceFile string   `yaml:"source_file"`
	IsCompiled bool     `yaml:"is_compiled"`
	BuildCmd   []string `yaml:"build_cmd"`
	RunCmd     []string `yaml:"run_cmd"`
}

var Registry map[string]LanguageDef

func LoadRegistry(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			Registry = make(map[string]LanguageDef)
			return nil
		}
		return err
	}
	return yaml.Unmarshal(data, &Registry)
}