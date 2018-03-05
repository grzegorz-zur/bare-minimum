package bm

type Modes struct {
	Mode
	Normal Mode
	Input  Mode
	Switch Mode
}

func (modes *Modes) SwitchMode(mode Mode) {
	if modes.Mode != nil {
		modes.Mode.Hide()
	}
	modes.Mode = mode
	modes.Mode.Show()
}