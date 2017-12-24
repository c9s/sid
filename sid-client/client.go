package main

import (
	"flag"
	"log"

	"golang.org/x/net/context"

	"github.com/c9s/sid"

	// "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"google.golang.org/grpc"
)

func main() {
	var addr string
	flag.StringVar(&addr, "connect", "localhost:51051", "the server address in the format host:port")
	flag.Parse()

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	oid := bson.NewObjectId()
	client := sid.NewSIDGeneratorClient(conn)
	reply, err := client.Generate(context.Background(), &sid.SIDRequest{Sequence: "jobs", Oid: []byte(oid.Hex())})
	if err != nil {
		log.Fatalf("%v.Generate(_) = _, %v: ", client, err)
	}
	log.Printf("Reply: %+v", reply)
}
