package gdal

// #cgo LDFLAGS: -lgdal
// #include "gdal.h"
// #include "gdalwarper.h"
// #include "ogr_srs_api.h"
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/brendan-ward/rastertiler/affine"
)

var DTypeStr = map[int]string{
	C.GDT_Byte:   "uint8",
	C.GDT_UInt16: "uint16",
	C.GDT_Int16:  "int16",
	C.GDT_UInt32: "uint32",
	C.GDT_Int32:  "int32",
}

var GDALDType = map[string]int{
	"byte":   C.GDT_Byte,
	"uint8":  C.GDT_Byte,
	"uint16": C.GDT_UInt16,
	"uint32": C.GDT_UInt32,
	"int8":   C.GDT_Byte, // Note: requires setting an option when creating dataset
	"int16":  C.GDT_Int16,
	"int32":  C.GDT_Int32,
}

type Dataset struct {
	path string
	ptr  C.GDALDatasetH
}

const RESAMPLING_NEAREST int = 0

const Version string = C.GDAL_RELEASE_NAME

func init() {
	C.GDALAllRegister()
}

func Open(filename string) (*Dataset, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	ptr := C.GDALOpen(cFilename, C.GA_ReadOnly)
	if ptr == nil {
		return nil, fmt.Errorf("could not open dataset: %v", filename)

	}
	return &Dataset{
		path: filename,
		ptr:  ptr,
	}, nil
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
func (d *Dataset) Bounds() ([4]float64, error) {
	d.mustBeOpen()

	var bounds [4]float64

	transform, err := d.Transform()
	if err != nil {
		return bounds, err
	}

	if transform.D > 0 {
		panic("rasters anchored from bottom left not yet supported")
	}

	// raster is anchored in upper left; this is the standard direction

	bounds[0] = transform.C
	bounds[1] = transform.F + transform.E*float64(d.Height())
	bounds[2] = transform.C + transform.A*float64(d.Width())
	bounds[3] = transform.F

	return bounds, nil
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

	b, err := d.Bounds()
	if err != nil {
		return bounds, err
	}

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
		C.double(b[0]),
		C.double(b[1]),
		C.double(b[2]),
		C.double(b[3]),
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
	return int(C.GDALGetRasterYSize(d.ptr))
}

// Get the width of the dataset, in number of pixels
func (d *Dataset) Width() int {
	d.mustBeOpen()
	return int(C.GDALGetRasterXSize(d.ptr))
}

// Return an Affine tranform object
func (d *Dataset) Transform() (*affine.Affine, error) {
	d.mustBeOpen()

	var transform [6]float64

	if d == nil {
		return nil, nil
	}
	if (C.GDALGetGeoTransform(d.ptr, (*C.double)(unsafe.Pointer(&transform[0])))) != C.CE_None {
		return nil, fmt.Errorf("could not get transform for: %v", d.path)
	}
	return affine.FromGDAL(transform), nil
}

// Get nodata value (cast to int) for first band, boolean to indicate if a nodata value is set
func (d *Dataset) Nodata() (int, bool, error) {
	d.mustBeOpen()

	// assume 1-band data
	band := C.GDALGetRasterBand(d.ptr, 1)
	if unsafe.Pointer(band) == nil {
		return 0, false, fmt.Errorf("could not get raster band")
	}

	var hasNodata int
	nodata := int(C.GDALGetRasterNoDataValue(band, (*C.int)(unsafe.Pointer(&hasNodata))))

	return nodata, hasNodata != 0, nil
}

func (d *Dataset) CRS() string {
	d.mustBeOpen()

	return C.GoString(C.GDALGetProjectionRef(d.ptr))
}

func (d *Dataset) DType() string {
	d.mustBeOpen()

	return DTypeStr[int(C.GDALGetRasterDataType(C.GDALGetRasterBand(d.ptr, 1)))]
}

func (d *Dataset) Window(bounds [4]float64) (*Window, error) {
	d.mustBeOpen()

	transform, err := d.Transform()
	if err != nil {
		return nil, err
	}

	return WindowFromBounds(transform, bounds), nil
}

func (d *Dataset) WindowTransform(window *Window) (*affine.Affine, error) {
	d.mustBeOpen()

	transform, err := d.Transform()
	if err != nil {
		return nil, err
	}

	return WindowTransform(window, transform), nil
}

func (d *Dataset) String() string {
	if d == nil {
		return ""
	}

	driver := C.GoString(C.GDALGetDriverShortName(C.GDALGetDatasetDriver(d.ptr)))
	dataType := d.DType()
	nodata, hasNodata, _ := d.Nodata()
	transform, _ := d.Transform()
	bounds, _ := d.Bounds()
	geoBounds, _ := d.GeoBounds()
	return fmt.Sprintf("%v (%v: %v, nodata: %v [set: %v])\ndimensions: %v x %v pixels\ntransform:\n%v\nbounds: %v\ngeographic bounds: %v", d.path, driver, dataType, nodata, hasNodata, d.Width(), d.Height(), transform, bounds, geoBounds)
}

func (d *Dataset) GetWarpedVRT(crs string) (*Dataset, error) {
	d.mustBeOpen()

	targetSRSName := C.CString(crs)
	defer C.free(unsafe.Pointer(targetSRSName))

	vrt := C.GDALAutoCreateWarpedVRT(
		d.ptr,
		C.GDALGetProjectionRef(d.ptr),
		targetSRSName,
		C.GDALResampleAlg(RESAMPLING_NEAREST),
		0,
		nil,
	)

	if unsafe.Pointer(vrt) == nil {
		return nil, fmt.Errorf("could not create WarpedVRT")
	}

	return &Dataset{
		path: fmt.Sprintf("WarpedVRT (src: %v)", d.path),
		ptr:  vrt,
	}, nil
}

func (d *Dataset) Read(offsetX int, offsetY int, width int, height int, bufferWidth int, bufferHeight int) (*Array, error) {
	d.mustBeOpen()

	gdalDataType := C.GDALGetRasterDataType(C.GDALGetRasterBand(d.ptr, 1))
	dtype := DTypeStr[int(gdalDataType)]

	var array *Array
	// var buffer interface{}
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
		C.int(bufferWidth), // TODO: can this be same as width and height when reading boundless?
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

func WriteGeoTIFF(filename string, data interface{}, width int, height int, transform *affine.Affine, crs string, dtype string, nodata int) error {

	isSignedByte := false
	// use type assertion switch to get data as indexable type
	var bufferPtr unsafe.Pointer
	switch bufferType := data.(type) {
	case []int8:
		isSignedByte = true
		bufferPtr = unsafe.Pointer(&bufferType[0])
	case []uint8:
		bufferPtr = unsafe.Pointer(&bufferType[0])
	case []uint16:
		bufferPtr = unsafe.Pointer(&bufferType[0])
	case []uint32:
		bufferPtr = unsafe.Pointer(&bufferType[0])
	case []int16:
		bufferPtr = unsafe.Pointer(&bufferType[0])
	case []int32:
		bufferPtr = unsafe.Pointer(&bufferType[0])
	}

	driverName := C.CString("GTiff")
	defer C.free(unsafe.Pointer(driverName))

	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	dataType := C.GDALDataType(GDALDType[dtype])

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
		C.int(width),
		C.int(height),
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

	if C.GDALSetRasterNoDataValue(band, C.double(nodata)) != C.CE_None {
		return fmt.Errorf("could not set NODATA")
	}

	// write data to band
	if C.GDALDatasetRasterIO(
		ptr,
		C.GF_Write,
		C.int(0),
		C.int(0),
		C.int(width),
		C.int(height),
		bufferPtr,
		C.int(width),
		C.int(height),
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
