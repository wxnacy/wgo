package color

import (
    "fmt"
)

// 前景 背景 颜色
// ---------------------------------------
// 30  40  黑色
// 31  41  红色
// 32  42  绿色
// 33  43  黄色
// 34  44  蓝色
// 35  45  紫红色
// 36  46  青蓝色
// 37  47  白色
//
// 代码 意义
// -------------------------
//  0  终端默认设置
//  1  高亮显示
//  4  使用下划线
//  5  闪烁
//  7  反白显示
//  8  不可见

const (
	TextBlack = iota + 30
	TextRed
	TextGreen
	TextYellow
	TextBlue
	TextMagenta
	TextCyan
	TextWhite

)

const (

	BackgroundBlack = iota + 40
	BackgroundRed
	BackgroundGreen
	BackgroundYellow
	BackgroundBlue
	BackgroundMagenta
	BackgroundCyan
	BackgroundWhite
)

func Black(msg string) string {
    return SetColor(msg, 0, 0, TextBlack)
}

func Red(msg string) string {
    return SetColor(msg, 0, 0, TextRed)
}

func Green(msg string) string {
    return SetColor(msg, 0, 0, TextGreen)
}

func Yellow(msg string) string {
    return SetColor(msg, 0, 0, TextYellow)
}

func Blue(msg string) string {
    return SetColor(msg, 0, 0, TextBlue)
}

func Magenta(msg string) string {
    return SetColor(msg, 0, 0, TextMagenta)
}

func Cyan(msg string) string {
    return SetColor(msg, 0, 0, TextCyan)
}

func White(msg string) string {
    return SetColor(msg, 0, 0, TextWhite)
}

func BgBlack(msg string) string {
    return SetColor(msg, 0, BackgroundBlack, TextWhite)
}

func BgRed(msg string) string {
    return SetColor(msg, 0, BackgroundRed, TextWhite)
}

func BgGreen(msg string) string {
    return SetColor(msg, 0, BackgroundGreen, TextWhite)
}

func BgYellow(msg string) string {
    return SetColor(msg, 0, BackgroundYellow, TextWhite)
}

func BgBlue(msg string) string {
    return SetColor(msg, 0, BackgroundBlue, TextWhite)
}

func BgMagenta(msg string) string {
    return SetColor(msg, 0, BackgroundMagenta, TextWhite)
}

func BgCyan(msg string) string {
    return SetColor(msg, 0, BackgroundCyan, TextWhite)
}

func BgWhite(msg string) string {
    return SetColor(msg, 0, BackgroundWhite, TextWhite)
}

func SetColor(msg string, conf, bg, text int) string {
    return fmt.Sprintf("%c[%d;%d;%dm%s%c[0m", 0x1B, conf, bg, text, msg, 0x1B)
}

func PrintColor() {

    text := "Hello World"
    fmt.Printf("Black()  \t%s\n", Black(text))
    fmt.Printf("Red()    \t%s\n", Red(text))
    fmt.Printf("Green()  \t%s\n", Green(text))
    fmt.Printf("Yellow()\t%s\n", Yellow(text))
    fmt.Printf("Blue()   \t%s\n", Blue(text))
    fmt.Printf("Magenta()\t%s\n", Magenta(text))
    fmt.Printf("Cyan()   \t%s\n", Cyan(text))
    fmt.Printf("White()  \t%s\n", White(text))
    fmt.Printf("BgBlack()\t%s\n", BgBlack(text))
    fmt.Printf("BgRed()  \t%s\n", BgRed(text))
    fmt.Printf("BgGreen()\t%s\n", BgGreen(text))
    fmt.Printf("BgYellow()\t%s\n", BgYellow(text))
    fmt.Printf("BgBlue()\t%s\n", BgBlue(text))
    fmt.Printf("BgMagenta()\t%s\n", BgMagenta(text))
    fmt.Printf("BgCyan()\t%s\n", BgCyan(text))
    fmt.Printf("BgWhite()\t%s\n", BgWhite(text))

    // fmt.Println( fmt.Sprintf("%c[%d;%d;%dm%s%c[0m", 0x1B, 0, BackgroundWhite, TextWhite, text, 0x1B) )
}
