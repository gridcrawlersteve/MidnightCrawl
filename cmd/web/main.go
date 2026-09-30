//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"syscall/js"
	"time"

	"github.com/gdamore/tcell/v2"
	"wizardry/data"
	"wizardry/engine"
	"wizardry/render"
)

var game *engine.GameState
var screen *render.Screen
var scenarioID = "1"

// Persist the complete session without duplicating embedded scenario assets or
// the title animation's channel. Roster references are reconnected on load.
type session struct {
	*engine.GameState
	Scenario *data.Scenario     `json:"Scenario"`
	Title    *engine.TitleState `json:"Title"`
}
type saveFile struct {
	Version  int           `json:"version"`
	Scenario string        `json:"scenario"`
	State    session       `json:"state"`
	Items    []data.Item   `json:"items"`
	Mazes    data.MazeData `json:"mazes"`
}

func encodeSave(g *engine.GameState) (string, error) {
	b, err := json.Marshal(saveFile{1, scenarioID, session{GameState: g}, g.Scenario.Items, g.Scenario.Mazes})
	return string(b), err
}
func saveBrowser(g *engine.GameState) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("save unavailable: %v", r)
		}
	}()
	raw, err := encodeSave(g)
	if err != nil {
		return err
	}
	js.Global().Get("localStorage").Call("setItem", "midnight-crawl-wizardry-"+scenarioID, raw)
	return nil
}
func reconnect(g *engine.GameState) {
	roster := map[string]*engine.Character{}
	for _, c := range g.Town.Roster.Characters {
		if c != nil {
			roster[c.Name] = c
		}
	}
	charType := reflect.TypeOf((*engine.Character)(nil))
	var walk func(reflect.Value)
	walk = func(v reflect.Value) {
		if !v.IsValid() {
			return
		}
		if v.Type() == charType {
			if !v.IsNil() && v.CanSet() {
				if c := roster[v.Interface().(*engine.Character).Name]; c != nil {
					v.Set(reflect.ValueOf(c))
				}
			}
			return
		}
		switch v.Kind() {
		case reflect.Pointer:
			if !v.IsNil() {
				walk(v.Elem())
			}
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				walk(v.Field(i))
			}
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				walk(v.Index(i))
			}
		}
	}
	walk(reflect.ValueOf(g.Town))
	walk(reflect.ValueOf(g.Combat))
	walk(reflect.ValueOf(g.MazeInspectFound))
}
func decodeSave(raw string) error {
	var saved saveFile
	if len(raw) > 5000000 {
		return fmt.Errorf("save too large")
	}
	if err := json.Unmarshal([]byte(raw), &saved); err != nil {
		return err
	}
	g := saved.State.GameState
	if saved.Version != 1 || saved.Scenario != scenarioID || g == nil || g.Town == nil || g.Town.Roster == nil || g.Town.Party == nil {
		return fmt.Errorf("invalid save")
	}
	if g.Phase < engine.PhaseTitle || g.Phase > engine.PhaseUtilities || g.PlayerX < 0 || g.PlayerX > 19 || g.PlayerY < 0 || g.PlayerY > 19 || g.Facing < 0 || g.Facing > 3 || len(g.Town.Party.Members) > 6 {
		return fmt.Errorf("invalid session")
	}
	if g.Phase == engine.PhaseCombat && g.Combat == nil {
		return fmt.Errorf("missing combat")
	}
	g.Scenario = game.Scenario
	if len(saved.Items) != len(g.Scenario.Items) || len(saved.Mazes.Levels) != len(g.Scenario.Mazes.Levels) {
		return fmt.Errorf("invalid scenario state")
	}
	for _, level := range saved.Mazes.Levels {
		if len(level.Cells) != 20 {
			return fmt.Errorf("invalid maze")
		}
		for _, row := range level.Cells {
			if len(row) != 20 {
				return fmt.Errorf("invalid maze row")
			}
		}
	}
	for _, c := range g.Town.Roster.Characters {
		if c != nil && (c.Class < 0 || c.Class > 7 || c.Race < 0 || c.Race > 4 || c.ItemCount < 0 || c.ItemCount > 8) {
			return fmt.Errorf("invalid character")
		}
	}
	if g.Phase == engine.PhaseCreation && (g.Town.Creation == nil || g.Town.Creation.Step < 0 || g.Town.Creation.Step > 6 || g.Town.Creation.StatCursor < 0 || g.Town.Creation.StatCursor > 5) {
		return fmt.Errorf("invalid creation")
	}
	g.Scenario.Items = saved.Items
	g.Scenario.Mazes = saved.Mazes
	if g.MazeLevel < 0 || g.MazeLevel >= len(g.Scenario.Mazes.Levels) {
		return fmt.Errorf("invalid dungeon level")
	}
	if g.Phase == engine.PhaseTitle {
		g.Title = &engine.TitleState{Step: engine.TitleMenu}
	}
	reconnect(g)
	if g.SearchCell != nil {
		g.SearchCell = g.CurrentCell()
	}
	game = g
	return nil
}
func frame() string {
	redraw(screen, game)
	raw, _ := json.Marshal(map[string]any{"cells": screen.BrowserFrame(), "text": screen.BrowserText(game.Phase == engine.PhaseUtilities || game.ShowMap), "phase": int(game.Phase), "location": int(game.Town.Location), "roster": game.Town.Roster.Characters, "party": game.Town.Party.Members, "combat": game.Combat, "creation": game.Town.Creation, "x": game.PlayerX, "y": game.PlayerY, "level": game.MazeLevel, "facing": int(game.Facing)})
	js.Global().Call("mcPaint", string(raw))
	return string(raw)
}
func keyInput(this js.Value, args []js.Value) any {
	if game == nil || len(args) == 0 {
		return nil
	}
	key := args[0].String()
	if game.Phase == engine.PhaseUtilities && game.Util != nil {
		if game.Util.Step == engine.UtilBackup && (key == "T" || key == "t" || key == "F" || key == "f") {
			game.Util.Step = engine.UtilMenu
			if key == "T" || key == "t" {
				js.Global().Call("mcDownloadSave")
			} else {
				js.Global().Call("mcChooseSave")
			}
			return frame()
		}
		if game.Util.Step == engine.UtilMenu && (key == "I" || key == "i") {
			js.Global().Call("mcChooseDisk")
			return frame()
		}
	}
	if game.Phase == engine.PhaseMaze && !game.ShowMap {
		switch key {
		case "ArrowUp":
			key = "F"
		case "ArrowLeft":
			key = "L"
		case "ArrowRight":
			key = "R"
		case "ArrowDown":
			game.TurnRight()
			key = "R"
		}
	}
	var ev *tcell.EventKey
	keys := map[string]tcell.Key{"Enter": tcell.KeyEnter, "Escape": tcell.KeyEscape, "Backspace": tcell.KeyBackspace2, "ArrowUp": tcell.KeyUp, "ArrowDown": tcell.KeyDown, "ArrowLeft": tcell.KeyLeft, "ArrowRight": tcell.KeyRight, "Tab": tcell.KeyTab}
	if k, ok := keys[key]; ok {
		ev = tcell.NewEventKey(k, 0, tcell.ModNone)
	} else {
		r := []rune(key)
		if len(r) != 1 {
			return nil
		}
		ev = tcell.NewEventKey(tcell.KeyRune, r[0], tcell.ModNone)
	}
	switch game.Phase {
	case engine.PhaseTitle:
		handleTitleInput(screen, game, ev)
	case engine.PhaseTown:
		if handleTownInput(game, ev) {
			game.Phase = engine.PhaseTitle
			game.Title = &engine.TitleState{Step: engine.TitleMenu}
		}
	case engine.PhaseCamp:
		handleCampInput(game, ev)
	case engine.PhaseMaze:
		handleMazeInput(screen, game, ev)
	case engine.PhaseCombat:
		handleCombatInput(game, ev)
	case engine.PhaseUtilities:
		handleUtilInput(game, ev)
	case engine.PhaseCreation:
		handleCreationInput(game, ev)
	}
	if err := game.Save(); err != nil {
		js.Global().Call("mcSaveStatus", err.Error())
	} else {
		js.Global().Call("mcSaveStatus", "Saved on this device")
	}
	return frame()
}
func initGame(this js.Value, args []js.Value) any {
	if len(args) > 0 {
		scenarioID = args[0].String()
	}
	s, err := loadScenario(scenarioID)
	if err != nil {
		return "error: " + err.Error()
	}
	game = engine.New(s)
	game.Version = "0.5"
	game.BuildDate = "29-SEP-26"
	game.Title.Step = engine.TitleMenu
	screen, err = render.NewScreen()
	if err != nil {
		return "error: " + err.Error()
	}
	// Storage denial (e.g. Safari private policy) must not prevent playing.
	func() {
		defer func() { recover() }()
		raw := js.Global().Get("localStorage").Call("getItem", "midnight-crawl-wizardry-"+scenarioID)
		if !raw.IsNull() {
			if err := decodeSave(raw.String()); err != nil {
				js.Global().Call("mcSaveStatus", "Save could not be loaded: "+err.Error())
			}
		}
	}()
	return frame()
}
func main() {
	engine.BrowserSave = saveBrowser
	engine.BrowserRosterReader = func(key string) (out []byte, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("storage unavailable")
			}
		}()
		raw := js.Global().Get("localStorage").Call("getItem", "midnight-crawl-wizardry-"+key)
		if raw.IsNull() {
			return nil, fmt.Errorf("no saved roster")
		}
		var saved saveFile
		if err = json.Unmarshal([]byte(raw.String()), &saved); err != nil {
			return nil, err
		}
		if saved.State.GameState == nil || saved.State.Town == nil || saved.State.Town.Roster == nil {
			return nil, fmt.Errorf("invalid roster")
		}
		return json.Marshal(engine.SaveState{Roster: saved.State.Town.Roster.Characters})
	}
	js.Global().Set("mcGameInit", js.FuncOf(initGame))
	js.Global().Set("mcGameKey", js.FuncOf(keyInput))
	js.Global().Set("mcGameState", js.FuncOf(func(js.Value, []js.Value) any { return frame() }))
	js.Global().Set("mcGameExport", js.FuncOf(func(js.Value, []js.Value) any { raw, _ := encodeSave(game); return raw }))
	js.Global().Set("mcGameImportDSK", js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) == 0 || args[0].Get("byteLength").Int() != 143360 {
			return "Expected a 143360-byte Apple II disk"
		}
		bytes := make([]byte, 143360)
		js.CopyBytesToGo(bytes, args[0])
		old := engine.ReadDSK
		engine.ReadDSK = func(string) ([]byte, error) { return bytes, nil }
		defer func() { engine.ReadDSK = old }()
		msgs, err := engine.ImportCharactersFromDSK(game, "upload.dsk")
		if err != nil {
			return err.Error()
		}
		game.Util = engine.NewUtilState()
		game.Util.Step = engine.UtilImportResult
		game.Util.Messages = msgs
		game.Phase = engine.PhaseUtilities
		game.Save()
		frame()
		return ""
	}))
	js.Global().Set("mcGameImport", js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) == 0 {
			return "Missing save"
		}
		if err := decodeSave(args[0].String()); err != nil {
			return err.Error()
		}
		game.Save()
		frame()
		return ""
	}))
	go timers()
	select {}
}

func timers() {
	combatTicker := time.NewTicker(100 * time.Millisecond)
	storyTicker := time.NewTicker(6 * time.Second)
	innTicker := time.NewTicker(700 * time.Millisecond)
	var lastCombatAdvance time.Time
	for {
		select {
		case <-combatTicker.C:
			if game != nil && game.Phase == engine.PhaseCombat && game.Combat != nil {
				combat := game.Combat

				// Compute delay based on combat phase.
				// PAUSE1 (CombatInit/CombatExecute): T)IME-controlled (MazeDelay 1-5000 → ms).
				// PAUSE2 (CombatChestResult/CombatVictory/CombatDefeat): fixed delay —
				//   Pascal PAUSE2 hardcodes 3000, independent of T)IME setting.
				delayMs := 1500 // default (PAUSE2 fixed, and PAUSE1 with no T)IME set)
				if combat.Phase == engine.CombatInit || combat.Phase == engine.CombatExecute {
					// PAUSE1: uses T)IME setting
					if game.MazeDelay > 0 {
						delayMs = game.MazeDelay * 3 / 10 // ~0.3ms per unit
						if delayMs < 50 {
							delayMs = 50 // minimum 50ms
						}
					}
				}
				elapsed := time.Since(lastCombatAdvance)
				if elapsed < time.Duration(delayMs)*time.Millisecond {
					continue // not enough time yet
				}

				if combat.Phase == engine.CombatInit {
					// Auto-advance after delay (simulates Apple II disk I/O time)
					lastCombatAdvance = time.Now()
					if combat.Surprised == 2 {
						combat.ExecuteRound(game)
					} else {
						combat.Phase = engine.CombatChoose
						combat.CurrentActor = findNextActor(game, -1)
					}
					frame()
					game.Save()
				} else if combat.Phase == engine.CombatExecute && !combat.HamanSelecting {
					lastCombatAdvance = time.Now()
					advanceCombatMessages(game)
					frame()
					game.Save()
				} else if combat.Phase == engine.CombatChestResult {
					lastCombatAdvance = time.Now()
					advanceChestMessages(game)
					frame()
					game.Save()
				} else if combat.Phase == engine.CombatVictory {
					lastCombatAdvance = time.Now()
					if !combat.ChestPauseUsed {
						combat.ChestPauseUsed = true
					} else {
						game.Combat = nil
						game.Phase = engine.PhaseMaze
						game.MazeMessage = ""
						game.MazeMessage2 = ""
					}
					frame()
					game.Save()
				} else if combat.Phase == engine.CombatDefeat {
					lastCombatAdvance = time.Now()
					if !combat.ChestPauseUsed {
						combat.ChestPauseUsed = true
					} else {
						game.Combat = nil
						game.Phase = engine.PhaseTown
						game.Town.Location = engine.Castle
						game.Town.Message = "YOUR PARTY HAS PERISHED IN THE MAZE..."
					}
					frame()
					game.Save()
				}
			}

		case <-storyTicker.C:
			// Auto-advance Wiz 3 story pages every 6 seconds (no input required)
			if game != nil && game.Phase == engine.PhaseTitle && game.Title != nil &&
				game.Title.Step == engine.TitleStory {
				title := game.Title
				title.StoryFrame++
				totalFrames := len(game.Scenario.TitleFrames)
				if totalFrames == 0 {
					totalFrames = len(game.Scenario.TitleStory)
				}
				if title.StoryFrame >= totalFrames {
					// No keypress through entire sequence → loop back to start
					// (p-code: cycling loop at offsets 855-871 in TITLELOA)
					title.StoryFrame = 0
				}
				frame()
				game.Save()
			}

		case <-innTicker.C:
			// Inn healing animation — Pascal HEALHP loop (CASTLE2.TEXT lines 423-448)
			// Each tick: heal HPADD HP, deduct gold, redraw. When done → level-up screen.
			// Pascal delay: FOR PAUSEX := 1 TO 500 DO ; (~0.5-1s on Apple II)
			if game != nil && game.Phase == engine.PhaseTown && game.Town.Location == engine.Inn &&
				game.Town.InnStep == engine.InnHealing && game.Town.InnChar != nil {
				c := game.Town.InnChar
				town := game.Town
				if town.InnHealAmt == 0 {
					// Stables: no healing animation, go straight to level-up
					innTransitionToLevelUp(game)
					frame()
					game.Save()
				} else if c.HP < c.MaxHP && c.Gold >= town.InnHealCost {
					// One heal step per tick
					c.HP += town.InnHealAmt
					if c.HP > c.MaxHP {
						c.HP = c.MaxHP
					}
					c.Gold -= town.InnHealCost
					frame()
					game.Save()
				} else {
					// Healing done — transition to level-up
					innTransitionToLevelUp(game)
					frame()
					game.Save()
				}
			}
		}
	}
}
