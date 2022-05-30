# rastertiler

A Go-based single-band GeoTIFF to PNG mbtiles creator.

Requires GDAL >= 3.4 to be installed on the system.

**WARNING** this project has been superseded by [rastertiler-rs](https://github.com/brendan-ward/rastertiler-rs), a better, faster implementation in Rust. This project will no longer be actively developed. See below for more information.

## Installation

`go get https://github.com/brendan-ward/rastertiler`

## Usage

### Create MBTiles from GeoTIFF

```bash
Create an MBTiles tileset from a single-band GeoTIFF

Usage:
  rastertiler create [IN.tiff] [OUT.mbtiles] [flags]

Flags:
  -a, --attribution string   tileset description
  -c, --colormap string      colormap '<value>:<hex>,<value>:<hex>'.  Only valid for 8-bit data
  -d, --description string   tileset description
  -h, --help                 help for create
  -z, --maxzoom uint8        maximum zoom level
  -Z, --minzoom uint8        minimum zoom level
  -n, --name string          tileset name
  -s, --tilesize int         tile size in pixels (default 256)
  -w, --workers int          number of workers to create tiles (default 4)
```

To create MBtiles from a single-band `uint8` GeoTIFF:

```bash
rastertiler create example.tif example.mbtiles --minzoom 0 --maxzoom 2
```

By default, this will render grayscale PNG tiles.

To use a colormap to render the `uint8` data to paletted PNG

```bash
rastertiler create example.tif example.mbtiles --minzoom 0 --maxzoom 2 --colormap "1:#686868,2:#fbb4b9,3:#c51b8a,4:#49006a"
```

## Porting to Rust

This project has been superseded by a port into Rust: https://github.com/brendan-ward/rastertiler-rs

Because this project uses GDAL for reading and warping data from a GeoTIFF to create
many tiles, the calls using CGO were in a hot loop and incurred more overhead
than is ideal. The Rust implementation does not incur this overhead and is
roughly 20% faster. The Rust PNG library also provided more flexibility for
writing RGB and paletted images.

## Credits

GDAL bindings inspired by:

-   [GDAL bindings for Go](https://github.com/lukeroth/gdal)
-   [rasterio](https://github.com/rasterio/rasterio)

See also [raster-tilecutter](https://github.com/brendan-ward/raster-tilecutter) which does much the same thing, in Python, using `rasterio`.

Parts of this project are adapted from [arrowtiler](https://github.com/brendan-ward/arrowtiler).

This project was developed with the support of the
[U.S. Fish and Wildlife Service](https://www.fws.gov/)
[Southeast Conservation Adaptation Strategy](https://secassoutheast.org/) for
use in the
[Southeast Conservation Blueprint Viewer](https://blueprint.geoplatform.gov/southeast/)
and
[South Atlantic Conservation Blueprint Simple Viewer](https://blueprint.geoplatform.gov/southatlantic/).
