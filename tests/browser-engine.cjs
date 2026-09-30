const fs=require('node:fs');
require('../wasm_exec.js');
const saves=new Map();global.localStorage={setItem:(k,v)=>saves.set(k,v),getItem:k=>saves.get(k)??null};
let frame;global.mcPaint=raw=>frame=JSON.parse(raw);global.mcSaveStatus=()=>{};
function screenText(){const rows=[];for(let y=0;y<24;y++){rows.push(frame.cells.slice(y*80,(y+1)*80).map(c=>c.ch).join('').replace(/(.) /g,'$1'))}return rows.join('\n')}
function key(k){mcGameKey(k)}
const assert=require('node:assert/strict');
function text(value){for(const c of value)key(c);key('Enter')}
function create(name,cl){
 text(name);key('Y');key('Enter');key('Enter');key('A');key(cl===3?'B':'A');
 assert.equal(frame.creation.Step,4);
 const stat=cl===0?0:cl===1?1:cl===2?2:4;
 while(frame.creation.StatCursor!==stat)key('Enter');
 while(frame.creation.Stats[stat]<11)key('+');
 let guard=100;while(frame.creation.BonusPoints>0&&guard-->0){if(frame.creation.Stats[frame.creation.StatCursor]<18)key('+');else key('Enter')}
 assert.equal(frame.creation.BonusPoints,0);key('Escape');key(String.fromCharCode(65+cl));assert.equal(frame.creation.Step,6);key('Y');assert.equal(frame.phase,1);
}
(async()=>{const go=new Go();const {instance}=await WebAssembly.instantiate(fs.readFileSync('../wizardry.wasm'),go.importObject);go.run(instance);mcGameInit('1');key('s');key('e');key('t');
 for(const [name,cl] of [['STEVE',0],['AMY',0],['BANDIT',0],['PRIEST',2],['MAGE',1],['THIEF',3]])create(name,cl);
 assert.equal(frame.roster.length,6);key('Enter');key('g');
 for(const c of frame.roster){key('a');text(c.Name);key('Enter')}
 assert.equal(frame.party.length,6);console.log('PASS: created six adventurers and recruited party');
 const exported=mcGameExport();assert(exported.length<1500000);assert.equal(mcGameImport(exported),'');assert.equal(frame.party.length,6);assert.equal(JSON.parse(exported).state.Scenario,null);console.log('PASS: complete session round trip',exported.length,'bytes');
 key('Enter');key('b');
 const gear=[[1,10,7],[1,10,7],[1,10,7],[3,10,7],[5,9],[2,10,7]];
 for(let who=0;who<6;who++){
  key(String(who+1));key('b');
  for(const item of gear[who]){let guard=50;while(!JSON.parse(mcGameExport()).state.Town.ShopPage.includes(item)&&guard-->0)key('f');const slot=JSON.parse(mcGameExport()).state.Town.ShopPage.indexOf(item);assert(slot>=0);key('p');key(String(slot+1))}
  key('l');key('Enter');
 }
 assert.equal(frame.party[0].item_count,3);assert.equal(frame.party[0].items[0].item_index,1);
 const bought=mcGameExport();assert.equal(mcGameImport(bought),'');key('Enter');key('e');key('m');assert.equal(frame.phase,2);
 key('e');for(let i=0;i<100&&JSON.parse(mcGameExport()).state.Town.InputMode!==0;i++){const t=JSON.parse(mcGameExport()).state.Town;key(t.EquipChoices?.length?'1':'Enter')}
 assert(frame.party[0].items[0].equipped);assert(frame.party[0].AC<10);console.log('PASS: bought and equipped party gear after loading a save');
 key('l');assert.equal(frame.phase,3);const stairsSave=mcGameExport();key('n');key('m');assert.match(screenText(),/PRESSANYKEY/);key('x');
 const dungeonSave=mcGameExport();mcGameInit('1');assert.equal(frame.phase,3);assert.equal(mcGameImport(dungeonSave),'');
 console.log('PASS: dungeon, map, and reload persistence');
 assert.equal(mcGameImport(stairsSave),'');key('y'); // return to castle
 assert.equal(frame.phase,1);key('a');key('4');console.log('INN\n'+screenText());key('a');
 await new Promise(r=>setTimeout(r,1600));console.log('REST\n'+screenText());
 for(let i=0;i<8;i++)key('Enter');
 fs.writeFileSync('/tmp/mc-party-save.json',mcGameExport());
 console.log('PARTY SPELLS',frame.party.map(c=>[c.Name,c.MageSpells,c.PriestSpells]));
 console.log('PASS: inn timer completed');process.exit(0)
})().catch(e=>{console.error(e);console.log(screenText());process.exit(1)});
