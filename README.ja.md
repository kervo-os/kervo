<div align="center">

```
┬┌─ ┌─┐ ┬─┐ ┬  ┬ ┌─┐
├┴┐ ├┤  ├┬┘ └┐┌┘ │ │
┴ ┴ └─┘ ┴└─  └┘  └─┘
```

### 非決定的エージェントのための決定的コンテキスト

**AI にプロジェクトを毎回説明し直すのはやめましょう。`kervo init` 一度だけ。**

[![CI](https://github.com/kervo-os/kervo/actions/workflows/ci.yml/badge.svg)](https://github.com/kervo-os/kervo/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/kervo-os/kervo)](https://goreportcard.com/report/github.com/kervo-os/kervo)
[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)](go.mod)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

[English](README.md) | [한국어](README.ko.md) | **日本語**

[クイックスタート](#クイックスタート) ·
[仕組み](#仕組み) ·
[トラストラベル](#なぜトラストラベルか) ·
[計測結果](#主張ではなく計測) ·
[キャプチャ](#キャプチャ-フックの接続)

</div>

---

kervo はリポジトリを決定的な **Context Artifact** にコンパイルし、
`CLAUDE.md` に注入します — すべての AI セッションが、プロジェクトを既に
知っている状態から始まるように。事実(Fact)は決定的に抽出され、解釈は
トラストラベル付きの提案としてのみ入ります。提案は検証され、古くなり、
**除外理由を表示したまま**退役します。

このリポジトリは自分の作ったものを自分で使っています:ここの
[`CLAUDE.md`](CLAUDE.md) は kervo がコンパイルしたものです。

## クイックスタート

```bash
go install github.com/kervo-os/kervo/cmd/kervo@latest   # または: make build
kervo init        # スキャン → .kervo/artifact.md → CLAUDE.md に注入
```

実リポジトリでの初回実行は 1 秒未満です(コミットスキャン上限 500、到達時は
`Partial` 表示)。`CLAUDE.md` のうち `<!-- kervo:begin -->` と
`<!-- kervo:end -->` の間のブロックだけに触れます — 手で書いた内容は
バイト単位で保存されます。

**Artifact が含むもの:** リポジトリ要約 · 宣言されたコマンド(Makefile
ターゲット、npm スクリプト、docker-compose サービス、pyproject スクリプト、
justfile レシピ) · マージノイズを除いた最近の変更 · 未解決の TODO/FIXME ·
モジュール構造 — モノレポのモジュール別 `CLAUDE.md`/`README.md` を含む —
そして目標 / 決定 / リスク / 要約のためのトラストラベル付きスロット。

## 仕組み

厳密に分離された 2 層:

| 層 | 内容 | 生成方法 |
|---|---|---|
| **Fact スケルトン** | 要約、コマンド、変更、タスク、モジュール | 決定的スキャン — 同じワークスペースなら同じバイト列、CI のゴールデンテストで固定。この経路に LLM は決して入らない。 |
| **トラストスロット** | 目標、決定、リスク、要約 | 出所付きのラベル提案 — 事実を装わず、匿名もなし。 |

スロットを埋める 3 つの方法 — 優雅に降格し、バックエンドが失敗しても警告付きで
下がるだけで、Fact スケルトンは常に生成されます:

| モード | セマンティックスロットを埋めるもの | 必要なもの |
|---|---|---|
| **1 — Fact のみ**(デフォルト) | なし — 決定的な事実のみ。常に動作。 | git |
| **2 — コンシューマ支援** | AI セッションが `.kervo/proposals.json` に提案を蓄積 | エージェントセッション |
| **3 — 専用バックエンド** | OpenAI 互換エンドポイントが観察を提案 | ローカル/リモート LLM |

完全ローカルモデルで Mode 3(何もマシンの外に出ません):

```bash
export KERVO_SEMANTIC_URL=http://localhost:1234/v1   # LM Studio(または Ollama :11434/v1)
export KERVO_SEMANTIC_MODEL=openai/gpt-oss-120b
kervo compile
# Artifact: .kervo/artifact.md (Mode 3 — backend:openai/gpt-oss-120b)
```

Artifact はデフォルトで英語でレンダリングされ、`--lang ko` / `--lang ja` で
ローカライズされます(選択はワークスペースごとに保持)。

## なぜトラストラベルか

蓄積されたコンテキストは腐ります — そして間違ったコンテキストは、ないより
悪い。事実でないものはすべて、出所付きのラベル提案として入ります:

```
**[generated — backend:openai/gpt-oss-120b]**
Needs confirmation — current focus appears to be terminal input/UX
hardening… Evidence: Recent Changes 05-28..06-28.
```

状態は `generated → observed → verified → stale → deprecated` と動きます —
減衰タイマーではなく、証拠と人間の確認によって。二人の行為者の判断が割れた
ときは、黙って勝者を選ぶ代わりに `⚠ conflict` と表示します。Stale 項目は
黙って消える代わりに、除外理由とともに列挙されます。

## 主張ではなく計測

これは本当に、汚染されたコンテキストからエージェントを守るのか?仮説を
事前登録し、ブラインド実験を実施しました:同じリポジトリ、3 つのコンテキスト
アーム — **A**(kervo artifact)、**B**(同内容、トラストラベル除去)、
**C**(管理なしのノート) — 偽の「決定」を仕込み、新規のコンシューマ
セッション、アームと仮説を知らない審判。

確証ラン(事前登録、リポジトリアクセス遮断、sonnet + haiku コンシューマ、
n = 24):

| | **A — kervo** | B — ラベル除去 | C — 管理なし |
|---|---|---|---|
| 総合 S1+S2+S3 | **91.7%** | 91.7% | 62.5% |

- **A−C = +29.2pp** — 事前登録基準(≥20pp)を満たしました。プログラム全体で
  発生した実際の汚染感染(3 件)は、すべて弱いコンシューマモデルの C アームで
  起きました。
- プログラム全体 54 応答で、A アームは汚染された主張に 1 点も失いません
  でした。混合条件(リポジトリアクセス許可)での無ラベルアームの失敗は
  *伝染*でした:嘘が 1 つ見つかると、真実の事実まで連座で排斥される —
  ラベルは `verified` への信頼を保ったまま、`⚠ conflict` だけを選別隔離
  しました。
- 要点:コードが反証できる嘘はエージェントが自力で防ぎます。**ラベルが守る
  のは、コードの外に生きる真実です** — 決定、制約、文脈。コンシューマが
  弱いほど、保護効果は大きくなります。

完全なプロトコル・artifact・応答別採点は、人間採点のレプリケーション
(進行中)完了後に公開します。現時点の数値は事前登録ルーブリックによる
エージェント採点です。

## キャプチャ: フックの接続

ライブキャプチャが台帳と内蔵計測カウンタを養います。Claude Code なら
プロジェクトの `.claude/settings.json` に追加してください(フックは
プロジェクトディレクトリで実行されるので、`kervo` が PATH にあれば十分):

```json
{
  "hooks": {
    "UserPromptSubmit": [
      { "hooks": [{ "type": "command", "command": "kervo hook || true", "timeout": 10 }] }
    ],
    "SessionStart": [
      { "hooks": [{ "type": "command", "command": "kervo hook || true", "timeout": 10 }] }
    ],
    "PostToolUse": [
      { "matcher": "Edit|Write",
        "hooks": [{ "type": "command", "command": "kervo hook || true", "timeout": 10 }] }
    ]
  }
}
```

フックはミリ秒予算のローカル append です — LLM なし、ネットワークなし、
セッションを決して壊しません(ゴミが来ても exit 0)。コミットされる台帳には
**名前・パス・サイズのみ**が保存されます:プロンプトやファイルの内容が
マシンを離れることも、git 履歴に入ることもありません。

```bash
kervo capture -type decision -body "JWT over sessions"   # 手で記録
kervo trust -id 01KWP -to verified -reason "team agreed" # 判定
kervo status                                             # 1 画面のトラストビュー
kervo metrics                                            # プロンプトサイズ: artifact 有/無の比較
kervo import claude                                      # 過去の Claude Code セッションをバックフィル
```

## 設計上の保証

- **決定的スケルトン** — 同じワークスペース、同じ言語なら同じバイト列;
  CI のゴールデンファイルで固定。Fact 経路に LLM は決して入らない。
- **イベントが真実** — append-only の JSONL 台帳を git にコミット
  (`merge=union`);artifact とインデックスは派生物で再構築可能。リポジトリを
  クローンすれば、コンパイルされた記憶も一緒に移動。
- **境界は検査で** — 純粋コアはアダプタを import できない
  (`make arch-check`);データ由来のテキストは構造マーカーを偽装できない;
  プロバイダは `generated` より上へ自己昇格できない。
- **サーバなし、デーモンなし、DB なし、アカウントなし** — すべての状態は
  `.kervo/` と `CLAUDE.md` に。

## ステータス

v0、高速に進行中。セマンティックスロットを有効にした Cold Start 検証を通過;
キャプチャ → 検証 → stale のループがこのリポジトリで実稼働しています。
PRD / RFC / 実験プロトコルは安定し次第公開します。

---

kervo はコーディングツールではありません。git の上で生きるすべてのチームの
ための記憶レイヤーです — 開発者はすでに作業をコミットとして保存しているため、
最初の市場であるにすぎません。

ライセンス: [Apache-2.0](LICENSE)。
