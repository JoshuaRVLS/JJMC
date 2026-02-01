package instances

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ModSearchResult struct {
	Hits []struct {
		ProjectID    string   `json:"project_id"`
		Slug         string   `json:"slug"`
		Title        string   `json:"title"`
		Description  string   `json:"description"`
		Categories   []string `json:"categories"`
		ClientSide   string   `json:"client_side"`
		ServerSide   string   `json:"server_side"`
		ProjectType  string   `json:"project_type"`
		Downloads    int      `json:"downloads"`
		IconUrl      string   `json:"icon_url"`
		Author       string   `json:"author"`
		Versions     []string `json:"versions"`
		Followers    int      `json:"followers"`
		DateCreated  string   `json:"date_created"`
		DateModified string   `json:"date_modified"`
		License      string   `json:"license"`
		Gallery      []string `json:"gallery"`
	} `json:"hits"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Total  int `json:"total_hits"`
}

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

	return SearchModrinth(query, resourceType, inst.Version, inst.Type, offset, sort, sides)
}

func SearchModrinth(query string, resourceType string, version string, loader string, offset int, sort string, sides []string) ([]interface{}, error) {
	ptype := "mod"
	if resourceType == "modpack" {
		ptype = "modpack"
	}

	u, _ := url.Parse("https://api.modrinth.com/v2/search")
	q := u.Query()
	q.Set("query", query)

	var facetList []string
	if loader != "vanilla" && loader != "" && loader != "spigot" && loader != "paper" && loader != "unknown" {
		facetList = append(facetList, fmt.Sprintf(`["categories:%s"]`, loader))
	}

	if version != "" {
		facetList = append(facetList, fmt.Sprintf(`["versions:%s"]`, version))
	}

	facetList = append(facetList, fmt.Sprintf(`["project_type:%s"]`, ptype))

	if len(sides) > 0 {
		var sideFacets []string
		for _, s := range sides {
			sideFacets = append(sideFacets, fmt.Sprintf(`"categories:%s"`, s))
		}
		facetList = append(facetList, fmt.Sprintf("[%s]", strings.Join(sideFacets, ",")))
	}

	q.Set("facets", fmt.Sprintf("[%s]", strings.Join(facetList, ",")))

	if sort != "" {
		q.Set("index", sort)
	} else if query == "" {
		q.Set("index", "downloads")
	} else {
		q.Set("index", "relevance")
	}

	q.Set("offset", fmt.Sprintf("%d", offset))
	q.Set("limit", "20")
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("modrinth error (%d): %s", resp.StatusCode, string(body))
	}

	var result ModSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var hits []interface{}
	for _, h := range result.Hits {
		hits = append(hits, h)
	}
	return hits, nil
}
