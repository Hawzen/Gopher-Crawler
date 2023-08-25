package main

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
)

func Db_setup() *dgo.Dgraph {
	// DB setup
	d, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	dg := dgo.NewDgraphClient(api.NewDgraphClient(d))

	// Drop all data
	err = dg.Alter(context.Background(), &api.Operation{DropAll: true})
	if err != nil {
		panic(err)
	}

	op := &api.Operation{}
	op.Schema = `
		url: string @index(exact) .
		domain: uid @reverse .
		title: string @index(exact) .
		related_pages: [uid] .
		is_crawled: bool @index(bool) .
		depth: int @index(int) .
		time_crawled: datetime @index(hour) .
		time_found: datetime @index(hour) .
		summary: string @index(fulltext) .
        keywords: [string] @index(fulltext) .
		name: string @index(exact) .
	`
	if err := dg.Alter(context.Background(), op); err != nil {
		log.Fatal(err)
	}

	return dg
}
