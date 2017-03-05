// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commander

import (
	"time"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/mixins"
)

const maxStatusAge = 5 * time.Second

var (
	cmdColor = gxui.Color{
		R: 0.3,
		G: 0.3,
		B: 0.6,
		A: 1,
	}
	displayColor = gxui.Color{
		R: 0.3,
		G: 1,
		B: 0.6,
		A: 1,
	}

	ColorErr = gxui.Color{
		R: 1,
		G: 0.2,
		B: 0,
		A: 1,
	}
	ColorWarn = gxui.Color{
		R: 0.8,
		G: 0.7,
		B: 0.1,
		A: 1,
	}
	ColorInfo = gxui.Color{
		R: 0.1,
		G: 0.8,
		B: 0,
		A: 1,
	}
)

type commandBox struct {
	mixins.LinearLayout

	driver     gxui.Driver
	controller Controller

	label   gxui.Label
	current Command
	display gxui.Control
	input   gxui.Focusable
	status  gxui.Control

	statusTimer *time.Timer
}

func newCommandBox(driver gxui.Driver, theme gxui.Theme, controller Controller) *commandBox {
	box := &commandBox{
		driver:     driver,
		controller: controller,
	}

	box.label = theme.CreateLabel()
	box.label.SetColor(cmdColor)

	box.LinearLayout.Init(box, theme)
	box.SetDirection(gxui.LeftToRight)
	box.AddChild(box.label)
	box.Clear()
	return box
}

func (b *commandBox) Finish() {
	defer b.controller.Editor().Focus()
	statuser, ok := b.current.(Statuser)
	if !ok {
		b.Clear()
		return
	}
	b.status = statuser.Status()
	if b.status == nil {
		b.Clear()
		return
	}
	b.clearDisplay()
	b.clearInput()
	b.AddChild(b.status)
	b.statusTimer = time.AfterFunc(maxStatusAge, func() {
		b.driver.CallSync(func() {
			b.Clear()
		})
	})
}

func (b *commandBox) Clear() {
	b.label.SetText("none")
	b.clearDisplay()
	b.clearInput()
	b.clearStatus()
	b.current = nil
}

func (b *commandBox) Run(command Command) (needsInput bool) {
	b.Clear()
	if b.statusTimer != nil {
		b.statusTimer.Stop()
	}
	b.current = command

	b.label.SetText(b.current.Name())
	b.startCurrent()
	return b.nextInput()
}

func (b *commandBox) startCurrent() {
	starter, ok := b.current.(Starter)
	if !ok {
		return
	}
	b.display = starter.Start(b.controller)
	if b.display == nil {
		return
	}
	if colorSetter, ok := b.display.(ColorSetter); ok {
		colorSetter.SetColor(displayColor)
	}
	b.AddChild(b.display)
}

func (b *commandBox) Current() Command {
	return b.current
}

func (b *commandBox) KeyPress(event gxui.KeyboardEvent) (consume bool) {
	if event.Modifier == 0 && event.Key == gxui.KeyEscape {
		return false
	}
	isEnter := event.Modifier == 0 && event.Key == gxui.KeyEnter
	complete := isEnter
	if completer, ok := b.input.(Completer); ok {
		complete = completer.Complete(event)
	}
	if complete {
		hasMore := b.nextInput()
		complete = !hasMore
	}
	return !(complete && isEnter)
}

func (b *commandBox) HasFocus() bool {
	if b.input == nil {
		return false
	}
	return b.input.HasFocus()
}

func (b *commandBox) clearDisplay() {
	if b.display == nil {
		return
	}
	b.RemoveChild(b.display)
	b.display = nil
}

func (b *commandBox) clearInput() {
	if b.input == nil {
		return
	}
	b.RemoveChild(b.input)
	b.input = nil
}

func (b *commandBox) clearStatus() {
	if b.status == nil {
		return
	}
	b.RemoveChild(b.status)
	b.status = nil
}

func (b *commandBox) nextInput() (more bool) {
	queue, ok := b.current.(InputQueue)
	if !ok {
		return false
	}
	next := queue.Next()
	if next == nil {
		return false
	}
	b.clearInput()
	b.input = next
	b.AddChild(b.input)
	gxui.SetFocus(b.input)
	return true
}
