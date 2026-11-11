package draw

import (
	gen "github.com/edanko/aveva-go"
)

func (d *Drawer) addGeometryData(gds []gen.Geometry) {
	if gds == nil {
		return
	}
	for _, gd := range gds {
		if gd.Type == "BEVEL_SYMBOL" {
			continue
		}
		d.drawContour(gd.Contour, geometryColor)
	}
}
