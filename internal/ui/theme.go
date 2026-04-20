package ui

import (
	"image/color"

	"avledger/internal/assets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// CustomTheme is a custom theme for AVLedger
type CustomTheme struct {
	// ForcedVariant overrides the OS theme variant when non-nil.
	ForcedVariant *fyne.ThemeVariant
	// lastVariant stores the variant requested by Fyne (OS default)
	lastVariant fyne.ThemeVariant
}

var _ fyne.Theme = (*CustomTheme)(nil)

func (m *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	m.lastVariant = variant
	v := variant
	if m.ForcedVariant != nil {
		v = *m.ForcedVariant
	}
	// Aviation blue primary color
	if name == theme.ColorNamePrimary {
		return color.NRGBA{R: 52, G: 152, B: 219, A: 255}
	}
	// Brighter disabled text in dark theme
	if name == theme.ColorNameDisabled {
		if v == theme.VariantDark {
			return color.NRGBA{R: 120, G: 120, B: 120, A: 255}
		}
	}
	return theme.DefaultTheme().Color(name, v)
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
