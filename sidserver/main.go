package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net"

	// config loading
	"encoding/json"

	// mysql backend
	"database/sql"
	_ "github.com/go-sql-driver/mysql"

	"github.com/c9s/sid"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

//go:generate protoc -I./ --go_out=plugins=grpc:. ./sid.proto

const (
	bind = ":51051"
)

const (
	ErrCodeSequenceRequired = iota
	ErrCodeSequenceUndefined
	ErrCodeStatementExecuteFailed
	ErrCodeFailedToGetLastInsertId
)

// server is used to implement helloworld.GreeterServer.
type SIDServer struct {
	config Config
	db     *sql.DB
}

// SayHello implements helloworld.GreeterServer
func (s *SIDServer) Generate(ctx context.Context, in *sid.SIDRequest) (*sid.SIDReply, error) {
	if len(in.Sequence) == 0 {
		return &sid.SIDReply{Code: ErrCodeSequenceRequired}, nil
	}

	seq, ok := s.config.Sequences[in.Sequence]
	if !ok {
		return &sid.SIDReply{Code: ErrCodeSequenceUndefined}, nil
	}

	res, err := seq.Stmt.Exec(in.Oid)
	if err != nil {
		return &sid.SIDReply{Code: ErrCodeStatementExecuteFailed}, nil
	}

	id, err := res.LastInsertId()
	if err != nil {
		return &sid.SIDReply{Code: ErrCodeFailedToGetLastInsertId}, nil
	}

	log.Printf("generated ID %d with %v", id, in.Oid)
	return &sid.SIDReply{Code: 0, Id: id, Oid: in.Oid}, nil
}

type MySQLBackendConfig struct {
	DSN string `json:"dsn"`
}

type BackendConfig struct {
	MySQL MySQLBackendConfig `json:"mysql"`
}

type SequenceConfig struct {
	Stmt *sql.Stmt
}

type Config struct {
	Backend   BackendConfig              `json:"backend"`
	Sequences map[string]*SequenceConfig `json:"sequences"`
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.json", "the path of the configuration file.")
	flag.Parse()

	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	var config Config
	if err := json.Unmarshal(configBytes, &config); err != nil {
		log.Fatal(err.Error())
	}

	log.Println("preparing mysql connection")
	// db, err := sql.Open("mysql", "user:password@/dbname")
	db, err := sql.Open("mysql", config.Backend.MySQL.DSN)
	if err != nil {
		log.Fatal(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	for sequence, sconf := range config.Sequences {
		log.Printf("checking sequence %s...", sequence)

		// oid requires 24 bytes
		q := `CREATE TABLE IF NOT EXISTS ` + sequence + ` (
				id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
				oid BINARY(14)
			)`
		if _, err = db.Query(q); err != nil {
			log.Fatal(err.Error())
		}

		stmt, err := db.Prepare(`INSERT INTO ` + sequence + ` (oid) VALUES (?)`)
		if err != nil {
			log.Fatal(err.Error())
		}
		sconf.Stmt = stmt
	}

	defer func() {
		for _, sconf := range config.Sequences {
			sconf.Stmt.Close()
		}
	}()

	c, err := net.Listen("tcp", bind)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	sid.RegisterSIDGeneratorServer(s, &SIDServer{config, db})

	// Register reflection service on gRPC server.
	reflection.Register(s)

	log.Printf("listening at %s", bind)
	if err := s.Serve(c); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
