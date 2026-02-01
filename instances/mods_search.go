package instances

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ModSearchResult reflects the structure returned by Modrinth Search API
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
		Gallery      []string `json:"gallery"` // Simplified
	} `json:"hits"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Total  int `json:"total_hits"`
}

// SearchMods queries Modrinth or Spiget
type InstalledPlugin struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Filename string `json:"filename"`
}

func (inst *Instance) SearchMods(query string, resourceType string, offset int, sort string, sides []string) ([]interface{}, error) {
	if resourceType == "plugin" {
		client := NewSpigetClient()
		// Page size 20, calculate page number from offset
		page := (offset / 20) + 1

		resources, err := client.SearchResources(query, 20, page)
		if err != nil {
			return nil, err
		}

		var hits []interface{}
		for _, r := range resources {
			// Map SpigetResource to frontend expected format
			hit := map[string]interface{}{
				"project_id":    fmt.Sprintf("%d", r.ID),
				"title":         r.Name,
				"description":   r.Tag,
				"icon_url":      "",                             // Default empty
				"author":        fmt.Sprintf("%d", r.Author.ID), // we only have ID initially
				"downloads":     r.Downloads,
				"date_modified": r.UpdateDate * 1000, // Unix timestamp to JS ms
				"categories":    []string{"plugin"},  // generic category
				"client_side":   "unsupported",
				"server_side":   "required",
			}

			// Handle Icon
			if r.Icon.Url != "" {
				if !strings.HasPrefix(r.Icon.Url, "http") {
					// Relative path, prepend spigotmc.org
					// Handle cases like "data/..." or "/data/..."
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

	// Modrinth logic (mod or modpack)
	loader := inst.Type
	mcVersion := inst.Version

	ptype := "mod"
	if resourceType == "modpack" {
		ptype = "modpack"
	}

	u, _ := url.Parse("https://api.modrinth.com/v2/search")
	q := u.Query()
	q.Set("query", query)

	// Modrinth facets: [["facet:value"], ["facet:value"]] for AND logic
	var facetList []string
	if loader != "vanilla" && loader != "" && loader != "spigot" && loader != "paper" && loader != "unknown" {
		// Only filter by loader if it maps to a Modrinth loader (fabric, forge, quilt, neoforge)
		facetList = append(facetList, fmt.Sprintf(`["categories:%s"]`, loader))
	} else if loader == "paper" || loader == "spigot" {
		// If on Paper/Spigot but searching Modrinth, we might want to find "bukkit" compatible plugins on Modrinth?
		// Modrinth uses "bukkit", "spigot", "paper" as categories.
	}

	facetList = append(facetList, fmt.Sprintf(`["versions:%s"]`, mcVersion))
	facetList = append(facetList, fmt.Sprintf(`["project_type:%s"]`, ptype))

	// Sides filter if provided (e.g. "client", "server")
	if len(sides) > 0 {
		var sideFacets []string
		for _, s := range sides {
			sideFacets = append(sideFacets, fmt.Sprintf(`"categories:%s"`, s))
		}
		facetList = append(facetList, fmt.Sprintf("[%s]", strings.Join(sideFacets, ",")))
	}

	q.Set("facets", fmt.Sprintf("[%s]", strings.Join(facetList, ",")))

	// Sorting
	if sort != "" {
		q.Set("index", sort) // relevance, downloads, follows, newest, updated
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
