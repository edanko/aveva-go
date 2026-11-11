package gen

import (
	"io"
)

// This writer covers the LISTED_PROFILE family; it is the counterpart of write.go.
//
// Key order comes from testdata/profiles.gen as the UNION of keys across all 440
// profiles in first-seen order, not the first profile alone: 20 bevel keys
// (BEVEL_CODE, E, GAP, ALPHA, ...) appear only on profiles that carry a bevel, so
// sampling one block silently dropped them.
//
// Floats use the shortest round-tripping representation rather than a fixed
// width, because this family carries values such as DENSITY=0.00000785 and
// AREA=2578.6708327727 that "%.5f" would corrupt.
//
// ENDCUT_MASK is a string: observed values are bit patterns such as
// 00010110000 whose leading zeros would not survive an integer round-trip.

// ProfileDocument is a parsed LISTED_PROFILE file.
type ProfileDocument struct {
	Version  string
	Usage    string
	Common   *CommonData
	Profiles []*Profile

	// WriteOptions, when set, is used by WriteProfiles unless an explicit
	// option is passed to that call.
	WriteOptions WriteOptions
}

// WriteProfiles writes a LISTED_PROFILE document.
func WriteProfiles(w io.Writer, doc *ProfileDocument, opts ...WriteOptions) error {
	o := doc.WriteOptions
	if len(opts) > 0 {
		o = opts[0]
	}

	e := newEncoder(w, o)

	e.str("TYPE_OF_GENERIC_FILE", "LISTED_PROFILE")
	e.str("VERSION", nonEmpty(doc.Version, "1.0"))
	e.str("USAGE", nonEmpty(doc.Usage, "LIST"))

	e.writeCommonData(doc.Common)

	for _, p := range doc.Profiles {
		e.writeProfileData(p)
		e.writeEnd("LEFT_END", p.LeftEnd)
		e.writeEnd("RIGHT_END", p.RightEnd)
	}

	for _, p := range doc.Profiles {
		for i := range p.RawSections {
			e.writeRawSection(&p.RawSections[i])
		}
	}

	return e.err()
}

func nonEmpty(v, fallback string) string {
	if v == "" {
		return fallback
	}

	return v
}

func (e *encoder) writeCommonData(cd *CommonData) {
	e.marker("COMMON_DATA")

	if cd == nil {
		cd = &CommonData{}
	}

	e.str("NEST_NAME", cd.NestName)
	e.str("SHAPE", cd.Shape)
	e.writeContour(cd.Contour)
	e.str("DIMENSION", cd.Dimension)
	e.integer("WEB_HEIGHT", cd.WebHeight)
	e.integer("WEB_THICKNESS", cd.WebThickness)
	e.integer("MAX_WEB_HEIGHT", cd.MaxWebHeight)
	e.exact("FLANGE_WIDTH", cd.FlangeWidth)
	e.exact("FLANGE_THICKNESS", cd.FlangeThickness)
	e.exact("RAW_LENGTH", cd.RawLength)
	e.exact("REST_LENGTH", cd.RestLength)
	e.exact("USED_LENGTH", cd.UsedLength)
	e.exact("LAST_SCRAP_LENGTH", cd.LastScrapLength)
	e.str("QUALITY", cd.Quality)
	e.exact("DENSITY", cd.Density)
	e.exact("AREA", cd.Area)
	e.exact("TOTAL_WEIGHT", cd.TotalWeight)
	e.str("RAW_BAR", cd.RawBar)
	e.str("BUYING_MARK", cd.BuyingMark)
	e.integer("CUTTING_STN", cd.CuttingStn)
	e.integer("TEXT_HEIGHT", cd.TextHeight)
	e.integer("TEXT_WIDTH", cd.TextWidth)
	e.integer("TEXT_PLACING", cd.TextPlacing)
	e.integer("TEXT_DIST_TOP", cd.TextDistTop)
	e.integer("TEXT_MAX1", cd.TextMax1)
	e.integer("TEXT_MAX2", cd.TextMax2)
	e.integer("TEXT_HOR1", cd.TextHor1)
	e.integer("TEXT_VER1", cd.TextVer1)
	e.integer("TEXT_HOR2", cd.TextHor2)
	e.integer("TEXT_VER2", cd.TextVer2)
	e.integer("TEXT_HOR3", cd.TextHor3)
	e.integer("TEXT_VER3", cd.TextVer3)
	e.integer("TSCRAP", cd.Tscrap)
	e.uintKey("NO_OF_PROFS", cd.NoOfProfs)
	e.integer("REST", cd.Rest)
	e.exact("GSD", cd.Gsd)
	e.exact("GSD_DIST", cd.GsdDist)
	e.integer("GSD_ONE_LIM", cd.GsdOneLim)
	e.marker("END_OF_COMMON_DATA")
}

func (e *encoder) writeProfileData(p *Profile) {
	e.marker("PROFILE_DATA")

	e.str("NAME", p.Name)
	e.str("GENERATED_AS", p.GeneratedAs)
	e.str("POSNO", p.PosNo)
	e.exact("MLENGTH", p.MLength)
	e.exact("MLENGTH_MANUAL", p.MLengthManual)
	e.exact("TLENGTH", p.TLength)
	e.exact("TLENGTH_MANUAL", p.TLengthManual)
	e.exact("END_POINT_DIST", p.EndPointDist)
	e.exact("END_POINT_DIST_PROD", p.EndPointDistProd)
	e.exact("TRACE_LENGTH", p.TraceLength)
	e.exact("TRACE_LENGTH_PROD", p.TraceLengthProd)
	e.exact("TWIST_ANGLE", p.TwistAngle)
	e.exact("APPROX_PART_WEIGHT", p.ApproxPartWeight)
	e.str("IDENT_STRING", p.IdentString)
	e.str("PARTNAME_LONG", p.PartnameLong)
	e.str("PARTNAME_SHORT", p.PartnameShort)
	e.str("SHIP_NO", p.ShipNo)
	e.str("SHIP_NAME", p.ShipName)
	e.str("PROF_SIDE", p.ProfSide)
	e.str("SIDE_UP", p.SideUp)
	e.str("DIRECTION", p.Direction)
	e.str("DIRECTION_SIGN", p.DirectionSign)
	e.str("DIRECTION_TRACE", p.DirectionTrace)
	e.integer("BEVEL_TRACE", p.BevelTrace)
	e.integer("GRINDING", p.Grinding)
	e.str("ASSEMBLY", p.Assembly)
	e.str("ASSEMBLY_HIGH", p.AssemblyHigh)
	e.str("ASSEMBLY_TOT", p.AssemblyTotal)
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
	e.str("SURFACE_TREATMENT", p.SurfaceTreatment)
	e.str("PLANNING_UNIT", p.PlanningUnit)
	e.str("WORKING_LOCATION", p.WorkingLocation)
	e.str("BUILDING_STRATEGY", p.BuildingStrategy)
	e.str("ASSEMBLY_SEQUENCE", p.AssemblySequence)
	e.str("BAUGRUPPE", p.Baugruppe)
	e.str("ZEICHNUNGSNO", p.Zeichnungsno)
	e.str("ZAEHLNO", p.Zaehlno)
	e.str("BLOCK_NO", p.BlockNo)
	e.str("BLOCK_NAME", p.BlockName)
	e.integer("NO_OF_REF_LINES", p.NoOfRefLines)
	e.str("MIRROR", p.Mirror)
	e.str("FORM", p.Form)
	e.str("BEND_DIRECTION", p.BendDirection)
	e.str("ADDITIONAL_INFO1", p.AdditionalInfo1)
	e.str("ADDITIONAL_INFO2", p.AdditionalInfo2)
	e.str("ADDITIONAL_INFO3", p.AdditionalInfo3)
	e.str("ADDITIONAL_INFO4", p.AdditionalInfo4)
	e.str("ADDITIONAL_INFO5", p.AdditionalInfo5)
	e.str("ADDITIONAL_INFO6", p.AdditionalInfo6)
	e.str("ADDITIONAL_INFO7", p.AdditionalInfo7)
	e.str("ADDITIONAL_INFO8", p.AdditionalInfo8)
	e.str("ADDITIONAL_INFO9", p.AdditionalInfo9)
	e.str("ADDITIONAL_INFO10", p.AdditionalInfo10)
	e.str("ADDITIONAL_INFO11", p.AdditionalInfo11)
	e.str("ADDITIONAL_INFO12", p.AdditionalInfo12)
	e.integer("MARKING_LEFT_EXCIST", p.MarkingLeftExcist)
	e.integer("MARKING_RIGHT_EXCIST", p.MarkingRightExcist)
	e.integer("NO_OF_MACS", p.NoOfMacs)
	e.integer("NO_OF_MARKS", p.NoOfMarks)
	e.integer("BEVEL_CODE", p.BevelCode)
	e.str("BEVEL_NAME", p.BevelName)
	e.integer("BEVEL_TYPE", p.BevelType)
	e.integer("BEVEL_VARIANT", p.BevelVariant)
	e.exact("E", p.E)
	e.exact("GAP", p.Gap)
	e.exact("CHAMFER", p.Chamfer)
	e.exact("ALPHA", p.Alpha)
	e.exact("BETA", p.Beta)
	e.exact("NOSE", p.Nose)
	e.exact("H", p.H)
	e.exact("H_FACT", p.HFact)
	e.exact("ANGLE_TS", p.AngleTS)
	e.exact("ANGLE_OS", p.AngleOS)
	e.exact("DEPTH_TS", p.DepthTS)
	e.exact("DEPTH_OS", p.DepthOS)
	e.exact("CHAMFER_WIDTH_TS", p.ChamferWidthTS)
	e.exact("CHAMFER_WIDTH_OS", p.ChamferWidthOS)
	e.exact("CHAMFER_HEIGHT_TS", p.ChamferHeightTS)
	e.exact("CHAMFER_HEIGHT_OS", p.ChamferHeightOS)

	e.marker("END_OF_PROFILE_DATA")
}

func (e *encoder) writeEnd(marker string, x *End) {
	if x == nil {
		return
	}

	e.marker(marker)

	e.exact("LEFT_ORIGIN", x.LeftOrigin)
	e.exact("LEFT_CLOSEST_POINT", x.LeftClosestPoint)
	e.exact("LEFT_FARTHEST_POINT", x.LeftFarthestPoint)
	e.exact("LEFT_V0", x.LeftV0)
	e.exact("LEFT_CLOSEST_POINT_WEB", x.LeftClosestPointWeb)
	e.exact("LEFT_FARTHEST_POINT_WEB", x.LeftFarthestPointWeb)
	e.integer("BURN_ECUT", x.BurnEcut)
	e.integer("STOSS", x.Stoss)
	e.exact("FABRICATION_EXCESS", x.FabricationExcess)
	e.integer("ENDCUT_TYPE", x.EndcutType)
	e.str("ENDCUT_MASK", x.EndcutMask)
	e.integer("ENDCUT_CODE", x.EndcutCode)
	e.exact("A", x.A)
	e.exact("B", x.B)
	e.exact("C", x.C)
	e.exact("R1", x.R1)
	e.exact("R2", x.R2)
	e.exact("V1", x.V1)
	e.exact("V2", x.V2)
	e.exact("V3", x.V3)
	e.exact("V4", x.V4)
	e.exact("KS", x.Ks)
	e.str("ENDCUT_USER_DEFINED_NAME", x.EndcutUserDefinedName)
	e.exact("ENDCUT_USER_DEFINED_PARAM1", x.EndcutUserDefinedParams[0])
	e.exact("ENDCUT_USER_DEFINED_PARAM2", x.EndcutUserDefinedParams[1])
	e.exact("ENDCUT_USER_DEFINED_PARAM3", x.EndcutUserDefinedParams[2])
	e.exact("ENDCUT_USER_DEFINED_PARAM4", x.EndcutUserDefinedParams[3])
	e.exact("ENDCUT_USER_DEFINED_PARAM5", x.EndcutUserDefinedParams[4])
	e.exact("ENDCUT_USER_DEFINED_PARAM6", x.EndcutUserDefinedParams[5])
	e.exact("ENDCUT_USER_DEFINED_PARAM7", x.EndcutUserDefinedParams[6])
	e.exact("ENDCUT_USER_DEFINED_PARAM8", x.EndcutUserDefinedParams[7])
	e.exact("ENDCUT_USER_DEFINED_PARAM9", x.EndcutUserDefinedParams[8])
	e.exact("ENDCUT_USER_DEFINED_PARAM10", x.EndcutUserDefinedParams[9])
	e.exact("ENDCUT_USER_DEFINED_PARAM11", x.EndcutUserDefinedParams[10])
	e.exact("ENDCUT_USER_DEFINED_PARAM12", x.EndcutUserDefinedParams[11])
	e.exact("ENDCUT_USER_DEFINED_PARAM13", x.EndcutUserDefinedParams[12])
	e.exact("ENDCUT_USER_DEFINED_PARAM14", x.EndcutUserDefinedParams[13])
	e.exact("ENDCUT_USER_DEFINED_PARAM15", x.EndcutUserDefinedParams[14])
	e.exact("ENDCUT_USER_DEFINED_PARAM16", x.EndcutUserDefinedParams[15])
	e.exact("ENDCUT_USER_DEFINED_PARAM17", x.EndcutUserDefinedParams[16])
	e.exact("ENDCUT_USER_DEFINED_PARAM18", x.EndcutUserDefinedParams[17])
	e.exact("ENDCUT_USER_DEFINED_PARAM19", x.EndcutUserDefinedParams[18])
	e.exact("ENDCUT_USER_DEFINED_PARAM20", x.EndcutUserDefinedParams[19])
	e.str("ENDCUT_USER_DEFINED_DATA1", x.EndcutUserDefinedData[0])
	e.str("ENDCUT_USER_DEFINED_DATA2", x.EndcutUserDefinedData[1])
	e.str("ENDCUT_USER_DEFINED_DATA3", x.EndcutUserDefinedData[2])
	e.str("ENDCUT_USER_DEFINED_DATA4", x.EndcutUserDefinedData[3])
	e.str("ENDCUT_USER_DEFINED_DATA5", x.EndcutUserDefinedData[4])
	e.str("ENDCUT_USER_DEFINED_DATA6", x.EndcutUserDefinedData[5])
	e.str("ENDCUT_USER_DEFINED_DATA7", x.EndcutUserDefinedData[6])
	e.str("ENDCUT_USER_DEFINED_DATA8", x.EndcutUserDefinedData[7])
	e.str("ENDCUT_USER_DEFINED_DATA9", x.EndcutUserDefinedData[8])
	e.str("ENDCUT_USER_DEFINED_DATA10", x.EndcutUserDefinedData[9])
	e.str("ENDCUT_USER_DEFINED_DATA11", x.EndcutUserDefinedData[10])
	e.str("ENDCUT_USER_DEFINED_DATA12", x.EndcutUserDefinedData[11])
	e.str("ENDCUT_USER_DEFINED_DATA13", x.EndcutUserDefinedData[12])
	e.str("ENDCUT_USER_DEFINED_DATA14", x.EndcutUserDefinedData[13])
	e.str("ENDCUT_USER_DEFINED_DATA15", x.EndcutUserDefinedData[14])
	e.str("ENDCUT_USER_DEFINED_DATA16", x.EndcutUserDefinedData[15])
	e.str("ENDCUT_USER_DEFINED_DATA17", x.EndcutUserDefinedData[16])
	e.str("ENDCUT_USER_DEFINED_DATA18", x.EndcutUserDefinedData[17])
	e.str("ENDCUT_USER_DEFINED_DATA19", x.EndcutUserDefinedData[18])
	e.str("ENDCUT_USER_DEFINED_DATA20", x.EndcutUserDefinedData[19])
	e.exact("EXCESS", x.Excess)
	e.integer("BEVEL_DEFINED", x.BevelDefined)
	e.integer("BEVEL_CODE", x.BevelCode)
	e.str("BEVEL_NAME", x.BevelName)
	e.integer("BEVEL_TYPE", x.BevelType)
	e.integer("BEVEL_VARIANT", x.BevelVariant)
	e.exact("E", x.E)
	e.exact("GAP", x.Gap)
	e.exact("CHAMFER", x.Chamfer)
	e.exact("ALPHA", x.Alpha)
	e.exact("BETA", x.Beta)
	e.exact("NOSE", x.Nose)
	e.exact("H", x.H)
	e.exact("H_FACT", x.HFact)
	e.exact("ANGLE_TS", x.AngleTS)
	e.exact("ANGLE_OS", x.AngleOS)
	e.exact("DEPTH_TS", x.DepthTS)
	e.exact("DEPTH_OS", x.DepthOS)
	e.exact("CHAMFER_WIDTH_TS", x.ChamferWidthTS)
	e.exact("CHAMFER_WIDTH_OS", x.ChamferWidthOS)
	e.exact("CHAMFER_HEIGHT_TS", x.ChamferHeightTS)
	e.exact("CHAMFER_HEIGHT_OS", x.ChamferHeightOS)
	e.exact("CONNECTION_ANGLE", x.ConnectionAngle)
	e.exact("WEB_SEG", x.WebSeg)
	e.exact("BEVEL_DEFINED_FLANGE", x.BevelDefinedFlange)
	e.exact("BEVEL_CODE_FLANGE", x.BevelCodeFlange)
	e.exact("BEVEL_TYPE_FLANGE", x.BevelTypeFlange)
	e.exact("BEVEL_VARIANT_FLANGE", x.BevelVariantFlange)
	e.exact("E_FLANGE", x.EFlange)
	e.exact("GAP_FLANGE", x.GapFlange)
	e.exact("CHAMFER_FLANGE", x.ChamferFlange)
	e.exact("ALPHA_FLANGE", x.AlphaFlange)
	e.exact("BETA_FLANGE", x.BetaFlange)
	e.exact("NOSE_FLANGE", x.NoseFlange)
	e.exact("H_FLANGE", x.HFlange)
	e.exact("H_FACT_FLANGE", x.HFactFlange)
	e.exact("ANGLE_TS_FLANGE", x.AngleTSFlange)
	e.exact("ANGLE_OS_FLANGE", x.AngleOSFlange)
	e.exact("DEPTH_TS_FLANGE", x.DepthTSFlange)
	e.exact("DEPTH_OS_FLANGE", x.DepthOSFlange)
	e.exact("CHAMFER_WIDTH_TS_FLANGE", x.ChamferWidthTSFlange)
	e.exact("CHAMFER_WIDTH_OS_FLANGE", x.ChamferWidthOSFlange)
	e.exact("CHAMFER_HEIGHT_TS_FLANGE", x.ChamferHeightTSFlange)
	e.exact("CHAMFER_HEIGHT_OS_FLANGE", x.ChamferHeightOSFlange)
	e.exact("CONNECTION_ANGLE_FLANGE", x.ConnectionAngleFlange)
	e.exact("FLA_SEG", x.FlaSeg)
	e.exact("BEVEL_DEFINED_FLANGE2", x.BevelDefinedFlange2)
	e.exact("BEVEL_CODE_FLANGE2", x.BevelCodeFlange2)
	e.exact("BEVEL_TYPE_FLANGE2", x.BevelTypeFlange2)
	e.exact("BEVEL_VARIANT_FLANGE2", x.BevelVariantFlange2)
	e.exact("E_FLANGE2", x.EFlange2)
	e.exact("GAP_FLANGE2", x.GapFlange2)
	e.exact("CHAMFER_FLANGE2", x.ChamferFlange2)
	e.exact("ALPHA_FLANGE2", x.AlphaFlange2)
	e.exact("BETA_FLANGE2", x.BetaFlange2)
	e.exact("NOSE_FLANGE2", x.NoseFlange2)
	e.exact("H_FLANGE2", x.HFlange2)
	e.exact("H_FACT_FLANGE2", x.HFactFlange2)
	e.exact("ANGLE_TS_FLANGE2", x.AngleTSFlange2)
	e.exact("ANGLE_OS_FLANGE2", x.AngleOSFlange2)
	e.exact("DEPTH_TS_FLANGE2", x.DepthTSFlange2)
	e.exact("DEPTH_OS_FLANGE2", x.DepthOSFlange2)
	e.exact("CHAMFER_WIDTH_TS_FLANGE2", x.ChamferWidthTSFlange2)
	e.exact("CHAMFER_WIDTH_OS_FLANGE2", x.ChamferWidthOSFlange2)
	e.exact("CHAMFER_HEIGHT_TS_FLANGE2", x.ChamferHeightTSFlange2)
	e.exact("CHAMFER_HEIGHT_OS_FLANGE2", x.ChamferHeightOSFlange2)
	e.exact("FLA2_SEG", x.Fla2Seg)
	e.exact("GSD", x.Gsd)
	e.exact("GSD_DIST", x.GsdDist)

	e.writeContour(x.Contour)
	e.writeFlangeContour(x.FlangeContour)

	e.marker("END_OF_" + marker)
}

func (e *encoder) writeFlangeContour(c *Contour) {
	if c == nil {
		return
	}

	e.marker("START_OF_FCONTOUR")

	e.integer("NO_OF_SEG", len(c.Segments))

	if len(c.Segments) > 0 && c.Segments[0].Start != nil {
		e.exact("START_U", c.Segments[0].Start.X)
		e.exact("START_V", c.Segments[0].Start.Y)
	} else {
		e.exact("START_U", 0)
		e.exact("START_V", 0)
	}

	for _, seg := range c.Segments {
		e.exact("AMP_U", seg.AmpU)
		e.exact("AMP_V", seg.AmpV)
		if seg.End != nil {
			e.exact("U", seg.End.X)
			e.exact("V", seg.End.Y)
		} else {
			e.exact("U", 0)
			e.exact("V", 0)
		}
	}

	e.marker("END_OF_FCONTOUR")
}
