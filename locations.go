package goatcounter

import (
	"context"
	"net"
	"strings"

	"zgo.at/errors"
	"zgo.at/goatcounter/v2/pkg/geo"
	"zgo.at/goatcounter/v2/pkg/log"
	"zgo.at/zdb"
)

type LocationID int32

type Location struct {
	ID LocationID `db:"location_id"`

	Country     string `db:"country"`
	Region      string `db:"region"`
	CountryName string `db:"country_name"`
	RegionName  string `db:"region_name"`

	// TODO: send patch to staticcheck to deal with this better. This shouldn't
	// errror since "ISO" is an initialism.
	ISO3166_2 string `db:"iso_3166_2"` //lint:ignore ST1003 staticcheck bug
}

// Derived from https://gist.github.com/ssskip/5a94bfcd2835bf1dea52
var iso3166_2_to_short_name = map[string]string{
  "AF": "Afghanistan",
  "AX": "Aland Islands",
  "AL": "Albania",
  "DZ": "Algeria",
  "AS": "American Samoa",
  "AD": "Andorra",
  "AO": "Angola",
  "AI": "Anguilla",
  "AQ": "Antarctica",
  "AG": "Antigua and Barbuda",
  "AR": "Argentina",
  "AM": "Armenia",
  "AW": "Aruba",
  "AU": "Australia",
  "AT": "Austria",
  "AZ": "Azerbaijan",
  "BS": "Bahamas",
  "BH": "Bahrain",
  "BD": "Bangladesh",
  "BB": "Barbados",
  "BY": "Belarus",
  "BE": "Belgium",
  "BZ": "Belize",
  "BJ": "Benin",
  "BM": "Bermuda",
  "BT": "Bhutan",
  "BO": "Bolivia",
  "BA": "Bosnia and Herzegovina",
  "BW": "Botswana",
  "BV": "Bouvet Island",
  "BR": "Brazil",
  "IO": "British Indian Ocean Territory",
  "BN": "Brunei Darussalam",
  "BG": "Bulgaria",
  "BF": "Burkina Faso",
  "BI": "Burundi",
  "KH": "Cambodia",
  "CM": "Cameroon",
  "CA": "Canada",
  "CV": "Cape Verde",
  "KY": "Cayman Islands",
  "CF": "Central African Republic",
  "TD": "Chad",
  "CL": "Chile",
  "CN": "China",
  "CX": "Christmas Island",
  "CC": "Cocos (Keeling) Islands",
  "CO": "Colombia",
  "KM": "Comoros",
  "CG": "Congo",
  "CD": "Democratic Republic of the Congo",
  "CK": "Cook Islands",
  "CR": "Costa Rica",
  "CI": "Cote D'Ivoire",
  "HR": "Croatia",
  "CU": "Cuba",
  "CY": "Cyprus",
  "CZ": "Czech Republic",
  "DK": "Denmark",
  "DJ": "Djibouti",
  "DM": "Dominica",
  "DO": "Dominican Republic",
  "EC": "Ecuador",
  "EG": "Egypt",
  "SV": "El Salvador",
  "GQ": "Equatorial Guinea",
  "ER": "Eritrea",
  "EE": "Estonia",
  "ET": "Ethiopia",
  "FK": "Falkland Islands (Malvinas)",
  "FO": "Faroe Islands",
  "FJ": "Fiji",
  "FI": "Finland",
  "FR": "France",
  "GF": "French Guiana",
  "PF": "French Polynesia",
  "TF": "French Southern Territories",
  "GA": "Gabon",
  "GM": "Gambia",
  "GE": "Georgia",
  "DE": "Germany",
  "GH": "Ghana",
  "GI": "Gibraltar",
  "GR": "Greece",
  "GL": "Greenland",
  "GD": "Grenada",
  "GP": "Guadeloupe",
  "GU": "Guam",
  "GT": "Guatemala",
  "GG": "Guernsey",
  "GN": "Guinea",
  "GW": "Guinea-Bissau",
  "GY": "Guyana",
  "HT": "Haiti",
  "HM": "Heard Island & Mcdonald Islands",
  "VA": "Holy See (Vatican City State)",
  "HN": "Honduras",
  "HK": "Hong Kong",
  "HU": "Hungary",
  "IS": "Iceland",
  "IN": "India",
  "ID": "Indonesia",
  "IR": "Iran",
  "IQ": "Iraq",
  "IE": "Ireland",
  "IM": "Isle Of Man",
  "IL": "Israel",
  "IT": "Italy",
  "JM": "Jamaica",
  "JP": "Japan",
  "JE": "Jersey",
  "JO": "Jordan",
  "KZ": "Kazakhstan",
  "KE": "Kenya",
  "KI": "Kiribati",
  "KR": "Korea",
  "KP": "North Korea",
  "KW": "Kuwait",
  "KG": "Kyrgyzstan",
  "LA": "Laos",
  "LV": "Latvia",
  "LB": "Lebanon",
  "LS": "Lesotho",
  "LR": "Liberia",
  "LY": "Libyan Arab Jamahiriya",
  "LI": "Liechtenstein",
  "LT": "Lithuania",
  "LU": "Luxembourg",
  "MO": "Macao",
  "MK": "Macedonia",
  "MG": "Madagascar",
  "MW": "Malawi",
  "MY": "Malaysia",
  "MV": "Maldives",
  "ML": "Mali",
  "MT": "Malta",
  "MH": "Marshall Islands",
  "MQ": "Martinique",
  "MR": "Mauritania",
  "MU": "Mauritius",
  "YT": "Mayotte",
  "MX": "Mexico",
  "FM": "Micronesia",
  "MD": "Moldova",
  "MC": "Monaco",
  "MN": "Mongolia",
  "ME": "Montenegro",
  "MS": "Montserrat",
  "MA": "Morocco",
  "MZ": "Mozambique",
  "MM": "Myanmar",
  "NA": "Namibia",
  "NR": "Nauru",
  "NP": "Nepal",
  "NL": "Netherlands",
  "AN": "Netherlands Antilles",
  "NC": "New Caledonia",
  "NZ": "New Zealand",
  "NI": "Nicaragua",
  "NE": "Niger",
  "NG": "Nigeria",
  "NU": "Niue",
  "NF": "Norfolk Island",
  "MP": "Northern Mariana Islands",
  "NO": "Norway",
  "OM": "Oman",
  "PK": "Pakistan",
  "PW": "Palau",
  "PS": "Palestinian Territory",
  "PA": "Panama",
  "PG": "Papua New Guinea",
  "PY": "Paraguay",
  "PE": "Peru",
  "PH": "Philippines",
  "PN": "Pitcairn",
  "PL": "Poland",
  "PT": "Portugal",
  "PR": "Puerto Rico",
  "QA": "Qatar",
  "RE": "Reunion",
  "RO": "Romania",
  "RU": "Russian Federation",
  "RW": "Rwanda",
  "BL": "Saint Barthelemy",
  "SH": "Saint Helena",
  "KN": "Saint Kitts and Nevis",
  "LC": "Saint Lucia",
  "MF": "Saint Martin",
  "PM": "Saint Pierre and Miquelon",
  "VC": "Saint Vincent and Grenadines",
  "WS": "Samoa",
  "SM": "San Marino",
  "ST": "Sao Tome and Principe",
  "SA": "Saudi Arabia",
  "SN": "Senegal",
  "RS": "Serbia",
  "SC": "Seychelles",
  "SL": "Sierra Leone",
  "SG": "Singapore",
  "SK": "Slovakia",
  "SI": "Slovenia",
  "SB": "Solomon Islands",
  "SO": "Somalia",
  "ZA": "South Africa",
  "GS": "South Georgia and Sandwich Isl.",
  "ES": "Spain",
  "LK": "Sri Lanka",
  "SD": "Sudan",
  "SR": "Suriname",
  "SJ": "Svalbard and Jan Mayen",
  "SZ": "Swaziland",
  "SE": "Sweden",
  "CH": "Switzerland",
  "SY": "Syrian Arab Republic",
  "TW": "Taiwan",
  "TJ": "Tajikistan",
  "TZ": "Tanzania",
  "TH": "Thailand",
  "TL": "Timor-Leste",
  "TG": "Togo",
  "TK": "Tokelau",
  "TO": "Tonga",
  "TT": "Trinidad and Tobago",
  "TN": "Tunisia",
  "TR": "Turkey",
  "TM": "Turkmenistan",
  "TC": "Turks and Caicos Islands",
  "TV": "Tuvalu",
  "UG": "Uganda",
  "UA": "Ukraine",
  "AE": "United Arab Emirates",
  "GB": "United Kingdom",
  "US": "United States",
  "UM": "United States Outlying Islands",
  "UY": "Uruguay",
  "UZ": "Uzbekistan",
  "VU": "Vanuatu",
  "VE": "Venezuela",
  "VN": "Vietnam",
  "VG": "British Virgin Islands",
  "VI": "U.S. Virgin Islands",
  "WF": "Wallis and Futuna",
  "EH": "Western Sahara",
  "YE": "Yemen",
  "ZM": "Zambia",
  "ZW": "Zimbabwe",
}

// ByCode gets a location by ISO-3166-2 code; e.g. "US" or "US-TX".
func (l *Location) ByCode(ctx context.Context, code string) error {
	if ll, ok := cacheLoc(ctx).Get(code); ok {
		*l = *ll
		return nil
	}

	err := zdb.Get(ctx, l, `select * from locations where iso_3166_2 = $1`, code)
	if zdb.ErrNoRows(err) {
		l.ISO3166_2 = code
		l.Country, l.Region, _ = strings.Cut(code, "-")
		l.CountryName, l.RegionName = findGeoName(ctx, l.Country, l.Region)
		err = l.insert(ctx)
	}
	if err != nil {
		return errors.Wrap(err, "Location.ByCode")
	}

	cacheLoc(ctx).Set(l.ISO3166_2, l)
	return nil
}

// Lookup a location by IPv4 or IPv6 address.
//
// This will insert a row in the locations table if one doesn't exist yet.
func (l *Location) Lookup(ctx context.Context, ip string) error {
	geodb := geo.Get(ctx)
	if geodb == nil {
		return errors.New("Location.Lookup: no geodb on context")
	}

	ip = "96.242.52.215"
	// mkember: Country ISO code, not city.
	iso, err := geodb.CountryISOCode(net.ParseIP(ip))
	if err != nil {
		return errors.Wrap(err, "Location.Lookup")
	}
	l.Country = iso
	l.CountryName = iso3166_2_to_short_name[iso]
	// if len(loc.Subdivisions) > 0 {
	// 	l.Region, l.RegionName = loc.Subdivisions[0].IsoCode, loc.Subdivisions[0].Names["en"]
	// }

	l.ISO3166_2 = iso
	if l.Region != "" {
		l.ISO3166_2 += "-" + l.Region
	}
	if ll, ok := cacheLoc(ctx).Get(l.ISO3166_2); ok {
		*l = *ll
		return nil
	}

	err = zdb.Get(ctx, l,
		`select * from locations where country = $1 and region = $2`,
		l.Country, l.Region)
	if zdb.ErrNoRows(err) {
		err = l.insert(ctx)
	}
	if err != nil {
		return errors.Wrap(err, "Location.Lookup")
	}

	cacheLoc(ctx).Set(l.ISO3166_2, l)
	return nil
}

// LookupIP is a shorthand for Lookup(); returns id 1 on errors ("unknown").
func (l Location) LookupIP(ctx context.Context, ip string) string {
	err := l.Lookup(ctx, ip)
	if err != nil {
		return "" // Special ID: "unknown".
	}
	return l.ISO3166_2
}

func (l *Location) insert(ctx context.Context) (err error) {
	l.ID, err = zdb.InsertID[LocationID](ctx, "location_id",
		`insert into locations (country, region, country_name, region_name) values (?, ?, ?, ?)`,
		l.Country, l.Region, l.CountryName, l.RegionName)
	if err != nil {
		return err
	}

	// Make sure there is an entry for the country as well.
	if l.Region != "" {
		err := (&Location{}).ByCode(ctx, l.Country)
		if err != nil {
			return err
		}
	}
	return nil
}

type Locations []Location

// ListCountries lists all counties. The region code/name will always be blank.
func (l *Locations) ListCountries(ctx context.Context) error {
	err := zdb.Select(ctx, l, `
		select country, country_name
        from locations
        where country != '' and country_name != '' and region = ''
        order by country_name`)
	return errors.Wrap(err, "Locations.ListCountries")
}

// This takes ~13s for a full iteration for the Cities database on my laptop
// (Countries is much faster, ~100ms) which is not a great worst case scenario,
// but in most cases it should be (much) faster, and this should get called
// extremely infrequently anyway, if ever.
func findGeoName(ctx context.Context, country, region string) (string, string) {
	geodb := geo.Get(ctx)
	if geodb == nil {
		panic("Location.Lookup: ")
	}

	_ = country
	_ = region

	// hasRegions := geodb.Metadata().DatabaseType == "City"
	iter := geodb.DB().Data()
	for iter.Next() {
		// var r struct {
		// 	Country struct {
		// 		ISOCode string            `maxminddb:"iso_code"`
		// 		Names   map[string]string `maxminddb:"names"`
		// 	} `maxminddb:"country"`
		// 	Subdivisions []struct {
		// 		ISOCode string            `maxminddb:"iso_code"`
		// 		Names   map[string]string `maxminddb:"names"`
		// 	} `maxminddb:"subdivisions"`
		// }
		var iso string
		err := iter.Data(&iso)
		if err != nil {
			log.Error(context.Background(), err)
			return "", ""
		}

		return iso3166_2_to_short_name[iso], ""

		// switch {
		// // Country database, no region.
		// case r.Country.ISOCode == country && !hasRegions:
		// 	return r.Country.Names["en"], ""
		// // City database, no region requested.
		// case r.Country.ISOCode == country && region == "":
		// 	return r.Country.Names["en"], ""
		// // Match region.
		// case r.Country.ISOCode == country && len(r.Subdivisions) > 0 && r.Subdivisions[0].ISOCode == region:
		// 	return r.Country.Names["en"], r.Subdivisions[0].Names["en"]
		// }
	}
	return "", ""
}
