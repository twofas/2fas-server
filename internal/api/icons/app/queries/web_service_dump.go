package queries

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"
)

type WebServiceDumpPresenter struct {
	WebService datatypes.JSON
}

type WebServicesDumpQuery struct {
	Limit  int `uri:"limit"`
	Offset int `uri:"offset"`
}

type WebServicesDumpQueryHandler struct {
	Database *sql.DB
}

func (h *WebServicesDumpQueryHandler) Dump(q *WebServicesDumpQuery) ([]json.RawMessage, error) {
	rows, err := h.Database.Query("SELECT * FROM web_services_dump")

	defer rows.Close()

	var emptyResult []json.RawMessage

	if err != nil {
		return emptyResult, err
	}

	var services []json.RawMessage

	for rows.Next() {
		var row json.RawMessage

		err := rows.Scan(&row)

		if err != nil {
			return emptyResult, err
		}

		var jsonRow json.RawMessage

		err = json.Unmarshal(row, &jsonRow)

		if err != nil {
			return emptyResult, err
		}

		out, err := json.Marshal(&jsonRow)
		if err != nil {
			return emptyResult, fmt.Errorf("could not marshal json: %w", err)
		}

		services = append(services, out)
	}

	if err = rows.Err(); err != nil {
		return emptyResult, err
	}

	return services, nil
}
