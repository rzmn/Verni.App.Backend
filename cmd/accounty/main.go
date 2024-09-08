package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"accounty/internal/apns"
	"accounty/internal/config"
	"accounty/internal/http-server/router/aasa"
	"accounty/internal/http-server/router/auth"
	"accounty/internal/http-server/router/avatars"
	"accounty/internal/http-server/router/friends"
	"accounty/internal/http-server/router/profile"
	"accounty/internal/http-server/router/spendings"
	"accounty/internal/http-server/router/users"
	"accounty/internal/storage"
	"accounty/internal/storage/ydbStorage"

	"github.com/gin-gonic/gin"
)

func main() {
	pushSender, err := apns.New("./internal/apns/apns_prod.p12", "./internal/apns/key.json")
	if err != nil {
		return
	}
	cfg := config.Load()

	storage, err := ydbStorage.New(os.Getenv("YDB_ENDPOINT"), "./internal/storage/ydbStorage/key.json")
	if err != nil {
		log.Fatalf("failed to init storage: %s", err)
	}
	defer storage.Close()
	// migrate(sqlStorage)
	// return

	gin.SetMode(cfg.Server.RunMode)
	router := gin.New()
	auth.RegisterRoutes(router, storage)
	users.RegisterRoutes(router, storage)
	friends.RegisterRoutes(router, storage, pushSender)
	spendings.RegisterRoutes(router, storage, pushSender)
	profile.RegisterRoutes(router, storage)
	aasa.RegisterRoutes(router, storage)
	avatars.RegisterRoutes(router, storage)

	address := ":" + os.Getenv("PORT")
	server := &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  cfg.Server.IdleTimeout,
		WriteTimeout: cfg.Server.IdleTimeout,
	}
	log.Printf("[info] start http server listening %s", address)

	server.ListenAndServe()
}

type MigrationSpendingItem struct {
	Date        string  `json:"Date"`
	Description string  `json:"Description"`
	Category    string  `json:"Category"`
	Cost        float32 `json:"Cost"`
	Currency    string  `json:"Currency"`
	Margo       float32 `json:"margo"`
	Rzmn        float32 `json:"rzmn"`
}

func migrate(db storage.Storage) {
	// jsonFile, err := os.Open("./data/migration.json")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println("Successfully Opened users.json")
	// defer jsonFile.Close()
	// byteValue, _ := io.ReadAll(jsonFile)
	// var items []MigrationSpendingItem
	// json.Unmarshal(byteValue, &items)
	// for i := 0; i < len(items); i++ {
	// 	format := "2006-01-02"
	// 	t, err := time.Parse(format, items[i].Date)
	// 	if err != nil {
	// 		fmt.Printf("time parse failed %v\n", err)
	// 		return
	// 	}
	// 	fmt.Printf("%s, %v\n", items[i].Date, t)

	// 	db.InsertDeal(storage.Deal{
	// 		Timestamp: t.Unix(),
	// 		Details:   items[i].Description,
	// 		Cost:      int(items[i].Cost * 100),
	// 		Currency:  items[i].Currency,
	// 		Spendings: []storage.Spending{
	// 			{
	// 				UserId: "margo",
	// 				Cost:   int(items[i].Margo * 100),
	// 			},
	// 			{
	// 				UserId: "rzmn",
	// 				Cost:   int(items[i].Rzmn * 100),
	// 			},
	// 		},
	// 	})
	// }
	counterpartiesMargo, err := db.GetCounterparties("margo")
	if err != nil {
		fmt.Printf("counterparties margo err: %v\n", err)
	} else {
		fmt.Printf("counterparties margo: %v\n", counterpartiesMargo)
	}
	counterpartiesRzmn, err := db.GetCounterparties("rzmn")
	if err != nil {
		fmt.Printf("counterparties margo err: %v\n", err)
	} else {
		fmt.Printf("counterparties margo: %v\n", counterpartiesRzmn)
	}
	deals, err := db.GetDeals("margo", "rzmn")
	if err != nil {
		fmt.Printf("deals err: %v\n", err)
	} else {
		for i := 0; i < len(deals); i++ {
			fmt.Printf("deal %d: %v\n", i, deals[i])
		}
	}
}
