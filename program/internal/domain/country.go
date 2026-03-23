package domain

type CountryMeta struct {
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
	IconURL string `json:"icon_url"`
}

type CountryInfo struct {
	Name    string `json:"name"`
	Count   int    `json:"count"`
	Pattern string `json:"pattern"`
	IconURL string `json:"icon_url"`
}
