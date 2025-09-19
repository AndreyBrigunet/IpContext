package geoip

import "strings"

// IsEUCountry reports whether the given ISO 3166-1 alpha-2 code belongs to the EU.
func IsEUCountry(code string) bool {
	cc := strings.ToUpper(code)
	_, ok := euCountries[cc]
	return ok
}

var euCountries = map[string]struct{} {
	"AT": {}, "BE": {}, "BG": {}, "HR": {}, "CY": {}, "CZ": {}, "DK": {}, "EE": {}, "FI": {}, "FR": {},
	"DE": {}, "GR": {}, "HU": {}, "IE": {}, "IT": {}, "LV": {}, "LT": {}, "LU": {}, "MT": {}, "NL": {},
	"PL": {}, "PT": {}, "RO": {}, "SK": {}, "SI": {}, "ES": {}, "SE": {},
}