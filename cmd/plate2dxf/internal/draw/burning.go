package draw

import (
	gen "github.com/edanko/aveva-go"
)

func (d *Drawer) addBurningData(bds []gen.BurningData) {
	if bds == nil {
		return
	}
	for _, bd := range bds {
		d.drawBurningContour(bd.Contour, burningColor)
	}
}
