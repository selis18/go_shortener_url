package model

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	ShortURL string `json:"result"`
}

type JSONStorage struct {
	UUID        string `json: "uuid"`
	ShortURL    string `json: "short_url"`
	OriginalURL string `json: "original_url"`
}
