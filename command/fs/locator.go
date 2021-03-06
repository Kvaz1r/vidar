// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package fs

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/scoring"
	"github.com/nelsam/vidar/setting"
)

const (
	minInputChars = 10
	metaNewFile   = "<new file>"
	metaCurrDir   = "<current dir>"
)

var (
	metaColor = gxui.Color{
		R: 0.4,
		G: 0.8,
		B: 0.3,
		A: 1,
	}
	dirColor = gxui.Color{
		R: 0.1,
		G: 0.3,
		B: 0.8,
		A: 1,
	}
	completionBG = gxui.Color{
		R: 1,
		G: 1,
		B: 1,
		A: 0.05,
	}
	completionPadding = math.Size{
		H: 7,
		W: 7,
	}
)

// Mod is a type to tell a Locator which types of files to locate.
type Mod int

const (
	// Files tells a Locator to only locate regular files.
	Files Mod = 1 << iota

	// Dirs tells a Locator to only locate directories.
	Dirs

	// All tells a Locator to locate all file types.
	All = Files | Dirs
)

// FileGetter is used to get the currently open file.
//
// TODO: replace this with hooks on opening a file.
type FileGetter interface {
	CurrentFile() string
}

// Projecter is used to get the currently open project.
//
// TODO: replace this with hooks on opening a project.
type Projecter interface {
	Project() setting.Project
}

type valueLabel interface {
	gxui.Label

	Value() string
}

// Locator is a type of UI element which prompts the user for a file path.  It has
// completion features to help with locating existing files and folders.
type Locator struct {
	mixins.LinearLayout

	lock sync.RWMutex

	theme       *basic.Theme
	driver      gxui.Driver
	dir         *dirLabel
	file        *fileBox
	completions []valueLabel
	files       []string
	mod         Mod
}

// NewLocator initializes and returns a *Locator.
func NewLocator(driver gxui.Driver, theme *basic.Theme, mod Mod) *Locator {
	f := &Locator{}
	f.Init(driver, theme, mod)
	return f
}

// Init is provided for legacy reasons, as a way to initialize an uninitialized
// *Locator value.  NewLocator should be used instead.
func (f *Locator) Init(driver gxui.Driver, theme *basic.Theme, mod Mod) {
	f.LinearLayout.Init(f, theme)
	f.theme = theme
	f.driver = driver
	f.mod = mod

	f.SetDirection(gxui.LeftToRight)
	f.dir = newDirLabel(driver, theme)
	f.AddChild(f.dir)
	f.file = newFileBox(driver, theme, f)
	f.AddChild(f.file)
	f.loadDirContents()
}

func (f *Locator) LoadDir(control gxui.Control) {
	startingPath := findStart(control)

	f.driver.Call(func() {
		defer f.loadDirContents()
		f.dir.SetText(startingPath)
		f.file.SetText("")
	})
}

func (f *Locator) Path() string {
	return filepath.Join(f.dir.Text(), f.file.Text())
}

func (f *Locator) SetPath(filePath string) {
	defer f.loadDirContents()
	dir, file := filepath.Split(filePath)

	f.dir.SetText(dir)
	f.file.SetText(file)
}

func (f *Locator) KeyPress(event gxui.KeyboardEvent) bool {
	return f.file.KeyPress(event)
}

func (f *Locator) KeyDown(event gxui.KeyboardEvent) {
	f.file.KeyDown(event)
}

func (f *Locator) KeyUp(event gxui.KeyboardEvent) {
	f.file.KeyUp(event)
}

func (f *Locator) KeyStroke(event gxui.KeyStrokeEvent) bool {
	return f.file.KeyStroke(event)
}

func (f *Locator) KeyRepeat(event gxui.KeyboardEvent) {
	f.file.KeyRepeat(event)
}

func (f *Locator) Paint(c gxui.Canvas) {
	f.LinearLayout.Paint(c)

	if f.HasFocus() {
		r := f.Size().Rect()
		s := f.theme.FocusedStyle
		c.DrawRoundedRect(r, 3, 3, 3, 3, s.Pen, s.Brush)
	}
}

func (f *Locator) IsFocusable() bool {
	return f.file.IsFocusable()
}

func (f *Locator) HasFocus() bool {
	return f.file.HasFocus()
}

func (f *Locator) GainedFocus() {
	f.file.GainedFocus()
}

func (f *Locator) LostFocus() {
	f.file.LostFocus()
}

func (f *Locator) OnGainedFocus(callback func()) gxui.EventSubscription {
	return f.file.OnGainedFocus(callback)
}

func (f *Locator) OnLostFocus(callback func()) gxui.EventSubscription {
	return f.file.OnLostFocus(callback)
}

func (f *Locator) updateCompletions() {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.clearCompletions(f.completions)

	f.completions = []valueLabel{metaLabel(f.driver, f.theme, f.file.Text())}
	newCompletions := scoring.Sort(f.files, f.file.Text())

	for _, comp := range newCompletions {
		if strings.TrimSuffix(comp, string(filepath.Separator)) == f.file.Text() {
			// the meta entry will be incorrect
			f.completions = f.completions[1:]
		}
		color := f.theme.LabelStyle.FontColor
		if len(comp) > 0 && comp[len(comp)-1] == filepath.Separator {
			color = dirColor
		}
		l := newCompletionLabel(f.driver, f.theme, color)
		l.text = comp
		f.completions = append(f.completions, l)
	}

	f.addCompletions(f.completions)
}

func (f *Locator) clearCompletions(completions []valueLabel) {
	cloned := append([]valueLabel{}, completions...)
	f.driver.Call(func() {
		for _, l := range cloned {
			f.RemoveChild(l)
		}
	})
}

func (f *Locator) addCompletions(completions []valueLabel) {
	cloned := append([]valueLabel{}, completions...)
	f.driver.Call(func() {
		for _, l := range cloned {
			f.AddChild(l)
			l.SetHorizontalAlignment(gxui.AlignCenter)
		}
	})
}

func (f *Locator) loadDirContents() {
	f.lock.Lock()
	defer func() {
		f.lock.Unlock()
		f.updateCompletions()
	}()
	f.files = nil
	dir := f.dir.Text()
	if dir == "" {
		// Should be Windows-only.  The drive hasn't been chosen yet.
		return
	}
	contents, err := ioutil.ReadDir(dir)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		log.Printf("Unexpected error trying to read directory %s: %s", f.dir.Text(), err)
		return
	}
	for _, finfo := range contents {
		if f.mod.match(finfo) {
			name := finfo.Name()
			if finfo.IsDir() {
				name += string(filepath.Separator)
			}
			f.files = append(f.files, name)
		}
	}
}

func (m Mod) match(finfo os.FileInfo) bool {
	switch m {
	case Files:
		return !finfo.IsDir()
	case Dirs:
		return finfo.IsDir()
	}
	return true
}

func metaLabel(d gxui.Driver, t gxui.Theme, comp string) valueLabel {
	l := newCompletionLabel(d, t, metaColor)
	l.value = comp
	if comp == "" {
		l.text = metaCurrDir
		return l
	}
	l.text = metaNewFile
	return l
}
