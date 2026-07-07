package games

type GameListItemResponse struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	ColorHex *string `json:"color_hex"`
}
