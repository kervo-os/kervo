package cli

import "html/template"

// dashTmpl is the fleet page. All dynamic data arrives as one JSON blob
// (json.Marshal escapes <>&, so it cannot break out of the script tag) and
// every data-derived DOM node is written via textContent — never innerHTML.
// Zero dependencies by design guarantee: no CDN, no framework, no build.
var dashTmpl = template.Must(template.New("dash").Parse(`<!doctype html>
<html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>kervo dash</title>
<link rel="icon" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 512 512'%3E%3Crect width='512' height='512' rx='112' fill='%231d2b3a'/%3E%3Crect x='96' y='120' width='34' height='272' rx='17' fill='%232ab7a9'/%3E%3Cg fill='%23eef2f6'%3E%3Crect x='170' y='104' width='92' height='304' rx='24'/%3E%3Cg transform='rotate(-46 262 256)'%3E%3Crect x='236' y='210' width='210' height='92' rx='24'/%3E%3C/g%3E%3Cg transform='rotate(46 262 256)'%3E%3Crect x='236' y='210' width='210' height='92' rx='24'/%3E%3C/g%3E%3C/g%3E%3Ccircle cx='400' cy='400' r='74' fill='%23f5a623'/%3E%3Cpath d='M368 402 l22 22 44 -50' fill='none' stroke='%231d2b3a' stroke-width='20' stroke-linecap='round' stroke-linejoin='round'/%3E%3C/svg%3E">
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
select{border:1px solid var(--line2);background:var(--card);color:var(--fg);border-radius:8px;
  padding:.4rem .5rem;font:inherit;font-size:.82rem;cursor:pointer}
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
#triage{display:none;grid-template-columns:minmax(0,1fr) 21rem;gap:1.3rem}
#triage.open{display:grid}
.tri-head{grid-column:1/-1;display:flex;align-items:center;gap:.8rem;margin-bottom:0}
.tri-main{min-width:0}
#overview{display:flex;flex-direction:column;gap:.8rem}
.ov{background:var(--card);border:1px solid var(--line);border-radius:12px;padding:.85rem 1rem}
.ov h4{margin:0 0 .5rem;font-size:.72rem;color:var(--faint);text-transform:uppercase;letter-spacing:.09em;font-weight:650}
.ov .li{display:flex;gap:.6rem;font-size:.78rem;padding:.14rem 0;color:var(--muted);min-width:0}
.ov .li .m{font-family:ui-monospace,monospace;color:var(--fg);flex:none;max-width:12rem;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.ov .li .t{flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;min-width:0}
.ov .chips{display:flex;flex-wrap:wrap;gap:.35rem}
.ov .more{color:var(--faint);font-size:.72rem;padding-top:.2rem}
.ov .link{display:flex;gap:.5rem;align-items:center;font-size:.78rem;padding:.14rem 0;color:var(--muted)}
.ov .link b{color:var(--fg);font-weight:600}
.ov .link .n{margin-left:auto;color:var(--faint);font-variant-numeric:tabular-nums}
@media(max-width:960px){#triage{grid-template-columns:1fr}}
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
.body-title{font-size:1.05rem;font-weight:650;letter-spacing:-.01em;line-height:1.5;max-width:60ch;margin-bottom:.45rem}
.body{white-space:pre-wrap;font-size:.94rem;line-height:1.75;max-width:64ch;margin-bottom:1rem;color:color-mix(in srgb,var(--fg) 82%,var(--muted))}
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
.rail .row.done{cursor:default;opacity:.85}
.rail .row.done:hover{background:transparent}
.rail .st{font-size:.7rem;font-weight:700;width:5.2rem;flex:none;font-variant-numeric:tabular-nums}
.rail .why{color:var(--faint);font-size:.74rem;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;max-width:18rem}
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
  <svg viewBox="0 0 512 512" width="22" height="22" style="border-radius:6px;flex:none" aria-hidden="true"><rect width="512" height="512" rx="112" fill="#1d2b3a"/><rect x="96" y="120" width="34" height="272" rx="17" fill="#2ab7a9"/><g fill="#eef2f6"><rect x="170" y="104" width="92" height="304" rx="24"/><g transform="rotate(-46 262 256)"><rect x="236" y="210" width="210" height="92" rx="24"/></g><g transform="rotate(46 262 256)"><rect x="236" y="210" width="210" height="92" rx="24"/></g></g><circle cx="400" cy="400" r="74" fill="#f5a623"/><path d="M368 402 l22 22 44 -50" fill="none" stroke="#1d2b3a" stroke-width="20" stroke-linecap="round" stroke-linejoin="round"/></svg>
  <div class="brand">kervo <em>dash</em></div>
  <div class="spacer"></div>
  <div class="progress"><i id="pbar"></i></div>
  <span class="hint"><kbd>?</kbd> <span id="keysHint"></span></span>
  <select id="langSel" onchange="setLang(this.value)">
    <option value="en">English</option><option value="ko">한국어</option><option value="ja">日本語</option>
  </select>
  <button class="primary" id="finishBtn" onclick="finish()"></button>
</header>
<main>
  <section id="fleet">
    <div class="page-head"><h1 id="wtitle"></h1><div class="sub" id="totals"></div>
      <div class="hint" id="localnote"></div></div>
    <div class="grid" id="grid"></div>
  </section>
  <section id="triage">
    <div class="tri-head">
      <button onclick="showFleet()"><span id="backLbl"></span> <kbd>Esc</kbd></button>
      <h2 id="tname"></h2><div class="sub" id="tsub"></div>
    </div>
    <div class="tri-main">
      <div id="tcard"></div>
      <div class="rail" id="rail"></div>
    </div>
    <aside id="overview"></aside>
  </section>
</main>
<div id="toast"></div>
<div id="help"><div class="box"><h3 id="helpTitle"></h3><div id="helpRows"></div></div></div>
<script>
const FLEET = {{.FleetJS}};
const TT = {{.TTJS}};
let LANG = "{{.Lang}}";
let T = TT[LANG] || TT.en;
// %[1]d-style verbs come straight from the Go string tables.
const F = (s,...a)=>s.replace(/%\[(\d+)\]d/g,(_,n)=>a[n-1]);
FLEET.forEach(r=>{ r.Items = r.Items || []; r.History = r.History || []; r.Counts = r.Counts || {} });
const STATES = ["verified","observed","generated","stale","deprecated"];
const SC = {verified:"var(--v)",observed:"var(--o)",generated:"var(--g)",stale:"var(--s)",deprecated:"var(--d)"};
let repo = null, idx = 0, judged = 0;
const total0 = FLEET.reduce((n,r)=>n+r.Items.length,0);
const el = (t,cls,txt)=>{const e=document.createElement(t); if(cls)e.className=cls; if(txt!==undefined)e.textContent=txt; return e};

function hue(s){ let h=0; for(const c of s) h=(h*31+c.codePointAt(0))>>>0; return h%360 }
function rel(iso){ if(!iso) return T.emptyledger;
  const s=(Date.now()-Date.parse(iso))/1e3;
  if(s<90) return T.justnow; if(s<5400) return F(T.minago,Math.round(s/60));
  if(s<129600) return F(T.hourago,Math.round(s/3600)); return F(T.dayago,Math.round(s/86400)) }

function totals(){
  const pend = FLEET.reduce((n,r)=>n+r.Items.length,0);
  document.getElementById("totals").textContent = F(T.totals, FLEET.length, pend);
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
                el("span","lbl", pend? T.awaiting : T.clear));
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
    foot.append(el("span","",F(T.events,r.Events)),
                el("span",fresh?"pulse":"", (fresh?"● ":"")+rel(r.LastEvent)));
    c.append(foot);
    c.onclick = ()=>openRepo(i);
    g.append(c);
  });
  totals();
}

function openRepo(i){ repo = FLEET[i]; idx = 0;
  document.getElementById("fleet").style.display="none";
  document.getElementById("triage").classList.add("open");
  renderOverview(); renderTriage() }

function showFleet(){ repo=null;
  document.getElementById("triage").classList.remove("open");
  document.getElementById("fleet").style.display="block";
  renderFleet() }

// The workspace's fact skeleton — same deterministic scan compile runs,
// capped for reading. Coupling pairs are proven by commit history.
function renderOverview(){
  const aside = document.getElementById("overview"); aside.textContent="";
  const ov = repo.Overview; if(!ov) return;
  const box = (title)=>{ const b=el("div","ov"); b.append(el("h4","",title)); aside.append(b); return b };

  const info = box(T.overview);
  const chips = el("div","chips");
  const seen = new Set();
  const chip = t=>{ if(!t||seen.has(t)) return; seen.add(t); chips.append(el("span","chip",t)) };
  chip(ov.Branch||"?");
  (ov.Languages||[]).forEach(chip);
  (ov.Frameworks||[]).forEach(chip);
  if(ov.Partial) chip(T.partialscan);
  info.append(chips);

  if((ov.Commands||[]).length){
    const b = box(T.commands);
    ov.Commands.forEach(c=>{ const r=el("div","li"); r.append(el("span","m",c.Run), el("span","t",c.Source)); b.append(r) });
    if(ov.TotalCommands>ov.Commands.length) b.append(el("div","more",F(T.more,ov.TotalCommands-ov.Commands.length)));
  }
  if((ov.Links||[]).length){
    const b = box(T.links);
    ov.Links.forEach(l=>{ const r=el("div","link");
      const pair=el("span"); pair.append(el("b","",l.A), document.createTextNode(" ↔ "), el("b","",l.B));
      r.append(pair, el("span","n","×"+l.N)); b.append(r) });
  }
  if((ov.Commits||[]).length){
    const b = box(T.recent);
    ov.Commits.forEach(c=>{ const r=el("div","li"); r.append(el("span","m",c.Date), el("span","t",c.Subject)); b.append(r) });
    if(ov.TotalCommits>ov.Commits.length) b.append(el("div","more",F(T.more,ov.TotalCommits-ov.Commits.length)));
  }
  if((ov.Tasks||[]).length){
    const b = box(T.tasks);
    ov.Tasks.forEach(t=>{ const r=el("div","li"); r.append(el("span","m",t.Loc), el("span","t",t.Text)); b.append(r) });
    if(ov.TotalTasks>ov.Tasks.length) b.append(el("div","more",F(T.more,ov.TotalTasks-ov.Tasks.length)));
  }
  if((ov.Modules||[]).length){
    const b = box(T.modules);
    const mc = el("div","chips");
    ov.Modules.forEach(m=>mc.append(el("span","chip",m.Name+"/ "+m.Files)));
    b.append(mc);
  }
}

function renderTriage(){
  document.getElementById("tname").textContent = repo.Name;
  const card = document.getElementById("tcard"); card.textContent="";
  if(!repo.Items.length){
    document.getElementById("tsub").textContent = T.clear;
    card.append(el("div","empty",T.cleared));
    renderRails(); return }
  if(idx>=repo.Items.length) idx=repo.Items.length-1;
  document.getElementById("tsub").textContent = F(T.pos, idx+1, repo.Items.length);
  const o = repo.Items[idx];
  const c = el("div","item");
  const m = el("div","meta");
  m.append(el("span","tag "+o.Type,o.Type), el("span","mono",o.ShortID), el("span","",o.State), el("span","",o.Actor));
  if(o.Conflict) m.append(el("span","conf","⚠ conflict"));
  c.append(m);
  // Claim-first display: the first line (or the pre-colon lead of a
  // one-paragraph body) becomes the headline, the rest reads as detail.
  const nl = o.Body.indexOf("\n"), col = o.Body.indexOf(":");
  let head = o.Body, rest = "";
  if(nl > 0){ head = o.Body.slice(0,nl); rest = o.Body.slice(nl+1).trim() }
  else if(col > 0 && col < 90){ head = o.Body.slice(0,col); rest = o.Body.slice(col+1).trim() }
  c.append(el("div","body-title",head));
  if(rest) c.append(el("div","body",rest));
  if(o.Evidence) c.append(el("div","evid",T.evidence+o.Evidence));
  const a = el("div","actions");
  const reason = el("input"); reason.id="reason"; reason.placeholder=T.reasonph;
  const mk=(cls,key,label,fn)=>{const b=el("button",cls); b.append(el("b","",key),document.createTextNode(label)); b.onclick=fn; return b};
  a.append(reason,
    mk("bv","v",T.verify,()=>judge("verified")),
    mk("bs","s",T.stale,()=>judge("stale")),
    mk("bd","d",T.deprecate,()=>judge("deprecated")),
    mk("","x",T.skip,skip));
  c.append(a); card.append(c);
  renderRails();
}

function renderRails(){
  const rail = document.getElementById("rail"); rail.textContent="";
  if(repo.Items.length){
    rail.append(el("div","rail-label",T.queue));
    repo.Items.forEach((it,i)=>{
      const row = el("div","row"+(i===idx?" cur":""));
      row.append(el("span","idx",String(i+1)), el("span","tag "+it.Type,it.Type), el("span","txt",it.Body));
      row.onclick=()=>{idx=i; renderTriage()};
      rail.append(row);
    });
  }
  if(repo.History.length){
    rail.append(el("div","rail-label",T.records));
    repo.History.forEach(it=>{
      const row = el("div","row done");
      const st = el("span","st",it.State); st.style.color = SC[it.State]||"var(--muted)";
      row.append(st, el("span","tag "+it.Type,it.Type), el("span","txt",it.Body));
      if(it.Reason) row.append(el("span","why","· "+it.Reason));
      rail.append(row);
    });
  }
}

async function judge(to){
  const o = repo.Items[idx]; if(!o) return;
  const reason = document.getElementById("reason")?.value || "";
  const res = await fetch("/judge",{method:"POST",headers:{"Content-Type":"application/json"},
    body:JSON.stringify({Workspace:repo.Path,ID:o.ID,To:to,Reason:reason})});
  if(!res.ok){ toast(await res.text(), "var(--d)"); return }
  repo.Items.splice(idx,1); repo.Counts[o.State]--; repo.Counts[to]=(repo.Counts[to]||0)+1; judged++;
  repo.History.unshift(Object.assign({},o,{State:to,Reason:reason}));
  toast(o.ShortID+" → "+to, SC[to]); totals(); renderTriage();
}
function skip(){ if(idx<repo.Items.length-1) idx++; renderTriage() }

let tmr;
function toast(msg,color){ const t=document.getElementById("toast");
  t.textContent=msg; t.style.borderColor=color||"var(--line)"; t.classList.add("show");
  clearTimeout(tmr); tmr=setTimeout(()=>t.classList.remove("show"),1800) }

function finish(){ fetch("/quit",{method:"POST"}).finally(()=>{
  document.body.textContent="";
  const box = el("div","empty"); box.style.paddingTop="5rem";
  box.append(el("h2","",F(T.donetitle,judged)), el("p","",T.donenote));
  document.body.append(box)}) }

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
// Static chrome + help overlay, all from the active string table — called
// again whenever the user switches languages.
function applyChrome(){
  for(const [id,key] of [["wtitle","workspaces"],["localnote","localnote"],
    ["keysHint","keys"],["finishBtn","finish"],["backLbl","back"],["helpTitle","helptitle"]])
    document.getElementById(id).textContent = T[key];
  document.getElementById("langSel").value = LANG;
  const rows = [[T.hopen,["1","–","9"]],[T.hmove,["j","/","k"]],[T.hjudge,["v","s","d"]],
    [T.hskip,["x"]],[T.hreason,["r"]],[T.hback,["Esc"]]];
  const hr = document.getElementById("helpRows"); hr.textContent="";
  for(const [label,keys] of rows){
    const d = el("div"); d.append(el("span","",label));
    const ks = el("span");
    keys.forEach(k=>{ (k==="–"||k==="/")? ks.append(document.createTextNode(" "+k+" ")) : ks.append(el("kbd","",k)) });
    d.append(ks); hr.append(d);
  }
}
function setLang(l){
  LANG = l; T = TT[l] || TT.en;
  applyChrome(); repo? renderTriage() : renderFleet(); totals();
  fetch("/lang",{method:"POST",headers:{"Content-Type":"application/json"},
    body:JSON.stringify({Lang:l})}); // persists for the next launch
}
applyChrome();
renderFleet();
// Deep link: #2 opens the second repo — refresh keeps your place.
const h = location.hash.match(/^#(\d+)$/);
if(h && FLEET[+h[1]-1]) openRepo(+h[1]-1);
</script>
</body></html>`))
