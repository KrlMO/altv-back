package games

type GameListItemResponse struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Slug     string  `json:"slug"`
	ColorHex *string `json:"color_hex"`
}
