/*
db_import.go is a script for downloading the latest zipped CSV from the MTGJSON source
and then importing it into a target SQL database. It unzips the downloaded file into
the project root and then attempts to import all of the cards in the `allprintings.csv`
file. It is written as a CLI utility with some convenience functionality wrapped around
it for verbose logging and targeting different databases.
*/

package main

import (
	"archive/zip"
	"bytes"
	"encoding/gob"
	"flag"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/openmtg/edh-go/persistence"

	"github.com/gocarina/gocsv"
)

// init runs gob registration functions
func init() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
}

// DBURL accesses a local Postgres instance running for local development.
var DBURL = "postgres://edhgo:edhgo@localhost:5432/edhgo?sslmode=disable"

// CSVCard is a struct that exactly follows the structure of the AllPrintingsCSV
// file that we get from MTGJSON.
// All of the fields are exported for ease of use.
type CSVCard struct {
	ID                     string `csv:"id"`
	Artist                 string `csv:"artist"`
	AsciiName              string `csv:"asciiName"`
	Availability           string `csv:"availability"`
	BorderColor            string `csv:"borderColor"`
	CardKingdomId          string `csv:"cardKingdomId"`
	ColorIdentity          string `csv:"colorIdentity"`
	ColorIndicator         string `csv:"colorIndicator"`
	Colors                 string `csv:"colors"`
	ConvertedManaCost      string `csv:"convertedManaCost"`
	FaceConvertedManaCost  string `csv:"faceConvertedManaCost"`
	FaceManaValue          string `csv:"faceManaValue"`
	FlavorName             string `csv:"flavorName"`
	FlavorText             string `csv:"flavorText"`
	Keywords               string `csv:"keywords"`
	MTGJSONV4ID            string `csv:"mtgjsonV4Id"`
	FaceName               string `csv:"faceName"`
	Name                   string `csv:"name"`
	Number                 string `csv:"number"`
	OriginalText           string `csv:"originalText"`
	OriginalType           string `csv:"originalType"`
	Power                  string `csv:"power"`
	ScryfallID             string `csv:"scryfallId"`
	ScryfallIllustrationID string `csv:"scryfallIllustrationId"`
	ScryfallOracleID       string `csv:"scryfallOracleId"`
	SetCode                string `csv:"setCode"`
	Side                   string `csv:"side"`
	Subtypes               string `csv:"subtypes"`
	Supertypes             string `csv:"supertypes"`
	TCGPlayerProductID     string `csv:"tcgplayerProductId"`
	Text                   string `csv:"text"`
	Toughness              string `csv:"toughness"`
	Type                   string `csv:"type"`
	Types                  string `csv:"types"`
	UUID                   string `csv:"uuid"`
}

type ImportReport struct {
	success int64
	errors  int64
}

func main() {
	var dburl = flag.String("db", DBURL, "connection URL for target import database. defaults to http://localhost:5432/vedh")
	var refresh = flag.Bool("refresh", false, "refresh specifies whether the card database should be downloaded fresh. defaults to false.")
	var verbose = flag.Bool("verbose", false, "specifies if the log output should be more verbose")

	flag.Parse()

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	parsed, err := url.Parse(*dburl)
	if err != nil {
		logger.Error("failed to parse url", "err", err)
		os.Exit(1)
	}

	logger.Debug("import targeting", "host", parsed.Host)

	conn, err := persistence.NewDB(*dburl)
	if err != nil {
		logger.Error("failed to connect to postgres for import", "err", err)
		os.Exit(1)
	}

	if *refresh {
		logger.Info("refreshing card sources before import")
		updateAndUnzip(logger)
	}

	// open up the file for reading
	f, err := os.Open("./cards.csv")
	if err != nil {
		logger.Error("failed to open file", "err", err)
		os.Exit(1)
	}
	defer f.Close()

	// build a csv reader
	cards := []*CSVCard{}
	err = gocsv.Unmarshal(f, &cards)
	if err != nil {
		logger.Error("failed to unmarshal csv", "err", err)
		os.Exit(1)
	}

	logger.Info("attempting to import cards", "count", len(cards), "db_host", strings.TrimSpace(parsed.Host))

	report := ImportReport{
		success: 0,
		errors:  0,
	}

	for _, c := range cards {
		// insert each card that we read in
		_, err := conn.Exec(insertSQL, c.ID, c.Artist, c.AsciiName, c.Availability,
			c.BorderColor, c.CardKingdomId, c.ColorIdentity, c.ColorIndicator,
			c.Colors, c.ConvertedManaCost, c.FaceConvertedManaCost,
			c.FaceManaValue, c.FlavorName, c.FlavorText, c.Keywords,
			c.MTGJSONV4ID, c.Name, c.FaceName, c.Number, c.OriginalText, c.OriginalType,
			c.Power, c.ScryfallID, c.ScryfallIllustrationID, c.ScryfallOracleID,
			c.SetCode, c.Side, c.Subtypes, c.Supertypes, c.TCGPlayerProductID,
			c.Text, c.Toughness, c.Type, c.Types, c.UUID,
		)
		if err != nil {
			logger.Warn("failed to insert card", "name", c.Name, "err", err)
			report.errors++
			continue
		} else {
			logger.Debug("imported", "name", c.Name)
			report.success++
		}
	}
	logger.Info("import report", "success", report.success, "errors", report.errors)
}

var insertSQL string = `INSERT INTO cards (
	ID, Artist, AsciiName, Availability, BorderColor, CardKingdomId,
	ColorIdentity, ColorIndicator, Colors, ConvertedManaCost,
	FaceConvertedManaCost, FaceManaValue, FlavorName, FlavorText, Keywords,
	MTGJSONV4ID, Name, FaceName, Number, OriginalText, OriginalType, Power, ScryfallID,
	ScryfallIllustrationID, ScryfallOracleID, SetCode, Side, Subtypes,
	Supertypes, TCGPLayerProductID, Text, Toughness, Type, Types, UUID)

	VALUES(
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
		$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
		$31, $32, $33, $34, $35)
	ON CONFLICT (id) DO
		UPDATE SET facename = EXCLUDED.facename;`

// DownloadURL will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadURL(filepath string, url string) error {
	// Download the latest card data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// GetBytes uses gob encoding to return the bytes of an interface.
func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnzipFile returns the path or an error of a folder or file that it unzipped.
// * Unzips zip files to a directory of the same name as the compressed file.
// * It handles directories and single files.
func UnzipFile(path string) error {
	// Create a reader out of the zip archive
	zipReader, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	// Iterate through each file/dir found in
	for _, file := range zipReader.Reader.File {
		// Open the file inside the zip archive
		// like a normal file
		zippedFile, err := file.Open()
		if err != nil {
			return err
		}
		defer zippedFile.Close()

		// Specify what the extracted file name should be.
		// You can specify a full path or a prefix
		// to move it to a different directory.
		// In this case, we will extract the file from
		// the zip to a file of the same name.
		targetDir := "./"
		extractedFilePath := filepath.Join(
			targetDir,
			file.Name,
		)

		// Extract the item (or create directory)
		if file.FileInfo().IsDir() {
			// Create directories to recreate directory
			// structure inside the zip archive. Also
			// preserves permissions
			slog.Default().Debug("creating directory", "path", extractedFilePath)
			if err := os.MkdirAll(extractedFilePath, file.Mode()); err != nil {
				return err
			}
		} else {
			// Extract regular file since not a directory
			slog.Default().Debug("extracting file", "name", file.Name)

			// Open an output file for writing
			outputFile, err := os.OpenFile(
				extractedFilePath,
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				file.Mode(),
			)
			if err != nil {
				return err
			}
			defer outputFile.Close()

			// "Extract" the file by copying zipped file
			// contents to the output file
			_, err = io.Copy(outputFile, zippedFile)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {
	// Get the request data at `url`
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func updateAndUnzip(logger *slog.Logger) {
	if logger == nil {
		logger = slog.Default()
	}
	logger.Info("refreshing allprintings file")
	// Let's take a stab at CSV rendering
	// Download the CSV AllPrintings zip file
	err := DownloadFile("./allprintings.csv.zip", "https://mtgjson.com/api/v5/AllPrintingsCSVFiles.zip")
	if err != nil {
		logger.Error("failed to download file", "err", err)
		os.Exit(1)
	}

	// // Commented out until cleanup is written.
	// // Unzip the CSV file and get the path
	err = UnzipFile("./allprintings.csv.zip")
	if err != nil {
		logger.Error("failed to unzip file", "err", err)
		os.Exit(1)
	}
	logger.Info("unzipped allprintings.csv file successfully")
}
