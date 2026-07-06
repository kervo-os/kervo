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
  --bg:#0a0c10;--panel:#101318;--card:#14181e;--card2:#171c23;--line:#242b35;--line2:#2f3846;
  --fg:#e9ecf1;--muted:#8b93a1;--faint:#5c6470;
  --v:#34d399;--o:#60a5fa;--g:#f5a623;--s:#6b7280;--d:#f87171;
  --accent:linear-gradient(135deg,#34d399,#22d3ee);
}
@media(prefers-color-scheme:light){:root{--bg:#f6f7f9;--panel:#fff;--card:#fff;--card2:#fbfcfd;
  --line:#e4e7ec;--line2:#d5dae2;--fg:#171b22;--muted:#667085;--faint:#98a2b3}}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--fg);
  font:14px/1.55 -apple-system,ui-sans-serif,system-ui,"Segoe UI",sans-serif;
  background-image:radial-gradient(1100px 480px at 50% -12%,rgba(52,211,153,.07),transparent 60%);
  background-repeat:no-repeat}
header{position:sticky;top:0;z-index:5;display:flex;align-items:center;gap:1rem;padding:.75rem 1.4rem;
  background:color-mix(in srgb,var(--bg) 78%,transparent);backdrop-filter:blur(12px);
  border-bottom:1px solid var(--line)}
.mark{width:22px;height:22px;border-radius:6px;background:var(--accent);
  box-shadow:0 0 14px rgba(52,211,153,.35)}
.brand{font-weight:750;letter-spacing:-.02em;font-size:.95rem}
.brand em{font-style:normal;color:var(--muted);font-weight:500}
.spacer{flex:1}
.progress{width:150px;height:4px;border-radius:2px;background:var(--line);overflow:hidden}
.progress i{display:block;height:100%;width:0;background:var(--accent);transition:width .3s}
kbd{border:1px solid var(--line2);border-bottom-width:2px;border-radius:5px;padding:.04em .38em;
  font:.74em ui-monospace,monospace;color:var(--muted);background:var(--panel)}
button{border:1px solid var(--line2);background:var(--card);color:var(--fg);border-radius:8px;
  padding:.42rem .85rem;font:inherit;font-size:.85rem;cursor:pointer;transition:.15s}
button:hover{border-color:var(--muted)}
button.primary{background:var(--accent);color:#07130e;font-weight:700;border:none;
  box-shadow:0 2px 12px rgba(52,211,153,.25)}
main{max-width:66rem;margin:0 auto;padding:1.6rem 1.4rem 3rem}
.page-head{display:flex;align-items:baseline;gap:1rem;margin:0 0 1.3rem}
.page-head h1{margin:0;font-size:1.5rem;letter-spacing:-.03em}
.page-head .sub{color:var(--muted);font-size:.86rem}
.hint{margin-left:auto;color:var(--faint);font-size:.78rem}
.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(21rem,1fr));gap:1rem}
.repo{position:relative;background:linear-gradient(180deg,var(--card2),var(--card));
  border:1px solid var(--line);border-radius:14px;padding:1.15rem 1.25rem;cursor:pointer;
  transition:transform .16s,box-shadow .16s,border-color .16s}
.repo:hover{transform:translateY(-2px);border-color:rgba(52,211,153,.4);
  box-shadow:0 12px 32px rgba(0,0,0,.3)}
.repo .key{position:absolute;top:.85rem;right:.9rem}
.rhead{display:flex;align-items:center;gap:.7rem;margin-bottom:.35rem}
.ava{width:36px;height:36px;border-radius:10px;display:flex;align-items:center;justify-content:center;
  font-weight:800;font-size:1rem;color:#fff;text-shadow:0 1px 2px rgba(0,0,0,.25);flex:none}
.rname{font-size:1.04rem;font-weight:700;letter-spacing:-.01em;display:flex;align-items:center;gap:.45rem}
.chip{font-size:.66rem;color:var(--muted);border:1px solid var(--line2);border-radius:5px;padding:.02rem .35rem;font-weight:600}
.path{color:var(--muted);font:.76rem ui-monospace,monospace;margin:.1rem 0 .85rem;
  overflow:hidden;text-overflow:ellipsis;white-space:nowrap;direction:rtl;text-align:left}
.path bdi{unicode-bidi:plaintext}
.attn{display:flex;align-items:baseline;gap:.45rem;margin-bottom:.7rem}
.attn .n{font-size:1.5rem;font-weight:800;font-variant-numeric:tabular-nums;letter-spacing:-.03em}
.attn.hot .n{color:var(--g)}
.attn.ok .n{color:var(--v)}
.attn .lbl{color:var(--muted);font-size:.8rem}
.bar{display:flex;height:6px;border-radius:3px;overflow:hidden;background:var(--line);margin:0 0 .5rem}
.bar i{display:block;height:100%}
.legend{color:var(--muted);font-size:.72rem;display:flex;flex-wrap:wrap;gap:.15rem .8rem}
.legend b{font-weight:650;font-variant-numeric:tabular-nums;color:var(--fg)}
.dot{display:inline-block;width:7px;height:7px;border-radius:99px;margin-right:.3rem;vertical-align:1px}
.foot{display:flex;justify-content:space-between;color:var(--faint);font-size:.73rem;margin-top:.7rem;
  padding-top:.65rem;border-top:1px solid var(--line);font-variant-numeric:tabular-nums}
.pulse{color:var(--v)}
/* triage */
#triage{display:none;max-width:52rem;margin:0 auto}
.tri-head{display:flex;align-items:center;gap:.8rem;margin-bottom:1rem}
.tri-head h2{margin:0;font-size:1.05rem;letter-spacing:-.02em}
.tri-head .sub{color:var(--muted);font-size:.8rem;font-variant-numeric:tabular-nums}
.item{background:linear-gradient(180deg,var(--card2),var(--card));border:1px solid var(--line);
  border-radius:16px;padding:1.3rem 1.5rem;animation:pop .18s ease-out;box-shadow:0 10px 30px rgba(0,0,0,.2)}
@keyframes pop{from{opacity:0;transform:translateY(6px)}to{opacity:1;transform:none}}
.meta{display:flex;flex-wrap:wrap;gap:.5rem;align-items:center;margin-bottom:.85rem;font-size:.76rem;color:var(--muted)}
.tag{border-radius:5px;padding:.07rem .5rem;font-weight:700;font-size:.7rem;letter-spacing:.02em;flex:none}
.tag.decision{background:#8b5cf626;color:#a78bfa}.tag.risk{background:#f8717126;color:#f87171}
.tag.summary{background:#60a5fa26;color:#60a5fa}.tag.goal{background:#34d39926;color:#34d399}
.tag.note,.tag.correction{background:#6b728026;color:#9aa1ac}
.conf{color:var(--d);font-weight:700}
.mono{font-family:ui-monospace,monospace;font-size:.72rem}
.body{white-space:pre-wrap;font-size:.98rem;line-height:1.75;max-width:64ch;margin-bottom:1rem}
.evid{font:.8rem/1.6 ui-monospace,monospace;color:var(--muted);border-left:3px solid var(--v);
  padding:.2rem 0 .2rem .75rem;margin-bottom:1.1rem;white-space:pre-wrap;opacity:.9}
.actions{display:flex;gap:.45rem;flex-wrap:wrap;align-items:center}
.actions input{flex:1;min-width:11rem;background:var(--panel);color:var(--fg);border:1px solid var(--line2);
  border-radius:8px;padding:.48rem .65rem;font:inherit;font-size:.85rem}
.actions input:focus{outline:none;border-color:var(--v)}
.actions button b{opacity:.5;font-weight:600;margin-right:.32rem}
.bv{border-color:rgba(52,211,153,.5)!important;color:var(--v)}
.bs{border-color:rgba(245,166,35,.5)!important;color:var(--g)}
.bd{border-color:rgba(248,113,113,.5)!important;color:var(--d)}
.rail{margin-top:1.15rem;display:flex;flex-direction:column;gap:.2rem}
.rail-label{color:var(--faint);font-size:.7rem;text-transform:uppercase;letter-spacing:.1em;margin:.25rem 0 .3rem}
.rail .row{display:flex;gap:.6rem;align-items:center;padding:.42rem .7rem;border-radius:8px;font-size:.8rem;
  color:var(--muted);cursor:pointer;border-left:3px solid transparent;transition:.12s}
.rail .row:hover{background:var(--card)}
.rail .row.cur{background:var(--card);color:var(--fg);border-left-color:var(--v)}
.rail .idx{color:var(--faint);font:.72rem ui-monospace,monospace;width:1.3rem;text-align:right;flex:none}
.rail .txt{flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;min-width:0}
#toast{position:fixed;left:50%;bottom:1.4rem;transform:translateX(-50%) translateY(20px);opacity:0;
  background:var(--card);border:1px solid var(--line2);border-radius:10px;padding:.55rem 1rem;font-size:.85rem;
  transition:.25s;pointer-events:none;box-shadow:0 10px 30px rgba(0,0,0,.4)}
#toast.show{opacity:1;transform:translateX(-50%)}
#help{position:fixed;inset:0;background:rgba(0,0,0,.55);display:none;align-items:center;justify-content:center;z-index:9;backdrop-filter:blur(3px)}
#help .box{background:var(--panel);border:1px solid var(--line2);border-radius:14px;padding:1.4rem 1.8rem;min-width:21rem}
#help h3{margin:0 0 .8rem}
#help div{display:flex;justify-content:space-between;gap:2rem;padding:.2rem 0;color:var(--muted);font-size:.86rem}
.empty{color:var(--muted);text-align:center;padding:3rem 0}
</style></head><body>
<header>
  <div class="mark"></div>
  <div class="brand">kervo <em>dash</em></div>
  <div class="spacer"></div>
  <div class="progress"><i id="pbar"></i></div>
  <span class="hint"><kbd>?</kbd> keys</span>
  <button class="primary" onclick="finish()">Finish</button>
</header>
<main>
  <section id="fleet">
    <div class="page-head"><h1>Workspaces</h1><div class="sub" id="totals"></div>
      <div class="hint">everything below is local — no server, no account</div></div>
    <div class="grid" id="grid"></div>
  </section>
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

function hue(s){ let h=0; for(const c of s) h=(h*31+c.codePointAt(0))>>>0; return h%360 }
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
    if(i<9){ const k=el("span","key"); k.append(el("kbd","",String(i+1))); c.append(k) }
    const head = el("div","rhead");
    const av = el("div","ava",(r.Name[0]||"?").toUpperCase());
    const H = hue(r.Name);
    av.style.background = "linear-gradient(135deg,hsl("+H+" 62% 46%),hsl("+((H+42)%360)+" 66% 56%))";
    const nm = el("div","rname"); nm.append(el("span","",r.Name), el("span","chip",r.Lang));
    head.append(av, nm); c.append(head);
    const path = el("div","path"); path.append(el("bdi","",r.DisplayPath||r.Path)); c.append(path);
    const pend = r.Items.length;
    const attn = el("div","attn "+(pend?"hot":"ok"));
    attn.append(el("span","n", pend? String(pend) : "✓"),
                el("span","lbl", pend? "awaiting judgment" : "all clear"));
    c.append(attn);
    const totalObs = STATES.reduce((n,st)=>n+(r.Counts[st]||0),0);
    if(totalObs){
      const bar = el("div","bar");
      STATES.forEach(st=>{ const n=r.Counts[st]||0; if(!n)return;
        const seg=el("i"); seg.style.width=(100*n/totalObs)+"%"; seg.style.background=SC[st]; bar.append(seg) });
      c.append(bar);
      const lg = el("div","legend");
      STATES.forEach(st=>{ const n=r.Counts[st]||0; if(!n)return;
        const it=el("span"); const dot=el("span","dot"); dot.style.background=SC[st];
        it.append(dot, el("b","",String(n)), document.createTextNode(" "+st)); lg.append(it) });
      c.append(lg);
    }
    const foot = el("div","foot");
    const fresh = r.LastEvent && (Date.now()-Date.parse(r.LastEvent))<36e5;
    foot.append(el("span","",r.Events+" events"),
                el("span",fresh?"pulse":"", (fresh?"● ":"")+rel(r.LastEvent)));
    c.append(foot);
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
  const card = document.getElementById("tcard"); card.textContent="";
  if(!repo.Items.length){
    document.getElementById("tsub").textContent = "all clear";
    card.append(el("div","empty","All judged here — Esc back to the fleet."));
    document.getElementById("rail").textContent=""; return }
  if(idx>=repo.Items.length) idx=repo.Items.length-1;
  document.getElementById("tsub").textContent = (idx+1)+" / "+repo.Items.length+" · j/k to move";
  const o = repo.Items[idx];
  const c = el("div","item");
  const m = el("div","meta");
  m.append(el("span","tag "+o.Type,o.Type), el("span","mono",o.ShortID), el("span","",o.State), el("span","",o.Actor));
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
  rail.append(el("div","rail-label","queue"));
  repo.Items.forEach((it,i)=>{
    const row = el("div","row"+(i===idx?" cur":""));
    row.append(el("span","idx",String(i+1)), el("span","tag "+it.Type,it.Type), el("span","txt",it.Body));
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
// Deep link: #2 opens the second repo — refresh keeps your place.
const h = location.hash.match(/^#(\d+)$/);
if(h && FLEET[+h[1]-1]) openRepo(+h[1]-1);
</script>
</body></html>`))
