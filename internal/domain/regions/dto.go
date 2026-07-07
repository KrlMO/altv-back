package regions

type RegionResponse struct {
	ID       int32   `json:"id"`
	Name     string  `json:"name"`
	IsoCode  string  `json:"iso_code"`
	FlagPath *string `json:"flag_path,omitempty"`
}
