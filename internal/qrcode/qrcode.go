package qrcode

import (
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/wxnacy/bdpan-cli/internal/logger"
)

func CreateQRCodeImage(text string, size int, filename string) error {
	scaleW := size
	scaleH := size
	// 生成二维码
	qrCode, err := qr.Encode(text, qr.M, qr.Auto)
	if err != nil {
		return err
	}

	// 可选：调整二维码的大小
	qrCode, err = barcode.Scale(qrCode, scaleW, scaleH)
	if err != nil {
		return err
	}

	// 创建图像文件
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 编码为PNG并保存到文件
	err = png.Encode(file, qrCode)
	if err != nil {
		return err
	}
	return nil
}

func ShowByUrl(uri string, timeout time.Duration) error {
	var images []image.Image
	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	image, _, err := image.Decode(resp.Body)
	if err != nil {
		return err
	}
	images = append(images, image)

	if err := ui.Init(); err != nil {
		return err
	}
	defer ui.Close()

	img := widgets.NewImage(nil)
	img.SetRect(0, 0, 100, 50)
	index := 0
	render := func() {
		img.Image = images[index]
		img.Monochrome = true
		img.Title = "BDPan"
		ui.Render(img)
	}
	render()

	// uiEvents := ui.PollEvents()
	for i := range int(timeout / time.Second) {
		deadline := int(timeout/time.Second) - i
		logger.Infof("二维码倒计时 %d", deadline)
		img.Title = fmt.Sprintf("BDPan %d", deadline)
		// e := <-uiEvents
		// switch e.ID {
		// case "q", "<C-c>":
		// return errors.New("Exit")
		// }
		render()
		time.Sleep(1 * time.Second)
	}
	return nil
}
