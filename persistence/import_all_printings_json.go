package persistence

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

type AllPrintingsImportOptions struct {
	JSONPath  string
	BatchSize int
	MaxRows   int
	Verbose   bool
	Logger    *slog.Logger
}

type jsonCard struct {
	Artist             string   `json:"artist"`
	Availability       []string `json:"availability"`
	ColorIdentity      []string `json:"colorIdentity"`
	Colors             []string `json:"colors"`
	ConvertedManaCost  float64  `json:"convertedManaCost"`
	Keywords           []string `json:"keywords"`
	ManaValue          float64  `json:"manaValue"`
	Name               string   `json:"name"`
	Number             string   `json:"number"`
	OriginalText       string   `json:"originalText"`
	OriginalType       string   `json:"originalType"`
	Power              string   `json:"power"`
	SetCode            string   `json:"setCode"`
	Side               string   `json:"side"`
	Subtypes           []string `json:"subtypes"`
	Supertypes         []string `json:"supertypes"`
	Text               string   `json:"text"`
	Toughness          string   `json:"toughness"`
	Type               string   `json:"type"`
	Types              []string `json:"types"`
	UUID               string   `json:"uuid"`
	Identifiers        ids      `json:"identifiers"`
	ScryfallID         string   `json:"scryfallId"`
	ScryfallOracleID   string   `json:"scryfallOracleId"`
	ScryfallIllustID   string   `json:"scryfallIllustrationId"`
	CardKingdomID      string   `json:"cardKingdomId"`
	TcgplayerProductID string   `json:"tcgplayerProductId"`
}

type ids struct {
	ScryfallID             string `json:"scryfallId"`
	ScryfallOracleID       string `json:"scryfallOracleId"`
	ScryfallIllustrationID string `json:"scryfallIllustrationId"`
	CardKingdomID          string `json:"cardKingdomId"`
	TcgplayerProductID     string `json:"tcgplayerProductId"`
}

type setData struct {
	Cards []jsonCard `json:"cards"`
}

func ImportAllPrintingsJSON(dbURL string, opts AllPrintingsImportOptions) (int, error) {
	jsonPath := opts.JSONPath
	if strings.TrimSpace(jsonPath) == "" {
		jsonPath = "All Printings.json"
	}
	batchSize := opts.BatchSize
	if batchSize <= 0 {
		batchSize = 2000
	}
	logger := opts.Logger
	if logger == nil {
		level := slog.LevelInfo
		if opts.Verbose {
			level = slog.LevelDebug
		}
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	}

	f, err := os.Open(jsonPath)
	if err != nil {
		return 0, fmt.Errorf("open json file: %w", err)
	}
	defer f.Close()

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return 0, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	insertSQL := `INSERT INTO cards (
		id, name, facename, colors, convertedmanacost, types, power, toughness, text,
		subtypes, supertypes, uuid, coloridentity, type, number, setcode, side, availability,
		keywords, artist, tcgplayerproductid, scryfallid, scryfallillustrationid, scryfalloracleid, cardkingdomid
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9,
		$10, $11, $12, $13, $14, $15, $16, $17, $18,
		$19, $20, $21, $22, $23, $24, $25
	)
	ON CONFLICT (id) DO UPDATE SET
		name = EXCLUDED.name,
		facename = EXCLUDED.facename,
		colors = EXCLUDED.colors,
		convertedmanacost = EXCLUDED.convertedmanacost,
		types = EXCLUDED.types,
		power = EXCLUDED.power,
		toughness = EXCLUDED.toughness,
		text = EXCLUDED.text,
		subtypes = EXCLUDED.subtypes,
		supertypes = EXCLUDED.supertypes,
		coloridentity = EXCLUDED.coloridentity,
		type = EXCLUDED.type,
		number = EXCLUDED.number,
		setcode = EXCLUDED.setcode,
		side = EXCLUDED.side,
		availability = EXCLUDED.availability,
		keywords = EXCLUDED.keywords,
		artist = EXCLUDED.artist,
		tcgplayerproductid = EXCLUDED.tcgplayerproductid,
		scryfallid = EXCLUDED.scryfallid,
		scryfallillustrationid = EXCLUDED.scryfallillustrationid,
		scryfalloracleid = EXCLUDED.scryfalloracleid,
		cardkingdomid = EXCLUDED.cardkingdomid;`

	dec := json.NewDecoder(f)
	tok, err := dec.Token()
	if err != nil || tok != json.Delim('{') {
		return 0, fmt.Errorf("invalid json root: %w", err)
	}

	tx, stmt, err := beginBatch(db, insertSQL)
	if err != nil {
		return 0, err
	}
	defer func() {
		if stmt != nil {
			_ = stmt.Close()
		}
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	var total int
	var batchCount int

	for dec.More() {
		keyToken, err := dec.Token()
		if err != nil {
			return 0, fmt.Errorf("read json key: %w", err)
		}
		key, _ := keyToken.(string)
		if key != "data" {
			if err := discardValue(dec); err != nil {
				return 0, err
			}
			continue
		}

		if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
			return 0, fmt.Errorf("invalid data object: %w", err)
		}

		for dec.More() {
			if _, err := dec.Token(); err != nil { // set code
				return 0, fmt.Errorf("read set key: %w", err)
			}
			var set setData
			if err := dec.Decode(&set); err != nil {
				return 0, fmt.Errorf("decode set data: %w", err)
			}
			for _, card := range set.Cards {
				if card.UUID == "" || card.Name == "" {
					continue
				}
				faceName := ""
				if parts := strings.Split(card.Name, " // "); len(parts) > 1 {
					faceName = strings.TrimSpace(parts[0])
				}
				colors := joinOrNil(card.Colors)
				if colors == nil {
					colors = joinOrNil(card.ColorIdentity)
				}
				availability := joinOrNil(card.Availability)
				keywords := joinOrNil(card.Keywords)
				types := joinOrNil(card.Types)
				subtypes := joinOrNil(card.Subtypes)
				supertypes := joinOrNil(card.Supertypes)

				cmc := fmt.Sprintf("%g", card.ConvertedManaCost)
				if card.ConvertedManaCost == 0 && card.ManaValue != 0 {
					cmc = fmt.Sprintf("%g", card.ManaValue)
				}

				identifiers := card.Identifiers
				if card.ScryfallID != "" && identifiers.ScryfallID == "" {
					identifiers.ScryfallID = card.ScryfallID
				}
				if card.ScryfallOracleID != "" && identifiers.ScryfallOracleID == "" {
					identifiers.ScryfallOracleID = card.ScryfallOracleID
				}
				if card.ScryfallIllustID != "" && identifiers.ScryfallIllustrationID == "" {
					identifiers.ScryfallIllustrationID = card.ScryfallIllustID
				}
				if card.CardKingdomID != "" && identifiers.CardKingdomID == "" {
					identifiers.CardKingdomID = card.CardKingdomID
				}
				if card.TcgplayerProductID != "" && identifiers.TcgplayerProductID == "" {
					identifiers.TcgplayerProductID = card.TcgplayerProductID
				}

				if _, err := stmt.Exec(
					card.UUID,
					textOrNil(card.Name),
					textOrNil(faceName),
					colors,
					textOrNil(cmc),
					types,
					textOrNil(card.Power),
					textOrNil(card.Toughness),
					textOrNil(card.Text),
					subtypes,
					supertypes,
					card.UUID,
					joinOrNil(card.ColorIdentity),
					textOrNil(card.Type),
					textOrNil(card.Number),
					textOrNil(card.SetCode),
					textOrNil(card.Side),
					availability,
					keywords,
					textOrNil(card.Artist),
					textOrNil(identifiers.TcgplayerProductID),
					textOrNil(identifiers.ScryfallID),
					textOrNil(identifiers.ScryfallIllustrationID),
					textOrNil(identifiers.ScryfallOracleID),
					textOrNil(identifiers.CardKingdomID),
				); err != nil {
					logger.Warn("insert failed", "err", err, "name", card.Name)
				}
				total++
				batchCount++
				if opts.MaxRows > 0 && total >= opts.MaxRows {
					if err := commitBatch(tx, stmt); err != nil {
						return total, err
					}
					logger.Info("import complete (limit)", "count", total)
					return total, nil
				}
				if batchCount >= batchSize {
					if err := commitBatch(tx, stmt); err != nil {
						return total, err
					}
					logger.Info("batch committed", "count", total)
					tx, stmt, err = beginBatch(db, insertSQL)
					if err != nil {
						return total, err
					}
					batchCount = 0
				}
			}
		}

		if tok, err := dec.Token(); err != nil || tok != json.Delim('}') {
			return total, fmt.Errorf("invalid data closure: %w", err)
		}
	}

	if err := commitBatch(tx, stmt); err != nil {
		return total, err
	}
	logger.Info("import complete", "count", total)
	return total, nil
}

func joinOrNil(values []string) any {
	if len(values) == 0 {
		return nil
	}
	var cleaned []string
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		cleaned = append(cleaned, v)
	}
	if len(cleaned) == 0 {
		return nil
	}
	return strings.Join(cleaned, ",")
}

func textOrNil(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func beginBatch(db *sql.DB, insertSQL string) (*sql.Tx, *sql.Stmt, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, nil, fmt.Errorf("begin transaction: %w", err)
	}
	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		_ = tx.Rollback()
		return nil, nil, fmt.Errorf("prepare statement: %w", err)
	}
	return tx, stmt, nil
}

func commitBatch(tx *sql.Tx, stmt *sql.Stmt) error {
	if stmt != nil {
		if err := stmt.Close(); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func discardValue(dec *json.Decoder) error {
	var discard any
	if err := dec.Decode(&discard); err != nil && err != io.EOF {
		return fmt.Errorf("skip json key: %w", err)
	}
	return nil
}
