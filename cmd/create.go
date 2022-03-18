package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/brendan-ward/rastertiler/encoding"
	"github.com/brendan-ward/rastertiler/gdal"
	"github.com/brendan-ward/rastertiler/mbtiles"
	"github.com/brendan-ward/rastertiler/tiles"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
)

var minzoom uint8
var maxzoom uint8
var tilesetName string
var description string
var attribution string
var numWorkers int
var tileSize int
var colormapStr string

var createCmd = &cobra.Command{
	Use:   "create [IN.tiff] [OUT.mbtiles]",
	Short: "Create an MBTiles tileset from a single-band GeoTIFF",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("GeoTIFF and mbtiles filenames are required")
		}
		if _, err := os.Stat(args[0]); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("input file '%s' does not exist", args[0])
		}
		outDir, _ := path.Split(args[1])
		if outDir != "" {
			if _, err := os.Stat(outDir); errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("output directory '%s' does not exist", outDir)
			}
		}
		if path.Ext(args[1]) != ".mbtiles" {
			return errors.New("mbtiles filename must end in '.mbtiles'")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// validate flags
		if numWorkers < 1 {
			numWorkers = 1
		}
		if maxzoom < minzoom {
			return errors.New("maxzoom must be no smaller than minzoom")
		}
		if maxzoom > 25 {
			return errors.New("maxzoom must be no greater than 24")
		}

		return create(args[0], args[1])
	},
	SilenceUsage: true,
}

func init() {
	createCmd.Flags().Uint8VarP(&minzoom, "minzoom", "Z", 0, "minimum zoom level")
	createCmd.Flags().Uint8VarP(&maxzoom, "maxzoom", "z", 0, "maximum zoom level")
	createCmd.Flags().IntVarP(&tileSize, "tilesize", "s", 256, "tile size in pixels")
	createCmd.Flags().StringVarP(&tilesetName, "name", "n", "", "tileset name")
	createCmd.Flags().StringVarP(&description, "description", "d", "", "tileset description")
	createCmd.Flags().StringVarP(&attribution, "attribution", "a", "", "tileset description")
	createCmd.Flags().IntVarP(&numWorkers, "workers", "w", 4, "number of workers to create tiles")
	createCmd.Flags().StringVarP(&colormapStr, "colormap", "c", "", "colormap '<value>:<hex>,<value>:<hex>'.  Only valid for 8-bit data")
}

func produce(minZoom uint8, maxZoom uint8, bounds [4]float64, queue chan<- *tiles.TileID) {
	defer close(queue)

	fmt.Println("Creating tiles")

	uiprogress.Start()

	for zoom := minZoom; zoom <= maxZoom; zoom++ {
		z := zoom
		minTile, maxTile := tiles.TileRange(zoom, bounds)
		count := ((maxTile.X - minTile.X) * (maxTile.Y - minTile.Y)) + 1
		bar := uiprogress.AddBar(int(count)).AppendCompleted().PrependElapsed()
		bar.PrependFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("zoom %2v (%8v/%8v)", z, b.Current(), count)
		})

		for x := minTile.X; x <= maxTile.X; x++ {
			for y := minTile.Y; y <= maxTile.Y; y++ {
				queue <- &tiles.TileID{Zoom: zoom, X: x, Y: y}
				bar.Incr()
			}
		}
	}
	uiprogress.Stop()
}

func create(infilename string, outfilename string) error {
	// default to input filename, without extension
	if tilesetName == "" {
		tilesetName = strings.TrimSuffix(path.Base(infilename), filepath.Ext(infilename))
	}

	d, err := gdal.Open(infilename)
	if err != nil {
		return err
	}
	defer d.Close()

	var encoder encoding.PNGEncoder

	switch d.DType() {
	case "uint8":
		if colormapStr != "" {
			encoder, err = encoding.NewColormapEncoder(colormapStr)
			if err != nil {
				return err
			}
		} else {
			encoder = encoding.NewGrayscaleEncoder()
		}

	default:
		panic("encoding not yet supported for other dtypes")
	}

	db, err := mbtiles.NewMBtilesWriter(outfilename, numWorkers)
	if err != nil {
		return err
	}
	defer db.Close()

	geoBounds, err := d.GeoBounds()
	if err != nil {
		return err
	}

	mercatorBounds, err := d.MercatorBounds()
	if err != nil {
		return err
	}

	d.Close()

	db.WriteMetadata(tilesetName, description, attribution, minzoom, maxzoom, geoBounds)

	queue := make(chan *tiles.TileID)
	var wg sync.WaitGroup

	go produce(minzoom, maxzoom, mercatorBounds, queue)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			con, err := db.GetConnection()
			if err != nil {
				panic(err)
			}
			defer db.CloseConnection(con)

			// get VRT once per goroutine
			ds, err := gdal.Open(infilename)
			defer ds.Close()

			vrt, err := ds.GetWarpedVRT("EPSG:3857")
			if err != nil {
				panic(err)
			}
			defer vrt.Close()

			for tileID := range queue {
				data, _, err := vrt.ReadTile(tileID, tileSize)
				if err != nil {
					panic(err)
				}

				if data != nil && !data.EqualsValue(vrt.Nodata()) {
					buffer, width, height, bits, err := data.Uint8Buffer()
					if err != nil {
						panic(err)
					}

					var png []byte
					png, err = encoder.Encode(buffer, width, height, bits)

					// FIXME: debug only
					// err = os.WriteFile(fmt.Sprintf("/tmp/png/%v_%v_%v.png", tileID.Zoom, tileID.X, tileID.Y), png, 0644)
					// if err != nil {
					// 	panic(err)
					// }

					mbtiles.WriteTile(con, tileID, png)

					// FIXME: debug only
					// gdal.WriteGeoTIFF(fmt.Sprintf("/tmp/tiffs/%v_%v_%v.tif", tileID.Zoom, tileID.X, tileID.Y), data, tileTransform, vrt.CRS(), vrt.Nodata())
				}
			}

		}()
	}

	wg.Wait()

	return nil
}

// TODO: remove
func Create(infilename string, outfilename string) error {
	return create(infilename, outfilename)
}
