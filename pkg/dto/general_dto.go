package dto

type CountryReponse struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	FlagPath string `json:"flag_path"`
	IsoCode  string `json:"iso_code"`
}
