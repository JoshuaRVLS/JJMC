package instances

import (
	"fmt"
	"strings"

	"jjmc/mods/modrinth"
)

type ModSearchResult = modrinth.SearchResult

type InstalledPlugin struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Filename string `json:"filename"`
}

func (inst *Instance) SearchMods(query string, resourceType string, offset int, sort string, sides []string) ([]interface{}, error) {
	if resourceType == "plugin" {
		client := NewSpigetClient()

		page := (offset / 20) + 1

		resources, err := client.SearchResources(query, 20, page)
		if err != nil {
			return nil, err
		}

		var hits []interface{}
		for _, r := range resources {

			hit := map[string]interface{}{
				"project_id":    fmt.Sprintf("%d", r.ID),
				"title":         r.Name,
				"description":   r.Tag,
				"icon_url":      "",
				"author":        fmt.Sprintf("%d", r.Author.ID),
				"downloads":     r.Downloads,
				"date_modified": r.UpdateDate * 1000,
				"categories":    []string{"plugin"},
				"client_side":   "unsupported",
				"server_side":   "required",
			}

			if r.Icon.Url != "" {
				if !strings.HasPrefix(r.Icon.Url, "http") {

					hit["icon_url"] = "https://www.spigotmc.org/" + strings.TrimPrefix(r.Icon.Url, "/")
				} else {
					hit["icon_url"] = r.Icon.Url
				}
			} else if r.Icon.Data != "" {
				hit["icon_url"] = "data:image/jpeg;base64," + r.Icon.Data
			}

			hits = append(hits, hit)
		}
		return hits, nil
	}

	return modrinth.Search(query, resourceType, inst.Version, inst.Type, offset, sort, sides)
}

func SearchModrinth(query string, resourceType string, version string, loader string, offset int, sort string, sides []string) ([]interface{}, error) {
	return modrinth.Search(query, resourceType, version, loader, offset, sort, sides)
}
