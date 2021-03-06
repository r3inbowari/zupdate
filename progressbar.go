package zupdate

// @author lycblank
// @change r3inbowari

import (
	"fmt"
	"github.com/fatih/color"
)

type PBOptions struct {
	graph string
}

type PBOption func(opts *PBOptions)

type ProgressBar struct {
	totalValue int64
	currValue  int64
	graph      string
	rate       string
}

func NewProgressBar(totalValue int64, options ...PBOption) *ProgressBar {
	opts := PBOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	if opts.graph == "" {
		opts.graph = "█"
	}
	bar := &ProgressBar{
		totalValue: totalValue,
		graph:      opts.graph,
	}
	return bar
}

func (bar *ProgressBar) Play(value int64) {
	val := float64(bar.totalValue) / 50
	prePercent := int32(float64(bar.currValue) / val)
	nowPercent := int32(float64(value) / val)
	for i := prePercent + 1; i <= nowPercent; i++ {
		bar.rate += bar.graph
	}
	bar.currValue = value

	_, _ = color.New(color.FgGreen).Printf(
		"\r[I] [%-50s]%0.2f%%   	%8d/%d",
		bar.rate,
		float64(bar.currValue)/float64(bar.totalValue)*100,
		bar.currValue,
		bar.totalValue)
}

func (bar *ProgressBar) Finish() {
	val := float64(bar.totalValue) / 50
	prePercent := int32(float64(bar.currValue) / val)
	for i := prePercent + 1; i <= 50; i++ {
		bar.rate += bar.graph
	}
	bar.currValue = bar.totalValue
	_, _ = color.New(color.FgGreen).Printf(
		"\r[I] [%-50s]%0.2f%%   	%8d/%d\n",
		bar.rate,
		float64(bar.currValue)/float64(bar.totalValue)*100,
		bar.currValue,
		bar.totalValue)
}

func (bar *ProgressBar) Stop() {
	fmt.Println()
}
