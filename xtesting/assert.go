package xtesting

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/buger/jsonparser"
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

type ValueChecker func(t *testing.T, v string)

func ValueTimeInPastDuration(dur time.Duration) ValueChecker {
	return func(t *testing.T, v string) {
		ti, err := time.Parse(time.RFC3339, strings.Trim(v, "\""))
		assert.NoError(t, err)
		assert.WithinRange(t, ti, time.Now().Add(-dur), time.Now())
	}
}
func ValueTime() ValueChecker {
	return func(t *testing.T, v string) {
		_, err := time.Parse(time.RFC3339, strings.Trim(v, "\""))
		assert.NoError(t, err)
	}
}

func ValueUUID() ValueChecker {
	return func(t *testing.T, v string) {
		_, err := uuid.Parse(strings.Trim(v, "\""))
		assert.NoError(t, err)
	}
}

func ValueRegexp(rx any) ValueChecker {
	return func(t *testing.T, v string) {
		assert.Regexp(t, rx, v)
	}
}

func ValueBase64Token(bitLength int) ValueChecker {
	return func(t *testing.T, v string) {

		token, err := base64.RawURLEncoding.DecodeString(strings.Trim(v, "\""))
		assert.NoError(t, err)
		assert.Len(t, token, bitLength/8)
	}
}

func ValueBcryptPassword(password string) ValueChecker {
	return func(t *testing.T, v string) {
		storedPassword := strings.Trim(v, "\"")
		err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
		assert.NoError(t, err)
	}
}

func ValueNotEqual(val string) ValueChecker {
	return func(t *testing.T, v string) {
		assert.NotEqual(t, val, v)
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
				result[fmt.Sprintf("[%d].%s", i, element)] = val
			}
		}
	}
	return result
}

func AssertGoldenJSON(t *testing.T, w *httptest.ResponseRecorder, ignore ...ValuesCheckers) {
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("content-type"))
	AssertGoldenJSONWithName(t, w.Body.Bytes(), "", ignore...)
}

func AssertGoldenJSONWithName(t *testing.T, got []byte, goldenName string, ignore ...ValuesCheckers) {
	if got != nil {
		if len(ignore) > 0 {
			for jsonPath, fn := range ignore[0] {
				value, err := jsonparser.GetString(got, strings.Split(jsonPath, ".")...)
				if errors.Is(err, jsonparser.KeyPathNotFoundError) {
					continue
				}
				assert.NoError(t, err)

				if fn != nil {
					fn(t, value)

					got, err = jsonparser.Set(got, []byte(`"-- Dynamic value --"`), strings.Split(jsonPath, ".")...)
					require.Nil(t, err)
				}
			}
		}
	}

	t.Log("Got: ", string(got))

	var indentedResult bytes.Buffer
	if string(got) != "" {
		err := json.Indent(&indentedResult, got, "", "\t")
		assert.NoError(t, err)
	}

	fileNamePath := fmt.Sprintf("testdata/%s%s.golden", t.Name(), goldenName)

	UpdateGoldenIfFlagSet(t, indentedResult.Bytes(), fileNamePath)

	f := ReadGoldenFile(t, fileNamePath)

	if indentedResult.String() != "" {
		assert.JSONEq(t, string(f), indentedResult.String())
	} else {
		assert.Equal(t, string(f), indentedResult.String())
	}

}

func AssertGoldenDatabaseTable(t *testing.T, db *gorm.DB, query any, ignore map[string]ValueChecker) {

	result := db.Order(clause.OrderByColumn{Column: clause.PrimaryColumn}).Find(&query)
	assert.NoError(t, result.Error)

	got, err := json.Marshal(query)
	assert.NoError(t, err)

	AssertGoldenJSONWithName(t, got, ".db."+result.Statement.Schema.Table, ignore)
}

func UpdateGoldenIfFlagSet(t *testing.T, data []byte, fileNamePath string) {
	if !flag.Parsed() {
		flag.Parse()
	}

	if *updateGoldenFiles {
		err := os.MkdirAll(path.Dir(fileNamePath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fileNamePath, data, 0644)
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
