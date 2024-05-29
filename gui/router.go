package gui

import (
	"net/http"
)

func Router() *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("/", listPage)
	router.HandleFunc("/part/", partPage)
	router.HandleFunc("/part/update-status", updatePartStatusHandler)
	router.HandleFunc("/part/generate-qr", generateQRHandler)
	router.HandleFunc("/part/export-to-excel", exportToExcelHandler)
	return router
}

func updatePartStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Handler implementation
}

func generateQRHandler(w http.ResponseWriter, r *http.Request) {
	// Handler implementation
}

func exportToExcelHandler(w http.ResponseWriter, r *http.Request) {
	// Handler implementation
}
