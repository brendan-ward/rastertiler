package main

import (
	"github.com/brendan-ward/rastertiler/cmd"
)

func main() {
	// fmt.Printf("GDAL version: %v\n", gdal.GDALVersion)

	// d, err := gdal.Open("/tmp/blueprint2021.tif")
	// if err != nil {
	// 	panic(err)
	// }
	// defer d.Close()

	// fmt.Println(d)

	cmd.Create("/tmp/blueprint2021.tif", "/tmp/blueprint2021.mbtiles")

}
