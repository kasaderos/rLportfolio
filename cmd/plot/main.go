package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/kasaderos/rLportfolio/pkg/agent"
	"github.com/kasaderos/rLportfolio/pkg/plot"
)

func main() {
	// Load series data
	prices, cashSeries, actions, err := plot.LoadSeriesData()
	if err != nil {
		log.Fatalf("Failed to load series data: %v", err)
	}

	fmt.Printf("Loaded %d data points\n", len(prices))
	fmt.Printf("Actions: %d non-empty actions\n", countNonEmptyActions(actions))

	// Create HTML with interactive Plotly chart
	html := generateInteractivePlot(prices, cashSeries, actions)

	// Save HTML file
	htmlPath := "rl/plot/plot.html"
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

func generateInteractivePlot(prices []float64, cashSeries []float64, actions []int) string {
	// Prepare data for JavaScript
	pricesJS := formatFloatArray(prices)
	cashJS := formatFloatArray(cashSeries)

	// Prepare action markers
	actionMarkers := prepareActionMarkers(prices, actions)

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
                <li><span style="color: #ff7f0e;">Orange line:</span> Cash series</li>
                <li><span style="color: #2ca02c;">Green markers:</span> Buy actions</li>
                <li><span style="color: #d62728;">Red markers:</span> Sell actions</li>
            </ul>
        </div>
    </div>

    <script>
        // Price data
        var prices = %s;
        var cashSeries = %s;
        var actionMarkers = %s;
        
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

        // Create cash trace
        var cashTrace = {
            x: time,
            y: cashSeries,
            type: 'scatter',
            mode: 'lines',
            name: 'Cash',
            line: {
                color: '#ff7f0e',
                width: 2
            },
            yaxis: 'y2'
        };

        // Create buy action markers
        var buyMarkers = {
            x: actionMarkers.buy.x,
            y: actionMarkers.buy.y,
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
            hovertemplate: '<b>Buy Action</b><br>Time: %%{x}<br>Price: %%{y:.2f}<extra></extra>'
        };

        // Create sell action markers
        var sellMarkers = {
            x: actionMarkers.sell.x,
            y: actionMarkers.sell.y,
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
            hovertemplate: '<b>Sell Action</b><br>Time: %%{x}<br>Price: %%{y:.2f}<extra></extra>'
        };

        var data = [priceTrace, cashTrace, buyMarkers, sellMarkers];

        var layout = {
            title: {
                text: 'RL Portfolio Trading - Price, Cash, and Actions',
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
            yaxis2: {
                title: 'Cash',
                side: 'right',
                overlaying: 'y',
                showgrid: false
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
</html>`, pricesJS, cashJS, actionMarkers)
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

func prepareActionMarkers(prices []float64, actions []int) string {
	var buyX []int
	var buyPrices []float64
	var sellX []int
	var sellPrices []float64

	for i, action := range actions {
		if i >= len(prices) {
			break
		}
		if action < 0 {
			continue
		}

		actionType := agent.Action(action)
		if actionType == agent.ActionBuySmall || actionType == agent.ActionBuyMedium || actionType == agent.ActionBuyLarge {
			buyX = append(buyX, i)
			buyPrices = append(buyPrices, prices[i])
		} else if actionType == agent.ActionSellSmall || actionType == agent.ActionSellMedium || actionType == agent.ActionSellLarge {
			sellX = append(sellX, i)
			sellPrices = append(sellPrices, prices[i])
		}
	}

	// Format as JavaScript object
	buyXJS := formatIntArray(buyX)
	buyYJS := formatFloatArray(buyPrices)
	sellXJS := formatIntArray(sellX)
	sellYJS := formatFloatArray(sellPrices)

	return fmt.Sprintf(`{
        "buy": {
            "x": %s,
            "y": %s
        },
        "sell": {
            "x": %s,
            "y": %s
        }
    }`, buyXJS, buyYJS, sellXJS, sellYJS)
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
