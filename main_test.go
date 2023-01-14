package main

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"megaCrawler/Crawler"
	"megaCrawler/Crawler/Tester"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestTester(t *testing.T) {
	buf, err := os.Create("table.txt")
	if err != nil {
		t.Error(err)
		return
	}

	Crawler.Test.WG.Add(1)
	target := os.Getenv("TARGET")
	if target == "" {
		_, _ = buf.WriteString("No target specified.\nFailed to run tests.\n")
		return
	}
	targets := strings.Split(target, ",")
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./log/debug.jsonl",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})
	var fileCore zapcore.Core
	ProductionEncoder := zap.NewProductionEncoderConfig()

	fileCore = zapcore.NewCore(
		zapcore.NewJSONEncoder(ProductionEncoder),
		w,
		zap.DebugLevel,
	)

	logger := zap.New(fileCore)

	Crawler.Sugar = logger.Sugar()

	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"Target", "Field", "Total", "Passed", "Coverage"})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)

	for _, target := range targets {
		Crawler.Test = &Tester.Tester{
			WG: &sync.WaitGroup{},
			News: Tester.Status{
				Name: "News",
			},
			Index: Tester.Status{
				Name: "Index",
			},
			Expert: Tester.Status{
				Name: "Expert",
			},
			Report: Tester.Status{
				Name: "Report",
			},
		}

		c, ok := Crawler.WebMap[target]
		if !ok {
			_, _ = fmt.Fprintf(buf, "No such target %s.\n\n", target)
			continue
		}

		go Crawler.StartEngine(c, true)
		Crawler.Test.WG.Wait()

		Crawler.Test.News.FillTable(table)
		Crawler.Test.Index.FillTable(table)
		Crawler.Test.Expert.FillTable(table)
		Crawler.Test.Report.FillTable(table)
	}

	table.Render()
}
