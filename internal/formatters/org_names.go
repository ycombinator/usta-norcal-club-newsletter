package formatters

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const orgNamesFile = "org_names.yaml"

type OrgNames struct {
	names map[string]string
}

func LoadOrgNames() (*OrgNames, error) {
	on := &OrgNames{names: make(map[string]string)}

	data, err := os.ReadFile(orgNamesFile)
	if os.IsNotExist(err) {
		return on, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", orgNamesFile, err)
	}

	var raw map[string]string
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", orgNamesFile, err)
	}
	for k, v := range raw {
		on.names[strings.ToUpper(k)] = v
	}
	return on, nil
}

func (on *OrgNames) Save() error {
	data, err := yaml.Marshal(on.names)
	if err != nil {
		return err
	}
	return os.WriteFile(orgNamesFile, data, 0644)
}

func (on *OrgNames) Resolve(reader io.Reader, writer io.Writer, ustaName string) string {
	key := strings.ToUpper(ustaName)
	if friendly, ok := on.names[key]; ok {
		return friendly
	}

	scanner := bufio.NewScanner(reader)
	for {
		fmt.Fprintf(writer, "Short display name for %q: ", ustaName)
		if !scanner.Scan() {
			return ustaName
		}
		name := strings.TrimSpace(scanner.Text())
		if name != "" {
			on.names[key] = name
			return name
		}
	}
}
