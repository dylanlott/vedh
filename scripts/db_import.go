/*
This is a script for downloading the latest zipped CSV from MTGJSON and then
importing it into our database.

It unzips the file into the project root, but will cleanup after itself.

// MAKE SURE THAT THIS IS TRUE.
It is idempotent and will upsert cards in the `cards` database based on
their ID's retrieved from the AllPrintings CSV file.
*/

package main

import (
	"archive/zip"
	"bytes"
	"encoding/gob"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dylanlott/edh-go/persistence"
	"github.com/gocarina/gocsv"
)

// init runs our gob registration functions
func init() {
	// Register gob types for getBytes decoding
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
}

const (
	// DB_URL accesses a local Postgres instance running for local development.
	// DB_URL = "postgres://edhgo:edhgodev@localhost:5432/edhgo?sslmode=disable"
	DB_URL = os.Getenv("EDHGO_PG_URL")
)

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

// TODO: Make this a commandline utility.
func main() {
	// declare flag variables
	var refresh bool = false
	var dburl string = DB_URL
	var drop bool = false

	// declare flags
	flag.Bool("refresh", refresh, "refresh specifies whether the card database should be downloaded fresh. defaults to false.")
	flag.StringVar(&dburl, "db", DB_URL, "import target database connection url. defaults to localhost:5432/edhgo")
	flag.Bool("drop", drop, "if drop is enabled then the database table will be dropped before the new set is loaded")

	// parse flags
	flag.Parse()

	// connect to Postgres
	conn, err := persistence.NewDB(DB_URL)
	if err != nil {
		log.Fatalf("failed to connecto to postgres for import: %s", err)
	}

	// handle refresh
	if refresh {
		updateAndUnzip()
	}

	// open up the file for reading
	f, err := os.Open("./cards.csv")
	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}
	defer f.Close()

	// build a csv reader
	cards := []*CSVCard{}
	err = gocsv.Unmarshal(f, &cards)
	if err != nil {
		log.Fatalf("failed to unmarshal csv: %s", err)
	}
	log.Printf("attempting to insert %d cards into postgres", len(cards))
	success := 0
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
			log.Printf("failed to insert card [%s]: %s", c.Name, err)
			continue
		}
		log.Printf("successfully imported %v", c.Name)
		success++
	}
	log.Printf("imported %d cards", success)
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
		log.Fatal(err)
	}
	defer zipReader.Close()

	// Iterate through each file/dir found in
	for _, file := range zipReader.Reader.File {
		// Open the file inside the zip archive
		// like a normal file
		zippedFile, err := file.Open()
		if err != nil {
			log.Fatal(err)
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
			log.Println("Creating directory:", extractedFilePath)
			os.MkdirAll(extractedFilePath, file.Mode())
		} else {
			// Extract regular file since not a directory
			log.Println("Extracting file:", file.Name)

			// Open an output file for writing
			outputFile, err := os.OpenFile(
				extractedFilePath,
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				file.Mode(),
			)
			if err != nil {
				log.Fatal(err)
			}
			defer outputFile.Close()

			// "Extract" the file by copying zipped file
			// contents to the output file
			_, err = io.Copy(outputFile, zippedFile)
			if err != nil {
				log.Fatal(err)
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
func updateAndUnzip() {
	log.Printf("refreshing allprintings.csv file")
	// Let's take a stab at CSV rendering
	// Download the CSV AllPrintings zip file
	err := DownloadFile("./allprintings.csv.zip", "https://mtgjson.com/api/v5/AllPrintingsCSVFiles.zip")
	if err != nil {
		log.Fatalf("failed to download file: %s", err)
	}

	// // Commented out until cleanup is written.
	// // Unzip the CSV file and get the path
	err = UnzipFile("./allprintings.csv.zip")
	if err != nil {
		log.Fatalf("failed to unzip file: %s", err)
	}
	log.Printf("unzipped allprintings.csv file successfully")
}
