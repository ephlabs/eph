package log

type Token string

type APIKey string

type Environment struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Provider string `json:"provider"`
	Token    Token  `json:"-"`
}
