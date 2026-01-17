package plot

import (
	"fmt"
	"image/color"

	"github.com/kasaderos/rLportfolio/pkg/agent"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// SavePlots writes price/actions and cash series using a plotting library.
func SavePlots(prices []float64, cashSeries []float64, actions []int) error {
	if len(prices) == 0 || len(actions) != len(prices) || len(cashSeries) != len(prices) {
		return fmt.Errorf("invalid input sizes for plot")
	}

	pricePlot := plot.New()
	pricePlot.Title.Text = "Price with actions"
	pricePlot.X.Label.Text = "t"
	pricePlot.Y.Label.Text = "price"

	priceXYs := make(plotter.XYs, len(prices))
	for i, v := range prices {
		priceXYs[i].X = float64(i)
		priceXYs[i].Y = v
	}

	priceLine, err := plotter.NewLine(priceXYs)
	if err != nil {
		return err
	}
	priceLine.Color = color.RGBA{R: 31, G: 119, B: 180, A: 255}
	pricePlot.Add(priceLine)

	for _, action := range []int{
		int(agent.ActionBuySmall),
		int(agent.ActionBuyLarge),
		int(agent.ActionSellSmall),
		int(agent.ActionSellLarge),
	} {
		points := make(plotter.XYs, 0)
		for i, a := range actions {
			if a != action {
				continue
			}
			points = append(points, plotter.XY{X: float64(i), Y: prices[i]})
		}
		if len(points) == 0 {
			continue
		}
		scatter, err := plotter.NewScatter(points)
		if err != nil {
			return err
		}
		scatter.GlyphStyle.Radius = vg.Points(3)
		scatter.GlyphStyle.Color = actionColorRGBA(action)
		pricePlot.Add(scatter)
	}

	if err := pricePlot.Save(12*vg.Inch, 4*vg.Inch, "policy_actions.png"); err != nil {
		return err
	}

	cashPlot := plot.New()
	cashPlot.Title.Text = "Cash series"
	cashPlot.X.Label.Text = "t"
	cashPlot.Y.Label.Text = "cash"

	cashXYs := make(plotter.XYs, len(cashSeries))
	for i, v := range cashSeries {
		cashXYs[i].X = float64(i)
		cashXYs[i].Y = v
	}
	cashLine, err := plotter.NewLine(cashXYs)
	if err != nil {
		return err
	}
	cashLine.Color = color.RGBA{R: 255, G: 127, B: 14, A: 255}
	cashPlot.Add(cashLine)

	return cashPlot.Save(12*vg.Inch, 4*vg.Inch, "cash_series.png")
}

// actionColorRGBA returns the color for a given action.
func actionColorRGBA(action int) color.RGBA {
	switch action {
	case int(agent.ActionBuySmall):
		return color.RGBA{R: 141, G: 223, B: 141, A: 255}
	case int(agent.ActionBuyLarge):
		return color.RGBA{R: 15, G: 125, B: 15, A: 255}
	case int(agent.ActionSellSmall):
		return color.RGBA{R: 255, G: 153, B: 153, A: 255}
	case int(agent.ActionSellLarge):
		return color.RGBA{R: 177, G: 18, B: 38, A: 255}
	default:
		return color.RGBA{R: 170, G: 170, B: 170, A: 255}
	}
}
