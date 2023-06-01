package pkg

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// ScreenerResult : base64 of the screenshot
type ScreenerResult struct {
	Path string `bson:"path" json:"path"`
}

// Screener : struct to take screenshots
type Screener struct{}

// NewScreenerResult : returns a new ScreenerResult struct
func NewScreenerResult(path string) *ScreenerResult {
	return &ScreenerResult{path}
}

// NewScreener : returns a new Screener struct
func NewScreener() *Screener {
	return &Screener{}
}

// Run : takes a screenshot using chrome headless
func (s Screener) Run(url string) (ScreenerResult, error) {

	// Start Chrome
	// Ignore ssl errors
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("ignore-certificate-errors", "1"),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	// Remove the 2nd param if you don't need debug information logged
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Run Tasks
	// List of actions to run in sequence (which also fills our image buffer)
	var imageBuf []byte
	if err := chromedp.Run(ctx, screenshotTasks(url, &imageBuf)); err != nil {
		return *NewScreenerResult(""), err
	}

	// Generate b64 image
	content := base64.StdEncoding.EncodeToString(imageBuf)

	return *NewScreenerResult(content), nil
}

func screenshotTasks(url string, imageBuf *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Sleep(5 * time.Second),
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			*imageBuf, err = page.CaptureScreenshot().WithQuality(90).Do(ctx)
			return err
		}),
	}
}
