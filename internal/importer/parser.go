package importer

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"firmwarehub/internal/domain"
)

func ParseCSV(input string, source string) ([]domain.ImportRow, []string, error) {
	reader := csv.NewReader(strings.NewReader(input))
	reader.FieldsPerRecord = 9
	rows := make([]domain.ImportRow, 0)
	warnings := make([]string, 0)
	line := 0
	for {
		fields, err := reader.Read()
		if err == io.EOF {
			break
		}
		line++
		if err != nil {
			return nil, warnings, fmt.Errorf("parse line %d: %w", line, err)
		}
		if line == 1 && strings.EqualFold(fields[0], "id") {
			continue
		}
		score, err := strconv.Atoi(fields[6])
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("line %d score ignored: %v", line, err))
			score = 0
		}
		rows = append(rows, domain.ImportRow{ID: fields[0], Product: fields[1], Version: fields[2], Checksum: fields[3], DownloadURL: fields[4], Platform: fields[5], Score: score, Owner: fields[7], Source: source})
	}
	return rows, warnings, nil
}

func EncodeCSV(rows []domain.ImportRow) string {
	var b strings.Builder
	b.WriteString("id,product,version,checksum,download_url,platform,score,owner,source\n")
	for _, row := range rows {
		b.WriteString(strings.Join([]string{row.ID, row.Product, row.Version, row.Checksum, row.DownloadURL, row.Platform, strconv.Itoa(row.Score), row.Owner, row.Source}, ","))
		b.WriteByte('\n')
	}
	return b.String()
}
