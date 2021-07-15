package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

type Credential struct {
	Uid         uuid.UUID
	Service     string `json:"service" validate:"required"`
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
	Description string
	Metadata    map[string]string
	CreatedAt   CustomTime
	UpdatedAt   CustomTime
}

type CustomTime time.Time

const ctLayout = time.RFC1123

func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	nt, err := time.Parse(ctLayout, s)
	*ct = CustomTime(nt)
	return
}

func (ct CustomTime) MarshalJSON() ([]byte, error) {
	return []byte(ct.String()), nil
}

// String returns the time in the custom format
func (ct *CustomTime) String() string {
	t := time.Time(*ct)
	return fmt.Sprintf("%q", t.Format(ctLayout))
}

//
//type NullTime sql.NullTime
//
//func NewNullTime() NullTime {
//	return NullTime{
//		Time:  time.Now(),
//		Valid: true,
//	}
//}
//
//// Value implements the driver Valuer interface.
//func (nt NullTime) Value() (driver.Value, error) {
//	if !nt.Valid {
//		return nil, nil
//	}
//	return nt.Time, nil
//}
//
//// Scan implements the Scanner interface for NullTime
//func (nt *NullTime) Scan(value interface{}) error {
//	var t sql.NullTime
//	if err := t.Scan(value); err != nil {
//		return err
//	}
//
//	// if nil then make Valid false
//	if reflect.TypeOf(value) == nil {
//		*nt = NullTime{Time: t.Time, Valid: false}
//	} else {
//		*nt = NullTime{Time: t.Time, Valid: true}
//	}
//
//	return nil
//}
//
//// MarshalJSON for NullTime
//func (nt *NullTime) MarshalJSON() ([]byte, error) {
//	if !nt.Valid {
//		return []byte("null"), nil
//	}
//	val := fmt.Sprintf("\"%s\"", nt.Time.Format(time.RFC3339))
//	return []byte(val), nil
//}
//
//// UnmarshalJSON for NullTime
//func (nt *NullTime) UnmarshalJSON(b []byte) error {
//	s, _ := strconv.Unquote(string(b))
//	x, err := time.Parse(time.RFC3339, s)
//	if err != nil {
//		nt.Valid = false
//		return nil
//	}
//
//	nt.Time = x
//	nt.Valid = true
//	return nil
//}
