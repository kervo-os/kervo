// Package i18n holds the artifact's human-language string tables.
// English is the default; the language is part of the compiler's input, so
// determinism holds per language (same workspace + same lang = same bytes,
// pinned by per-language golden files). Structural tokens — markers, trust
// states, the artifact header shape — are never localized.
package i18n

import "fmt"

type Lang string

const (
	EN Lang = "en"
	JA Lang = "ja"
	KO Lang = "ko"
)

// Supported returns all languages in stable order (used in error text).
func Supported() []Lang { return []Lang{EN, KO, JA} }

// Parse validates a user-supplied language code. Empty means default (en).
func Parse(s string) (Lang, error) {
	switch Lang(s) {
	case "", EN:
		return EN, nil
	case KO:
		return KO, nil
	case JA:
		return JA, nil
	default:
		return "", fmt.Errorf("i18n: unsupported language %q (supported: en, ko, ja)", s)
	}
}

// T returns the string for key in lang, falling back to English. A missing
// English key returns the key itself — loud enough to catch in any test.
func T(l Lang, key string) string {
	if v, ok := tables[l][key]; ok {
		return v
	}
	if v, ok := tables[EN][key]; ok {
		return v
	}
	return key
}

// Keys returns every key defined for English — the completeness reference.
func Keys() []string {
	ks := make([]string, 0, len(tables[EN]))
	for k := range tables[EN] {
		ks = append(ks, k)
	}
	return ks
}

// Has reports whether lang defines key natively (no fallback).
func Has(l Lang, key string) bool {
	_, ok := tables[l][key]
	return ok
}

var tables = map[Lang]map[string]string{
	EN: {
		"hdr.quote": "> Machine-generated context for AI agents. Fact sections are deterministic;\n> slot sections carry trust-labeled observations. Regenerate with `kervo compile`\n> — do not edit by hand.\n>\n> Start from this artifact: state what it already answers for your task,\n> then explore only the gaps — and verify only what your task puts in\n> question.",

		"sec.summary": "Repository Summary",

		"brief.title":    "Brief",
		"brief.focus":    "Focus",
		"brief.run":      "Run",
		"brief.edges":    "Open edges",
		"brief.unpushed": "Unpushed commits",
		"lbl.name":       "Name",
		"lbl.branch":     "Branch",
		"lbl.languages":  "Languages",
		"lbl.frameworks": "Frameworks",
		"lbl.docs":       "Docs",
		"excerpt.suffix": "(excerpt)",

		"sec.commands":   "Commands",
		"commands.empty": "_No declared commands found (Makefile targets, package.json scripts)._",

		"sec.recent":     "Recent Changes",
		"recent.empty":   "_No commits found._",
		"recent.showing": "Showing %[1]d of %[2]d analyzed commits.",
		"recent.capped":  " Scan capped — older history not analyzed (Partial).",
		"sec.hotfiles":   "Frequently Changed Files",

		"sec.tasks":     "Open Tasks",
		"tasks.empty":   "_No TODO/FIXME comments found._",
		"tasks.showing": "Showing %[1]d of %[2]d open tasks.",

		"sec.modules":   "Related Modules",
		"modules.empty": "_No top-level modules (flat repository)._",
		"modules.line":  "- %[1]s/ (%[2]d files)",

		"sec.facts":      "Workspace Facts",
		"facts.commits":  "Commits analyzed",
		"facts.complete": "complete",
		"facts.partial":  "partial — caps hit",
		"facts.tasks":    "Open tasks (TODO/FIXME)",
		"facts.modules":  "Top-level modules",
		"facts.docs":     "Docs captured",

		"slot.goal":      "Possible Current Goal",
		"slot.decisions": "Known Decisions",
		"slot.risks":     "Known Risks",
		"slot.summaries": "Doc Summaries",
		"ph.goal":        "_No proposal yet. A confirmed goal becomes the first Verified observation._",
		"ph.generic":     "_None proposed yet. Semantic providers (Mode 2/3) attach labeled observations here._",

		"sec.stale":   "Deprecated / Stale Notes",
		"stale.empty": "_None recorded. Stale or deprecated observations are listed here with their\nexclusion reason instead of being silently dropped._",

		"dash.workspaces":  "Workspaces",
		"dash.totals":      "%[1]d workspaces · %[2]d awaiting judgment",
		"dash.localnote":   "everything below is local — no server, no account",
		"dash.finish":      "Finish",
		"dash.keys":        "keys",
		"dash.awaiting":    "awaiting judgment",
		"dash.clear":       "all clear",
		"dash.events":      "%[1]d events",
		"dash.emptyledger": "empty ledger",
		"dash.queue":       "queue",
		"dash.records":     "judged records",
		"dash.overview":    "overview",
		"dash.links":       "change together",
		"dash.more":        "+%[1]d more",
		"dash.partialscan": "partial scan",
		"dash.connected":   "connected",
		"dash.knowledge":   "knowledge — verified & observed",
		"dash.retired":     "retired — with reasons",
		"dash.jhint":       "v signs it as true for every future session (reversible) · d records why it is wrong · skip keeps it an unverified proposal",
		"dash.jtitle":      "What a judgment does",
		"dash.jv":          "verify — signed true; future sessions trust it. Reversible: corrections stay ledgered.",
		"dash.js":          "stale — not current anymore; listed with its reason instead of deleted",
		"dash.jd":          "deprecate — wrong; your reason teaches the next agent",
		"dash.jx":          "skip — no signature; consumers keep treating it as unverified",
		"dash.pos":         "%[1]d / %[2]d · j/k to move",
		"dash.cleared":     "All judged here — Esc back to the fleet.",
		"dash.reasonph":    "reason (optional) — r to focus",
		"dash.verify":      "verify",
		"dash.stale":       "stale",
		"dash.deprecate":   "deprecate",
		"dash.skip":        "skip",
		"dash.back":        "← fleet",
		"dash.donetitle":   "Done — %[1]d judged.",
		"dash.donenote":    "Run kervo compile in the affected repos, then close this tab.",
		"dash.evidence":    "evidence: ",
		"dash.justnow":     "just now",
		"dash.minago":      "%[1]dm ago",
		"dash.hourago":     "%[1]dh ago",
		"dash.dayago":      "%[1]dd ago",
		"dash.helptitle":   "Keys",
		"dash.hopen":       "open repo",
		"dash.hmove":       "next / prev item",
		"dash.hjudge":      "verify · stale · deprecate",
		"dash.hskip":       "skip",
		"dash.hreason":     "reason field",
		"dash.hback":       "back / close",

		"sec.writeback":  "Write-back Protocol",
		"writeback.body": "> For AI consumers — close the loop. If this session taught you a durable\n> fact this artifact does not carry (how to run something, what a component\n> does, a decision made in conversation), stage it for human judgment:\n>\n> `kervo capture -type decision|risk|summary|goal -actor \"agent:<you>\" -body \"<the fact>\" -evidence \"<how you verified it>\"`\n>\n> Rules: facts you observed, not speculation · conclusions, not corpus —\n> what lives in a file agents can read stays there; cite it as evidence\n> instead of mirroring it · start the body with a\n> one-line claim, details after it · attach evidence — the command you\n> ran, the doc you read — so the human can sign in one keystroke · one\n> capture per fact · skip what this artifact already says · never include\n> secrets or file contents. Duplicates are dropped automatically.\n>\n> The conversation is the review: if the human affirmed a fact in this\n> session, relay their judgment with the capture (`kervo trust -to\n> verified -reason \"<their words>\"`) — only undiscussed facts wait in\n> the queue (`kervo review`). If evidence contradicts a verified entry,\n> raise it with the human and record their updated judgment instead of\n> re-proposing. When you relay this artifact's knowledge, carry its\n> trust label with it — the reader must be able to tell record from\n> inference.",
	},
	KO: {
		"hdr.quote": "> AI 에이전트용으로 기계 생성된 컨텍스트. Fact 섹션은 결정적으로 생성되며,\n> 슬롯 섹션에는 신뢰 라벨이 달린 관찰이 담긴다. `kervo compile`로 재생성 —\n> 손으로 수정하지 말 것.\n>\n> 이 artifact에서 시작하라: 과제에 대해 이미 답하는 것을 먼저 밝히고,\n> 빈 곳만 탐색하라 — 검증은 과제가 의심하는 것에만.",

		"sec.summary": "저장소 요약",

		"brief.title":    "브리핑",
		"brief.focus":    "집중",
		"brief.run":      "실행",
		"brief.edges":    "미완의 가장자리",
		"brief.unpushed": "미푸시 커밋",
		"lbl.name":       "이름",
		"lbl.branch":     "브랜치",
		"lbl.languages":  "언어",
		"lbl.frameworks": "프레임워크",
		"lbl.docs":       "문서",
		"excerpt.suffix": "(발췌)",

		"sec.commands":   "명령어",
		"commands.empty": "_선언된 명령을 찾지 못함 (Makefile 타깃, package.json scripts)._",

		"sec.recent":     "최근 변경",
		"recent.empty":   "_커밋이 없음._",
		"recent.showing": "분석된 커밋 %[2]d개 중 %[1]d개 표시.",
		"recent.capped":  " 스캔 캡 도달 — 더 오래된 이력은 분석되지 않음 (Partial).",
		"sec.hotfiles":   "자주 변경된 파일",

		"sec.tasks":     "미해결 작업",
		"tasks.empty":   "_TODO/FIXME 주석을 찾지 못함._",
		"tasks.showing": "미해결 작업 %[2]d개 중 %[1]d개 표시.",

		"sec.modules":   "관련 모듈",
		"modules.empty": "_최상위 모듈 없음 (평면 구조 저장소)._",
		"modules.line":  "- %[1]s/ (파일 %[2]d개)",

		"sec.facts":      "워크스페이스 사실",
		"facts.commits":  "분석된 커밋",
		"facts.complete": "전체",
		"facts.partial":  "부분 — 캡 도달",
		"facts.tasks":    "미해결 작업 (TODO/FIXME)",
		"facts.modules":  "최상위 모듈",
		"facts.docs":     "수집된 문서",

		"slot.goal":      "추정 현재 목표",
		"slot.decisions": "알려진 결정",
		"slot.risks":     "알려진 리스크",
		"slot.summaries": "문서 요약",
		"ph.goal":        "_아직 제안 없음. 확인된 목표가 첫 Verified 관찰이 된다._",
		"ph.generic":     "_아직 제안 없음. Semantic 제공자(Mode 2/3)가 라벨 달린 관찰을 여기에 부착한다._",

		"sec.stale":   "폐기·낡음 기록",
		"stale.empty": "_기록 없음. 낡거나 폐기된 관찰은 조용히 삭제되는 대신\n제외 사유와 함께 여기에 나열된다._",

		"dash.workspaces":  "워크스페이스",
		"dash.totals":      "워크스페이스 %[1]d개 · 판정 대기 %[2]d건",
		"dash.localnote":   "아래 전부 로컬입니다 — 서버 없음, 계정 없음",
		"dash.finish":      "마치기",
		"dash.keys":        "단축키",
		"dash.awaiting":    "판정 대기",
		"dash.clear":       "모두 처리됨",
		"dash.events":      "이벤트 %[1]d건",
		"dash.emptyledger": "빈 장부",
		"dash.queue":       "대기열",
		"dash.records":     "판정 기록",
		"dash.overview":    "개요",
		"dash.links":       "함께 변경되는 모듈",
		"dash.more":        "+%[1]d건 더",
		"dash.partialscan": "부분 스캔",
		"dash.connected":   "연결됨",
		"dash.knowledge":   "지식 — verified·observed",
		"dash.retired":     "퇴역 — 사유와 함께",
		"dash.jhint":       "v = 사실로 서명, 이후 모든 세션이 신뢰 (되돌릴 수 있음) · d = 오류, 사유 기록 · skip = 미검증 제안으로 유지",
		"dash.jtitle":      "판정의 의미",
		"dash.jv":          "승인 — 사실로 서명. 이후 세션이 신뢰. 되돌릴 수 있고 정정도 장부에 남음",
		"dash.js":          "낡음 — 더는 유효하지 않음. 삭제 대신 사유와 함께 나열",
		"dash.jd":          "폐기 — 틀림. 남긴 사유가 다음 에이전트의 같은 실수를 막음",
		"dash.jx":          "건너뜀 — 서명 없음. 소비자는 계속 미검증 제안으로 취급",
		"dash.pos":         "%[1]d / %[2]d · 이동 j/k",
		"dash.cleared":     "여기는 다 판정했습니다 — Esc로 전체 화면에 돌아갑니다.",
		"dash.reasonph":    "사유 (선택) — r로 포커스",
		"dash.verify":      "승인",
		"dash.stale":       "낡음",
		"dash.deprecate":   "폐기",
		"dash.skip":        "건너뜀",
		"dash.back":        "← 전체",
		"dash.donetitle":   "완료 — %[1]d건 판정.",
		"dash.donenote":    "해당 레포에서 kervo compile을 실행한 뒤 탭을 닫으세요.",
		"dash.evidence":    "증거: ",
		"dash.justnow":     "방금",
		"dash.minago":      "%[1]d분 전",
		"dash.hourago":     "%[1]d시간 전",
		"dash.dayago":      "%[1]d일 전",
		"dash.helptitle":   "단축키",
		"dash.hopen":       "레포 열기",
		"dash.hmove":       "다음 / 이전 항목",
		"dash.hjudge":      "승인 · 낡음 · 폐기",
		"dash.hskip":       "건너뜀",
		"dash.hreason":     "사유 입력",
		"dash.hback":       "뒤로 / 닫기",

		"sec.writeback":  "Write-back 프로토콜",
		"writeback.body": "> AI 소비자용 — 루프를 닫을 것. 이 세션에서 이 artifact에 없는 지속적\n> 사실을 알게 됐다면(무언가의 실행 방법, 컴포넌트의 역할, 대화 중 내려진\n> 결정), 사람의 판정을 받도록 올려둘 것:\n>\n> `kervo capture -type decision|risk|summary|goal -actor \"agent:<너>\" -body \"<사실>\" -evidence \"<검증 방법>\"`\n>\n> 규칙: 관찰한 사실만, 추측 금지 · 코퍼스가 아니라 결론만 — 에이전트가\n> 읽을 수 있는 파일에 있는 내용은 거기 두고, 미러링 대신 증거로 인용할\n> 것 · 본문 첫 줄은 한 줄 주장, 세부는 마크다운으로(불릿, `코드`) · 증거를 붙일 것 — 실행한 명령, 읽은 문서 — 사람이 키 한\n> 번으로 서명할 수 있도록 · 사실 하나당 캡처 하나 · 이 artifact에 이미\n> 있는 내용은 생략 · 비밀값·파일 내용은 절대 넣지 말 것. 중복은 자동으로\n> 버려진다.\n>\n> 대화가 곧 검토다: 사람이 이 세션에서 사실에 명시적으로 동의했다면\n> 캡처와 함께 그 판정을 중계할 것(`kervo trust -to verified -reason\n> \"<동의 인용>\"`) — 논의되지 않은 사실만 큐에서 기다린다(`kervo\n> review`). verified 항목과 어긋나는 증거를 발견하면 재제안하지 말고\n> 사람에게 물어 갱신된 판정을 기록할 것. 이 artifact의 지식을 중계할\n> 때는 신뢰 라벨을 함께 인용할 것 — 읽는 사람이 기록과 추론을 구분할 수\n> 있어야 한다.",
	},
	JA: {
		"hdr.quote": "> AIエージェント向けに機械生成されたコンテキスト。Factセクションは決定的に生成され、\n> スロットセクションには信頼ラベル付きの観察が入る。`kervo compile`で再生成 —\n> 手で編集しないこと。\n>\n> この artifact から始めること: タスクに対して既に答えているものを先に\n> 述べ、足りない部分だけを探索する — 検証はタスクが疑問とするものだけに。",

		"sec.summary": "リポジトリ概要",

		"brief.title":    "ブリーフ",
		"brief.focus":    "フォーカス",
		"brief.run":      "実行",
		"brief.edges":    "未完のエッジ",
		"brief.unpushed": "未プッシュコミット",
		"lbl.name":       "名前",
		"lbl.branch":     "ブランチ",
		"lbl.languages":  "言語",
		"lbl.frameworks": "フレームワーク",
		"lbl.docs":       "ドキュメント",
		"excerpt.suffix": "(抜粋)",

		"sec.commands":   "コマンド",
		"commands.empty": "_宣言されたコマンドが見つからない (Makefileターゲット、package.json scripts)._",

		"sec.recent":     "最近の変更",
		"recent.empty":   "_コミットがない._",
		"recent.showing": "分析済みコミット%[2]d件のうち%[1]d件を表示.",
		"recent.capped":  " スキャン上限に到達 — それ以前の履歴は未分析 (Partial).",
		"sec.hotfiles":   "変更頻度の高いファイル",

		"sec.tasks":     "未解決タスク",
		"tasks.empty":   "_TODO/FIXMEコメントが見つからない._",
		"tasks.showing": "未解決タスク%[2]d件のうち%[1]d件を表示.",

		"sec.modules":   "関連モジュール",
		"modules.empty": "_トップレベルモジュールなし (フラットなリポジトリ)._",
		"modules.line":  "- %[1]s/ (ファイル%[2]d件)",

		"sec.facts":      "ワークスペースの事実",
		"facts.commits":  "分析済みコミット",
		"facts.complete": "完全",
		"facts.partial":  "部分 — 上限到達",
		"facts.tasks":    "未解決タスク (TODO/FIXME)",
		"facts.modules":  "トップレベルモジュール",
		"facts.docs":     "収集済みドキュメント",

		"slot.goal":      "推定される現在の目標",
		"slot.decisions": "既知の決定",
		"slot.risks":     "既知のリスク",
		"slot.summaries": "ドキュメント要約",
		"ph.goal":        "_まだ提案なし。確認された目標が最初のVerified観察になる._",
		"ph.generic":     "_まだ提案なし。Semanticプロバイダ(Mode 2/3)がラベル付き観察をここに付ける._",

		"sec.stale":   "廃止・陳腐化ノート",
		"stale.empty": "_記録なし。陳腐化・廃止された観察は黙って消される代わりに、\n除外理由と共にここへ列挙される._",

		"dash.workspaces":  "ワークスペース",
		"dash.totals":      "ワークスペース %[1]d 件 · 判定待ち %[2]d 件",
		"dash.localnote":   "以下はすべてローカルです — サーバなし、アカウントなし",
		"dash.finish":      "終了",
		"dash.keys":        "キー",
		"dash.awaiting":    "判定待ち",
		"dash.clear":       "すべて処理済み",
		"dash.events":      "イベント %[1]d 件",
		"dash.emptyledger": "空の台帳",
		"dash.queue":       "キュー",
		"dash.records":     "判定記録",
		"dash.overview":    "概要",
		"dash.links":       "共に変更されるモジュール",
		"dash.more":        "+%[1]d 件",
		"dash.partialscan": "部分スキャン",
		"dash.connected":   "接続",
		"dash.knowledge":   "ナレッジ — verified·observed",
		"dash.retired":     "退役 — 理由と共に",
		"dash.jhint":       "v = 事実として署名、以後のセッションが信頼 (取り消し可能) · d = 誤り、理由を記録 · skip = 未検証の提案のまま",
		"dash.jtitle":      "判定の意味",
		"dash.jv":          "承認 — 事実として署名。以後のセッションが信頼。取り消し可能で訂正も台帳に残る",
		"dash.js":          "陳腐化 — もう有効ではない。削除の代わりに理由と共に列挙",
		"dash.jd":          "廃止 — 誤り。残した理由が次のエージェントの同じ誤りを防ぐ",
		"dash.jx":          "スキップ — 署名なし。コンシューマは未検証の提案として扱い続ける",
		"dash.pos":         "%[1]d / %[2]d · 移動 j/k",
		"dash.cleared":     "ここは判定済みです — Esc で全体に戻ります。",
		"dash.reasonph":    "理由 (任意) — r でフォーカス",
		"dash.verify":      "承認",
		"dash.stale":       "陳腐化",
		"dash.deprecate":   "廃止",
		"dash.skip":        "スキップ",
		"dash.back":        "← 全体",
		"dash.donetitle":   "完了 — %[1]d 件判定。",
		"dash.donenote":    "該当リポジトリで kervo compile を実行し、このタブを閉じてください。",
		"dash.evidence":    "証拠: ",
		"dash.justnow":     "たった今",
		"dash.minago":      "%[1]d分前",
		"dash.hourago":     "%[1]d時間前",
		"dash.dayago":      "%[1]d日前",
		"dash.helptitle":   "キー",
		"dash.hopen":       "リポジトリを開く",
		"dash.hmove":       "次 / 前の項目",
		"dash.hjudge":      "承認 · 陳腐化 · 廃止",
		"dash.hskip":       "スキップ",
		"dash.hreason":     "理由フィールド",
		"dash.hback":       "戻る / 閉じる",

		"sec.writeback":  "Write-back プロトコル",
		"writeback.body": "> AI コンシューマ向け — ループを閉じること。このセッションでこの artifact\n> にない持続的事実を知ったなら(何かの実行方法、コンポーネントの役割、会話中\n> の決定)、人間の判定を受けるよう登録すること:\n>\n> `kervo capture -type decision|risk|summary|goal -actor \"agent:<あなた>\" -body \"<事実>\" -evidence \"<検証方法>\"`\n>\n> ルール: 観察した事実のみ、推測禁止 · コーパスではなく結論のみ —\n> エージェントが読めるファイルにある内容はそこに置き、ミラーリングの\n> 代わりに証拠として引用すること · 本文の 1 行目は一行の主張に(詳細は\n> その後に) · 証拠を付けること — 実行したコマンド、読んだドキュメント —\n> 人間がキー一つで署名できるように · 事実ひとつにつきキャプチャひとつ ·\n> この artifact に既にある内容は省略 · 秘密情報・ファイル内容は決して\n> 入れない。重複は自動的に破棄される。\n>\n> 会話こそがレビューである: 人間がこのセッションで事実に明示的に同意\n> したなら、キャプチャと共にその判定を中継すること(`kervo trust -to\n> verified -reason \"<同意の引用>\"`) — 議論されていない事実だけがキュー\n> で待つ(`kervo review`)。verified 項目と矛盾する証拠を見つけたら、\n> 再提案せず人間に確認して更新された判定を記録すること。この artifact\n> の知識を伝えるときはトラストラベルを併記すること — 読み手が記録と\n> 推論を区別できなければならない。",
	},
}
