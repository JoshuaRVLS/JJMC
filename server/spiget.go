package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const SpigetBaseURL = "https://api.spiget.org/v2"

type SpigetClient struct {
	BaseURL string
}

func NewSpigetClient() *SpigetClient {
	return &SpigetClient{
		BaseURL: SpigetBaseURL,
	}
}

// SpigetResource represents a plugin from Spiget
type SpigetResource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"tag"`
	File struct {
		Type     string  `json:"type"`
		Size     float64 `json:"size"`
		SizeUnit string  `json:"sizeUnit"`
		Url      string  `json:"url"`
	} `json:"file"`
	Rating struct {
		Count   int     `json:"count"`
		Average float64 `json:"average"`
	} `json:"rating"`
	Author struct {
		ID int `json:"id"`
	} `json:"author"`
	Downloads  int   `json:"downloads"`
	UpdateDate int64 `json:"updateDate"` // Unix timestamp
	Icon       struct {
		Url  string `json:"url"`
		Data string `json:"data"`
	} `json:"icon"`
}

// SpigetAuthor represents an author from Spiget
type SpigetAuthor struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

// SearchResources searches for plugins
func (c *SpigetClient) SearchResources(query string, size int, page int) ([]SpigetResource, error) {
	var urlStr string
	if query == "" {
		// List resources (sort by likes or downloads by default)
		urlStr = fmt.Sprintf("%s/resources?size=%d&page=%d&sort=-likes", c.BaseURL, size, page)
	} else {
		// Search resources
		urlStr = fmt.Sprintf("%s/search/resources/%s?size=%d&page=%d&sort=-likes", c.BaseURL, url.PathEscape(query), size, page)
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "JJMC/1.0")

	return c.doRequest(req)
}

func (c *SpigetClient) doRequest(req *http.Request) ([]SpigetResource, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// If 404 on search, it might just mean no results or empty, return empty list safely?
		// Spiget returns 404 for some empty searches?
		if resp.StatusCode == 404 {
			return []SpigetResource{}, nil
		}
		return nil, fmt.Errorf("spiget api returned %d: %s", resp.StatusCode, resp.Status)
	}

	var resources []SpigetResource
	if err := json.NewDecoder(resp.Body).Decode(&resources); err != nil {
		return nil, err
	}

	return resources, nil
}

// GetAuthor fetches author details
func (c *SpigetClient) GetAuthor(id int) (*SpigetAuthor, error) {
	reqUrl := fmt.Sprintf("%s/authors/%d", c.BaseURL, id)
	resp, err := http.Get(reqUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("spiget api returned %d", resp.StatusCode)
	}

	var author SpigetAuthor
	if err := json.NewDecoder(resp.Body).Decode(&author); err != nil {
		return nil, err
	}
	return &author, nil
}

// GetDownloadURL returns the download URL for a resource
// Note: Spiget download URLs might require being behind Cloudflare or user agent mimicking.
// Direct download: https://api.spiget.org/v2/resources/{id}/download
func (c *SpigetClient) GetDownloadURL(resourceID int) string {
	return fmt.Sprintf("%s/resources/%d/download", c.BaseURL, resourceID)
}

// GetResourceDetails fetches full details for a resource
func (c *SpigetClient) GetResourceDetails(id int) (*SpigetResource, error) {
	reqUrl := fmt.Sprintf("%s/resources/%d", c.BaseURL, id)
	resp, err := http.Get(reqUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("spiget api returned %d", resp.StatusCode)
	}

	var resource SpigetResource
	if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
		return nil, err
	}
	return &resource, nil
}

// DownloadResource downloads the resource to a writer
func (c *SpigetClient) DownloadResource(resourceID int, w io.Writer) error {
	downloadUrl := c.GetDownloadURL(resourceID)

	// We might need to handle redirects manually if Spiget does weird things,
	// but http.Get follows redirects by default.
	resp, err := http.Get(downloadUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: status %d", resp.StatusCode)
	}

	_, err = io.Copy(w, resp.Body)
	return err
}
