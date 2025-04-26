package aggregate

// ZoneInfo represents basic information about a zone
type ZoneInfo struct {
	zone     string
	category string
}

// NewZoneInfo creates a new ZoneInfo
func NewZoneInfo() *ZoneInfo {
	return &ZoneInfo{}
}

// GetZone returns the zone name
func (z *ZoneInfo) GetZone() string {
	return z.zone
}

// GetCategory returns the category
func (z *ZoneInfo) GetCategory() string {
	return z.category
}

// SetZone sets the zone name
func (z *ZoneInfo) SetZone(zone string) {
	z.zone = zone
}

// SetCategory sets the category
func (z *ZoneInfo) SetCategory(category string) {
	z.category = category
}
