package main

import (
	"risk-service/services"

	"gorm.io/gorm"
)

// seedCosoCatalogue runs in `risk up` (autoMigrate). Insert-if-missing
// reference data; see services.SeedCosoCatalogue.
func seedCosoCatalogue(db *gorm.DB) error {
	return services.SeedCosoCatalogue(db)
}

// seedCosoPofCatalogue runs in `risk up` after seedCosoCatalogue. Insert-if-
// missing, validated before any write; see services.SeedCosoPofCatalogue.
func seedCosoPofCatalogue(db *gorm.DB) error {
	return services.SeedCosoPofCatalogue(db)
}
