package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"swiftTask/handlers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
)

const testDBFile = "test_swift_codes.db"
const testBucketName = "swift_code"

func setupRouter(db *bbolt.DB) *gin.Engine {
	router := gin.Default()
	router.GET("/v1/swift-codes/:swift-code", handlers.GetSwiftCodeDetails(db, testBucketName))
	router.DELETE("/v1/swift-codes/:swift-code", handlers.DeleteSwiftCode(db, testBucketName))
	router.GET("/v1/swift-codes/country/:countryISO2code", handlers.GetSwiftCodesByCountry(db, testBucketName))
	router.POST("/v1/swift-codes", handlers.AddSwiftCode(db, testBucketName))
	return router
}

func TestAddSwiftCode(t *testing.T) {
	db, _ := bbolt.Open(testDBFile, 0600, nil)
	defer db.Close()
	router := setupRouter(db)

	newSwiftCode := handlers.SwiftCode{
		Address:       "Test Address",
		BankName:      "Test Bank",
		CountryISO2:   "US",
		CountryName:   "United States",
		IsHeadquarter: true,
		SwiftCode:     "TESTUS33XXX",
	}

	jsonValue, _ := json.Marshal(newSwiftCode)
	req, _ := http.NewRequest("POST", "/v1/swift-codes", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "SWIFT code added successfully")
}

func TestGetSwiftCodeDetails(t *testing.T) {
	db, _ := bbolt.Open(testDBFile, 0600, nil)
	defer db.Close()
	router := setupRouter(db)

	// Add a test SWIFT code
	db.Update(func(tx *bbolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(testBucketName))
		swiftCode := handlers.SwiftCode{
			Address:       "Test Address",
			BankName:      "Test Bank",
			CountryISO2:   "US",
			CountryName:   "United States",
			IsHeadquarter: true,
			SwiftCode:     "TESTUS33XXX",
		}
		data, _ := json.Marshal(swiftCode)
		b.Put([]byte(swiftCode.SwiftCode), data)
		return nil
	})

	req, _ := http.NewRequest("GET", "/v1/swift-codes/TESTUS33XXX", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test Bank")
}

func TestDeleteSwiftCode(t *testing.T) {
	db, _ := bbolt.Open(testDBFile, 0600, nil)
	defer db.Close()
	router := setupRouter(db)

	// Add a test SWIFT code
	db.Update(func(tx *bbolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(testBucketName))
		swiftCode := handlers.SwiftCode{
			Address:       "Test Address",
			BankName:      "Test Bank",
			CountryISO2:   "US",
			CountryName:   "United States",
			IsHeadquarter: true,
			SwiftCode:     "TESTUS33XXX",
		}
		data, _ := json.Marshal(swiftCode)
		b.Put([]byte(swiftCode.SwiftCode), data)
		return nil
	})

	req, _ := http.NewRequest("DELETE", "/v1/swift-codes/TESTUS33XXX", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "SWIFT code deleted successfully")
}

func TestGetSwiftCodesByCountry(t *testing.T) {
	db, _ := bbolt.Open(testDBFile, 0600, nil)
	defer db.Close()
	router := setupRouter(db)

	// Add test SWIFT codes
	db.Update(func(tx *bbolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(testBucketName))
		swiftCodes := []handlers.SwiftCode{
			{
				Address:       "Test Address 1",
				BankName:      "Test Bank 1",
				CountryISO2:   "US",
				CountryName:   "United States",
				IsHeadquarter: true,
				SwiftCode:     "TESTUS33XXX",
			},
			{
				Address:       "Test Address 2",
				BankName:      "Test Bank 2",
				CountryISO2:   "US",
				CountryName:   "United States",
				IsHeadquarter: false,
				SwiftCode:     "TESTUS33YYY",
			},
		}
		for _, swiftCode := range swiftCodes {
			data, _ := json.Marshal(swiftCode)
			b.Put([]byte(swiftCode.SwiftCode), data)
		}
		return nil
	})

	req, _ := http.NewRequest("GET", "/v1/swift-codes/country/US", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test Bank 1")
	assert.Contains(t, w.Body.String(), "Test Bank 2")
}
