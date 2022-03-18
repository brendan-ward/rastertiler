package gdal

// #cgo LDFLAGS: -lgdal
// #include "gdal.h"
// #include "gdalwarper.h"
// #include "ogr_srs_api.h"
import "C"
import (
	"fmt"
	"math"
	"unsafe"

	"github.com/brendan-ward/rastertiler/affine"
	"github.com/brendan-ward/rastertiler/tiles"
)

const Version string = C.GDAL_RELEASE_NAME

const RESAMPLING_NEAREST int = 0

// mapping of GDAL
var gdalDtypeStr = map[int]string{
	C.GDT_Byte:   "uint8",
	C.GDT_UInt16: "uint16",
	C.GDT_Int16:  "int16",
	C.GDT_UInt32: "uint32",
	C.GDT_Int32:  "int32",
}

var gdalDtype = map[string]int{
	"byte":   C.GDT_Byte,
	"uint8":  C.GDT_Byte,
	"uint16": C.GDT_UInt16,
	"uint32": C.GDT_UInt32,
	"int8":   C.GDT_Byte, // Note: requires setting an option when creating dataset
	"int16":  C.GDT_Int16,
	"int32":  C.GDT_Int32,
}

func init() {
	C.GDALAllRegister()
}

type Dataset struct {
	path      string
	ptr       C.GDALDatasetH
	driver    string
	dtype     string
	crs       string
	transform *affine.Affine
	width     int
	height    int
	nodata    interface{} // value is in dtype
	bounds    [4]float64
}

func newDataset(filename string, ptr C.GDALDatasetH) (*Dataset, error) {
	// assume 1-band data
	band := C.GDALGetRasterBand(ptr, 1)
	if unsafe.Pointer(band) == nil {
		return nil, fmt.Errorf("could not get raster band")
	}

	driver := C.GoString(C.GDALGetDriverShortName(C.GDALGetDatasetDriver(ptr)))
	crs := C.GoString(C.GDALGetProjectionRef(ptr))
	dtype := gdalDtypeStr[int(C.GDALGetRasterDataType(C.GDALGetRasterBand(ptr, 1)))]
	width := int(C.GDALGetRasterXSize(ptr))
	height := int(C.GDALGetRasterYSize(ptr))

	var rawTransform [6]float64
	if (C.GDALGetGeoTransform(ptr, (*C.double)(unsafe.Pointer(&rawTransform[0])))) != C.CE_None {
		return nil, fmt.Errorf("could not read transform")
	}
	transform := affine.FromGDAL(rawTransform)

	// var hasNodata int
	// rawNodata := int(C.GDALGetRasterNoDataValue(band, (*C.int)(unsafe.Pointer(&hasNodata))))
	rawNodata := int(C.GDALGetRasterNoDataValue(band, nil))
	var nodata interface{}

	switch dtype {
	case "int8":
		nodata = int8(rawNodata)
	case "uint8":
		nodata = uint8(rawNodata)
	case "int16":
		nodata = int16(rawNodata)
	case "uint16":
		nodata = uint16(rawNodata)
	case "int32":
		nodata = int32(rawNodata)
	case "uint32":
		nodata = uint32(rawNodata)
	default:
		panic("Nodata() not yet supported for other dtypes")
	}

	if transform.D > 0 {
		panic("rasters anchored from bottom left not yet supported")
	}

	// raster is anchored in upper left; this is the standard direction
	var bounds [4]float64
	bounds[0] = transform.C
	bounds[1] = transform.F + transform.E*float64(height)
	bounds[2] = transform.C + transform.A*float64(width)
	bounds[3] = transform.F

	return &Dataset{
		path:      filename,
		ptr:       ptr,
		driver:    driver,
		crs:       crs,
		transform: transform,
		width:     width,
		height:    height,
		dtype:     dtype,
		nodata:    nodata,
		bounds:    bounds,
	}, nil
}

func Open(filename string) (*Dataset, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	ptr := C.GDALOpen(cFilename, C.GA_ReadOnly)
	if ptr == nil {
		return nil, fmt.Errorf("could not open dataset: %v", filename)
	}

	return newDataset(filename, ptr)
}

func (d *Dataset) Close() {
	if d != nil && unsafe.Pointer(d.ptr) != nil {
		C.GDALClose(d.ptr)
	}
	// clear out previous references
	*d = Dataset{}
}

func (d *Dataset) mustBeOpen() {
	if d == nil {
		panic("dataset not initialized")
	}
}

// Get bounds of dataset: [xmin, ymin, xmax, ymax]
func (d *Dataset) Bounds() [4]float64 {
	d.mustBeOpen()

	return d.bounds
}

// Get geographic bounds of dataset: [xmin, ymin, xmax, ymax]
func (d *Dataset) GeoBounds() ([4]float64, error) {
	return d.transformBounds("EPSG:4326")
}

// Get Mercator bounds of dataset: [xmin, ymin, xmax, ymax]
func (d *Dataset) MercatorBounds() ([4]float64, error) {
	return d.transformBounds("EPSG:3857")
}

// Project dataset bounds to CRS
func (d *Dataset) transformBounds(crs string) ([4]float64, error) {
	d.mustBeOpen()

	var bounds [4]float64

	srcSRS := C.GDALGetSpatialRef(d.ptr)

	targetSRSName := C.CString(crs)
	defer C.free(unsafe.Pointer(targetSRSName))
	targetSRS := C.OSRNewSpatialReference(nil)
	defer C.OSRDestroySpatialReference(targetSRS)
	C.OSRSetFromUserInput(targetSRS, targetSRSName)
	if unsafe.Pointer(targetSRS) == nil {
		return bounds, fmt.Errorf("could not set target SRS to WGS84")
	}
	// make sure that coords are always returned in long/lat order (otherwise EPSG:4326 returns in opposite order)
	C.OSRSetAxisMappingStrategy(targetSRS, C.OAMS_TRADITIONAL_GIS_ORDER)

	transform := C.OCTNewCoordinateTransformation(srcSRS, targetSRS)
	if unsafe.Pointer(transform) == nil {
		return bounds, fmt.Errorf("could not create coordinate transform")
	}
	defer C.OCTDestroyCoordinateTransformation(transform)

	if C.OCTTransformBounds(
		transform,
		C.double(d.bounds[0]),
		C.double(d.bounds[1]),
		C.double(d.bounds[2]),
		C.double(d.bounds[3]),
		(*C.double)(unsafe.Pointer(&bounds[0])),
		(*C.double)(unsafe.Pointer(&bounds[1])),
		(*C.double)(unsafe.Pointer(&bounds[2])),
		(*C.double)(unsafe.Pointer(&bounds[3])),
		21,
	) == 0 {
		return bounds, fmt.Errorf("error transforming bounds to WGS84 coordinates")
	}

	return bounds, nil
}

// Get the height of the dataset, in number of pixels
func (d *Dataset) Height() int {
	d.mustBeOpen()

	return d.height
}

// Get the width of the dataset, in number of pixels
func (d *Dataset) Width() int {
	d.mustBeOpen()

	return d.width
}

// Return an Affine tranform object
func (d *Dataset) Transform() *affine.Affine {
	d.mustBeOpen()

	return d.transform
}

func (d *Dataset) Nodata() interface{} {
	d.mustBeOpen()

	return d.nodata
}

func (d *Dataset) CRS() string {
	d.mustBeOpen()

	return d.crs
}

func (d *Dataset) DType() string {
	return d.dtype
}

func (d *Dataset) Window(bounds [4]float64) *Window {
	d.mustBeOpen()

	return WindowFromBounds(d.transform, bounds)
}

func (d *Dataset) WindowTransform(window *Window) *affine.Affine {
	d.mustBeOpen()

	return WindowTransform(window, d.transform)
}

func (d *Dataset) String() string {
	if d == nil {
		return ""
	}

	geoBounds, _ := d.GeoBounds()
	return fmt.Sprintf("%v (%v: %v, nodata: %v)\ndimensions: %v x %v pixels\ntransform:\n%v\nbounds: %v\ngeographic bounds: %v", d.path, d.driver, d.dtype, d.nodata, d.Width(), d.Height(), d.transform, d.bounds, geoBounds)
}

func (d *Dataset) GetWarpedVRT(crs string) (*Dataset, error) {
	d.mustBeOpen()

	targetSRSName := C.CString(crs)
	defer C.free(unsafe.Pointer(targetSRSName))

	ptr := C.GDALAutoCreateWarpedVRT(
		d.ptr,
		C.GDALGetProjectionRef(d.ptr),
		targetSRSName,
		C.GDALResampleAlg(RESAMPLING_NEAREST),
		0,
		nil,
	)

	if unsafe.Pointer(ptr) == nil {
		return nil, fmt.Errorf("could not create WarpedVRT")
	}

	return newDataset(fmt.Sprintf("WarpedVRT (src: %v)", d.path), ptr)
}

func (d *Dataset) Read(offsetX int, offsetY int, width int, height int, bufferWidth int, bufferHeight int) (*Array, error) {
	d.mustBeOpen()

	gdalDataType := C.GDALGetRasterDataType(C.GDALGetRasterBand(d.ptr, 1))
	dtype := gdalDtypeStr[int(gdalDataType)]

	var array *Array
	var bufferPtr unsafe.Pointer
	size := bufferWidth * bufferHeight

	switch dtype {
	// TODO: other data types
	case "uint8":
		uint8Buffer := make([]uint8, size)
		array = &Array{
			DType:  dtype,
			Width:  bufferWidth,
			Height: bufferHeight,
			buffer: uint8Buffer,
		}
		bufferPtr = unsafe.Pointer(&uint8Buffer[0])
	default:
		panic("Other dtypes not yet supported for reading")
	}

	if C.GDALDatasetRasterIO(
		d.ptr,
		C.GF_Read,
		C.int(offsetX),
		C.int(offsetY),
		C.int(width),
		C.int(height),
		bufferPtr,
		C.int(bufferWidth),
		C.int(bufferHeight),
		gdalDataType,
		C.int(1), // number of bands being written
		nil,      // default to selecting first band for writing
		0,        // pixel spacing (same as underlying data type)
		0,        // line spacing (default)
		0,        // band spacing (default)
	) != C.CE_None {
		return nil, fmt.Errorf("could not read data")
	}

	return array, nil
}

// Read a tile of data from a Mercator-projection VRT or dataset
func (d *Dataset) ReadTile(tileID *tiles.TileID, tileSize int) (*Array, *affine.Affine, error) {
	size := float64(tileSize)
	vrtWidth := float64(d.width)
	vrtHeight := float64(d.height)

	tileBounds := tileID.MercatorBounds()
	window := d.Window(tileBounds)
	tileTransform := d.WindowTransform(window)

	// scale transform for tile
	tileTransform = tileTransform.Scale(window.Width/size, window.Height/size)

	xres, yres := tileTransform.Resolution()
	leftOffset := math.Max(math.Round((d.bounds[0]-tileBounds[0])/xres), 0)
	rightOffset := math.Max(math.Round((tileBounds[2]-d.bounds[2])/xres), 0)
	bottomOffset := math.Max(math.Round((d.bounds[1]-tileBounds[1])/yres), 0)
	topOffset := math.Max(math.Round((tileBounds[3]-d.bounds[3])/yres), 0)
	width := int(size - leftOffset - rightOffset)
	height := int(size - topOffset - bottomOffset)

	// crop the window to the available pixels and convert to integer values
	xStart := math.Round(math.Min(math.Max(window.XOffset, 0), vrtWidth))
	yStart := math.Round(math.Min(math.Max(window.YOffset, 0), vrtHeight))
	xStop := math.Max(math.Min(window.XOffset+window.Width, vrtWidth), 0)
	yStop := math.Max(math.Min(window.YOffset+window.Height, vrtHeight), 0)
	readWidth := int(math.Floor((xStop - xStart) + 0.5))
	readHeight := int(math.Floor((yStop - yStart) + 0.5))

	// fmt.Println(window)
	// fmt.Println(tileTransform)
	// fmt.Printf("xoff: %v, yoff: %v, width: %v, height:%v\n", xStart, yStart, readWidth, readHeight)
	// fmt.Printf("paste offsets: %v, %v\n", int(topOffset), int(leftOffset))

	if readWidth <= 0 || readHeight <= 0 {
		// no tile available
		return nil, nil, nil
	}

	data, err := d.Read(int(xStart), int(yStart), readWidth, readHeight, width, height)
	if err != nil {
		return nil, nil, err
	}

	if data.EqualsValue(d.nodata) {
		// empty tile
		return nil, nil, nil
	}

	if width != tileSize || height != tileSize {
		out := NewArray(tileSize, tileSize, data.DType, d.nodata)
		err := out.Paste(data, int(topOffset), int(leftOffset))
		if err != nil {
			return nil, nil, err
		}
		data = out
	}

	return data, tileTransform, nil
}

func WriteGeoTIFF(filename string, data *Array, transform *affine.Affine, crs string, nodata interface{}) error {

	isSignedByte := false
	// use type assertion switch to get data as indexable type
	var bufferPtr unsafe.Pointer
	switch bufferType := data.buffer.(type) {
	case []int8:
		isSignedByte = true
		bufferPtr = unsafe.Pointer(&bufferType[0])
	case []uint8:
		bufferPtr = unsafe.Pointer(&bufferType[0])
	case []int16:
		bufferPtr = unsafe.Pointer(&bufferType[0])
	case []uint16:
		bufferPtr = unsafe.Pointer(&bufferType[0])
	case []int32:
		bufferPtr = unsafe.Pointer(&bufferType[0])
	case []uint32:
		bufferPtr = unsafe.Pointer(&bufferType[0])
	}

	driverName := C.CString("GTiff")
	defer C.free(unsafe.Pointer(driverName))

	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	dataType := C.GDALDataType(gdalDtype[data.DType])

	// set sensible default options
	options := []string{
		"TILED=YES",
		"BLOCKXSIZE=256",
		"BLOCKYSIZE=256",
		"COMPRESS=lzw",
	}
	if isSignedByte {
		options = append(options, "PIXELTYPE=SIGNEDBYTE")
	}

	// create a null-terminated C string array
	length := len(options)
	gdalOpts := make([]*C.char, length+1)
	for i := 0; i < len(options); i++ {
		gdalOpts[i] = C.CString(options[i])
		defer C.free(unsafe.Pointer(gdalOpts[i]))
	}
	gdalOpts[length] = (*C.char)(unsafe.Pointer(nil))

	ptr := C.GDALCreate(
		C.GDALGetDriverByName(driverName),
		cFilename,
		C.int(data.Width),
		C.int(data.Height),
		1, // number of bands
		dataType,
		(**C.char)(unsafe.Pointer(&gdalOpts[0])),
	)
	if unsafe.Pointer(ptr) == nil {
		return fmt.Errorf("could not open dataset for writing: %v", filename)
	}

	band := C.GDALGetRasterBand(ptr, 1)
	if unsafe.Pointer(band) == nil {
		return fmt.Errorf("could not get raster band")
	}

	outCRS := C.CString(crs)
	defer C.free(unsafe.Pointer(outCRS))
	if C.GDALSetProjection(ptr, outCRS) != C.CE_None {
		return fmt.Errorf("could not set CRS")
	}

	gdalTransform := transform.ToGDAL()
	if C.GDALSetGeoTransform(
		ptr,
		(*C.double)(unsafe.Pointer(&gdalTransform[0])),
	) != C.CE_None {
		return fmt.Errorf("could not set transform")
	}

	var cNodata C.double
	switch typedNodata := nodata.(type) {
	case int8:
		cNodata = C.double(typedNodata)
	case uint8:
		cNodata = C.double(typedNodata)
	case int16:
		cNodata = C.double(typedNodata)
	case uint16:
		cNodata = C.double(typedNodata)
	case int32:
		cNodata = C.double(typedNodata)
	case uint32:
		cNodata = C.double(typedNodata)
	}

	if C.GDALSetRasterNoDataValue(band, cNodata) != C.CE_None {
		return fmt.Errorf("could not set NODATA")
	}

	// write data to band
	if C.GDALDatasetRasterIO(
		ptr,
		C.GF_Write,
		C.int(0),
		C.int(0),
		C.int(data.Width),
		C.int(data.Height),
		bufferPtr,
		C.int(data.Width),
		C.int(data.Height),
		dataType,
		C.int(1), // number of bands being written
		nil,      // default to selecting first band for writing
		0,        // pixel spacing
		0,        // line spacing
		0,        // band spacing
	) != C.CE_None {
		return fmt.Errorf("could not write data")
	}

	C.GDALClose(ptr)

	return nil
}
