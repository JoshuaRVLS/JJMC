package instances

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func ResolveForgeVersion(mcVersion string) (string, error) {
	resp, err := http.Get("https://files.minecraftforge.net/net/minecraftforge/forge/promotions_slim.json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var promotions struct {
		Promos map[string]string `json:"promos"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&promotions); err != nil {
		return "", fmt.Errorf("failed to decode promotions: %v", err)
	}

	if ver, ok := promotions.Promos[mcVersion+"-recommended"]; ok {
		return ver, nil
	}
	if ver, ok := promotions.Promos[mcVersion+"-latest"]; ok {
		return ver, nil
	}

	return "", fmt.Errorf("no forge version found for %s", mcVersion)
}

func ResolveNeoForgeVersion(mcVersion string) (string, error) {
	var prefix string
	if len(mcVersion) > 2 && mcVersion[:2] == "1." {
		prefix = mcVersion[2:]
	} else {
		return "", fmt.Errorf("unsupported mc version format: %s", mcVersion)
	}

	resp, err := http.Get("https://maven.neoforged.net/releases/net/neoforged/neoforge/maven-metadata.xml")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	sBody := string(body)

	parts := strings.Split(sBody, "<version>")
	var bestVer string

	for i := 1; i < len(parts); i++ {

		end := strings.Index(parts[i], "</version>")
		if end == -1 {
			continue
		}
		ver := parts[i][:end]

		if strings.HasPrefix(ver, prefix) {
			bestVer = ver
		}
	}

	if bestVer != "" {
		return bestVer, nil
	}

	return "", fmt.Errorf("no neoforge version found for mc %s (prefix %s)", mcVersion, prefix)
}
