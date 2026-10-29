package games

// for the search endpoint, many games can come in the response
type SteamSearchResponse struct {
    Total int                `json:"total"`
    Items []SteamSearchItem  `json:"items"`
}

type SteamSearchItem struct {
    Type       string          `json:"type"`
    Name       string          `json:"name"`
    ID         int             `json:"id"`
    TinyImage  string          `json:"tiny_image"`
    Metascore  string          `json:"metascore"`
    Platforms  SteamPlatforms  `json:"platforms"`
    Price      *SteamPrice     `json:"price"` // puntero porque no siempre viene
}

type SteamPlatforms struct {
    Windows bool `json:"windows"`
    Mac     bool `json:"mac"`
    Linux   bool `json:"linux"`
}

// for the search by id endpoint, only one game come in the response, and the data (or nil if the game doesnt exists)
type SteamAppDetails struct {
    Success bool `json:"success"`
    Data *SteamAppData `json:"data"`
}

type SteamAppData struct {
	ID	int `json:"steam_appid"`
	Name string `json:"name"`
    IsFree bool `json:"is_free"`
    PriceOverview *SteamPrice `json:"price_overview"`
}

type SteamPrice struct {
    Final float64 `json:"final"`
    Currency string `json:"currency"`
}
