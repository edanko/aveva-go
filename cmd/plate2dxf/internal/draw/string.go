package draw

import (
	"github.com/edanko/dxf/color"

	gen "github.com/edanko/aveva-go"
)

func (d *Drawer) addStringData(sds []gen.StringData) {
	if sds == nil {
		return
	}
	for _, sd := range sds {
		// Position is optional in the parsed model; skip rather than panic.
		if sd.Type == "EXCESS_GEOMETRY" && sd.Pos != nil {
			d.drawText(sd.Pos.X, sd.Pos.Y, 25, sd.Angle, sd.String, color.Red)
		}
	}
}
