package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"verni/internal/apns"
	"verni/internal/auth/jwt"
	"verni/internal/config"
	"verni/internal/http-server/handlers/auth"
	"verni/internal/http-server/handlers/avatars"
	"verni/internal/http-server/handlers/friends"
	"verni/internal/http-server/handlers/profile"
	"verni/internal/http-server/handlers/spendings"
	"verni/internal/http-server/handlers/users"
	"verni/internal/storage"
	"verni/internal/storage/ydbStorage"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := ydbStorage.New(os.Getenv("YDB_ENDPOINT"), "./internal/storage/ydbStorage/key.json")
	if err != nil {
		log.Fatalf("failed to init storage: %s", err)
	}
	defer db.Close()
	// migrate(sqlStorage)
	// return

	pushService, err := apns.DefaultService(db, "./internal/apns/apns_prod.p12", "./internal/apns/key.json")

	jwtService := jwt.DefaultService(
		time.Hour*24*30,
		time.Hour,
		func() time.Time {
			return time.Now()
		},
	)

	gin.SetMode(cfg.Server.RunMode)
	router := gin.New()

	auth.RegisterRoutes(router, db, jwtService)
	profile.RegisterRoutes(router, db, jwtService)
	avatars.RegisterRoutes(router, db)
	users.RegisterRoutes(router, db, jwtService)
	spendings.RegisterRoutes(router, db, jwtService)
	friends.RegisterRoutes(router, db, jwtService)

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
