package geo

// Member present the Coordinate and the key of data
type Member struct {
	Name       string     `json:"name"`
	Coordinate Coordinate `json:"coordinate"`
}

// NewMember create a meta dta
func NewMember(name string, lat, lon float64) *Member {
	return &Member{
		Name: name,
		Coordinate: Coordinate{
			Lat: lat,
			Lon: lon,
		},
	}
}
