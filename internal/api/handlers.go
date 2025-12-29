package api

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"project_sem/internal/db"
	"project_sem/internal/models"
)

type Router struct {
	DB *db.PG
}

func NewRouter(pg *db.PG) *http.ServeMux {
	r := &Router{DB: pg}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v0/prices", r.HandlePrices)
	return mux
}

func (api *Router) HandlePrices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		api.uploadPrices(w, r)
	case http.MethodGet:
		api.getPrices(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (api *Router) uploadPrices(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	body, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Read error", http.StatusInternalServerError)
		return
	}

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		http.Error(w, "Bad zip", http.StatusBadRequest)
		return
	}

	var csvFile *zip.File
	for _, f := range zipReader.File {
		if f.Name == "data.csv" || (len(f.Name) > 4 && f.Name[len(f.Name)-4:] == ".csv") {
			csvFile = f
			break
		}
	}

	if csvFile == nil {
		http.Error(w, "data.csv not found", http.StatusBadRequest)
		return
	}

	rc, err := csvFile.Open()
	if err != nil {
		http.Error(w, "Open csv error", http.StatusInternalServerError)
		return
	}
	defer rc.Close()

	reader := csv.NewReader(rc)
	if _, err := reader.Read(); err != nil {
		api.sendJSON(w, 0, 0, 0)
		return
	}

	var prices []models.Price
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "CSV parse error", http.StatusBadRequest)
			return
		}

		if len(record) < 5 {
			continue
		}

		id, _ := strconv.ParseInt(record[0], 10, 64)
		priceVal, _ := strconv.ParseFloat(record[3], 64)
		
		cDate, _ := time.Parse("2006-01-02", record[4])
		if cDate.IsZero() {
			cDate = time.Now()
		}

		prices = append(prices, models.Price{
			ID:        id,
			Name:      record[1],
			Category:  record[2],
			Price:     priceVal,
			CreatedAt: cDate,
		})
	}

	insertedCount, totalCat, totalPrice, err := db.InsertPrices(api.DB, prices)
	if err != nil {
		http.Error(w, fmt.Sprintf("DB Insert error: %v", err), http.StatusInternalServerError)
		return
	}

	api.sendJSON(w, insertedCount, totalCat, totalPrice)
}

func (api *Router) sendJSON(w http.ResponseWriter, items, cats int, price float64) {
	resp := map[string]interface{}{
		"total_items":      items,
		"total_categories": cats,
		"total_price":      price,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (api *Router) getPrices(w http.ResponseWriter, r *http.Request) {
	items, err := db.GetAllPrices(api.DB)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	f, err := zw.Create("data.csv")
	if err != nil {
		http.Error(w, "Zip create error", http.StatusInternalServerError)
		return
	}

	cw := csv.NewWriter(f)
	cw.Write([]string{"id", "name", "category", "price", "create_date"})

	for _, item := range items {
		cw.Write([]string{
			fmt.Sprintf("%d", item.ID),
			item.Name,
			item.Category,
			fmt.Sprintf("%.0f", item.Price),
			item.CreatedAt.Format("2006-01-02"),
		})
	}
	cw.Flush()
	zw.Close()

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"data.zip\"")
	w.Write(buf.Bytes())
}