package gen

import (
	"bufio"
	"io"
	"strconv"
)

// This is a writer for the plate-part family. The field ORDER and numeric widths
// are not guesswork: they come from the AVEVA serialisers decompiled in
// TbGenPostProcessor.dll and confirmed against the shipped samples.
//
//	CGeneralData::Write   (FUN_1015a4b0)  -> GENERAL_DATA field order
//	CPartData::Write      (FUN_1015fbc0)  -> PART_DATA field order
//	CMarkingData::Write   (FUN_1015e4b0)  -> MARKING_DATA field order
//	contour writer        (FUN_10157380)  -> NO_OF_SEG / START_U / START_V
//
// Numeric widths are the ones the AVEVA writers use: "%.2f" for areas, lengths
// and totals, "%.5f" for coordinates and amplitudes, plain integers otherwise.
const (
	fmtCoord = "%.5f"
	fmtTotal = "%.2f"
	// fmtExact is the shortest decimal that parses back to the same float64.
	// Used where the shipped writers do not use a fixed width: LISTED_PROFILE
	// carries values such as DENSITY=0.00000785 and AREA=2578.6708327727, which
	// "%.5f" would corrupt to 0.00001 and 2578.67083.
	fmtExact = "exact"

	lineEnd = "\r\n"
)

// WriteOptions controls output details.
type WriteOptions struct {
	// LF emits "\n" instead of the CRLF the AVEVA writers use.
	LF bool
	// OmitEmptyValues skips empty strings and zero numbers. AVEVA always emits
	// them, so the default (false) is the faithful setting.
	OmitEmptyValues bool
}

type encoder struct {
	w    *bufio.Writer
	opts WriteOptions
	eol  string

	// writeErr keeps the first failure so a broken writer is reported once,
	// rather than silently truncating the output.
	writeErr error
}

func newEncoder(w io.Writer, opts WriteOptions) *encoder {
	eol := lineEnd
	if opts.LF {
		eol = "\n"
	}

	return &encoder{w: bufio.NewWriter(w), opts: opts, eol: eol}
}

// raw is the single write path; every byte written is checked here.
func (e *encoder) raw(s string) {
	if e.writeErr != nil {
		return
	}

	if _, err := e.w.WriteString(s); err != nil {
		e.writeErr = err
	}
}

func (e *encoder) marker(name string) {
	e.raw(name)
	e.raw(e.eol)
}

func (e *encoder) str(key, value string) {
	if e.opts.OmitEmptyValues && value == "" {
		return
	}
	e.raw(key)
	e.raw("=")
	e.raw(value)
	e.raw(e.eol)
}

func (e *encoder) f(key string, value float64, format string) {
	if e.opts.OmitEmptyValues && value == 0 {
		return
	}
	e.raw(key)
	e.raw("=")
	e.raw(formatFloat(value, format))
	e.raw(e.eol)
}

func (e *encoder) coord(key string, v float64) { e.f(key, v, fmtCoord) }
func (e *encoder) total(key string, v float64) { e.f(key, v, fmtTotal) }

// exact writes the shortest decimal that round-trips to v.
func (e *encoder) exact(key string, v float64) { e.f(key, v, fmtExact) }

// uintKey writes an unsigned value without a lossy conversion to int.
func (e *encoder) uintKey(key string, v uint64) {
	if e.opts.OmitEmptyValues && v == 0 {
		return
	}
	e.raw(key)
	e.raw("=")
	e.raw(strconv.FormatUint(v, 10))
	e.raw(e.eol)
}

func (e *encoder) integer(key string, v int) {
	if e.opts.OmitEmptyValues && v == 0 {
		return
	}
	e.raw(key)
	e.raw("=")
	e.raw(strconv.Itoa(v))
	e.raw(e.eol)
}

func (e *encoder) boolean(key string, v bool) {
	if v {
		e.integer(key, 1)
		return
	}
	e.integer(key, 0)
}

func (e *encoder) err() error {
	if err := e.w.Flush(); err != nil && e.writeErr == nil {
		return err
	}

	return e.writeErr
}

func formatFloat(v float64, format string) string {
	switch format {
	case fmtCoord:
		return strconv.FormatFloat(v, 'f', 5, 64)
	case fmtTotal:
		return strconv.FormatFloat(v, 'f', 2, 64)
	case fmtExact:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
}

// WriteParts writes a plate-part document for the given parts.
func WriteParts(w io.Writer, parts []*Part, opts ...WriteOptions) error {
	return WriteDocument(w, &Document{Parts: parts}, opts...)
}

// WriteDocument writes a plate-part generic file.
//
// The layout mirrors what the AVEVA writers emit: GENERAL_DATA, then every
// PART_DATA block, then each part's PART_INFORMATION scope followed by its
// payload sections. Document-level labels (the nest-wide labelling pass) are
// emitted after the last part, under a nameless scope, exactly as they were read.
//
// The LISTED_PROFILE family is not writable; this package reads it only.
func WriteDocument(w io.Writer, d *Document, opts ...WriteOptions) error {
	e := newEncoder(w, d.opts(opts))

	e.writeGeneralData(d.General)

	for _, p := range d.Parts {
		e.writePartData(p)
	}

	for _, p := range d.Parts {
		e.writePartInformation(p)
		e.writePayload(p)
	}

	if len(d.LabelTexts) > 0 || len(d.LabelSymbols) > 0 {
		// Nest-wide labelling pass: a scope with no owning part.
		e.marker("PART_INFORMATION")
		e.str("PART_NAME", "")
		e.str("TYPE_OF_WORK", "LABELLING")
		e.marker("END_OF_PART_INFORMATION")

		for _, lt := range d.LabelTexts {
			e.writeLabelText(lt)
		}
		for _, ls := range d.LabelSymbols {
			e.writeLabelSymbol(ls)
		}
	}

	return e.err()
}

func (d *Document) opts(overrides []WriteOptions) WriteOptions {
	if len(overrides) > 0 {
		return overrides[0]
	}

	return d.WriteOptions
}

func (e *encoder) writeGeneralData(g *GeneralData) {
	e.marker("GENERAL_DATA")

	if g == nil {
		g = &GeneralData{}
	}

	e.str("TYPE_OF_GENERIC_FILE", g.TypeOfGenericFile)
	e.str("TYPE_OF_MANUFACT", g.TypeOfManufact)
	e.str("BURNER_ID", g.BurnerID)
	e.str("NEST_NAME", g.NestName)
	e.str("RAW_NAME", g.RawName)
	e.total("RAW_LENGTH", g.RawLength)
	e.total("RAW_WIDTH", g.RawWidth)
	e.total("RAW_THICKNESS", g.Thickness)
	e.total("RAW_AREA", g.RawArea)
	e.str("QUALITY", g.Quality)
	e.total("DENSITY", g.Density)
	e.str("SURF_TREAT_TS", g.SurfTreatTS)
	e.str("SURF_TREAT_OS", g.SurfTreatOS)
	e.total("REST_LENGTH", g.RestLength)
	e.str("SPLIT_DATE", g.SplitDate)
	e.str("PRODUCTION_DATE", g.ProductionDate)
	e.integer("NO_OF_PARTS", g.NoOfParts)
	e.total("TOTAL_IDLE", g.TotalIdle)
	e.total("TOTAL_MARKING", g.TotalMarking)
	e.total("TOTAL_BURNING", g.TotalBurning)
	e.total("TOTAL_BLASTING", g.TotalBlasting)
	e.total("TOTAL_MARK_BLAST", g.TotalMarkBlast)
	e.integer("TOTAL_NO_OF_STARTS", g.TotalNoOfStarts)
	e.integer("NO_OF_BURNING_STARTS", g.NoOfBurningStarts)
	e.integer("NO_OF_MARKING_STARTS", g.NoOfMarkingStarts)
	e.integer("NO_OF_BLASTING_STARTS", g.NoOfBlastingStarts)
	e.integer("NO_OF_MARK_BLAST_STARTS", g.NoOfMarkBlastStarts)
	e.total("STARTING_POS_X", g.StartingPosX)
	e.total("STARTING_POS_Y", g.StartingPosY)
	e.integer("QUANTITY_NORMAL", g.QuantityNormal)
	e.integer("QUANTITY_MIRRORED", g.QuantityMirrored)
	e.integer("TOOL_ID", g.ToolID)
	e.str("PROCESS_ID", g.ProcessID)
	e.str("SEQUENCE", g.Sequence)
	e.total("DISTANCE_Y1_Y2", g.DistanceY1Y2)
	e.integer("CUTTING_GAS", g.CuttingGas)
	e.str("MATERIAL_CODE", g.MaterialCode)
	e.str("PURCHASE_INFO", g.PurchaseInfo)
	e.str("PLATE_ALIGNMENT", g.PlateAlignment)
	e.str("VERIFICATION_DATE", g.VerificationDate)

	e.marker("END_OF_GENERAL_DATA")
}

func (e *encoder) writePartData(p *Part) {
	e.marker("PART_DATA")

	e.str("NAME", p.Name)
	e.str("TYPE_OF_WORK", p.TypeOfWork)
	e.str("POSNO", p.PosNo)
	e.str("PARTNAME_LONG", p.PartnameLong)
	e.str("PARTNAME_LONG_SB", p.PartnameLongSB)
	e.str("SHIP_NO", p.ShipNo)
	e.str("ASSEMBLY_LOW", p.AssemblyLow)
	e.str("ASSEMBLY_LOW_SB", p.AssemblyLowSB)
	e.str("ASSEMBLY_HIGH", p.AssemblyHigh)
	e.str("ASSEMBLY_HIGH_SB", p.AssemblyHighSB)
	e.str("BLOCK_NO", p.BlockNo)
	e.total("PART_AREA", p.PartArea)
	e.boolean("MIRRORED", p.Mirrored)

	e.str("CLIP_GRINDING_SIDE", p.ClipGrindingSide)
	e.str("CUST_DATA_1", p.CustData1)
	e.str("CUST_DATA_2", p.CustData2)
	e.str("CUST_DATA_3", p.CustData3)
	e.str("CUST_DATA_4", p.CustData4)
	e.str("CUST_TEXT_1", p.CustText1)
	e.str("CUST_TEXT_2", p.CustText2)
	e.str("CUST_TEXT_3", p.CustText3)
	e.str("CUST_TEXT_4", p.CustText4)
	e.str("CUST_TEXT_5", p.CustText5)
	e.str("CUST_TEXT_6", p.CustText6)
	e.str("CUST_TEXT_7", p.CustText7)
	e.str("CUST_TEXT_8", p.CustText8)
	e.str("CUST_TEXT_9", p.CustText9)
	e.str("CUST_TEXT_10", p.CustText10)

	e.str("WORKING_LOCATION", p.WorkingLocation)
	e.str("BUILDING_STRATEGY", p.BuildingStrategy)
	e.str("PLANNING_UNIT", p.PlanningUnit)
	e.coord("EXTENSION_U", p.ExtensionU)
	e.coord("EXTENSION_V", p.ExtensionV)
	e.str("PLATE_SIDE", p.PlateSide)
	e.str("PANEL_NAME", p.PanelName)
	if p.PartCog != nil {
		e.coord("PART_COG_U", p.PartCog.X)
		e.coord("PART_COG_V", p.PartCog.Y)
	} else {
		e.coord("PART_COG_U", 0)
		e.coord("PART_COG_V", 0)
	}
	e.str("NESTED_ON", p.NestedOn)
	e.str("KNUCKLED", p.Knuckled)
	e.str("RAW_NAME", p.RawName)
	e.str("DESTINATION", p.Destination)
	e.integer("PART_ID", p.PartID)
	e.str("COMMENT", p.Comment)
	e.str("FUNCTIONAL_DESCR_CODE", p.FunctionalDescrCode)
	e.str("FUNCTIONAL_DESCR_TEXT", p.FunctionalDescrText)

	if p.Transformation != nil {
		e.marker("TRANSFORMATION_DATA")
		e.str("TRANSFORMATION_TYPE", p.Transformation.Type)
		e.coord("TRANSFORMATION_1", p.Transformation.Transformation1)
		e.coord("TRANSFORMATION_2", p.Transformation.Transformation2)
		e.coord("TRANSFORMATION_3", p.Transformation.Transformation3)
		e.coord("TRANSFORMATION_4", p.Transformation.Transformation4)
		e.coord("TRANSFORMATION_5", p.Transformation.Transformation5)
		e.coord("TRANSFORMATION_6", p.Transformation.Transformation6)
		e.marker("END_OF_TRANSFORMATION_DATA")
	}

	e.integer("NO_INTERVALS_EXCESS_1", p.NoIntervalsExcess1)
	e.coord("LENGTH_EXCESS_1", p.LengthExcess1)
	e.integer("NO_INTERVALS_EXCESS_2", p.NoIntervalsExcess2)
	e.coord("LENGTH_EXCESS_2", p.LengthExcess2)
	e.integer("NO_INTERVALS_EXCESS_3", p.NoIntervalsExcess3)
	e.coord("LENGTH_EXCESS_3", p.LengthExcess3)

	e.marker("END_OF_PART_DATA")
}

func (e *encoder) writePartInformation(p *Part) {
	// A part with no scope had no payload either (payload is attached through a
	// scope), so inventing one would change the document's shape.
	if len(p.Information) == 0 {
		return
	}

	for _, info := range p.Information {
		e.marker("PART_INFORMATION")

		name := info.PartName
		if name == "" {
			name = p.Name
		}
		e.str("PART_NAME", name)

		if info.TypeOfWork != "" {
			e.str("TYPE_OF_WORK", info.TypeOfWork)
		}

		e.marker("END_OF_PART_INFORMATION")
	}
}

func (e *encoder) writePayload(p *Part) {
	for i := range p.BurningData {
		e.writeBurningData(&p.BurningData[i])
	}

	for i := range p.IdleData {
		e.writeIdleData(p.IdleData[i])
	}

	for i := range p.MarkingData {
		e.writeMarking(&p.MarkingData[i])
	}

	for i := range p.GeometryData {
		e.writeGeometry(&p.GeometryData[i])
	}

	for i := range p.StringData {
		e.writeStringData(&p.StringData[i])
	}

	for i := range p.BumpData {
		e.writeBumpData(p.BumpData[i])
	}

	for i := range p.LabelTexts {
		e.writeLabelText(p.LabelTexts[i])
	}

	for i := range p.LabelSymbols {
		e.writeLabelSymbol(p.LabelSymbols[i])
	}

	for i := range p.Edges {
		e.writeEdge(&p.Edges[i])
	}

	for i := range p.Auxiliaries {
		e.writeAuxiliary(p.Auxiliaries[i])
	}

	for i := range p.RawSections {
		e.writeRawSection(&p.RawSections[i])
	}
}

func (e *encoder) writeBurningData(b *BurningData) {
	e.marker("BURNING_DATA")

	e.str("SHAPE", b.Shape)
	e.str("START_END_IN_GAP", b.StartEndInGap)
	e.str("BEVEL_DEFINED", b.BevelDefined)
	e.integer("DIRECTION", b.Direction)
	e.integer("NUMBER_OF_HEADS", b.NumberOfHeads)
	e.integer("GEOMETRY_VALID_FOR", b.GeometryValidFor)
	e.total("DISTANCE_Y1_Y2", b.DistanceY1Y2)

	if b.HoleShipside != "" {
		e.str("HOLE_SHIPSIDE", b.HoleShipside)
	}
	if b.Designation != "" {
		e.str("DESIGNATION", b.Designation)
	}
	if b.PartID != 0 {
		e.integer("PART_ID", b.PartID)
	}

	if b.GeometryData != nil {
		e.writeGeometry(b.GeometryData)
	}

	e.writeContour(b.Contour)

	e.marker("END_OF_BURNING_DATA")
}

func (e *encoder) writeContour(c *Contour) {
	if c == nil {
		return
	}

	e.marker("START_OF_CONTOUR")

	// The AVEVA writers put the contour's bevel before NO_OF_SEG.
	if c.ExcessData != nil {
		e.marker("EXCESS_DATA")
		e.total("EXCESS_VALUE", c.ExcessData.Value)
		e.marker("END_OF_EXCESS_DATA")
	}

	// Observed order:
	//
	//	START_OF_CONTOUR
	//	[EXCESS_DATA]
	//	[BEVEL_DATA]        <- bevel of segment 0
	//	NO_OF_SEG
	//	START_U / START_V
	//	segment 0 keys
	//	[BEVEL_DATA]        <- bevel of segment 1
	//	segment 1 keys
	//	...
	//	[AUXILIARY_DATA]
	//	END_OF_CONTOUR
	//
	// Bevels are per-segment and sparse: a segment inherits the last bevel seen,
	// so only changes are written.
	var last *Bevel

	if len(c.Segments) > 0 && c.Segments[0].Bevel != nil {
		e.writeBevel(c.Segments[0].Bevel)
		last = c.Segments[0].Bevel
	}

	e.integer("NO_OF_SEG", len(c.Segments))

	if len(c.Segments) > 0 && c.Segments[0].Start != nil {
		e.coord("START_U", c.Segments[0].Start.X)
		e.coord("START_V", c.Segments[0].Start.Y)
	} else {
		e.coord("START_U", 0)
		e.coord("START_V", 0)
	}

	for i := range c.Segments {
		seg := c.Segments[i]

		if i > 0 && !sameBevel(last, seg.Bevel) {
			if seg.Bevel != nil {
				e.writeBevel(seg.Bevel)
			}
			last = seg.Bevel
		}

		e.writeSegment(seg)
	}

	for i := range c.Auxiliaries {
		e.writeAuxiliary(c.Auxiliaries[i])
	}

	e.marker("END_OF_CONTOUR")
}

func (e *encoder) writeSegment(s *Segment) {
	e.coord("AMP_U", s.AmpU)
	e.coord("AMP_V", s.AmpV)
	e.coord("AMP", s.Amp)
	e.coord("RADIUS", s.Radius)
	e.coord("SWEEP", s.Sweep)

	if s.Origin != nil {
		e.coord("ORIGIN_U", s.Origin.X)
		e.coord("ORIGIN_V", s.Origin.Y)
	} else {
		e.coord("ORIGIN_U", 0)
		e.coord("ORIGIN_V", 0)
	}

	if s.End != nil {
		e.coord("U", s.End.X)
		e.coord("V", s.End.Y)
	} else {
		e.coord("U", 0)
		e.coord("V", 0)
	}
}

func (e *encoder) writeBevel(b *Bevel) {
	e.marker("BEVEL_DATA")

	e.str("BEVEL", b.Bevel)
	if b.BevelCVBA != "" {
		e.str("BEVEL_CVBA", b.BevelCVBA)
	}
	e.total("BEVEL_CODE", b.BevelCode)
	e.str("BEVEL_NAME", b.BevelName)
	e.integer("BEVEL_TYPE", b.BevelType)
	e.integer("BEVEL_VARIANT", b.BevelVariant)
	e.total("PLATE_THICKNESS", b.PlateThickness)
	e.total("E", b.E)
	e.total("ANGLE_TS", b.AngleTS)
	e.total("ANGLE2_TS", b.Angle2TS)
	e.total("ANGLE_OS", b.AngleOS)
	e.total("ANGLE2_OS", b.Angle2OS)
	e.total("DEPTH_TS", b.DepthTS)
	e.total("DEPTH_OS", b.DepthOS)
	e.total("CHAMFER_WIDTH_TS", b.ChamferWidthTS)
	e.total("CHAMFER_WIDTH_OS", b.ChamferWidthOS)
	e.total("ANGLE2_WTS", b.Angle2Wts)
	e.total("ANGLE2_WOS", b.Angle2Wos)
	e.total("CHAMFER_HEIGHT_TS", b.ChamferHeightTS)
	e.total("CHAMFER_HEIGHT_OS", b.ChamferHeightOS)

	if b.WeldTS != 0 {
		e.total("WELD_TS", b.WeldTS)
	}
	if b.WeldOS != 0 {
		e.total("WELD_OS", b.WeldOS)
	}
	if b.Manual != "" {
		e.str("MANUAL", b.Manual)
	}

	e.marker("END_OF_BEVEL_DATA")
}

func (e *encoder) writeMarking(m *Marking) {
	e.marker("MARKING_DATA")

	e.str("MARKING_SIDE", m.MarkingSide)
	e.str("MARKING_TYPE", m.MarkingType)
	e.str("MARKING_NAME", m.MarkingName)
	e.integer("MARKING_USER_TYPE", m.MarkingUserType)
	e.str("MARKING_ATTRI", m.MarkingAttri)
	e.str("MARKING_ATTRI_SB", m.MarkingAttriSb)
	e.str("MARKING_SHIPSIDE", m.MarkingShipside)
	e.str("MARKING_ASSEMBLY_LOW", m.MarkingAssemblyLow)
	e.total("MARKING_PROFILE_THICKNESS", m.MarkingProfileThickness)
	e.total("MARKING_WELD", m.MarkingWeld)
	e.integer("MARKING_DIRECTION", m.MarkingDirection)
	e.integer("NUMBER_OF_HEADS", m.NumberOfHeads)
	e.integer("GEOMETRY_VALID_FOR", m.GeometryValidFor)
	e.total("DISTANCE_Y1_Y2", m.DistanceY1Y2)
	e.coord("MTRL_SIDE_U", m.MtrlSideU)
	e.coord("MTRL_SIDE_V", m.MtrlSideV)
	e.str("COMMENT", m.Comment)
	e.total("INCLINATION_ANGLE", m.InclinationAngle)

	e.writeContour(m.Contour)

	e.marker("END_OF_MARKING_DATA")
}

func (e *encoder) writeGeometry(g *Geometry) {
	e.marker("GEOMETRY_DATA")

	e.str("GEOMETRY_TYPE", g.Type)
	if g.RollAxisNo != 0 {
		e.integer("ROLL_AXIS_NO", g.RollAxisNo)
	}

	e.writeContour(g.Contour)

	e.marker("END_OF_GEOMETRY_DATA")
}

func (e *encoder) writeStringData(s *StringData) {
	e.marker("STRING_DATA")

	e.str("STRING_TYPE", s.Type)
	if s.Pos != nil {
		e.coord("STRING_POSITION_U", s.Pos.X)
		e.coord("STRING_POSITION_V", s.Pos.Y)
	} else {
		e.coord("STRING_POSITION_U", 0)
		e.coord("STRING_POSITION_V", 0)
	}
	e.total("STRING_ANGLE", s.Angle)
	e.total("STRING_HEIGHT", s.Height)
	e.str("STRING", s.String)

	e.marker("END_OF_STRING_DATA")
}

func (e *encoder) writeLabelText(l LabelText) {
	e.marker("LABELTEXT_DATA")

	e.str("TEXT_TYPE", l.Type)
	e.coord("TEXT_POSITION_U", l.PositionU)
	e.coord("TEXT_POSITION_V", l.PositionV)
	e.total("TEXT_ANGLE", l.Angle)
	e.total("TEXT_HEIGHT", l.Height)
	e.integer("TEXT_DIRECTION", l.Direction)
	e.str("TEXT_SIDE", l.Side)
	e.str("TEXT", l.Text)
	e.integer("TEXT_ELEMENTS", len(l.Geometry))

	for i := range l.Geometry {
		e.writeGeometry(&l.Geometry[i])
	}

	e.marker("END_OF_LABELTEXT_DATA")
}

func (e *encoder) writeLabelSymbol(s LabelSymbol) {
	e.marker("LABELSYMBOL_DATA")

	e.str("SYMBOL_TYPE", s.Type)
	e.coord("SYMBOL_POSITION_U", s.PositionU)
	e.coord("SYMBOL_POSITION_V", s.PositionV)
	e.total("SYMBOL_ANGLE", s.Angle)
	e.total("SYMBOL_HEIGHT", s.Height)
	e.integer("SYMBOL_FONT", s.Font)
	e.integer("SYMBOL_NUMBER", s.Number)

	e.marker("END_OF_LABELSYMBOL_DATA")
}

func (e *encoder) writeEdge(edge *Edge) {
	e.marker("EDGE_DATA")

	e.str("NAME", edge.Name)
	e.coord("START_U", edge.StartU)
	e.coord("START_V", edge.StartV)
	e.coord("END_U", edge.EndU)
	e.coord("END_V", edge.EndV)
	e.total("ALENGTH", edge.ALength)
	e.total("MLENGTH", edge.MLength)

	e.marker("END_OF_EDGE_DATA")
}

func (e *encoder) writeAuxiliary(a Auxiliary) {
	e.marker("AUXILIARY_DATA")

	e.str("AUXILIARY_FUNCTION", a.Function)
	if a.Parameter != "" {
		e.str("AUXILIARY_PARAMETER", a.Parameter)
	}

	e.marker("END_OF_AUXILIARY_DATA")
}

func (e *encoder) writeRawSection(rs *RawSection) {
	e.marker(rs.Name)

	for _, f := range rs.Fields {
		if f.Value == "" {
			e.marker(f.Key)
			continue
		}
		e.str(f.Key, f.Value)
	}

	e.marker("END_OF_" + rs.Name)
}

// sameBevel reports whether two bevels would serialise identically. The reader
// makes every segment inherit the last BEVEL_DATA, so the writer only needs to
// emit a block when the value actually changes.
func sameBevel(a, b *Bevel) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}

	return *a == *b
}

func (e *encoder) writeIdleData(d IdleData) {
	e.marker("IDLE_DATA")

	e.writeContour(d.Contour)

	e.marker("END_OF_IDLE_DATA")
}

func (e *encoder) writeBumpData(d BumpData) {
	e.marker("BUMP_DATA")

	e.str("SHAPE", d.Shape)

	e.writeContour(d.Contour)

	e.marker("END_OF_BUMP_DATA")
}
