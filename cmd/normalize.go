package cmd

import (
	"strings"

	"github.com/staticbackendhq/backend-go"
)

func cleanConfigValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	return value
}

func normalizeBackendRegion(region string) string {
	region = cleanConfigValue(region)
	switch region {
	case backend.RegionNorthAmerica1:
		return "https://na1.staticbackend.dev"
	case backend.RegionLocalDev, "":
		return backend.RegionLocalDev
	}

	if strings.Contains(region, "://") {
		return region
	}

	return "https://" + region
}
