package utils

type IPinfoJson struct {
	IP       string `json:"ip,omitempty"`
	HOSTNAME string `json:"hostname,omitempty"`
	CITY     string `json:"city,omitempty"`
	REGION   string `json:"region,omitempty"`
	COUNTRY  string `json:"country,omitempty"`
	LOC      string `json:"loc,omitempty"`
	POSTAL   string `json:"postal,omitempty"`
	ORG      string `json:"org,omitempty"`
}
