package spiget

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const SpigetBaseURL = "https://api.spiget.org/v2"

type Client struct {
	BaseURL string
}

func New() *Client {
	return &Client{
		BaseURL: SpigetBaseURL,
	}
}

type Resource struct {
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
	UpdateDate int64 `json:"updateDate"`
	Icon       struct {
		Url  string `json:"url"`
		Data string `json:"data"`
	} `json:"icon"`
}

type Author struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

func (c *Client) SearchResources(query string, size int, page int) ([]Resource, error) {
	var urlStr string
	if query == "" {
		urlStr = fmt.Sprintf("%s/resources?size=%d&page=%d&sort=-likes", c.BaseURL, size, page)
	} else {
		urlStr = fmt.Sprintf("%s/search/resources/%s?size=%d&page=%d&sort=-likes", c.BaseURL, url.PathEscape(query), size, page)
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "JJMC/1.0")

	return c.doRequest(req)
}

func (c *Client) doRequest(req *http.Request) ([]Resource, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 404 {
			return []Resource{}, nil
		}
		return nil, fmt.Errorf("spiget api returned %d: %s", resp.StatusCode, resp.Status)
	}

	var resources []Resource
	if err := json.NewDecoder(resp.Body).Decode(&resources); err != nil {
		return nil, err
	}

	return resources, nil
}

func (c *Client) GetAuthor(id int) (*Author, error) {
	reqUrl := fmt.Sprintf("%s/authors/%d", c.BaseURL, id)
	resp, err := http.Get(reqUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("spiget api returned %d", resp.StatusCode)
	}

	var author Author
	if err := json.NewDecoder(resp.Body).Decode(&author); err != nil {
		return nil, err
	}
	return &author, nil
}

func (c *Client) GetDownloadURL(resourceID int) string {
	return fmt.Sprintf("%s/resources/%d/download", c.BaseURL, resourceID)
}

func (c *Client) GetResourceDetails(id int) (*Resource, error) {
	reqUrl := fmt.Sprintf("%s/resources/%d", c.BaseURL, id)
	resp, err := http.Get(reqUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("spiget api returned %d", resp.StatusCode)
	}

	var resource Resource
	if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
		return nil, err
	}
	return &resource, nil
}

func (c *Client) DownloadResource(resourceID int, w io.Writer) error {
	downloadUrl := c.GetDownloadURL(resourceID)

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

type Version struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Date int64  `json:"date"`
}

func (c *Client) GetResourceVersions(resourceID int) ([]Version, error) {
	reqUrl := fmt.Sprintf("%s/resources/%d/versions?size=100&sort=-id", c.BaseURL, resourceID)
	resp, err := http.Get(reqUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("spiget api returned %d", resp.StatusCode)
	}

	var versions []Version
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, err
	}
	return versions, nil
}

func (c *Client) GetVersionDownloadURL(resourceID int, versionID int) string {
	return fmt.Sprintf("%s/resources/%d/versions/%d/download", c.BaseURL, resourceID, versionID)
}
