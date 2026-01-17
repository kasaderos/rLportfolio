package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/kasaderos/rLportfolio/pkg/agent"
	ma "github.com/kasaderos/rLportfolio/pkg/moving-average"
	"github.com/kasaderos/rLportfolio/pkg/plot"
	"github.com/kasaderos/rLportfolio/pkg/state"
)

const (
	// Moving average parameters (must match training parameters)
	minStartIdx = 120 // Need at least 120 prices for MA120
)

func main() {
	// Load series data
	prices, portfolioSeries, actions, err := plot.LoadSeriesData()
	if err != nil {
		log.Fatalf("Failed to load series data: %v", err)
	}

	fmt.Printf("Loaded %d data points\n", len(prices))
	fmt.Printf("Actions: %d non-empty actions\n", countNonEmptyActions(actions))

	// Create HTML with interactive Plotly chart
	html := generateInteractivePlot(prices, portfolioSeries, actions)

	// Save HTML file
	htmlPath := "templates/plot.html"
	if err := os.MkdirAll(filepath.Dir(htmlPath), 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	if err := os.WriteFile(htmlPath, []byte(html), 0644); err != nil {
		log.Fatalf("Failed to write HTML file: %v", err)
	}

	fmt.Printf("Interactive plot saved to %s\n", htmlPath)
	fmt.Println("Opening in browser...")

	// Start a simple HTTP server to serve the HTML
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, htmlPath)
	})

	fmt.Println("Server running at http://localhost:8080")
	fmt.Println("Open http://localhost:8080 in your browser")
	fmt.Println("Press Ctrl+C to stop the server")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func countNonEmptyActions(actions []int) int {
	count := 0
	for _, a := range actions {
		if a >= 0 {
			count++
		}
	}
	return count
}

func generateInteractivePlot(prices []float64, portfolioSeries []float64, actions []int) string {
	// Prepare data for JavaScript
	pricesJS := formatFloatArray(prices)

	// Calculate moving averages
	maDataJS := calculateMAsForPlot(prices)

	// Prepare action markers with state information
	actionMarkers := prepareActionMarkers(prices, portfolioSeries, actions)

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>RL Portfolio Trading - Interactive Plot</title>
    <script src="https://cdn.plot.ly/plotly-latest.min.js"></script>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background-color: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            margin-bottom: 20px;
        }
        #plot {
            width: 100%%;
            height: 800px;
        }
        .info {
            margin-top: 20px;
            padding: 10px;
            background-color: #e8f4f8;
            border-radius: 4px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>RL Portfolio Trading - Interactive Plot</h1>
        <div id="plot"></div>
        <div class="info">
            <h3>Controls:</h3>
            <ul>
                <li><strong>Zoom:</strong> Click and drag to select a region, or use mouse wheel</li>
                <li><strong>Pan:</strong> Click and drag on the plot background</li>
                <li><strong>Reset:</strong> Double-click on the plot</li>
                <li><strong>Hover:</strong> Hover over points to see details</li>
            </ul>
            <h3>Legend:</h3>
            <ul>
                <li><span style="color: #1f77b4;">Blue line:</span> Price series</li>
                <li><span style="color: #ff7f0e;">Orange dashed:</span> MA5</li>
                <li><span style="color: #9467bd;">Purple dashed:</span> MA10</li>
                <li><span style="color: #8c564b;">Brown dashed:</span> MA20</li>
                <li><span style="color: #e377c2;">Pink dashed:</span> MA40</li>
                <li><span style="color: #7f7f7f;">Gray dashed:</span> MA80</li>
                <li><span style="color: #bcbd22;">Olive dashed:</span> MA120</li>
                <li><span style="color: #2ca02c;">Green markers:</span> Buy actions</li>
                <li><span style="color: #d62728;">Red markers:</span> Sell actions</li>
            </ul>
        </div>
    </div>

    <script>
        // Price data
        var prices = %s;
        var actionMarkers = %s;
        var maData = %s;
        
        var time = [];
        for (var i = 0; i < prices.length; i++) {
            time.push(i);
        }

        // Create price trace
        var priceTrace = {
            x: time,
            y: prices,
            type: 'scatter',
            mode: 'lines',
            name: 'Price',
            line: {
                color: '#1f77b4',
                width: 2
            },
            yaxis: 'y'
        };

        // Create MA traces
        var maTraces = [];
        var maColors = ['#ff7f0e', '#9467bd', '#8c564b', '#e377c2', '#7f7f7f', '#bcbd22'];
        var maPeriods = [5, 10, 20, 40, 80, 120];
        
        for (var i = 0; i < maPeriods.length; i++) {
            var period = maPeriods[i];
            var maValues = maData[period];
            var maTime = [];
            var maY = [];
            
            // MA arrays are shorter, need to offset x values
            var offset = period - 1;
            for (var j = 0; j < maValues.length; j++) {
                maTime.push(offset + j);
                maY.push(maValues[j]);
            }
            
            maTraces.push({
                x: maTime,
                y: maY,
                type: 'scatter',
                mode: 'lines',
                name: 'MA' + period,
                line: {
                    color: maColors[i],
                    width: 1.5,
                    dash: 'dash'
                },
                yaxis: 'y',
                hovertemplate: 'MA' + period + '<br>Time: %%{x}<br>Value: %%{y:.2f}<extra></extra>'
            });
        }

        // Create buy action markers
        var buyMarkers = {
            x: actionMarkers.buy.x,
            y: actionMarkers.buy.y,
            text: actionMarkers.buy.labels,
            type: 'scatter',
            mode: 'markers',
            name: 'Buy Actions',
            marker: {
                color: '#2ca02c',
                size: 8,
                symbol: 'triangle-up',
                line: {
                    color: '#1f7f1f',
                    width: 1
                }
            },
            yaxis: 'y',
            hovertemplate: '<b>%%{text}</b><br>Time: %%{x}<br>Price: %%{y:.2f}<br>State: %%{customdata}<extra></extra>',
            customdata: actionMarkers.buy.states
        };

        // Create sell action markers
        var sellMarkers = {
            x: actionMarkers.sell.x,
            y: actionMarkers.sell.y,
            text: actionMarkers.sell.labels,
            type: 'scatter',
            mode: 'markers',
            name: 'Sell Actions',
            marker: {
                color: '#d62728',
                size: 8,
                symbol: 'triangle-down',
                line: {
                    color: '#7f0f0f',
                    width: 1
                }
            },
            yaxis: 'y',
            hovertemplate: '<b>%%{text}</b><br>Time: %%{x}<br>Price: %%{y:.2f}<br>State: %%{customdata}<extra></extra>',
            customdata: actionMarkers.sell.states
        };

        var data = [priceTrace].concat(maTraces).concat([buyMarkers, sellMarkers]);

        var layout = {
            title: {
                text: 'RL Portfolio Trading - Price and Actions',
                font: {
                    size: 18
                }
            },
            xaxis: {
                title: 'Time',
                showgrid: true,
                gridcolor: '#e0e0e0'
            },
            yaxis: {
                title: 'Price',
                side: 'left',
                showgrid: true,
                gridcolor: '#e0e0e0'
            },
            hovermode: 'closest',
            legend: {
                x: 0,
                y: 1,
                bgcolor: 'rgba(255,255,255,0.8)'
            },
            plot_bgcolor: 'white',
            paper_bgcolor: 'white'
        };

        var config = {
            responsive: true,
            displayModeBar: true,
            modeBarButtonsToRemove: ['lasso2d', 'select2d'],
            displaylogo: false
        };

        Plotly.newPlot('plot', data, layout, config);
    </script>
</body>
</html>`, pricesJS, actionMarkers, maDataJS)
}

func formatFloatArray(arr []float64) string {
	if len(arr) == 0 {
		return "[]"
	}
	result := "["
	for i, v := range arr {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%.6f", v)
	}
	result += "]"
	return result
}

func prepareActionMarkers(prices []float64, portfolioSeries []float64, actions []int) string {
	var buyX []int
	var buyPrices []float64
	var buyLabels []string
	var buyStates []string
	var sellX []int
	var sellPrices []float64
	var sellLabels []string
	var sellStates []string

	for i, action := range actions {
		if i >= len(prices) || i >= len(portfolioSeries) {
			break
		}
		if action < 0 {
			continue
		}

		// Compute state for this point
		stateStr := computeStateString(prices, portfolioSeries, i)

		actionType := agent.Action(action)
		if actionType == agent.ActionBuySmall || actionType == agent.ActionBuyLarge {
			buyX = append(buyX, i)
			buyPrices = append(buyPrices, prices[i])
			buyLabels = append(buyLabels, actionType.String())
			buyStates = append(buyStates, stateStr)
		} else if actionType == agent.ActionSellSmall || actionType == agent.ActionSellLarge {
			sellX = append(sellX, i)
			sellPrices = append(sellPrices, prices[i])
			sellLabels = append(sellLabels, actionType.String())
			sellStates = append(sellStates, stateStr)
		}
	}

	// Format as JavaScript object
	buyXJS := formatIntArray(buyX)
	buyYJS := formatFloatArray(buyPrices)
	buyLabelsJS := formatStringArray(buyLabels)
	buyStatesJS := formatStringArray(buyStates)
	sellXJS := formatIntArray(sellX)
	sellYJS := formatFloatArray(sellPrices)
	sellLabelsJS := formatStringArray(sellLabels)
	sellStatesJS := formatStringArray(sellStates)

	return fmt.Sprintf(`{
        "buy": {
            "x": %s,
            "y": %s,
            "labels": %s,
            "states": %s
        },
        "sell": {
            "x": %s,
            "y": %s,
            "labels": %s,
            "states": %s
        }
    }`, buyXJS, buyYJS, buyLabelsJS, buyStatesJS, sellXJS, sellYJS, sellLabelsJS, sellStatesJS)
}

// computeStateString computes the state string for a given point in the series.
func computeStateString(prices []float64, portfolioSeries []float64, idx int) string {
	if idx < minStartIdx || idx >= len(prices) {
		return "N/A"
	}

	// Need at least 120 prices for all MAs to be available
	if idx < 120 {
		return "N/A"
	}

	// Get moving average ordering state
	maState := ma.GetMAStateForIndex(prices, idx)
	if maState < 0 {
		return "N/A"
	}

	// Get MA convergence/divergence state
	maDivergence := ma.GetMADivergenceState(prices, idx)
	divergenceStr := "N"
	switch maDivergence {
	case 0:
		divergenceStr = "C" // Converging
	case 1:
		divergenceStr = "N" // Neutral
	case 2:
		divergenceStr = "D" // Diverging
	}

	// Get portfolio position categories
	// portfolioSeries now contains portfolio value (cash + price * shares) directly
	portfolioValue := portfolioSeries[idx]

	// Estimate cash and shares from portfolio value
	// This is approximate but gives reasonable state categorization
	// We estimate that cash is roughly proportional to portfolio value
	// and shares value is the remainder
	initialPortfolioValue := portfolioSeries[0]
	if initialPortfolioValue <= 0 {
		initialPortfolioValue = 10000.0 // Default estimate
	}

	// Rough estimate: assume cash is a portion of portfolio value
	// This is an approximation since we don't have exact cash/shares breakdown
	estimatedCash := portfolioValue * 0.5 // Conservative estimate
	estimatedSharesValue := portfolioValue - estimatedCash
	if portfolioValue <= 0 {
		portfolioValue = 1.0 // Avoid division by zero
	}

	cashCat := state.GetCashCategory(estimatedCash, portfolioValue)
	sharesCat := state.GetSharesCategory(estimatedSharesValue, portfolioValue)

	return fmt.Sprintf("MA:%d %s C:%d S:%d",
		maState, divergenceStr, cashCat, sharesCat)
}

func formatIntArray(arr []int) string {
	if len(arr) == 0 {
		return "[]"
	}
	result := "["
	for i, v := range arr {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%d", v)
	}
	result += "]"
	return result
}

func formatStringArray(arr []string) string {
	if len(arr) == 0 {
		return "[]"
	}
	result := "["
	for i, v := range arr {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf(`"%s"`, v)
	}
	result += "]"
	return result
}

// calculateMAsForPlot calculates all moving averages for plotting.
func calculateMAsForPlot(prices []float64) string {
	mas := ma.CalculateAllMAs(prices)

	// Format as JavaScript object
	result := "{"
	first := true
	for _, period := range ma.MAPeriods {
		if !first {
			result += ","
		}
		first = false
		maValues := mas[period]
		result += fmt.Sprintf(`"%d":%s`, period, formatFloatArray(maValues))
	}
	result += "}"
	return result
}
