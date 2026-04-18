package ghostty

// Mods represents keyboard modifier key flags (ghostty_input_mods_e).
type Mods int

const (
	ModsNone       Mods = 0
	ModsShift      Mods = 1 << 0
	ModsCtrl       Mods = 1 << 1
	ModsAlt        Mods = 1 << 2
	ModsSuper      Mods = 1 << 3
	ModsCaps       Mods = 1 << 4
	ModsNum        Mods = 1 << 5
	ModsShiftRight Mods = 1 << 6
	ModsCtrlRight  Mods = 1 << 7
	ModsAltRight   Mods = 1 << 8
	ModsSuperRight Mods = 1 << 9
)

// InputAction represents a key event action (ghostty_input_action_e).
type InputAction int

const (
	InputActionRelease InputAction = iota // GHOSTTY_ACTION_RELEASE
	InputActionPress                      // GHOSTTY_ACTION_PRESS
	InputActionRepeat                     // GHOSTTY_ACTION_REPEAT
)

// MouseButtonState is press or release (ghostty_input_mouse_state_e).
type MouseButtonState int

const (
	MouseStateRelease MouseButtonState = iota // GHOSTTY_MOUSE_RELEASE
	MouseStatePress                           // GHOSTTY_MOUSE_PRESS
)

// MouseButton identifies a mouse button (ghostty_input_mouse_button_e).
type MouseButton int

const (
	MouseUnknown MouseButton = iota
	MouseLeft
	MouseRight
	MouseMiddle
	MouseFour
	MouseFive
	MouseSix
	MouseSeven
	MouseEight
	MouseNine
	MouseTen
	MouseEleven
)

// ScrollMods are packed modifier flags for scroll events
// (ghostty_input_scroll_mods_t).
type ScrollMods int

// Clipboard identifies the clipboard type (ghostty_clipboard_e).
type Clipboard int

const (
	ClipboardStandard  Clipboard = iota // GHOSTTY_CLIPBOARD_STANDARD
	ClipboardSelection                  // GHOSTTY_CLIPBOARD_SELECTION
)

// ClipboardRequest is the type of clipboard request
// (ghostty_clipboard_request_e).
type ClipboardRequest int

const (
	ClipboardRequestPaste      ClipboardRequest = iota // GHOSTTY_CLIPBOARD_REQUEST_PASTE
	ClipboardRequestOSC52Read                          // GHOSTTY_CLIPBOARD_REQUEST_OSC_52_READ
	ClipboardRequestOSC52Write                         // GHOSTTY_CLIPBOARD_REQUEST_OSC_52_WRITE
)

// ClipboardContent is a single clipboard content item
// (ghostty_clipboard_content_s).
type ClipboardContent struct {
	MIME string
	Data string
}

// TargetTag identifies whether an action targets the app or a surface
// (ghostty_target_tag_e).
type TargetTag int

const (
	TargetApp     TargetTag = iota // GHOSTTY_TARGET_APP
	TargetSurface                  // GHOSTTY_TARGET_SURFACE
)

// Target is an action target (ghostty_target_s).
type Target struct {
	Tag     TargetTag
	Surface *Surface
}

// ActionTag identifies a ghostty runtime action (ghostty_action_tag_e).
type ActionTag int

const (
	ActionQuit ActionTag = iota
	ActionNewWindow
	ActionNewTab
	ActionCloseTab
	ActionNewSplit
	ActionCloseAllWindows
	ActionToggleMaximize
	ActionToggleFullscreen
	ActionToggleTabOverview
	ActionToggleWindowDecorations
	ActionToggleQuickTerminal
	ActionToggleCommandPalette
	ActionToggleVisibility
	ActionToggleBackgroundOpacity
	ActionMoveTab
	ActionGotoTab
	ActionGotoSplit
	ActionGotoWindow
	ActionResizeSplit
	ActionEqualizeSplits
	ActionToggleSplitZoom
	ActionPresentTerminal
	ActionSizeLimit
	ActionResetWindowSize
	ActionInitialSize
	ActionCellSize
	ActionScrollbar
	ActionRender
	ActionInspector
	ActionShowGtkInspector
	ActionRenderInspector
	ActionDesktopNotification
	ActionSetTitle
	ActionSetTabTitle
	ActionPromptTitle
	ActionPwd
	ActionMouseShape
	ActionMouseVisibility
	ActionMouseOverLink
	ActionRendererHealth
	ActionOpenConfig
	ActionQuitTimer
	ActionFloatWindow
	ActionSecureInput
	ActionKeySequence
	ActionKeyTable
	ActionColorChange
	ActionReloadConfig
	ActionConfigChange
	ActionCloseWindow
	ActionRingBell
	ActionUndo
	ActionRedo
	ActionCheckForUpdates
	ActionOpenUrl
	ActionShowChildExited
	ActionProgressReport
	ActionShowOnScreenKeyboard
	ActionCommandFinished
	ActionStartSearch
	ActionEndSearch
	ActionSearchTotal
	ActionSearchSelected
	ActionReadonly
	ActionCopyTitleToClipboard
)

// Action is a ghostty runtime action (ghostty_action_s).
type Action struct {
	Tag ActionTag
}

// Key represents a keyboard key (ghostty_input_key_e).
type Key int

const (
	KeyUnidentified Key = iota

	// Writing System Keys
	KeyBackquote
	KeyBackslash
	KeyBracketLeft
	KeyBracketRight
	KeyComma
	KeyDigit0
	KeyDigit1
	KeyDigit2
	KeyDigit3
	KeyDigit4
	KeyDigit5
	KeyDigit6
	KeyDigit7
	KeyDigit8
	KeyDigit9
	KeyEqual
	KeyIntlBackslash
	KeyIntlRo
	KeyIntlYen
	KeyA
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyQ
	KeyR
	KeyS
	KeyT
	KeyU
	KeyV
	KeyW
	KeyX
	KeyY
	KeyZ
	KeyMinus
	KeyPeriod
	KeyQuote
	KeySemicolon
	KeySlash

	// Functional Keys
	KeyAltLeft
	KeyAltRight
	KeyBackspace
	KeyCapsLock
	KeyContextMenu
	KeyControlLeft
	KeyControlRight
	KeyEnter
	KeyMetaLeft
	KeyMetaRight
	KeyShiftLeft
	KeyShiftRight
	KeySpace
	KeyTab
	KeyConvert
	KeyKanaMode
	KeyNonConvert

	// Control Pad
	KeyDelete
	KeyEnd
	KeyHelp
	KeyHome
	KeyInsert
	KeyPageDown
	KeyPageUp

	// Arrow Pad
	KeyArrowDown
	KeyArrowLeft
	KeyArrowRight
	KeyArrowUp

	// Numpad
	KeyNumLock
	KeyNumpad0
	KeyNumpad1
	KeyNumpad2
	KeyNumpad3
	KeyNumpad4
	KeyNumpad5
	KeyNumpad6
	KeyNumpad7
	KeyNumpad8
	KeyNumpad9
	KeyNumpadAdd
	KeyNumpadBackspace
	KeyNumpadClear
	KeyNumpadClearEntry
	KeyNumpadComma
	KeyNumpadDecimal
	KeyNumpadDivide
	KeyNumpadEnter
	KeyNumpadEqual
	KeyNumpadMemoryAdd
	KeyNumpadMemoryClear
	KeyNumpadMemoryRecall
	KeyNumpadMemoryStore
	KeyNumpadMemorySubtract
	KeyNumpadMultiply
	KeyNumpadParenLeft
	KeyNumpadParenRight
	KeyNumpadSubtract
	KeyNumpadSeparator
	KeyNumpadUp
	KeyNumpadDown
	KeyNumpadRight
	KeyNumpadLeft
	KeyNumpadBegin
	KeyNumpadHome
	KeyNumpadEnd
	KeyNumpadInsert
	KeyNumpadDelete
	KeyNumpadPageUp
	KeyNumpadPageDown

	// Function Keys
	KeyEscape
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyF13
	KeyF14
	KeyF15
	KeyF16
	KeyF17
	KeyF18
	KeyF19
	KeyF20
	KeyF21
	KeyF22
	KeyF23
	KeyF24
	KeyF25
	KeyFn
	KeyFnLock
	KeyPrintScreen
	KeyScrollLock
	KeyPause

	// Media Keys
	KeyBrowserBack
	KeyBrowserFavorites
	KeyBrowserForward
	KeyBrowserHome
	KeyBrowserRefresh
	KeyBrowserSearch
	KeyBrowserStop
	KeyEject
	KeyLaunchApp1
	KeyLaunchApp2
	KeyLaunchMail
	KeyMediaPlayPause
	KeyMediaSelect
	KeyMediaStop
	KeyMediaTrackNext
	KeyMediaTrackPrevious
	KeyPower
	KeySleep
	KeyAudioVolumeDown
	KeyAudioVolumeMute
	KeyAudioVolumeUp
	KeyWakeUp

	// Legacy/Special
	KeyCopy
	KeyCut
	KeyPaste
)

// KeyEvent describes a keyboard input event (ghostty_input_key_s).
type KeyEvent struct {
	Action             InputAction
	Mods               Mods
	ConsumedMods       Mods
	Keycode            uint32
	Text               string
	UnshiftedCodepoint uint32
	Composing          bool
}

// ColorScheme is light or dark (ghostty_color_scheme_e).
type ColorScheme int

const (
	ColorSchemeLight ColorScheme = 0
	ColorSchemeDark  ColorScheme = 1
)

// SurfaceContext is the context for a new surface
// (ghostty_surface_context_e).
type SurfaceContext int

const (
	SurfaceContextWindow SurfaceContext = iota
	SurfaceContextTab
	SurfaceContextSplit
)
