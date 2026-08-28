package store

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"

	"examvault/internal/domain"
)

func EncodeEnvelope(kind string, value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.BigEndian, uint16(len(kind))); err != nil {
		return nil, err
	}
	buf.WriteString(kind)
	buf.Write(raw)
	return buf.Bytes(), nil
}

func DecodeEnvelope(data []byte) (string, []byte, error) {
	if len(data) < 2 {
		return "", nil, fmt.Errorf("envelope too short")
	}
	size := int(binary.BigEndian.Uint16(data[:2]))
	if len(data) < size+2 {
		return "", nil, fmt.Errorf("envelope header truncated")
	}
	return string(data[2 : 2+size]), append([]byte(nil), data[2+size:]...), nil
}

func EncodeImportRows(rows []domain.ImportRow) string {
	parts := make([]string, 0, len(rows))
	for _, row := range rows {
		parts = append(parts, strings.Join([]string{row.ID, row.Title, row.Permission, row.Payload, row.Owner}, "|"))
	}
	return strings.Join(parts, "\n")
}

func DecodeImportRows(raw string) []domain.ImportRow {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	rows := make([]domain.ImportRow, 0, len(lines))
	for _, line := range lines {
		fields := strings.Split(line, "|")
		if len(fields) != 5 {
			continue
		}
		rows = append(rows, domain.ImportRow{ID: fields[0], Title: fields[1], Permission: fields[2], Payload: fields[3], Owner: fields[4]})
	}
	return rows
}
