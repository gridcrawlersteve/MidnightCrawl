//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"wizardry/engine"
	"wizardry/scenarios/wiz1"
)

var game *engine.GameState

func stateJSON(this js.Value, args []js.Value) any {
	if game == nil { return "{}" }
	cell := game.CurrentCell()
	// The upstream Wizardry renderer uses the original Pascal maze convention:
	// North increases Y. The browser renderer uses screen/grid convention:
	// North decreases Y. Mirror Y here so movement state and rendered wall
	// sampling share one coordinate system.
	out := map[string]any{
		"phase": int(game.Phase), "level": game.MazeLevel,
		"x": game.PlayerX, "y": 19-game.PlayerY, "facing": int(game.Facing),
		"wallAhead": string(game.WallAhead()),
		"partySize": len(game.Town.Party.Members),
		"inCombat": game.Combat != nil,
	}
	if cell != nil { out["cellType"] = string(cell.Type) }
	if level := game.CurrentLevel(); level != nil {
		h := len(level.Cells)
		walls := make([][]map[string]string, h)
		for engineY, row := range level.Cells {
			browserY := h - 1 - engineY
			walls[browserY] = make([]map[string]string, len(row))
			for x, c := range row {
				walls[browserY][x] = map[string]string{"n":string(c.N),"e":string(c.E),"s":string(c.S),"w":string(c.W)}
			}
		}
		out["walls"] = walls
	}
	b, _ := json.Marshal(out)
	return string(b)
}

func initGame(this js.Value, args []js.Value) any {
	s, err := wiz1.Load()
	if err != nil { return "error: " + err.Error() }
	game = engine.New(s)
	return stateJSON(this, nil)
}

func move(this js.Value, args []js.Value) any {
	if game != nil { game.MoveForward() }
	return stateJSON(this, nil)
}

func turn(this js.Value, args []js.Value) any {
	if game != nil && len(args)>0 {
		if args[0].String()=="left" { game.TurnLeft() } else { game.TurnRight() }
	}
	return stateJSON(this, nil)
}

func main() {
	js.Global().Set("mcEngineInit", js.FuncOf(initGame))
	js.Global().Set("mcEngineState", js.FuncOf(stateJSON))
	js.Global().Set("mcEngineMove", js.FuncOf(move))
	js.Global().Set("mcEngineTurn", js.FuncOf(turn))
	select {}
}
