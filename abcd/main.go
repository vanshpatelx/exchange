// package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"github.com/go-redis/redis/v8"
// )

// type Message struct{
// 	ID string `json:"id"`
// 	Content string `json:"content"`
// 	Timestamp string `json:"timestamp"`
// }

// func simulatePubSub(messages chan<- string){
// 	dummyMessages := []string{
// 		`{"id": "1", "content": "Hello, World!", "timestamp": "2024-12-31T10:00:00Z"}`,
//         `{"id": "2", "content": "Go is awesome!", "timestamp": "2024-12-31T10:01:00Z"}`,
//         `{"id": "3", "content": "Pub-Sub simulation.", "timestamp": "2024-12-31T10:02:00Z"}`,
// 	}

// 	for _,msg := range dummyMessages{
// 		messages <- msg
// 		time.Sleep(1 * time.Second)
// 	}
// 	close(messages)
// }

// func main() {
// 	messages := make(chan string)

// 	go simulatePubSub(messages)

// 	for msg := range messages{
// 		fmt.Println(msg)
// 	}

// }


// type ABCD struct {
// 	Key   string
// 	Value string
// }


// func ConnectCache() *redis.Client{
// 	return redis.NewClient(&redis.Options{
// 		Addr: "localhost:6379",
// 	})
// }

// func CreateABCDObjects(ctx context.Context, client *redis.Client) ([]ABCD, error){
// 	keys, err := client.Keys(ctx, "*").Result()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch keys: %v", err)
// 	}

// 	var objects []ABCD

// 	for _,key := range keys{
// 		value,err := client.Get(ctx, key).Result()
// 		if err != nil {
// 			log.Printf("failed to fetch value for key %s: %v", key, err)
// 			continue
// 		}

// 		abcd:= ABCD{
// 			Key:   key,
// 			Value: value,
// 		}
// 		objects = append(objects, abcd)
// 	}
// 	return objects, nil
// }

// func main() {
// 	ctx := context.Background()

// 	client := ConnectCache()


// 	_, err := client.Ping(ctx).Result()
// 	if err != nil{
// 		log.Fatalf("failed to connect to cache: %v", err)
// 	}

// 	objects, err := CreateABCDObjects(ctx, client)
// 	if err != nil {
// 		log.Fatalf("error while creating ABCD objects: %v", err)
// 	}

// 	for _, obj := range objects {
// 		fmt.Printf("ABCD Object - Key: %s, Value: %s\n", obj.Key, obj.Value)
// 	}

// }


// // Let's Design how our system look like
// // Exchange Class => StockManager[], addStock, removeStock
// // StockManager Class => buyQueue, sellQueue, ownMatching Engine, settlements
// // Main function => Which are doing orchestators, every connection done by here
// // Subscriber Class => Which go lots of events, also filterout, and calling Exchange class based on event
// // connectitvty class => Redis, Publisher also conncted to Main


package main

import (
	"context"
	"fmt"
	"log"
	"project/pkg/cache"

)

func main() {
	redisURL := "localhost:6379"
	cacheInstance := cache.NewCache(redisURL)

	ctx := context.Background()

	userId := "use123"
	balance := 100.50

	err := cacheInstance.SetBalance(ctx, userId, balance)

	if err != nil {
		log.Fatalf("Error setting balance: %v", err)
	}

	balance, err = cacheInstance.GetBalance(ctx, userId)
	if err != nil {
		log.Fatalf("Error getting balance: %v", err)
	}

	fmt.Printf("The balance for user %s is: %.2f\n", userId, balance)
}
