package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"swiftTask/handlers"

	"github.com/gin-gonic/gin"
	"go.etcd.io/bbolt"
)

const dbFile = "swift_codes.db"
const bucketName = "swift_code"

func main() {
	// Initialize BoltDB
	db, err := bbolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal("Ошибка открытия базы данных:", err)
	}
	defer db.Close()

	// Populate DB (only if empty)
	populateDatabase(db, "swiftCodes.csv")

	// Setup API with Gin
	router := gin.Default()

	// Define endpoints
	router.GET("/v1/swift-codes/:swift-code", handlers.GetSwiftCodeDetails(db, bucketName))
	router.DELETE("/v1/swift-codes/:swift-code", handlers.DeleteSwiftCode(db, bucketName))
	router.GET("/v1/swift-codes/country/:countryISO2code", handlers.GetSwiftCodesByCountry(db, bucketName))
	router.POST("/v1/swift-codes", handlers.AddSwiftCode(db, bucketName))

	// Start server
	fmt.Println("Server running on port 8080...")
	router.Run(":8080")
}

// Populate database from CSV
func populateDatabase(db *bbolt.DB, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("❌ Ошибка открытия CSV:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	records, err := reader.ReadAll()

	if err != nil {
		log.Println("❌ Ошибка чтения CSV:", err)
		return
	}

	if len(records) == 0 {
		log.Println("❌ Файл CSV пуст, загрузка данных невозможна")
		return
	}

	log.Printf("📄 Всего строк в CSV: %d\n", len(records))

	db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}

		count := 0
		for i, record := range records {
			if i == 0 {
				continue // Пропускаем заголовки
			}

			if len(record) < 6 {
				log.Printf("❌ Пропущена строка %d: %v (не хватает данных)\n", i, record)
				continue
			}

			swiftData := handlers.SwiftCode{
				Address:       record[4],
				BankName:      record[3],
				CountryISO2:   strings.ToUpper(record[0]),
				CountryName:   strings.ToUpper(record[6]),
				IsHeadquarter: handlers.IsHeadquarter(record[1]),
				SwiftCode:     record[1],
			}

			data, err := json.Marshal(swiftData)
			if err != nil {
				log.Printf("❌ Ошибка сериализации данных для строки %d: %v\n", i, err)
				continue
			}

			// Вставляем в BoltDB
			err = b.Put([]byte(record[1]), data)
			if err != nil {
				log.Printf("❌ Ошибка вставки в BoltDB: %v\n", err)
			} else {
				log.Printf("✅ Вставлен SWIFT-код: %s", record[1])
				count++
			}
		}

		log.Printf("✅ Загружено в базу записей: %d\n", count)
		return nil
	})
}
