package draw

import (
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"

	gen "github.com/edanko/aveva-go"
)

const (
	posNoTextHeight        float64 = 25
	markingTextHeight      float64 = 20
	bevelMarkingTextHeight float64 = 20
	attrTextHeight         float64 = 5.0
	attrSpacing            float64 = 10
	addAngleLength         float64 = 10
	attrColor                      = color.White
	burningColor                   = color.Green
	geometryColor                  = color.Cyan
	markingTSColor                 = color.Cyan
	markingTextTSColor             = color.Red
	markingOSColor                 = 30 // orange
	markingTextOSColor             = 30 // orange
)

type Drawer struct {
	drawing *drawing.Drawing
	writer  io.Writer

	// autoMirror enables the historical heuristic: flip the part when the OS
	// marking length exceeds the TS length. It is on by default and is disabled
	// when the caller has decided the orientation explicitly.
	autoMirror bool
}

func NewDrawer(writer io.Writer) (*Drawer, error) {
	d, err := drawing.New()
	if err != nil {
		return nil, err
	}
	return &Drawer{
		drawing:    d,
		writer:     writer,
		autoMirror: true,
	}, nil
}

// SetAutoMirror enables or disables the OS-longer-than-TS mirror heuristic.
func (d *Drawer) SetAutoMirror(enabled bool) {
	d.autoMirror = enabled
}

func (d *Drawer) PartToDXF(p *gen.Part) error {
	var tsLength, osLength float64

	for _, m := range p.MarkingData {
		var l float64

		for _, s := range m.Contour.Segments {
			l += s.Length()
		}

		switch m.MarkingSide {
		case "TS":
			tsLength += l
		case "OS":
			osLength += l
		case "BOTH":
			tsLength += l
			osLength += l
		}
	}

	if d.autoMirror && osLength > tsLength {
		p.Mirror()
	}

	quantity := strconv.Itoa(p.Quantity)
	thickness := strconv.FormatFloat(p.Thickness, 'g', -1, 64)
	order := p.ShipNo
	section := p.BlockNo
	posno := p.PosNo
	ina := order + section + posno
	err := d.addAttrs(p.Quality, quantity, thickness, order, section, posno, ina)
	if err != nil {
		return err
	}

	d.addPosNo(p)
	d.addBurningData(p.BurningData)
	d.addMarkingData(p.MarkingData)
	d.addStringData(p.StringData)
	d.addGeometryData(p.GeometryData)

	_, err = d.drawing.WriteTo(d.writer)
	if err != nil {
		return err
	}

	err = d.drawing.Close()
	if err != nil {
		return err
	}

	return nil
}

func (d *Drawer) drawContour(c *gen.Contour, cl color.ColorNumber) {
	if c == nil {
		return
	}

	vertices := c.Vertices()
	if len(vertices) == 0 {
		return
	}

	bulges := c.Bulges()

	polyline, err := d.drawing.LwPolyline(false, vertices...)
	if err != nil {
		panic(err)
	}
	polyline.Bulges = bulges
	polyline.SetColor(cl)
}

func (d *Drawer) addBevel(c *gen.Contour, from, to int) {
	if c == nil || len(c.Segments) == 0 || from >= len(c.Segments) {
		return
	}

	longestIdx := from
	longestLen := c.Segments[from].Length()

	var x, y, rotation float64

	for i := from + 1; i <= to; i++ {
		l := c.Segments[i].Length()

		if l > longestLen {
			longestLen = l
			longestIdx = i
		}
	}

	x, y = gen.SegmentMidPoint(c.Segments[longestIdx])
	rotation = gen.ToDeg(gen.SegmentAngle(c.Segments[longestIdx]))

	text, attrText := parseBevel(c.Segments[from].Bevel)

	// nestix

	n, err := d.drawing.Text(attrText, x, y, 0, 1)
	if err != nil {
		panic(err)
	}
	n.Rotation = rotation
	n.SetColor(color.White)
	// ---

	textRotation := rotation

	if c.IsClockwise() {
		x, y = gen.AddAngle(x, y, rotation-90, addAngleLength)
		textRotation = rotation + 180
	} else {
		x, y = gen.AddAngle(x, y, rotation+90, addAngleLength)
	}
	t, err := d.drawing.Text(text, x, y, 0, bevelMarkingTextHeight)
	if err != nil {
		panic(err)
	}
	t.Rotation = textRotation
	t.SetColor(color.Red)
	t.Anchor(entity.CENTER_BOTTOM)
}

func (d *Drawer) drawBurningContour(c *gen.Contour, cl color.ColorNumber) {
	from := 0

	// to is assigned at the top of every loop iteration before it is read.
	var to int

	lastBevelCode := 0.0
	if c.Segments[0].Bevel != nil {
		lastBevelCode = c.Segments[0].Bevel.BevelCode
	}

	vertices := make([][]float64, 0, len(c.Segments))
	bulges := make([]float64, 0, len(c.Segments))

	for i := 0; i < len(c.Segments); i++ {
		to = i

		if c.Segments[i].Bevel != nil && c.Segments[i].Bevel.BevelCode != lastBevelCode || i == len(c.Segments)-1 {
			vertices = append(vertices, c.Segments[from].Start.Vertice())
			b := c.Segments[from].Bulge()
			bulges = append(bulges, b)

			for j := from; j < to; j++ {
				b = 0.0
				if j < len(c.Segments)-1 {
					b = c.Segments[j+1].Bulge()
				}

				vertices = append(vertices, c.Segments[j].End.Vertice())
				bulges = append(bulges, b)
			}

			if i == len(c.Segments)-1 {
				vertices = append(vertices, c.Segments[to].End.Vertice())
				bulges = append(bulges, 0.0)
			}

			d.addBevel(c, from, to-1)

			if c.Segments[i].Bevel != nil {
				lastBevelCode = c.Segments[i].Bevel.BevelCode
			}
			from = i
		}
	}

	polyline, err := d.drawing.LwPolyline(false, vertices...)
	if err != nil {
		panic(err)
	}
	polyline.Bulges = bulges
	polyline.SetColor(cl)
}

func (d *Drawer) drawMarkingText(md gen.Marking, color color.ColorNumber) {
	x, y, rot := getMarkingTextPosition(md.Contour)

	if strings.Contains(md.MarkingName, "ROLL AXIS") {
		d.drawText(x, y, markingTextHeight, rot, md.MarkingName, color)
		return
	}

	if md.MarkingType == "INCLINATION_ANGLE_LINE" {
		angle := strconv.FormatFloat(md.InclinationAngle, 'f', -1, 64)

		if md.Contour == nil || len(md.Contour.Segments) == 0 {
			return
		}

		s := md.Contour.Segments[0].Start
		e := md.Contour.Segments[len(md.Contour.Segments)-1].End

		x, y = gen.LineMidPoint(s, e)
		rot = gen.Angle(s, e)
		rot = gen.ToDeg(rot)

		d.drawText(x, y, markingTextHeight, rot, angle, color)
		return
	}

	d.drawText(x, y, markingTextHeight, rot, filterPos(md.MarkingAttri), color)
}

func (d *Drawer) drawText(x, y, height, rotation float64, text string, color color.ColorNumber) {
	t, err := d.drawing.Text(text, x, y, 0, height)
	if err != nil {
		panic(err)
	}
	t.SetColor(color)
	t.Rotation = rotation
	t.Anchor(entity.CENTER_BOTTOM)
}

func getMarkingTextPosition(c *gen.Contour) (x float64, y float64, rot float64) {
	if c == nil || len(c.Segments) == 0 {
		return 0, 0, 0
	}

	if len(c.Segments) > 1 {
		segmentsCount := len(c.Segments) - 1

		found := false

		for i := 0; i < len(c.Segments); i++ {
			if math.Round(c.Segments[i].Radius) == 10 {
				if i < segmentsCount-2 {
					if math.Round(c.Segments[i+1].Radius) == 30 || math.Round(c.Segments[i+1].Radius) == 35 || math.Round(c.Segments[i+1].Radius) == 50 {
						if math.Round(c.Segments[i+2].Radius) == 10 {
							found = true

							// i may be 0, so fall back to the first segment.
							var s *gen.Point
							if i > 0 {
								s = c.Segments[i-1].End
							} else {
								s = c.Segments[0].Start
							}

							e := c.Segments[i+2].End

							x, y = gen.LineMidPoint(s, e)
							rot = gen.Angle(s, e)
							rot = gen.ToDeg(rot)
							break
						}
					}
				}
			}
		}

		if !found {
			longestIdx := 0
			longestLen := c.Segments[0].Length()

			for i := 1; i < len(c.Segments); i++ {
				l := c.Segments[i].Length()
				if l > longestLen {
					longestLen = l
					longestIdx = i
				}
			}

			seg := c.Segments[longestIdx]
			x, y = gen.SegmentMidPoint(seg)
			rot = gen.SegmentAngle(seg)
			rot = gen.ToDeg(rot)
		}
	} else {
		seg := c.Segments[0]
		x, y = gen.SegmentMidPoint(seg)
		rot = gen.SegmentAngle(seg)
		rot = gen.ToDeg(rot)
	}

	x, y = gen.AddAngle(x, y, rot, addAngleLength)
	return
}

func filterPos(pos string) string {
	ss := strings.Split(pos, "-")
	s := ss[len(ss)-1]
	s = strings.ReplaceAll(s, "P", "")
	s = strings.ReplaceAll(s, "S", "")
	s = strings.ReplaceAll(s, "B", "")
	s = strings.ReplaceAll(s, "C", "")

	return s
}
