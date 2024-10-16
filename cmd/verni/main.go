package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"verni/internal/apns"
	"verni/internal/auth/confirmation"
	"verni/internal/auth/jwt"
	"verni/internal/http-server/handlers/auth"
	"verni/internal/http-server/handlers/avatars"
	"verni/internal/http-server/handlers/friends"
	"verni/internal/http-server/handlers/profile"
	"verni/internal/http-server/handlers/spendings"
	"verni/internal/http-server/handlers/users"
	"verni/internal/http-server/longpoll"
	"verni/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	configFile, err := os.Open("users.json")
	if err != nil {
		log.Fatalf("failed to open config file: %s", err)
	}
	defer configFile.Close()
	configData, err := io.ReadAll(configFile)
	if err != nil {
		log.Fatalf("failed to read config file: %s", err)
	}
	type Module struct {
		Type   string                 `json:"type"`
		Config map[string]interface{} `json:"config"`
	}
	type Config struct {
		Storage     Module `json:"storage"`
		Apns        Module `json:"apns"`
		EmailSender Module `json:"emailSender"`
		Server      Module `json:"server"`
	}
	var config Config
	json.Unmarshal([]byte(configData), &config)
	log.Printf("initializing with config %v", config)

	db := func() storage.Storage {
		switch config.Storage.Type {
		case "yandex":
			type YDBConfig struct {
				Endpoint        string `json:"endpoint"`
				CredentialsPath string `json:"credentialsPath"`
			}
			data, err := json.Marshal(config.Storage.Config)
			if err != nil {
				log.Fatalf("failed to serialize ydb config err: %v", err)
			}
			var ydbConfig YDBConfig
			json.Unmarshal(data, &ydbConfig)
			log.Printf("creating ydb with config %v", ydbConfig)
			db, err := storage.YDB(ydbConfig.Endpoint, ydbConfig.CredentialsPath)
			if err != nil {
				log.Fatalf("failed to initialize ydb err: %v", err)
			}
			log.Printf("initialized ydb")
			return db
		default:
			log.Fatalf("unknown storage type %s", config.Storage.Type)
			return nil
		}
	}()
	defer db.Close()
	pushService := func() apns.Service {
		switch config.Apns.Type {
		case "apple":
			type AppleConfig struct {
				CertificatePath string `json:"certificatePath"`
				CredentialsPath string `json:"credentialsPath"`
			}
			data, err := json.Marshal(config.Apns.Config)
			if err != nil {
				log.Fatalf("failed to serialize apple apns config err: %v", err)
			}
			var appleConfig AppleConfig
			json.Unmarshal(data, &appleConfig)
			log.Printf("creating apple apns service with config %v", appleConfig)
			service, err := apns.AppleService(
				db,
				appleConfig.CertificatePath,
				appleConfig.CredentialsPath,
			)
			if err != nil {
				log.Fatalf("failed to initialize apple apns service err: %v", err)
			}
			log.Printf("initialized apple apns service")
			return service
		default:
			log.Fatalf("unknown apns type %s", config.Apns.Type)
			return nil
		}
	}()
	jwtService := jwt.DefaultService(
		time.Hour*24*30,
		time.Hour,
		func() time.Time {
			return time.Now()
		},
	)
	emailConfirmation := func() confirmation.Service {
		switch config.EmailSender.Type {
		case "yandex":
			type YandexConfig struct {
				Address  string `json:"address"`
				Password string `json:"password"`
				Host     string `json:"host"`
				Port     string `json:"port"`
			}
			data, err := json.Marshal(config.EmailSender.Config)
			if err != nil {
				log.Fatalf("failed to serialize yandex email sender config err: %v", err)
			}
			var yandexConfig YandexConfig
			json.Unmarshal(data, &yandexConfig)
			log.Printf("creating yandex email sender with config %v", yandexConfig)
			return confirmation.YandexService(
				db,
				yandexConfig.Address,
				yandexConfig.Password,
				yandexConfig.Host,
				yandexConfig.Port,
			)
		default:
			log.Fatalf("unknown email sender type %s", config.EmailSender.Type)
			return nil
		}
	}()
	server := func() http.Server {
		switch config.Server.Type {
		case "gin":
			type GinConfig struct {
				TimeoutSec     int    `json:"timeoutSec"`
				IdleTimeoutSec int    `json:"idleTimeoutSec"`
				RunMode        string `json:"runMode"`
				Port           string `json:"port"`
			}
			data, err := json.Marshal(config.Apns.Config)
			if err != nil {
				log.Fatalf("failed to serialize default server config err: %v", err)
			}
			var ginConfig GinConfig
			json.Unmarshal(data, &ginConfig)
			log.Printf("creating gin server with config %v", ginConfig)
			gin.SetMode(ginConfig.RunMode)
			router := gin.New()
			longpollService := longpoll.DefaultService(router, db, jwtService)

			longpollService.RegisterRoutes()

			auth.RegisterRoutes(router, db, jwtService, emailConfirmation)
			profile.RegisterRoutes(router, db, jwtService)
			avatars.RegisterRoutes(router, db)
			users.RegisterRoutes(router, db, jwtService)
			spendings.RegisterRoutes(router, db, jwtService, pushService, longpollService)
			friends.RegisterRoutes(router, db, jwtService, pushService, longpollService)

			address := ":" + ginConfig.Port
			return http.Server{
				Addr:         address,
				Handler:      router,
				ReadTimeout:  time.Duration(ginConfig.IdleTimeoutSec),
				WriteTimeout: time.Duration(ginConfig.IdleTimeoutSec),
			}
		default:
			log.Fatalf("unknown server type %s", config.Apns.Type)
			return http.Server{}
		}
	}()
	log.Printf("[info] start http server listening %s", server.Addr)
	server.ListenAndServe()
}

// type MigrationSpendingItem struct {
// 	Date        string  `json:"Date"`
// 	Description string  `json:"Description"`
// 	Category    string  `json:"Category"`
// 	Cost        float32 `json:"Cost"`
// 	Currency    string  `json:"Currency"`
// 	Margo       float32 `json:"margo"`
// 	Rzmn        float32 `json:"rzmn"`
// }

// func migrate(db storage.Storage) {
// 	// jsonFile, err := os.Open("./data/migration.json")
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// 	return
// 	// }
// 	// fmt.Println("Successfully Opened users.json")
// 	// defer jsonFile.Close()
// 	// byteValue, _ := io.ReadAll(jsonFile)
// 	// var items []MigrationSpendingItem
// 	// json.Unmarshal(byteValue, &items)
// 	// for i := 0; i < len(items); i++ {
// 	// 	format := "2006-01-02"
// 	// 	t, err := time.Parse(format, items[i].Date)
// 	// 	if err != nil {
// 	// 		fmt.Printf("time parse failed %v\n", err)
// 	// 		return
// 	// 	}
// 	// 	fmt.Printf("%s, %v\n", items[i].Date, t)

// 	// 	db.InsertDeal(storage.Deal{
// 	// 		Timestamp: t.Unix(),
// 	// 		Details:   items[i].Description,
// 	// 		Cost:      int(items[i].Cost * 100),
// 	// 		Currency:  items[i].Currency,
// 	// 		Spendings: []storage.Spending{
// 	// 			{
// 	// 				UserId: "margo",
// 	// 				Cost:   int(items[i].Margo * 100),
// 	// 			},
// 	// 			{
// 	// 				UserId: "rzmn",
// 	// 				Cost:   int(items[i].Rzmn * 100),
// 	// 			},
// 	// 		},
// 	// 	})
// 	// }
// 	counterpartiesMargo, err := db.GetCounterparties("margo")
// 	if err != nil {
// 		fmt.Printf("counterparties margo err: %v\n", err)
// 	} else {
// 		fmt.Printf("counterparties margo: %v\n", counterpartiesMargo)
// 	}
// 	counterpartiesRzmn, err := db.GetCounterparties("rzmn")
// 	if err != nil {
// 		fmt.Printf("counterparties margo err: %v\n", err)
// 	} else {
// 		fmt.Printf("counterparties margo: %v\n", counterpartiesRzmn)
// 	}
// 	deals, err := db.GetDeals("margo", "rzmn")
// 	if err != nil {
// 		fmt.Printf("deals err: %v\n", err)
// 	} else {
// 		for i := 0; i < len(deals); i++ {
// 			fmt.Printf("deal %d: %v\n", i, deals[i])
// 		}
// 	}
// }
