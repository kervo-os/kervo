package cli

import "html/template"

// dashTmpl is the fleet page. All dynamic data arrives as one JSON blob
// (json.Marshal escapes <>&, so it cannot break out of the script tag) and
// every data-derived DOM node is written via textContent — never innerHTML.
// Zero dependencies by design guarantee: no CDN, no framework, no build.
var dashTmpl = template.Must(template.New("dash").Parse(`<!doctype html>
<html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>kervo dash</title>
<style>
:root{
  --bg:#0b0d10;--panel:#12151a;--card:#171b21;--line:#262c35;--fg:#e7eaee;--muted:#8b93a1;
  --v:#34d399;--o:#60a5fa;--g:#f5a623;--s:#6b7280;--d:#f87171;
  --accent:linear-gradient(135deg,#34d399,#22d3ee);
}
@media(prefers-color-scheme:light){:root{--bg:#f7f8f9;--panel:#fff;--card:#fff;--line:#e3e6ea;--fg:#181c22;--muted:#68707d}}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--fg);font:14px/1.55 ui-sans-serif,system-ui,-apple-system,sans-serif}
header{position:sticky;top:0;z-index:5;display:flex;align-items:center;gap:1rem;padding:.7rem 1.2rem;
  background:color-mix(in srgb,var(--bg) 80%,transparent);backdrop-filter:blur(10px);border-bottom:1px solid var(--line)}
.brand{font-weight:700;letter-spacing:-.02em}
.brand em{font-style:normal;background:var(--accent);-webkit-background-clip:text;background-clip:text;color:transparent}
.htotals{color:var(--muted);font-size:.82rem}
.spacer{flex:1}
.progress{width:160px;height:4px;border-radius:2px;background:var(--line);overflow:hidden}
.progress i{display:block;height:100%;width:0;background:var(--accent);transition:width .3s}
kbd{border:1px solid var(--line);border-bottom-width:2px;border-radius:4px;padding:0 .32em;font:.78em ui-monospace,monospace;color:var(--muted)}
button{border:1px solid var(--line);background:var(--card);color:var(--fg);border-radius:8px;padding:.4rem .8rem;font:inherit;font-size:.85rem;cursor:pointer;transition:.15s}
button:hover{border-color:var(--muted)}
button.primary{background:var(--accent);color:#08110d;font-weight:650;border:none}
main{max-width:64rem;margin:0 auto;padding:1.2rem}
.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(17rem,1fr));gap:.9rem}
.repo{background:var(--card);border:1px solid var(--line);border-radius:12px;padding:1rem;cursor:pointer;
  transition:transform .15s,box-shadow .15s,border-color .15s}
.repo:hover{transform:translateY(-2px);border-color:#3a4250;box-shadow:0 8px 24px rgba(0,0,0,.25)}
.repo h3{margin:0;font-size:.98rem;display:flex;align-items:center;gap:.5rem}
.repo .path{color:var(--muted);font:.72rem ui-monospace,monospace;margin:.15rem 0 .7rem;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.badge{margin-left:auto;background:var(--g);color:#161104;font-size:.72rem;font-weight:700;border-radius:99px;padding:.05rem .5rem}
.badge.zero{background:var(--line);color:var(--muted)}
.chip{font-size:.68rem;color:var(--muted);border:1px solid var(--line);border-radius:5px;padding:0 .35rem}
.bar{display:flex;height:6px;border-radius:3px;overflow:hidden;background:var(--line);margin:.5rem 0 .4rem}
.bar i{display:block;height:100%}
.legend,.last{color:var(--muted);font-size:.72rem}
.dot{display:inline-block;width:7px;height:7px;border-radius:99px;margin:0 .25rem 0 .6rem}
.dot:first-child{margin-left:0}
/* triage */
#triage{display:none}
.tri-head{display:flex;align-items:center;gap:.8rem;margin-bottom:1rem}
.tri-head h2{margin:0;font-size:1.05rem}
.tri-head .sub{color:var(--muted);font-size:.8rem}
.item{background:var(--card);border:1px solid var(--line);border-radius:14px;padding:1.2rem 1.3rem;
  animation:pop .18s ease-out}
@keyframes pop{from{opacity:0;transform:translateY(6px)}to{opacity:1;transform:none}}
.meta{display:flex;flex-wrap:wrap;gap:.4rem;align-items:center;margin-bottom:.8rem;font-size:.75rem;color:var(--muted)}
.tag{border-radius:5px;padding:.06rem .45rem;font-weight:650;font-size:.72rem}
.tag.decision{background:#8b5cf633;color:#a78bfa}.tag.risk{background:#f8717133;color:#f87171}
.tag.summary{background:#60a5fa33;color:#60a5fa}.tag.goal{background:#34d39933;color:#34d399}
.tag.note,.tag.correction{background:#6b728033;color:#9aa1ac}
.conf{color:var(--d);font-weight:700}
.body{white-space:pre-wrap;font-size:.95rem;margin-bottom:.8rem}
.evid{font:.8rem ui-monospace,monospace;color:var(--muted);border-left:3px solid var(--line);padding:.15rem 0 .15rem .7rem;margin-bottom:1rem;white-space:pre-wrap}
.actions{display:flex;gap:.45rem;flex-wrap:wrap;align-items:center}
.actions input{flex:1;min-width:11rem;background:var(--panel);color:var(--fg);border:1px solid var(--line);border-radius:8px;padding:.45rem .6rem;font:inherit;font-size:.85rem}
.actions button b{opacity:.55;font-weight:600;margin-right:.3rem}
.bv{border-color:var(--v)!important;color:var(--v)}.bs{border-color:var(--g)!important;color:var(--g)}
.bd{border-color:var(--d)!important;color:var(--d)}
.rail{margin-top:1rem;display:flex;flex-direction:column;gap:.3rem}
.rail .row{display:flex;gap:.6rem;align-items:center;padding:.35rem .6rem;border-radius:8px;font-size:.8rem;color:var(--muted);cursor:pointer}
.rail .row.cur{background:var(--card);color:var(--fg);border:1px solid var(--line)}
.rail .row .st{margin-left:auto;font-size:.72rem;font-weight:650}
#toast{position:fixed;left:50%;bottom:1.4rem;transform:translateX(-50%) translateY(20px);opacity:0;
  background:var(--card);border:1px solid var(--line);border-radius:10px;padding:.5rem .9rem;font-size:.85rem;
  transition:.25s;pointer-events:none;box-shadow:0 8px 24px rgba(0,0,0,.35)}
#toast.show{opacity:1;transform:translateX(-50%)}
#help{position:fixed;inset:0;background:rgba(0,0,0,.55);display:none;align-items:center;justify-content:center;z-index:9}
#help .box{background:var(--panel);border:1px solid var(--line);border-radius:14px;padding:1.4rem 1.8rem;min-width:20rem}
#help h3{margin:0 0 .8rem}
#help div{display:flex;justify-content:space-between;gap:2rem;padding:.18rem 0;color:var(--muted);font-size:.86rem}
.empty{color:var(--muted);text-align:center;padding:3rem 0}
</style></head><body>
<header>
  <div class="brand">kervo <em>dash</em></div>
  <div class="htotals" id="totals"></div>
  <div class="spacer"></div>
  <div class="progress"><i id="pbar"></i></div>
  <span class="htotals"><kbd>?</kbd> keys</span>
  <button class="primary" onclick="finish()">Finish</button>
</header>
<main>
  <section id="fleet"><div class="grid" id="grid"></div></section>
  <section id="triage">
    <div class="tri-head">
      <button onclick="showFleet()">← fleet <kbd>Esc</kbd></button>
      <h2 id="tname"></h2><div class="sub" id="tsub"></div>
    </div>
    <div id="tcard"></div>
    <div class="rail" id="rail"></div>
  </section>
</main>
<div id="toast"></div>
<div id="help"><div class="box"><h3>Keys</h3>
  <div><span>open repo</span><span><kbd>1</kbd>–<kbd>9</kbd></span></div>
  <div><span>next / prev item</span><span><kbd>j</kbd> / <kbd>k</kbd></span></div>
  <div><span>verify · stale · deprecate</span><span><kbd>v</kbd> <kbd>s</kbd> <kbd>d</kbd></span></div>
  <div><span>skip</span><span><kbd>x</kbd></span></div>
  <div><span>reason field</span><span><kbd>r</kbd></span></div>
  <div><span>back / close</span><span><kbd>Esc</kbd></span></div>
</div></div>
<script>
const FLEET = {{.FleetJS}};
FLEET.forEach(r=>{ r.Items = r.Items || []; r.Counts = r.Counts || {} });
const STATES = ["verified","observed","generated","stale","deprecated"];
const SC = {verified:"var(--v)",observed:"var(--o)",generated:"var(--g)",stale:"var(--s)",deprecated:"var(--d)"};
let repo = null, idx = 0, judged = 0;
const total0 = FLEET.reduce((n,r)=>n+r.Items.length,0);
const el = (t,cls,txt)=>{const e=document.createElement(t); if(cls)e.className=cls; if(txt!==undefined)e.textContent=txt; return e};

function rel(iso){ if(!iso) return "empty ledger";
  const s=(Date.now()-Date.parse(iso))/1e3;
  if(s<90) return "just now"; if(s<5400) return Math.round(s/60)+"m ago";
  if(s<129600) return Math.round(s/3600)+"h ago"; return Math.round(s/86400)+"d ago" }

function totals(){
  const pend = FLEET.reduce((n,r)=>n+r.Items.length,0);
  document.getElementById("totals").textContent =
    FLEET.length+" workspaces · "+pend+" awaiting judgment";
  document.getElementById("pbar").style.width = total0? (100*judged/total0)+"%" : "100%";
}

function renderFleet(){
  const g = document.getElementById("grid"); g.textContent = "";
  FLEET.forEach((r,i)=>{
    const c = el("div","repo");
    const h = el("h3"); h.append(el("span","", (i<9? (i+1)+"  " : "")+r.Name), el("span","chip",r.Lang));
    const b = el("span", "badge"+(r.Items.length?"":" zero"), r.Items.length? r.Items.length+" pending":"clear");
    h.append(b); c.append(h);
    c.append(el("div","path",r.Path));
    const bar = el("div","bar");
    const totalObs = STATES.reduce((n,st)=>n+(r.Counts[st]||0),0)||1;
    STATES.forEach(st=>{ const n=r.Counts[st]||0; if(!n)return;
      const seg=el("i"); seg.style.width=(100*n/totalObs)+"%"; seg.style.background=SC[st]; bar.append(seg) });
    c.append(bar);
    const lg = el("div","legend");
    STATES.forEach(st=>{ const n=r.Counts[st]||0; if(!n)return;
      const dot=el("span","dot"); dot.style.background=SC[st]; lg.append(dot, document.createTextNode(n+" "+st)) });
    c.append(lg);
    c.append(el("div","last", r.Events+" events · "+rel(r.LastEvent)));
    c.onclick = ()=>openRepo(i);
    g.append(c);
  });
  totals();
}

function openRepo(i){ repo = FLEET[i]; idx = 0;
  document.getElementById("fleet").style.display="none";
  document.getElementById("triage").style.display="block";
  renderTriage() }

function showFleet(){ repo=null;
  document.getElementById("triage").style.display="none";
  document.getElementById("fleet").style.display="block";
  renderFleet() }

function renderTriage(){
  document.getElementById("tname").textContent = repo.Name;
  document.getElementById("tsub").textContent = repo.Items.length+" awaiting";
  const card = document.getElementById("tcard"); card.textContent="";
  if(!repo.Items.length){ card.append(el("div","empty","All judged here — Esc back to the fleet."));
    document.getElementById("rail").textContent=""; return }
  if(idx>=repo.Items.length) idx=repo.Items.length-1;
  const o = repo.Items[idx];
  const c = el("div","item");
  const m = el("div","meta");
  m.append(el("span","tag "+o.Type,o.Type), el("span","",o.ShortID), el("span","",o.State), el("span","",o.Actor));
  if(o.Conflict) m.append(el("span","conf","⚠ conflict"));
  c.append(m, el("div","body",o.Body));
  if(o.Evidence) c.append(el("div","evid","evidence: "+o.Evidence));
  const a = el("div","actions");
  const reason = el("input"); reason.id="reason"; reason.placeholder="reason (optional) — r to focus";
  const mk=(cls,key,label,fn)=>{const b=el("button",cls); b.append(el("b","",key),document.createTextNode(label)); b.onclick=fn; return b};
  a.append(reason,
    mk("bv","v","verify",()=>judge("verified")),
    mk("bs","s","stale",()=>judge("stale")),
    mk("bd","d","deprecate",()=>judge("deprecated")),
    mk("","x","skip",skip));
  c.append(a); card.append(c);
  const rail = document.getElementById("rail"); rail.textContent="";
  repo.Items.forEach((it,i)=>{
    const row = el("div","row"+(i===idx?" cur":""));
    row.append(el("span","tag "+it.Type,it.Type), el("span","",it.Body.slice(0,80)));
    row.onclick=()=>{idx=i; renderTriage()};
    rail.append(row);
  });
}

async function judge(to){
  const o = repo.Items[idx]; if(!o) return;
  const reason = document.getElementById("reason")?.value || "";
  const res = await fetch("/judge",{method:"POST",headers:{"Content-Type":"application/json"},
    body:JSON.stringify({Workspace:repo.Path,ID:o.ID,To:to,Reason:reason})});
  if(!res.ok){ toast(await res.text(), "var(--d)"); return }
  repo.Items.splice(idx,1); repo.Counts[o.State]--; repo.Counts[to]=(repo.Counts[to]||0)+1; judged++;
  toast(o.ShortID+" → "+to, SC[to]); totals(); renderTriage();
}
function skip(){ if(idx<repo.Items.length-1) idx++; renderTriage() }

let tmr;
function toast(msg,color){ const t=document.getElementById("toast");
  t.textContent=msg; t.style.borderColor=color||"var(--line)"; t.classList.add("show");
  clearTimeout(tmr); tmr=setTimeout(()=>t.classList.remove("show"),1800) }

function finish(){ fetch("/quit",{method:"POST"}).finally(()=>{
  document.body.innerHTML="<div class='empty' style='padding-top:5rem'><h2>Done — "+judged+" judged.</h2>"+
  "<p>Run <code>kervo compile</code> in the affected repos, then close this tab.</p></div>"}) }

document.addEventListener("keydown",e=>{
  const typing = e.target.tagName==="INPUT";
  if(e.key==="Escape"){ typing? e.target.blur() : (document.getElementById("help").style.display="none", repo&&showFleet()); return }
  if(typing) return;
  if(e.key==="?"){ const h=document.getElementById("help"); h.style.display=h.style.display==="flex"?"none":"flex"; return }
  if(!repo){ const n=parseInt(e.key); if(n>=1&&n<=Math.min(9,FLEET.length)) openRepo(n-1); return }
  if(e.key==="j") skip();
  else if(e.key==="k"){ if(idx>0){idx--; renderTriage()} }
  else if(e.key==="v") judge("verified");
  else if(e.key==="s") judge("stale");
  else if(e.key==="d") judge("deprecated");
  else if(e.key==="x") skip();
  else if(e.key==="r"){ e.preventDefault(); document.getElementById("reason")?.focus() }
});
renderFleet();
</script>
</body></html>`))
