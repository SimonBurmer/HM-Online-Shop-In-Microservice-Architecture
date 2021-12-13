import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"gitlab.lrz.de/vss/semester/ob-21ws/ob-21ws/Code/microservices/api"
	"google.golang.org/grpc"
)

const (
	name = "World"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
	})
	redisVal := rdb.Get(context.TODO(), "greeter")
	if redisVal == nil {
		log.Fatal("service not registered")
	}
	address, err := redisVal.Result()
	if err != nil {
		log.Fatalf("error while trying to get the result %v", err)
	}

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := api.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &api.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
}
