package xtesting

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/Jeffail/gabs/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	updateGoldenFiles = flag.Bool("update", false, "update the golden files of this test")
)

type ValueChecker func(t *testing.T, v *gabs.Container)

func ValueTimeInPastDuration(dur time.Duration) ValueChecker {
	return func(t *testing.T, v *gabs.Container) {
		ti, err := time.Parse(time.RFC3339, strings.Trim(v.String(), "\""))
		assert.NoError(t, err)
		assert.WithinRange(t, ti, time.Now().Add(-dur), time.Now())
	}
}
func ValueTime() ValueChecker {
	return func(t *testing.T, v *gabs.Container) {
		_, err := time.Parse(time.RFC3339, strings.Trim(v.String(), "\""))
		assert.NoError(t, err)
	}
}

func ValueUUID() ValueChecker {
	return func(t *testing.T, v *gabs.Container) {
		_, err := uuid.Parse(strings.Trim(v.String(), "\""))
		assert.NoError(t, err)
	}
}

func ValueRegexp(rx any) ValueChecker {
	return func(t *testing.T, v *gabs.Container) {
		assert.Regexp(t, rx, v.String())
	}
}

func ValueBase64Token(bitLength int) ValueChecker {
	return func(t *testing.T, v *gabs.Container) {

		token, err := base64.RawURLEncoding.DecodeString(strings.Trim(v.String(), "\""))
		assert.NoError(t, err)
		assert.Len(t, token, bitLength/8)
	}
}

func ValueBcryptPassword(password string) ValueChecker {
	return func(t *testing.T, v *gabs.Container) {
		storedPassword := bytes.Trim(v.Bytes(), "\"")
		err := bcrypt.CompareHashAndPassword(storedPassword, []byte(password))
		assert.NoError(t, err)
	}
}

func ValueNotEqual(val string) ValueChecker {
	return func(t *testing.T, v *gabs.Container) {
		assert.NotEqual(t, val, v.String())
	}
}

type ValuesCheckers map[string]ValueChecker

func GenerateValueCheckersForArrays(checkers map[string]ValueChecker, n int) ValuesCheckers {
	return GenerateValueCheckersForArraysWithOffset(checkers, n, 0)
}

func GenerateValueCheckersForArraysWithOffset(checkers map[string]ValueChecker, n int, offset int) ValuesCheckers {
	result := ValuesCheckers{}

	for element := range checkers {
		val, ok := checkers[element]
		if ok {
			for i := offset; i < n+offset; i++ {
				result[fmt.Sprintf("%d.%s", i, element)] = val
			}
		}
	}
	return result
}

func AssertGoldenJSON(t *testing.T, w *httptest.ResponseRecorder, ignore ...ValuesCheckers) {
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("content-type"))
	AssertGoldenJSONWithName(t, w.Body.String(), "", ignore...)
}

func AssertGoldenJSONWithName(t *testing.T, got string, goldenName string, ignore ...ValuesCheckers) {
	dataString := ""
	dataStringIndent := ""
	if got != "" {

		j, err := gabs.ParseJSONBuffer(strings.NewReader(got))
		assert.NoError(t, err)

		if len(ignore) > 0 {
			for jsonPath, fn := range ignore[0] {
				if !j.ExistsP(jsonPath) {
					continue
				}

				if fn != nil {
					fn(t, j.Path(jsonPath))

					_, err := j.SetP("-- Dynamic value --", jsonPath)
					require.Nil(t, err)
				}
			}
		}
		dataString = j.String()
		dataStringIndent = j.StringIndent("", "  ")
	}

	fileNamePath := fmt.Sprintf("testdata/%s%s.golden", t.Name(), goldenName)

	UpdateGoldenIfFlagSet(t, dataStringIndent, fileNamePath)

	f := ReadGoldenFile(t, fileNamePath)

	if dataString != "" {
		assert.JSONEq(t, string(f), dataString)
	} else {
		assert.Equal(t, string(f), dataString)
	}

}

func AssertGoldenDatabaseTable(t *testing.T, db *gorm.DB, query any, ignore map[string]ValueChecker) {

	result := db.Order(clause.OrderByColumn{Column: clause.PrimaryColumn}).Find(&query)
	assert.NoError(t, result.Error)

	got, err := json.Marshal(query)
	assert.NoError(t, err)

	AssertGoldenJSONWithName(t, string(got), ".db."+result.Statement.Schema.Table, ignore)
}

func UpdateGoldenIfFlagSet(t *testing.T, got, fileNamePath string) {
	if !flag.Parsed() {
		flag.Parse()
	}

	if *updateGoldenFiles {
		err := os.MkdirAll(path.Dir(fileNamePath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fileNamePath, []byte(got), 0644)
		if err != nil {
			t.Fatalf("Error writing to file %s: %s", fileNamePath, err)
		}
		return
	}
}

func ReadGoldenFile(t *testing.T, fileNamePath string) []byte {
	f, err := os.ReadFile(fileNamePath)
	assert.NoError(t, err, "Error loading golden file")
	return f
}
