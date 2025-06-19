package readconfig

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Function to read configuration from YAML file
func ReadConfigFromFile[T any](filename string, config *T) error {
	// Read YAML file
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Unmarshal YAML data into config struct
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return err
	}

	return nil
}
