# Midnight Crawl v0.5 — full Wizardry browser interface

The browser runs the upstream Go game with its original town, creation, camp,
dungeon, combat, and utilities screens. The previous movement-only bridge and
prototype UI are replaced. Upstream source: https://github.com/sshoecraft/wizardry
(MIT; see THIRD_PARTY_LICENSES). The render package and input handlers are ported
from upstream commit 39faa42d05356aadd915c23330d8f1e0292845d6.

Build with Go 1.25.1:

    GOOS=js GOARCH=wasm go build -o wizardry.wasm ./cmd/web
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" wasm_exec.js

Serve this directory over HTTP. No third-party browser scripts or game server
are required. The original terminal renders to a tcell simulation screen whose
cells are painted to canvas. Terminal Sixel output is disabled. The title page
uses the original static bitmap; terminal animation is not run in the browser.

## Persistence

Each scenario stores a versioned session in localStorage. Saves retain roster,
party references, dungeon position, encounters and their actions, creation and
town prompts, item stocks, altered maze cells, and timers' gameplay transitions.
On restore, Character references are reconnected to the roster. Embedded title
and monster assets are loaded from the compiled scenario rather than duplicated
in the save. Export/import use JSON files. Apple II roster import uses a browser
file picker. Scenario transfer reads other browser saves and applies upstream
transfer restrictions. Scenarios II and III require existing characters.

The prototype's `mc_save` is separate and is not imported into Wizardry.

## Checks

    go test ./engine ./data ./scenarios/...
    cd tests
    node browser-engine.cjs
    node combat-loop.cjs

The second test exercises randomized real encounters and can end in defeat,
as original Wizardry is deadly for a level-one party. Neither test establishes
that all dungeon events, rare spells, or entire campaigns have been completed.
Browser checks cover desktop and phone viewports, touch menu actions, keyboard,
reload persistence, and absence of JavaScript startup errors.

## Publishing

The v0.5 workflow builds and tests the browser game, commits the matching WASM
runtime, then copies only the three generated preview assets to `main/v05/`.
The main root's v0.45 UI is untouched. A GitHub App commit on main can trigger
the branch-based Pages build after the workflow-generated asset commit.
