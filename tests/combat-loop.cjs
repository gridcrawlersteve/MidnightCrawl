const fs=require('node:fs'),assert=require('node:assert/strict');require('../wasm_exec.js');const saves=new Map();global.localStorage={setItem:(k,v)=>saves.set(k,v),getItem:k=>saves.get(k)??null};let frame;global.mcPaint=raw=>frame=JSON.parse(raw);global.mcSaveStatus=()=>{};global.mcChooseSave=global.mcDownloadSave=global.mcChooseDisk=()=>{};const key=k=>mcGameKey(k),state=()=>JSON.parse(mcGameExport()).state;const sleep=ms=>new Promise(r=>setTimeout(r,ms));
function screen(){return frame.cells.map(c=>c.ch).join('')}
(async()=>{const go=new Go();const {instance}=await WebAssembly.instantiate(fs.readFileSync('../wizardry.wasm'),go.importObject);go.run(instance);mcGameInit('1');assert.equal(mcGameImport(fs.readFileSync('/tmp/mc-party-save.json','utf8')),'');
 key('e');key('m');assert.equal(frame.phase,2);key('l');key('n');key('t');for(const c of '167')key(c);key('Enter'); // 50ms battle message timing
 let moves=0;
 for(let i=0;i<100&&frame.phase===3;i++){
  const st=state();if(st.MazeMsgWait){key('Enter');continue}if(st.MazeSearchYN){key('y');continue}if(st.MazeMessage2?.includes('Y/N')){key('n');continue}
  const cells=JSON.parse(mcGameExport()).mazes.levels[frame.level].cells,queue=[[frame.x,frame.y,[]]],seen=new Set();let route;
  while(queue.length){const [x,y,path]=queue.shift(),id=x+','+y;if(seen.has(id))continue;seen.add(id);if(cells[y][x].encounter&&path.length){route=path;break}for(let d=0;d<4;d++){if(['open','door'].includes(cells[y][x][['n','e','s','w'][d]])){queue.push([(x+[0,1,0,-1][d]+20)%20,(y+[1,0,-1,0][d]+20)%20,[...path,d]])}}}
  assert(route);const d=route[0];while(frame.facing!==d)key('r');const wall=cells[frame.y][frame.x][['n','e','s','w'][d]];key(wall==='door'?'k':'f');moves++;
 }

 assert.equal(frame.phase,4,'Encounter should trigger during exploration');console.log('PASS: natural encounter after',moves,'moves');
 const combatSave=mcGameExport();mcGameInit('1');assert.equal(frame.phase,4);assert.equal(mcGameImport(combatSave),'');console.log('PASS: combat reload preserves encounter');
 let choseSpell=false,executed=false,loops=0;
 while(frame.phase===4&&loops++<500){const c=frame.combat;
  if(c.Phase===0||c.Phase===4){await sleep(110);continue}
  if(c.Phase===1){key('f');continue}
  if(c.Phase===2){
   if(c.SelectingGroup||c.SelectingSpellGroup||c.SelectingSpellTarget){key('1');continue}
   if(c.InputtingSpell){for(const ch of 'HALITO')key(ch);key('Enter');continue}
   const actor=c.CurrentActor;
   if(actor===4&&!choseSpell&&frame.party[actor].MageSpells[0]>0){key('s');choseSpell=true;continue}
   key(actor<3?'f':'p');continue
  }
  if(c.Phase===3){key('Enter');executed=true;continue}
  if(c.Phase===5){key('l');continue} // Leave chest, keep earned XP.
  key('Enter');
 }
 assert(executed);assert(choseSpell);assert.notEqual(frame.phase,4,'Combat must finish');console.log('PASS: fight, spell selection, timed round execution, and battle completion',frame.party.map(c=>[c.Name,c.XP,c.HP]));
 if(frame.phase===3){
  // Walk a shortest route through open/door cells back to the castle stairs.
  for(let n=0;n<150&&!(frame.x===0&&frame.y===0);n++){
   if(frame.phase!==3)break;
   const cells=JSON.parse(mcGameExport()).mazes.levels[0].cells,queue=[[frame.x,frame.y,[]]],seen=new Set();let route;
   while(queue.length){const [x,y,path]=queue.shift(),id=x+','+y;if(seen.has(id))continue;seen.add(id);if(x===0&&y===0){route=path;break}for(let d=0;d<4;d++){if(['open','door'].includes(cells[y][x][['n','e','s','w'][d]])){queue.push([(x+[0,1,0,-1][d]+20)%20,(y+[1,0,-1,0][d]+20)%20,[...path,d]])}}}
   const d=route[0];while(frame.facing!==d)key('r');const cell=cells[frame.y][frame.x];key(cell[['n','e','s','w'][d]]==='door'?'k':'f');
  }
  if(frame.phase===3&&frame.x===0&&frame.y===0){key('y');assert.equal(frame.phase,1);console.log('PASS: returned to castle after combat')}
 }
 assert.equal(mcGameImport(fs.readFileSync('/tmp/mc-party-save.json','utf8')),'');
 // Cross-scenario utilities can read persisted browser rosters.
 mcGameInit('2');key('u');key('t');key('1');assert.equal(frame.roster.length,6);console.log('PASS: scenario II transfers the saved roster');
 mcGameInit('3');key('u');key('t');key('1');assert.equal(frame.roster.length,6);console.log('PASS: scenario III transfers the saved roster');
 process.exit(0)
})().catch(e=>{console.error(e);console.log(screen());process.exit(1)});
