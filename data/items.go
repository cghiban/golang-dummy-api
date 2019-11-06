package data

import (
	"database/sql"
	"fmt"
	"strings"

	"../model"
)

type Items struct {
	DB *sql.DB
}

const sqlQueryTemplate = `SELECT at_id, at_name, at_shortdesc, at_keywords
			FROM atoms
			WHERE at_public = 'yes' AND at_type != ''
			AND at_id IN (15140, 15688, 16578, 15579, 15551, 16453, 16942, 15087, 15506, 16151, 2398, 15310, 16704, 17044, 16444, 15476,  16929, 16627, 16529, 16489, 16987, 16934, 16541)
			%s
			ORDER BY RAND() DESC LIMIT 25`

func toUtf8(s string) string {
	iso8859_1_buf := []byte(s)
	buf := make([]rune, len(iso8859_1_buf))
	for i, b := range iso8859_1_buf {
		buf[i] = rune(b)
	}
	return string(buf)
}

// GetAll - retrieves items from the database
func (a *Items) GetAll(lastUpdated ...string) ([]model.Item, error) {

	dbh := a.DB

	var sth *sql.Rows
	var sqlQuery string
	var err error

	if len(lastUpdated) > 0 {
		sqlQuery = fmt.Sprintf(sqlQueryTemplate, "AND at_date_update >= ?")
		sth, err = dbh.Query(sqlQuery, lastUpdated[0])
	} else {
		sqlQuery = fmt.Sprintf(sqlQueryTemplate, "")
		sth, err = dbh.Query(sqlQuery)
	}

	if err != nil {
		panic(err.Error())
	}

	items := []model.Item{}
	for sth.Next() {
		var itemID int32
		var itemIDStr string
		var item model.Item
		err := sth.Scan(&itemID, &item.Name, &item.ShortDesc, &item.Keywords)
		if err != nil {
			panic(err.Error())
		}

		itemIDStr = fmt.Sprintf("%d", itemID)
		item.ID = &itemIDStr

		*item.Name = strings.TrimSpace(*item.Name)
		*item.ShortDesc = strings.TrimSpace(*item.ShortDesc)

		if "" == *item.Keywords {
			item.Keywords = nil
		} else {
			*item.Keywords = strings.TrimSpace(*item.Keywords)
		}

		items = append(items, item)
	}
	return items, nil
}
