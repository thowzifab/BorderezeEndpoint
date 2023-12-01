package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/denisenkom/go-mssqldb"
)

// ScanStatistic represents the structure of the ScanStatistics table
type ScanStatistic struct {
    ScanID           int    `json:"scanID"`
    TotalScans       int    `json:"totalScans"`
    CBSAHolds        int    `json:"cbsaHolds"`
    OGDSHolds        int    `json:"ogdsHolds"`
    CurrentDate      string `json:"currentDate"` // Change the type if it's not a string
    ConveyorBeltName string `json:"conveyorBeltName"`
    TargetTotalScans int    `json:"targetTotalScans"`
    CurrentSpeed     int    `json:"currentSpeed"`
    CurrentStatus    string `json:"currentStatus"`
}

func fetchDataFromDB() ([]ScanStatistic, error) {
    // Connection string for Windows Authentication
    connString := "server=DESKTOP-OMNRLEK\\SQLEXPRESS;database=telnet_project;trusted_connection=yes;"

    // Open the database connection
    db, err := sql.Open("sqlserver", connString)
    if err != nil {
        return nil, err
    }
    defer db.Close()

    // SQL query
    rows, err := db.Query(`
        SELECT [ScanID], [TotalScans], [CBSAHolds], [OGDSHolds], 
                              [CurrentDate], [ConveyorBeltName], [target_total_scans], 
                              [current_speed], [current_status]
        FROM [dbo].[ScanStatistics]
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var results []ScanStatistic

    // Iterate through the rows
    for rows.Next() {
        var r ScanStatistic
        err := rows.Scan(
            &r.ScanID, &r.TotalScans, &r.CBSAHolds, &r.OGDSHolds,
            &r.CurrentDate, &r.ConveyorBeltName, &r.TargetTotalScans,
            &r.CurrentSpeed, &r.CurrentStatus,
        )
        if err != nil {
            return nil, err
        }
        results = append(results, r)
    }

    // Check for errors from iterating over rows
    if err = rows.Err(); err != nil {
        return nil, err
    }

    return results, nil
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
    data, err := fetchDataFromDB()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    jsonData, err := json.Marshal(data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonData)
}

// corsMiddleware adds CORS headers to the response
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Set CORS headers
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

        // If it's a preflight OPTIONS request, send a 200 response and return
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        // Otherwise, serve the next handler
        next.ServeHTTP(w, r)
    })
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/data", dataHandler)

    handler := corsMiddleware(mux)

    fmt.Println("Server is running on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", handler))
}
