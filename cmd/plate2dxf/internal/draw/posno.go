package draw

import (
	"fmt"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"

	gen "github.com/edanko/aveva-go"
)

func (d *Drawer) addPosNo(p *gen.Part) {
	// No centre of gravity means no place to anchor the label. Parts are
	// expected to carry one, but this is a library call on a public type, so a
	// missing cog must not panic.
	if p.PartCog == nil {
		return
	}

	text := fmt.Sprintf("%s-%s-%s", p.ShipNo, p.BlockNo, p.PosNo)

	if p.Mirrored {
		text += "(m)"
	}

	t, err := d.drawing.Text(text, p.PartCog.X, p.PartCog.Y, 0, posNoTextHeight)
	if err != nil {
		panic(err)
	}
	t.SetColor(color.Red)
	t.Anchor(entity.CENTER_TOP)
}
