package ui

import (
	"image/color"

	"avledger/internal/assets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// CustomTheme is a custom theme for AVLedger
type CustomTheme struct{}

var _ fyne.Theme = (*CustomTheme)(nil)

func (m *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Aviation blue primary color
	if name == theme.ColorNamePrimary {
		return color.NRGBA{R: 52, G: 152, B: 219, A: 255}
	}
	return theme.DefaultTheme().Color(name, variant)
}

// Font returns the bundled custom font (Roboto)
func (m *CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	return assets.ResourceRobotoRegularTtf
}

// Icon returns the default icons
func (m *CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns the default sizes
func (m *CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
