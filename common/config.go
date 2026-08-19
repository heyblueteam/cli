package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// ReadConfigFile reads the global config.env into a map
func ReadConfigFile() (map[string]string, error) {
	path := ConfigPath()
	if path == "" {
		return nil, fmt.Errorf("could not determine config path")
	}
	envMap, err := godotenv.Read(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}
	return envMap, nil
}

// LoadDefaultWorkspace returns the configured default workspace ID or slug.
// Environment variables win over local/global env files, matching LoadConfig.
func LoadDefaultWorkspace() string {
	_ = godotenv.Load()
	if globalConfig := ConfigPath(); globalConfig != "" {
		_ = godotenv.Load(globalConfig)
	}
	return strings.TrimSpace(os.Getenv("DEFAULT_WORKSPACE_ID"))
}

// WriteConfigFile writes a map back to config.env. Mode 0600 — the file holds
// credentials (AUTH_TOKEN / REFRESH_TOKEN).
func WriteConfigFile(values map[string]string) error {
	path := ConfigPath()
	if path == "" {
		return fmt.Errorf("could not determine config path")
	}
	content, err := godotenv.Marshal(values)
	if err != nil {
		return fmt.Errorf("could not serialize config: %w", err)
	}
	if err := os.MkdirAll(ConfigDir(), 0700); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}
	return os.WriteFile(path, []byte(content), 0600)
}

// GetCompanies returns the list of known company slugs from config
func GetCompanies() ([]string, error) {
	envMap, err := ReadConfigFile()
	if err != nil {
		return nil, err
	}
	raw := envMap["COMPANIES"]
	if raw == "" {
		return []string{}, nil
	}
	var companies []string
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			companies = append(companies, s)
		}
	}
	return companies, nil
}

// GetActiveCompany returns the active COMPANY_ID from config
func GetActiveCompany() (string, error) {
	envMap, err := ReadConfigFile()
	if err != nil {
		return "", err
	}
	return envMap["COMPANY_ID"], nil
}

// SetActiveCompany updates COMPANY_ID in config.env
func SetActiveCompany(slug string) error {
	envMap, err := ReadConfigFile()
	if err != nil {
		return err
	}
	envMap["COMPANY_ID"] = slug
	return WriteConfigFile(envMap)
}

// GetDefaultWorkspace returns DEFAULT_WORKSPACE_ID from config.env.
func GetDefaultWorkspace() (string, error) {
	envMap, err := ReadConfigFile()
	if err != nil {
		return "", err
	}
	return envMap["DEFAULT_WORKSPACE_ID"], nil
}

// SetDefaultWorkspace updates DEFAULT_WORKSPACE_ID in config.env.
func SetDefaultWorkspace(workspace string) error {
	envMap, err := ReadConfigFile()
	if err != nil {
		return err
	}
	envMap["DEFAULT_WORKSPACE_ID"] = workspace
	return WriteConfigFile(envMap)
}

// ClearDefaultWorkspace removes DEFAULT_WORKSPACE_ID from config.env.
func ClearDefaultWorkspace() error {
	envMap, err := ReadConfigFile()
	if err != nil {
		return err
	}
	delete(envMap, "DEFAULT_WORKSPACE_ID")
	return WriteConfigFile(envMap)
}

// AddCompany adds a slug to the known companies list if not already present
func AddCompany(slug string) error {
	if strings.Contains(slug, ",") {
		return fmt.Errorf("company slug cannot contain commas")
	}
	companies, err := GetCompanies()
	if err != nil {
		return err
	}
	for _, c := range companies {
		if c == slug {
			return nil // already present
		}
	}
	companies = append(companies, slug)
	return setCompanies(companies)
}

// RemoveCompany removes a slug from the known companies list
func RemoveCompany(slug string) error {
	companies, err := GetCompanies()
	if err != nil {
		return err
	}
	var updated []string
	found := false
	for _, c := range companies {
		if c == slug {
			found = true
			continue
		}
		updated = append(updated, c)
	}
	if !found {
		return fmt.Errorf("company %q not found in known companies", slug)
	}
	return setCompanies(updated)
}

func setCompanies(slugs []string) error {
	envMap, err := ReadConfigFile()
	if err != nil {
		return err
	}
	envMap["COMPANIES"] = strings.Join(slugs, ",")
	return WriteConfigFile(envMap)
}
