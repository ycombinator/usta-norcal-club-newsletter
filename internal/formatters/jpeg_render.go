package formatters

import (
	"context"
	"math"
	"net/url"
	"time"

	"github.com/chromedp/chromedp"
)

func renderHTMLToJPEG(htmlContent string, quality int) ([]byte, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("force-color-profile", "srgb"),
		chromedp.Flag("force-device-scale-factor", "2"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	dataURI := "data:text/html;charset=utf-8," + url.PathEscape(htmlContent)

	var dims []any
	err := chromedp.Run(ctx,
		chromedp.Navigate(dataURI),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Evaluate(`(() => {
			const body = document.body;
			const rect = body.getBoundingClientRect();
			return [rect.width, rect.height];
		})()`, &dims),
	)
	if err != nil {
		return nil, err
	}

	width := dims[0].(float64)
	height := dims[1].(float64)

	w := int64(math.Ceil(width))
	h := int64(math.Ceil(height))
	if w < 100 {
		w = 100
	}
	if h < 100 {
		h = 100
	}

	err = chromedp.Run(ctx,
		chromedp.EmulateViewport(w, h, chromedp.EmulateScale(2)),
		chromedp.Navigate(dataURI),
		chromedp.WaitVisible("body", chromedp.ByQuery),
	)
	if err != nil {
		return nil, err
	}

	var buf []byte
	err = chromedp.Run(ctx,
		chromedp.FullScreenshot(&buf, quality),
	)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
