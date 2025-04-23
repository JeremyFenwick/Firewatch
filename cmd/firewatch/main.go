package main

import (
	"github.com/JeremyFenwick/firewatch/internal/budgetchat"
	"github.com/JeremyFenwick/firewatch/internal/linereversal"
	"github.com/JeremyFenwick/firewatch/internal/meanstoanend"
	"github.com/JeremyFenwick/firewatch/internal/mobinthemiddle"
	"github.com/JeremyFenwick/firewatch/internal/primetime"
	"github.com/JeremyFenwick/firewatch/internal/smoketest"
	"github.com/JeremyFenwick/firewatch/internal/speeddaemon"
	"github.com/JeremyFenwick/firewatch/internal/unusualdatabase"
)

func main() {
	go smoketest.Listen(5000)
	go primetime.Listen(5001)
	go meanstoanend.Listen(5002)
	go budgetchat.Listen(5003)
	go unusualdatabase.Listen(5004)
	go mobinthemiddle.Listen(5005, "chat.protohackers.com", 16963)
	go speeddaemon.Listen(5006)
	go linereversal.Listen(5007)
	select {}
}
