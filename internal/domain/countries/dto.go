package countries

type CountryResponse struct {
	ID       int32   `json:"id"`
	Name     string  `json:"name"`
	FlagPath *string `json:"flag_path,omitempty"`
	IsoCode  string  `json:"iso_code"`
}
