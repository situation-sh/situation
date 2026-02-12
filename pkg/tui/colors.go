package tui

import "github.com/charmbracelet/lipgloss"

var PrimaryColor = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#1c1c1c", ANSI256: "234", ANSI: "0"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#eeeeee ", ANSI256: "255", ANSI: "15"},
}

var InvPrimaryColor = lipgloss.CompleteAdaptiveColor{
	Light: PrimaryColor.Dark,
	Dark:  PrimaryColor.Light,
}

// var BgPrimaryColor = lipgloss.CompleteAdaptiveColor{
// 	Light: PrimaryColor.Dark,
// 	Dark:  PrimaryColor.Light,
// }

var PrimaryMutedColor = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#d0d0d0", ANSI256: "252", ANSI: "8"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#121212", ANSI256: "233", ANSI: "7"},
}

var AccentColor = lipgloss.CompleteColor{
	TrueColor: "#00c573",
	ANSI256:   "42",
	ANSI:      "10",
}

var AccentMutedColor = lipgloss.CompleteAdaptiveColor{
	Light: lipgloss.CompleteColor{TrueColor: "#00530d", ANSI256: "22", ANSI: "2"},
	Dark:  lipgloss.CompleteColor{TrueColor: "#00530d", ANSI256: "22", ANSI: "2"},
}
