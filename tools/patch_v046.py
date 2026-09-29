from pathlib import Path

p = Path('index.html')
s = p.read_text()

old_label = "Sword & sorcery dungeon prototype · v0.45 RESTART MAP FIX"
new_label = "Sword & sorcery dungeon prototype · v0.46 ENGINE AUTOMAP"
old_describe = "function describe(){let open=[];for(let i=0;i<4;i++){let[dxx,dyy]=dirs[i];if(!wallAt(x+dxx,y+dyy))open.push(i)}if(open.length===1)return'The passage ends at a blank wall of fitted stone.';if(open.length===2&&((open[0]+2)%4===open[1]))return'A narrow stone corridor runs through the darkness.';if(open.length===2)return'The passage bends sharply here.';if(open.length===3)return'A three-way junction opens through worn stonework.';return'Cold air moves through the ancient masonry.'}"
new_describe = "function describe(){let open=[];for(let i=0;i<4;i++)if(!cellWall(x,y,i))open.push(i);if(open.length===1)return'The passage ends at a blank wall of fitted stone.';if(open.length===2&&((open[0]+2)%4===open[1]))return'A narrow stone corridor runs through the darkness.';if(open.length===2)return'The passage bends sharply here.';if(open.length===3)return'A three-way junction opens through worn stonework.';return'Cold air moves through the ancient masonry.'}"
old_render = "function render(){document.getElementById('rewardBox').style.display='none';if(inCombat){renderCombat();return}document.getElementById('controls').style.display='block';document.getElementById('combat').style.display='none';document.getElementById('mapPanel').style.display='none';document.getElementById('locationTitle').textContent='YOUR SURROUNDINGS';renderView();document.getElementById('description').textContent=describe();let out='';for(let yy=0;yy<grid.length;yy++){for(let xx=0;xx<grid[0].length;xx++){let k=xx+','+yy;if(xx==x&&yy==y)out+='^>v<'[dir];else if(!seen.has(k))out+=' ';else out+=grid[yy][xx]=='#'?'█':'·'}out+='\\n'}document.getElementById('map').textContent=out}"
new_render = "function renderMap(){if(!engineReady||!engineWalls){let out='';for(let yy=0;yy<grid.length;yy++){for(let xx=0;xx<grid[0].length;xx++){let k=xx+','+yy;if(xx==x&&yy==y)out+='^>v<'[dir];else if(!seen.has(k))out+=' ';else out+=grid[yy][xx]=='#'?'█':'·'}out+='\\n'}document.getElementById('map').textContent=out;return}const h=engineWalls.length,w=Math.max(...engineWalls.map(r=>r.length)),rows=Array.from({length:h*2+1},()=>Array(w*2+1).fill(' '));for(let yy=0;yy<h;yy++)for(let xx=0;xx<w;xx++){const k=xx+','+yy;if(!seen.has(k)&&!(xx===x&&yy===y))continue;const sy=yy*2+1,sx=xx*2+1;rows[sy][sx]=(xx===x&&yy===y)?'^>v<'[dir]:'·';if(engineWall(xx,yy,0))rows[sy-1][sx]='─';if(engineWall(xx,yy,2))rows[sy+1][sx]='─';if(engineWall(xx,yy,3))rows[sy][sx-1]='│';if(engineWall(xx,yy,1))rows[sy][sx+1]='│'}document.getElementById('map').textContent=rows.map(r=>r.join('')).join('\\n')}function render(){document.getElementById('rewardBox').style.display='none';if(inCombat){renderCombat();return}document.getElementById('controls').style.display='block';document.getElementById('combat').style.display='none';document.getElementById('mapPanel').style.display='none';document.getElementById('locationTitle').textContent='YOUR SURROUNDINGS';renderView();document.getElementById('description').textContent=describe();renderMap()}"

for name, old in [('label', old_label), ('describe', old_describe), ('render', old_render)]:
    count = s.count(old)
    if count != 1:
        raise SystemExit(f'Guard failed: expected exactly one {name} match, found {count}')

s = s.replace(old_label, new_label).replace(old_describe, new_describe).replace(old_render, new_render)
if s.count('<!doctype html>') != 1:
    raise SystemExit('Guard failed: document duplication detected')
p.write_text(s)
