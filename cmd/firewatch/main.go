package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/JeremyFenwick/firewatch/internal/budgetchat"
	"github.com/JeremyFenwick/firewatch/internal/insecuresocketslayer"
	"github.com/JeremyFenwick/firewatch/internal/jobcenter"
	"github.com/JeremyFenwick/firewatch/internal/linereversal"
	"github.com/JeremyFenwick/firewatch/internal/meanstoanend"
	"github.com/JeremyFenwick/firewatch/internal/mobinthemiddle"
	"github.com/JeremyFenwick/firewatch/internal/primetime"
	"github.com/JeremyFenwick/firewatch/internal/smoketest"
	"github.com/JeremyFenwick/firewatch/internal/speeddaemon"
	"github.com/JeremyFenwick/firewatch/internal/unusualdatabase"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe(":8080", nil)) // used for the pprof profiler
	}()
	go smoketest.Listen(5000)
	go primetime.Listen(5001)
	go meanstoanend.Listen(5002)
	go budgetchat.Listen(5003)
	go unusualdatabase.Listen(5004)
	go mobinthemiddle.Listen(5005, "chat.protohackers.com", 16963)
	go speeddaemon.Listen(5006)
	go linereversal.Listen(5007)
	go insecuresocketslayer.Listen(5008)
	go jobcenter.Listen(5009)
	select {}
}
