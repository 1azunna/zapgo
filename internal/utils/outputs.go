package utils

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

func SummaryOutput(data [][]string) {

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Alert Title", "Risk", "Confidence", "Instances", "Site", "Alert ID"})
	table.AppendBulk(data)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render() // Send output
}

func PrintAlerts(data [][]string) {

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Alert", "Risk", "Confidence", "Url", "Params"})
	table.AppendBulk(data)
	table.SetAutoMergeCellsByColumnIndex([]int{0, 1, 3})
	table.SetRowLine(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render() // Send output
}
