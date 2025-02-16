package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.etcd.io/bbolt"
)

// SwiftCode struct represents a SWIFT entry
type SwiftCode struct {
	Address       string `json:"address"`
	BankName      string `json:"bankName"`
	CountryISO2   string `json:"countryISO2"`
	CountryName   string `json:"countryName"`
	IsHeadquarter bool   `json:"isHeadquarter"`
	SwiftCode     string `json:"swiftCode"`
}
type BranchesCode struct {
	Address       string `json:"address"`
	BankName      string `json:"bankName"`
	CountryISO2   string `json:"countryISO2"`
	IsHeadquarter bool   `json:"isHeadquarter"`
	SwiftCode     string `json:"swiftCode"`
}

// HeadquarterResponse struct for headquarters with branches
// type HeadquarterResponse struct {
// 	SwiftCode
// 	Branches []SwiftCode `json:"branches,omitempty"`
// }

type HeadquarterResponse struct {
	SwiftCode
	Branches []BranchesCode `json:"branches,omitempty"` // Використовуємо BranchesCode
}

// Function to determine if the SWIFT code represents a headquarters
// IsHeadquarter checks if a SWIFT code represents a headquarters
func IsHeadquarter(swiftCode string) bool {
	return len(swiftCode) >= 8 && strings.HasSuffix(swiftCode, "XXX")
}

func GetSwiftCodeDetails(db *bbolt.DB, bucketName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		swiftCode := c.Param("swift-code")
		log.Printf("API Request Received: SWIFT Code: %s", swiftCode)

		var swiftData SwiftCode
		var branches []BranchesCode // Використовуємо BranchesCode замість SwiftCode

		err := db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte(bucketName))
			if b == nil {
				log.Println("❌ Bucket not found!")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database bucket not found"})
				return nil
			}

			// Retrieve the requested SWIFT code's data
			data := b.Get([]byte(swiftCode))
			if data == nil {
				log.Printf("❌ SWIFT code %s not found in database", swiftCode)
				c.JSON(http.StatusNotFound, gin.H{"error": "SWIFT code not found"})
				return nil
			}

			// Parse SWIFT data
			if err := json.Unmarshal(data, &swiftData); err != nil {
				log.Println("❌ Error parsing SWIFT data:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing data"})
				return err
			}

			// Ensure country fields are uppercase
			swiftData.CountryISO2 = strings.ToUpper(swiftData.CountryISO2)
			swiftData.CountryName = strings.ToUpper(swiftData.CountryName)

			// If it's a headquarters, find its branches
			if swiftData.IsHeadquarter {
				log.Printf("🏦 Found Headquarters: %s", swiftCode)

				b.ForEach(func(k, v []byte) error {
					code := string(k)
					if code[:8] == swiftCode[:8] && code != swiftCode { // Додаємо перевірку, щоб не включати головний офіс
						var branch SwiftCode
						if err := json.Unmarshal(v, &branch); err == nil {
							// Конвертуємо SwiftCode до BranchesCode, виключаючи CountryName
							branches = append(branches, BranchesCode{
								Address:       branch.Address,
								BankName:      branch.BankName,
								CountryISO2:   strings.ToUpper(branch.CountryISO2),
								IsHeadquarter: branch.IsHeadquarter,
								SwiftCode:     branch.SwiftCode,
							})
						}
					}
					return nil
				})

				log.Printf("🏢 Found %d branches for HQ: %s", len(branches), swiftCode)

				c.JSON(http.StatusOK, HeadquarterResponse{
					SwiftCode: swiftData,
					Branches:  branches, // Використовуємо branches типу BranchesCode
				})
				return nil
			}

			// If it's a branch, return only the branch details
			log.Printf("🏢 Found Branch: %s", swiftCode)
			c.JSON(http.StatusOK, swiftData)
			return nil
		})

		if err != nil {
			log.Println("❌ Database error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
	}
}

// Handler for DELETE /v1/swift-codes/{swift-code}
func DeleteSwiftCode(db *bbolt.DB, bucketName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		swiftCode := c.Param("swift-code")
		log.Printf("API Request Received: Delete SWIFT Code: %s", swiftCode)

		err := db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte(bucketName))
			if b == nil {
				log.Println("❌ Bucket not found!")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database bucket not found"})
				return nil
			}

			// Delete the requested SWIFT code's data
			err := b.Delete([]byte(swiftCode))
			if err != nil {
				log.Printf("❌ Error deleting SWIFT code %s: %v", swiftCode, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting SWIFT code"})
				return err
			}

			log.Printf("✅ SWIFT code %s deleted successfully", swiftCode)
			c.JSON(http.StatusOK, gin.H{"message": "SWIFT code deleted successfully"})
			return nil
		})

		if err != nil {
			log.Println("❌ Database error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
	}
}

// CountryResponse struct for returning SWIFT codes for a specific country
type CountryResponse struct {
	CountryISO2 string      `json:"countryISO2"`
	CountryName string      `json:"countryName"`
	SwiftCodes  []SwiftCode `json:"swiftCodes"`
}

// Handler for GET /v1/swift-codes/country/{countryISO2code}
func GetSwiftCodesByCountry(db *bbolt.DB, bucketName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		countryISO2 := strings.ToUpper(c.Param("countryISO2code"))
		log.Printf("API Request Received: SWIFT Codes for Country: %s", countryISO2)

		var swiftCodes []SwiftCode
		var countryName string

		err := db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte(bucketName))
			if b == nil {
				log.Println("❌ Bucket not found!")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database bucket not found"})
				return nil
			}

			b.ForEach(func(k, v []byte) error {
				var swiftData SwiftCode
				if err := json.Unmarshal(v, &swiftData); err == nil {
					if swiftData.CountryISO2 == countryISO2 {
						swiftCodes = append(swiftCodes, swiftData)
						countryName = swiftData.CountryName
					}
				}
				return nil
			})

			return nil
		})

		if err != nil {
			log.Println("❌ Database error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if len(swiftCodes) == 0 {
			log.Printf("❌ No SWIFT codes found for country: %s", countryISO2)
			c.JSON(http.StatusNotFound, gin.H{"error": "No SWIFT codes found for country"})
			return
		}

		c.JSON(http.StatusOK, CountryResponse{
			CountryISO2: countryISO2,
			CountryName: countryName,
			SwiftCodes:  swiftCodes,
		})
	}
}

// Handler for POST /v1/swift-codes
func AddSwiftCode(db *bbolt.DB, bucketName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var newSwiftCode SwiftCode
		if err := c.ShouldBindJSON(&newSwiftCode); err != nil {
			log.Println("❌ Error parsing request body:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		err := db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte(bucketName))
			if b == nil {
				log.Println("❌ Bucket not found!")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database bucket not found"})
				return nil
			}

			data, err := json.Marshal(newSwiftCode)
			if err != nil {
				log.Println("❌ Error serializing SWIFT code data:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error serializing data"})
				return err
			}

			err = b.Put([]byte(newSwiftCode.SwiftCode), data)
			if err != nil {
				log.Println("❌ Error inserting SWIFT code into database:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting data"})
				return err
			}

			log.Printf("✅ SWIFT code %s added successfully", newSwiftCode.SwiftCode)
			c.JSON(http.StatusOK, gin.H{"message": "SWIFT code added successfully"})
			return nil
		})

		if err != nil {
			log.Println("❌ Database error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
	}
}
