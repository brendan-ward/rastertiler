package gdal

// #cgo LDFLAGS: -lgdal
// #include "gdal.h"
// #include "ogr_srs_api.h"
import "C"
import (
	"fmt"
	"unsafe"
)

type Dataset struct {
	path string
	ptr  C.GDALDatasetH
}

// type SpatialReference C.OGRSpatialReferenceH

const Version string = C.GDAL_RELEASE_NAME

func init() {
	C.GDALAllRegister()
}

func Open(filename string) (*Dataset, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	dataset := C.GDALOpen(cFilename, C.GA_ReadOnly)
	if dataset == nil {
		return nil, fmt.Errorf("could not open dataset: %v", filename)

	}
	return &Dataset{
		path: filename,
		ptr:  dataset,
	}, nil
}

func (d *Dataset) Close() {
	if d != nil {
		C.GDALClose(d.ptr)
	}
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

	// raster is anchored in upper left; this is the standard direction
	if transform[5] < 0 {
		bounds[0] = transform[0]
		bounds[1] = transform[3] + transform[5]*float64(d.Height())
		bounds[2] = transform[0] + transform[1]*float64(d.Width())
		bounds[3] = transform[3]
	} else {
		panic("rasters anchored from bottom left not yet supported")
	}

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

// Get 6-part geo transform array of dataset
func (d *Dataset) Transform() ([6]float64, error) {
	d.mustBeOpen()

	var transform [6]float64

	if d == nil {
		return transform, nil
	}
	if (C.GDALGetGeoTransform(d.ptr, (*C.double)(unsafe.Pointer(&transform[0])))) != C.CE_None {
		return transform, fmt.Errorf("could not get transform for: %v", d.path)
	}
	return transform, nil
}

func (d *Dataset) String() string {
	if d == nil {
		return ""
	}

	driver := C.GoString(C.GDALGetDriverShortName(C.GDALGetDatasetDriver(d.ptr)))
	// crs := C.GoString(C.GDALGetProjectionRef(d.ptr))
	transform, _ := d.Transform()
	bounds, _ := d.Bounds()
	geoBounds, _ := d.GeoBounds()
	return fmt.Sprintf("%v (%v): %v x %v pixels\ntransform: %v\nbounds: %v\ngeographic bounds: %v", d.path, driver, d.Width(), d.Height(), transform, bounds, geoBounds)
}
