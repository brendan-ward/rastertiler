package mbtiles

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/brendan-ward/rastertiler/affine"
	"github.com/brendan-ward/rastertiler/tiles"
)

var emptyContext context.Context

type MBtilesWriter struct {
	pool *sqlitex.Pool
}

// Create a tiles table structure that allows us to de-duplicate tile images
// shared by multiple tileIDs (e.g., blank tiles, ocean tiles)
const init_sql = `
CREATE TABLE IF NOT EXISTS metadata (name text, value text);
CREATE UNIQUE INDEX IF NOT EXISTS name ON metadata (name);

CREATE TABLE IF NOT EXISTS map (
	zoom_level INTEGER,
	tile_column INTEGER,
	tile_row INTEGER,
	tile_id TEXT
);
CREATE UNIQUE INDEX IF NOT EXISTS map_index ON map (zoom_level, tile_column, tile_row);

CREATE TABLE IF NOT EXISTS images (tile_data blob, tile_id text);
CREATE UNIQUE INDEX IF NOT EXISTS images_id ON images (tile_id);
CREATE VIEW IF NOT EXISTS tiles AS
	SELECT zoom_level, tile_column, tile_row, tile_data
	FROM map JOIN images ON images.tile_id = map.tile_id;
`

func NewMBtilesWriter(path string, poolsize int) (*MBtilesWriter, error) {
	ext := filepath.Ext(path)
	if ext != ".mbtiles" {
		return nil, fmt.Errorf("path must end in .mbtiles")
	}

	// always overwrite
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		os.Remove(path)
	}

	// check flags: this may not be safe for multiple goroutines (only one write  per connection though)
	pool, err := sqlitex.Open(path, sqlite.SQLITE_OPEN_CREATE|sqlite.SQLITE_OPEN_READWRITE|sqlite.SQLITE_OPEN_NOMUTEX|sqlite.SQLITE_OPEN_WAL, poolsize)
	if err != nil {
		return nil, err
	}

	db := &MBtilesWriter{
		pool: pool,
	}

	con, err := db.GetConnection()
	if err != nil {
		return nil, err
	}
	defer db.CloseConnection(con)

	// create tables
	err = sqlitex.ExecScript(con, init_sql)
	if err != nil {
		return nil, fmt.Errorf("could not initialize database: %q", err)
	}

	return db, nil
}

func (db *MBtilesWriter) Close() {
	if db.pool != nil {

		// make sure that anything pending is written
		con, err := db.GetConnection()
		if err != nil {
			panic(err)
		}
		// flush the WAL
		err = sqlitex.Exec(con, "PRAGMA wal_checkpoint;", nil)
		if err != nil {
			panic(err)
		}

		db.CloseConnection(con)

		db.pool.Close()
	}
}

// GetConnection gets a sqlite.Conn from an open connection pool.
// CloseConnection(con) must be called to release the connection.
func (db *MBtilesWriter) GetConnection() (*sqlite.Conn, error) {
	con := db.pool.Get(emptyContext)
	if con == nil {
		return nil, fmt.Errorf("connection could not be opened")
	}
	return con, nil
}

// CloseConnection closes an open sqlite.Conn and returns it to the pool.
func (db *MBtilesWriter) CloseConnection(con *sqlite.Conn) {
	if con != nil {
		db.pool.Put(con)
	}
}

func writeMetadataItem(con *sqlite.Conn, key string, value interface{}) error {
	return sqlitex.Exec(con, "INSERT INTO metadata (name,value) VALUES (?, ?)", nil, key, value)
}

func (db *MBtilesWriter) WriteMetadata(name string, description string, attribution string, minZoom uint8, maxZoom uint8, bounds *affine.Bounds) (err error) {
	if db == nil || db.pool == nil {
		return fmt.Errorf("cannot write to closed mbtiles database")
	}

	con, e := db.GetConnection()
	if e != nil {
		return e
	}
	defer db.CloseConnection(con)

	// create savepoint
	defer sqlitex.Save(con)(&err)

	if err = writeMetadataItem(con, "name", name); err != nil {
		return err
	}
	if description != "" {
		if err = writeMetadataItem(con, "description", description); err != nil {
			return err
		}
	}
	if attribution != "" {
		if err = writeMetadataItem(con, "attribution", attribution); err != nil {
			return err
		}
	}
	if err = writeMetadataItem(con, "minzoom", minZoom); err != nil {
		return err
	}
	if err = writeMetadataItem(con, "maxzoom", maxZoom); err != nil {
		return err
	}
	if err = writeMetadataItem(con, "center", fmt.Sprintf("%.5f,%.5f,%v", (bounds.Xmax-bounds.Xmin)/2.0, (bounds.Ymax-bounds.Ymin)/2.0, minZoom)); err != nil {
		return err
	}
	if err = writeMetadataItem(con, "bounds", fmt.Sprintf("%.5f,%.5f,%.5f,%.5f", bounds.Xmax, bounds.Ymin, bounds.Xmax, bounds.Ymax)); err != nil {
		return err
	}
	if err = writeMetadataItem(con, "type", "overlay"); err != nil {
		return err
	}
	if err = writeMetadataItem(con, "format", "png"); err != nil {
		return err
	}
	if err = writeMetadataItem(con, "version", "1.0.0"); err != nil {
		return err
	}

	return nil
}

func (db *MBtilesWriter) WriteTile(tile *tiles.TileID, data []byte) error {
	con, err := db.GetConnection()
	if err != nil {
		return err
	}
	defer db.CloseConnection(con)

	return WriteTile(con, tile, data)
}

// Write the tile to the open connection
func WriteTile(con *sqlite.Conn, tile *tiles.TileID, png []byte) (err error) {
	// flip tile Y to match mbtiles spec
	y := (1 << tile.Zoom) - 1 - tile.Y

	defer sqlitex.Save(con)(&err)

	h := sha1.New()
	h.Write(png)
	id := hex.EncodeToString(h.Sum(nil))

	err = sqlitex.Exec(con, "INSERT OR REPLACE INTO images (tile_id, tile_data) values (?, ?)",
		nil, id, png)
	if err != nil {
		return fmt.Errorf("could not write tile %v to mbtiles: %q", tile, err)
	}

	err = sqlitex.Exec(con, "INSERT OR REPLACE INTO map (zoom_level, tile_column, tile_row, tile_id) values(?, ?, ?, ?)",
		nil, tile.Zoom, tile.X, y, id)
	if err != nil {
		return fmt.Errorf("could not write tile %v to mbtiles: %q", tile, err)
	}

	return nil
}
