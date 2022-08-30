package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

type Promotion struct {
	id             string
	price          float64
	expirationDate string
}

func (a *App) Initialize(user, password, dbname string) {

	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	resetTable(a.DB)
	f, _ := os.Open("promotions.csv")
	promotions := basicRead(f)

	// both functions take about the same time ~2secs
	//bulkImport(promotions, a.DB)
	unnestInsert(promotions, a.DB)

	a.Router = mux.NewRouter()

	a.initializeRoutes()
}

func basicRead(f *os.File) []Promotion {
	fcsv := csv.NewReader(f)

	var promotions []Promotion

	for {
		rStr, err := fcsv.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("ERROR: ", err.Error())
			break
		}
		price, _ := strconv.ParseFloat(rStr[1], 64)
		p := Promotion{
			rStr[0], price, rStr[2],
		}
		promotions = append(promotions, p)
	}
	return promotions
}

func resetTable(db *sql.DB) {
	createTableQueryPath := "create_table.sql"
	executeQuery(db, createTableQueryPath)
	truncateTableQueryPath := "truncate_table.sql"
	executeQuery(db, truncateTableQueryPath)

	/*log.Println("Starting import")
	q := "COPY promotions FROM 'D:\\Rob\\Resume\\verveGroup\\main\\promotions.csv' DELIMITER ',' CSV"
	_, err := db.Exec(q)
	if err != nil {
		return
	}
	log.Println("Done import")*/
}

func basicReadAll() []Promotion {
	filePath := "promotions.csv"
	open, err2 := os.Open(filePath)
	f, err := open, err2
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}
	promotions := make([]Promotion, len(records))
	for i, r := range records {
		price, _ := strconv.ParseFloat(r[1], 64)
		p := Promotion{
			r[0], price, r[2],
		}
		promotions[i] = p
	}

	return promotions
}

func bulkImport(promotions []Promotion, db *sql.DB) {
	txn, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := txn.Prepare(pq.CopyIn("promotions", "id", "price", "expiration_date"))
	if err != nil {
		log.Fatal(err)
	}

	for _, promotion := range promotions {
		_, err = stmt.Exec(promotion.id, promotion.price, promotion.expirationDate)
		if err != nil {
			log.Fatal(err)
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Fatal(err)
	}

	err = stmt.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = txn.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func unnestInsert(unsavedRows []Promotion, db *sql.DB) {
	var ids []string
	var prices []float64
	var expirationDates []string
	for _, v := range unsavedRows {
		ids = append(ids, v.id)
		prices = append(prices, v.price)
		expirationDates = append(expirationDates, v.expirationDate)
	}
	query := `INSERT INTO promotions
    (id, price, expiration_date)
    (
      select * from unnest($1::text[], $2::numeric(9,6)[], $3::text[])
    )`

	if _, err := db.Exec(query, pq.Array(ids), pq.Array(prices), pq.Array(expirationDates)); err != nil {
		fmt.Println(err)
	}
	//err := db.Close()
	//if err != nil {
	//	log.Fatal(err)
	//}
}

func executeQuery(db *sql.DB, createTableQueryPath string) {
	path := filepath.Join(createTableQueryPath)

	c, ioErr := os.ReadFile(path)
	if ioErr != nil {
		// handle error.
	}
	query := string(c)
	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("failed to initialize the DB: %s", err)
	}
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) getPromotion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	p := promotionModel{ID: id}
	if err := p.getPromotion(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Promotion not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

func (a *App) createPromotion(w http.ResponseWriter, r *http.Request) {
	var p promotionModel
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := p.createPromotion(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, p)
}

func (a *App) updatePromotion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var p promotionModel
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	p.ID = id

	if err := p.updatePromotion(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

func (a *App) deletePromotion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	p := promotionModel{ID: id}
	if err := p.deletePromotion(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) getPromotions(w http.ResponseWriter, _ *http.Request) {
	promotions, err := getPromotions(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, promotions)
}

func (a *App) helloWorld(writer http.ResponseWriter, request *http.Request) {
	respondWithError(writer, http.StatusOK, "Hello World")
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/", a.helloWorld).Methods("GET")
	a.Router.HandleFunc("/promotions", a.getPromotions).Methods("GET")
	a.Router.HandleFunc("/promotions", a.createPromotion).Methods("POST")
	a.Router.HandleFunc("/promotions/{id}", a.getPromotion).Methods("GET")
	a.Router.HandleFunc("/promotions/{id}", a.updatePromotion).Methods("PUT")
	a.Router.HandleFunc("/promotions/{id}", a.deletePromotion).Methods("DELETE")
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
