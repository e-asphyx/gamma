package gamma

import (
    "image"
    "image/color"
    "math"
)

type TransferFunction func(x float64) float64

type Transfer struct {
    src     image.Image
    op      func(t *Transfer, x, y int) color.NRGBA64
    stride  int
    minX    int
    minY    int
    pix     []uint8
    ramp    []uint16
}

func makeRamp(fn TransferFunction) (ramp []uint16) {
    ramp = make([]uint16, 0x10000)

    for c := 0; c < 0x10000; c++ {
        v := int(fn(float64(c) / 0xffff) * 0xffff)

        if v > 0xffff {
            ramp[c] = 0xffff
        } else if(v < 0) {
            ramp[c] = 0
        } else {
            ramp[c] = uint16(v)
        }
    }

    return
}

func NewTransfer(img image.Image, fn TransferFunction) *Transfer {
    t := Transfer{
        src:    img,
        ramp:   makeRamp(fn),
    }

    switch i := img.(type) {
    case *image.NRGBA:
        t.op = (*Transfer).opNRGBA
        t.pix = i.Pix
        t.stride = i.Stride
        t.minX = i.Rect.Min.X
        t.minY = i.Rect.Min.Y

    case *image.NRGBA64:
        t.op = (*Transfer).opNRGBA64
        t.pix = i.Pix
        t.stride = i.Stride
        t.minX = i.Rect.Min.X
        t.minY = i.Rect.Min.Y

    case *image.RGBA:
        t.op = (*Transfer).opRGBA
        t.pix = i.Pix
        t.stride = i.Stride
        t.minX = i.Rect.Min.X
        t.minY = i.Rect.Min.Y

    case *image.RGBA64:
        t.op = (*Transfer).opRGBA64
        t.pix = i.Pix
        t.stride = i.Stride
        t.minX = i.Rect.Min.X
        t.minY = i.Rect.Min.Y

    default:
        t.op = (*Transfer).opGen
    }

    return &t
}

func (t *Transfer) ColorModel() color.Model {
    return color.NRGBA64Model
}

func (t *Transfer) Bounds() image.Rectangle {
    return t.src.Bounds()
}

func (t *Transfer) At(x, y int) color.Color {
    return t.op(t, x, y)
}

func Gamma(g float64) TransferFunction {
    return func(x float64) float64 {
        return math.Pow(x, g)
    }
}

func SrgbReverse(c float64) float64 {
    if c <= 0.04045 {
        return c / 12.92
    } else {
        return math.Pow((c + 0.055) / 1.055, 2.4)
    }
}

func SrgbForward(c float64) float64 {
    if c <= 0.0031308 {
        return c * 12.92
    } else {
        return 1.055 * math.Pow(c, 1 / 2.4) - 0.055
    }
}

/* --------------------------------------------------------------------------------- */

func (t *Transfer) opNRGBA(x, y int) color.NRGBA64 {
    offs := (y - t.minY) * t.stride + (x - t.minX) * 4

    r := uint16(t.pix[offs + 0]) << 8
    g := uint16(t.pix[offs + 1]) << 8
    b := uint16(t.pix[offs + 2]) << 8
    a := uint16(t.pix[offs + 3]) << 8

    return color.NRGBA64{t.ramp[r], t.ramp[g], t.ramp[b], a}
}

func (t *Transfer) opNRGBA64(x, y int) color.NRGBA64 {
    offs := (y - t.minY) * t.stride + (x - t.minX) * 8

    r := uint16(t.pix[offs + 0]) << 8 | uint16(t.pix[offs + 1])
    g := uint16(t.pix[offs + 2]) << 8 | uint16(t.pix[offs + 3])
    b := uint16(t.pix[offs + 4]) << 8 | uint16(t.pix[offs + 5])
    a := uint16(t.pix[offs + 6]) << 8 | uint16(t.pix[offs + 7])

    return color.NRGBA64{t.ramp[r], t.ramp[g], t.ramp[b], a}
}

func (t *Transfer) opRGBA(x, y int) color.NRGBA64 {
    offs := (y - t.minY) * t.stride + (x - t.minX) * 4

    r := uint32(t.pix[offs + 0]) << 8
    g := uint32(t.pix[offs + 1]) << 8
    b := uint32(t.pix[offs + 2]) << 8
    a := uint32(t.pix[offs + 3]) << 8

    if a != 0 {
        r = (r * 0xffff) / a
        g = (g * 0xffff) / a
        b = (b * 0xffff) / a
    }

    return color.NRGBA64{t.ramp[r], t.ramp[g], t.ramp[b], uint16(a)}
}

func (t *Transfer) opRGBA64(x, y int) color.NRGBA64 {
    offs := (y - t.minY) * t.stride + (x - t.minX) * 8

    r := uint32(t.pix[offs + 0]) << 8 | uint32(t.pix[offs + 1])
    g := uint32(t.pix[offs + 2]) << 8 | uint32(t.pix[offs + 3])
    b := uint32(t.pix[offs + 4]) << 8 | uint32(t.pix[offs + 5])
    a := uint32(t.pix[offs + 6]) << 8 | uint32(t.pix[offs + 7])

    if a != 0 {
        r = (r * 0xffff) / a
        g = (g * 0xffff) / a
        b = (b * 0xffff) / a
    }

    return color.NRGBA64{t.ramp[r], t.ramp[g], t.ramp[b], uint16(a)}
}

func (t *Transfer) opGen(x, y int) color.NRGBA64 {
    r, g, b, a := t.src.At(x, y).RGBA()
 
    if a != 0 {
        r = (r * 0xffff) / a
        g = (g * 0xffff) / a
        b = (b * 0xffff) / a
    }

    return color.NRGBA64{t.ramp[r], t.ramp[g], t.ramp[b], uint16(a)}
}