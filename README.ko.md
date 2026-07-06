<div align="center">

<img src="assets/logo.svg" width="140" alt="kervo">

### 비결정적 에이전트를 위한 결정적 컨텍스트

**AI에게 프로젝트를 매번 다시 설명하지 마세요. `kervo init` 한 번이면 됩니다.**

[![CI](https://github.com/kervo-os/kervo/actions/workflows/ci.yml/badge.svg)](https://github.com/kervo-os/kervo/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/kervo-os/kervo?color=34d399)](https://github.com/kervo-os/kervo/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/kervo-os/kervo)](https://goreportcard.com/report/github.com/kervo-os/kervo)
[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)](go.mod)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

[English](README.md) | **한국어** | [日本語](README.ja.md)

[빠른 시작](#빠른-시작) ·
[팀 사용](#팀-레포에서) ·
[동작 방식](#동작-방식) ·
[트러스트 라벨](#왜-트러스트-라벨인가) ·
[측정 결과](#주장이-아니라-측정) ·
[캡처](#캡처-훅-연결) ·
[명령어](#명령어)

</div>

---

kervo는 저장소를 결정적 **Context Artifact**로 컴파일해 `CLAUDE.md`에
주입합니다 — 모든 AI 세션이 프로젝트를 이미 아는 상태에서 시작하도록.
사실(Fact)은 결정적으로 추출되고, 해석은 오직 신뢰 라벨이 달린 제안으로만
들어옵니다. 제안은 검증되고, 낡아지고, **제외 사유가 표시된 채** 퇴역합니다.

이 저장소는 자기가 만든 것을 자기가 씁니다: 여기의
[`CLAUDE.md`](CLAUDE.md)는 kervo가 컴파일한 것입니다.

루프 전체:

```mermaid
flowchart LR
  S[결정적 스캔] --> A[Context Artifact]
  A -->|CLAUDE.md · AGENTS.md · MCP| C[AI 세션]
  C -->|write-back: 사실 + 증거| L[(이벤트 장부 · git)]
  P[외부 생산자] -->|proposals.json| L
  L --> R{사람의 판정}
  R -->|verified / deprecated + 사유| L
  L --> A
```

그리고 `kervo dash` — 머신의 모든 워크스페이스를 한 페이지에:

<p align="center"><img src="assets/dash-fleet.png" width="860" alt="kervo dash — 전체 워크스페이스, 대기 판정, 활동"></p>

## 빠른 시작

```bash
brew install kervo-os/tap/kervo   # macOS & Linux — 프리빌트 바이너리
# 또는: go install github.com/kervo-os/kervo/cmd/kervo@latest
kervo init        # 스캔 → .kervo/artifact.md → CLAUDE.md에 주입
```

macOS·Linux·Windows용 프리빌트 바이너리는
[릴리스 페이지](https://github.com/kervo-os/kervo/releases)에 있습니다 —
Go 툴체인이 필요 없습니다. 실제 저장소 기준 첫 실행은 1초 미만입니다(커밋
스캔 상한 500개, 도달 시 `Partial` 표시). `CLAUDE.md`에서 `<!-- kervo:begin -->`과
`<!-- kervo:end -->` 사이 블록만 건드립니다 — 손으로 쓴 내용은
바이트 단위로 보존됩니다.

코덱스처럼 `AGENTS.md`를 읽는 에이전트를 쓴다면? 레포 루트에 그 파일이
있으면 같은 마커 블록을 같은 계약으로 거기에도 주입합니다. 파일 존재가 곧
opt-in이고(`touch AGENTS.md`), kervo가 이 파일을 스스로 만들지는 않습니다.

`CLAUDE.md`를 깔끔하게 유지하고 싶다면? `kervo compile -inject import`가
전체 블록 대신 `@.kervo/artifact.md` 한 줄만 주입합니다(Claude Code가 로드
시점에 확장). 트레이드오프는 의도된 것입니다: artifact 파일은 파생물이라
gitignore 대상이므로, 신선한 클론은 `kervo compile` 한 번 전까지 아무것도
보지 못합니다 — 전체 블록이 기본값인 이유입니다. 선택은 워크스페이스별로
유지됩니다(`.kervo/inject`, 커밋됨). `@` 줄은 Claude Code 문법이라
AGENTS.md 소비자는 확장하지 못할 수 있습니다.

**Artifact가 담는 것:** 저장소 요약 · 선언된 명령어(Makefile 타깃, npm
스크립트, docker-compose 서비스, pyproject 스크립트, justfile 레시피) ·
머지 노이즈를 제외한 최근 변경 · 열린 TODO/FIXME · 모듈 구조 — 모노레포의
모듈별 `CLAUDE.md`/`README.md` 포함 — 그리고 목표 / 결정 / 리스크 / 요약을
담는 트러스트 라벨 슬롯. 인용 트랜스크립트 같은 보관 자료는
`.kervoignore`(한 줄에 경로 접두 하나)로 TODO 스캔에서 제외할 수 있습니다.

## 팀 레포에서

커밋되는 진실과 파생 상태의 분리가 컨텍스트를 이동 가능하게 만듭니다:

| 상태 | 경로 | git 커밋? |
|---|---|---|
| 이벤트 장부 — 진실 | `.kervo/events/*.jsonl` | **예** — append-only, `merge=union`: 브랜치 머지는 장부의 합집합 |
| Artifact 언어 | `.kervo/lang` | **예** |
| 주입된 컨텍스트 블록 | `CLAUDE.md` | **예** |
| 컴파일된 artifact | `.kervo/artifact.md` | 아니오 — 파생물, `compile`이 재생성 |
| 인덱스·캐시 | `.kervo/index.db`, `.kervo/cache/` | 아니오 — 파생물 |

수명주기:

1. **최초 도입** — 한 사람이 `kervo init`을 한 번 실행하고 결과(장부,
   `.kervo/lang`, 주입된 `CLAUDE.md`, gitignore 항목)를 커밋합니다.
2. **팀원이 클론** — 컨텍스트는 이미 살아 있습니다: `CLAUDE.md`에 마지막
   컴파일 블록이 있고 장부 전체가 클론에 포함됩니다. AI 세션은 **명령 0개**로
   바로 읽고, `kervo status` / `metrics`도 클론된 장부에서 즉시 동작합니다.
3. **라이브 전환** — 바이너리를 설치하고(`brew install kervo-os/tap/kervo`)
   `kervo compile`(재`init` 아님)로 재스캔·사실 갱신. `init`도 멱등이라 습관적으로 실행해도 깨지지 않습니다.
4. **훅** — `.claude/settings.json`을 커밋해두면 `kervo`가 PATH에 있는
   팀원 전원에게 캡처가 자동 발화합니다.

이 저장소의 신선한 클론으로 검증: `compile`이 커밋된 장부(이벤트 112개,
관찰 4개)를 재생했고, 트러스트 상태·언어 유지, artifact 재생성 확인.

## 동작 방식

엄격히 분리된 2계층:

| 계층 | 내용 | 생산 방식 |
|---|---|---|
| **Fact 스켈레톤** | 요약, 명령어, 변경, 태스크, 모듈 | 결정적 스캔 — 같은 워크스페이스면 같은 바이트, CI 골든 테스트로 고정. 이 경로에 LLM은 절대 없음. |
| **트러스트 슬롯** | 목표, 결정, 리스크, 요약 | 출처가 달린 라벨 제안 — 사실로 위장하지 않고, 익명도 없음. |

슬롯을 채우는 세 가지 방법 — 우아하게 강등되며, 백엔드가 실패해도 경고와
함께 내려갈 뿐 Fact 스켈레톤은 항상 생산됩니다:

| 모드 | 시맨틱 슬롯을 채우는 것 | 필요한 것 |
|---|---|---|
| **1 — Fact 전용** (기본) | 없음 — 결정적 사실만. 항상 동작. | git |
| **2 — 소비자 보조** | AI 세션이 `.kervo/proposals.json`에 제안을 적재 | 에이전트 세션 |
| **3 — 전용 백엔드** | OpenAI 호환 엔드포인트가 관찰을 제안 | 로컬/원격 LLM |

Mode 3는 부트스트랩 채널입니다: 캡처하는 에이전트가 아직 일한 적 없는
레포에서 빈 슬롯을 채우는 용도. Mode 2 세션 캡처가 살아 있으면 env를
설정하지 마세요 — artifact만 읽는 추론은 의도가 아니라 이력을 읽고,
대부분 리뷰 소음만 늘립니다(실레포 실측).

완전 로컬 모델로 Mode 3 (아무것도 머신 밖으로 나가지 않음):

```bash
export KERVO_SEMANTIC_URL=http://localhost:1234/v1   # LM Studio (또는 Ollama :11434/v1)
export KERVO_SEMANTIC_MODEL=openai/gpt-oss-120b
kervo compile
# Artifact: .kervo/artifact.md (Mode 3 — backend:openai/gpt-oss-120b)
```

**외부 생산자.** 레포 지식을 만들어내는 어떤 도구든 — 그래프 빌더, 메모리
스토어, 위키 생성기 — `.kervo/proposals.json`에 항목을 적재하는 것만으로
kervo에 공급할 수 있습니다:

```json
[{ "slot": "summaries", "body": "AuthService는 TokenStore에 의존", "source": "graphify" }]
```

`compile`이 출처를 달아 `generated`로 장부에 수용하고, `review`가 다른
제안과 똑같이 관문이 됩니다. 이 형식에 state 필드가 없는 건 설계입니다:
생산자는 자가 승격할 수 없습니다 — 다른 도구는 기억을 만들고, kervo는
어떤 기억을 다음 세션에 가져갈지 판정합니다.

Artifact는 기본 영어로 렌더링되며 `--lang ko` / `--lang ja`로 현지화됩니다
(선택은 워크스페이스별로 유지).

## 왜 트러스트 라벨인가

축적된 컨텍스트는 부패합니다 — 그리고 틀린 컨텍스트는 없는 것보다
나쁩니다. 사실이 아닌 모든 것은 출처가 달린 라벨 제안으로 들어옵니다:

```
**[generated — backend:openai/gpt-oss-120b]**
Needs confirmation — current focus appears to be terminal input/UX
hardening… Evidence: Recent Changes 05-28..06-28.
```

상태는 `generated → observed → verified → stale → deprecated`로 움직입니다
— 감쇠 타이머가 아니라 증거와 사람의 확인으로. 두 행위자의 판단이 갈리면
조용히 승자를 고르는 대신 `⚠ conflict`로 표시합니다. Stale 항목은 조용히
사라지는 대신 제외 사유와 함께 나열됩니다.

역할 분담은 의도된 설계입니다: **에이전트가 캡처·제안·관리하고, 사람은
판정만 합니다.** `kervo review`가 그 판정 표면입니다 — 결정을 기다리는
모든 것을 하나씩 보여주는 검토 큐.

모든 artifact는 탐색의 루프를 닫는 **write-back 프로토콜**로 끝납니다:
AI 소비자가 어렵게 알아낸 지속적 사실 — 실행 방법, 컴포넌트 역할, 대화 중
결정 — 을 제안으로 캡처하라고 지시합니다. `kervo review`로 한 번 판정하면,
이후 모든 세션이 — 어떤 에이전트든, 어떤 팀원이든 — 도구 호출 0회로 그
답을 받습니다. 제안에는 **증거**가 실립니다 — 에이전트가 실행한 명령, 읽은
문서 — 검증 노동은 에이전트가 지고, 사람의 서명은 키 한 번이면 됩니다.
중복 본문은 자동으로 버려져 큐가 깨끗하게 유지됩니다.

## 주장이 아니라 측정

이게 실제로 오염된 컨텍스트에서 에이전트를 보호하나? 가설을 사전 등록하고
블라인드 실험을 돌렸습니다: 같은 저장소, 세 가지 컨텍스트 암 — **A**(kervo
artifact), **B**(같은 내용, 트러스트 라벨 제거), **C**(관리 없는 노트) —
거짓 "결정"을 심고, 신선한 소비자 세션, 암과 가설을 모르는 심판.

확증 런 (사전 등록, 레포 접근 차단, sonnet + haiku 소비자, n = 24):

| | **A — kervo** | B — 라벨 제거 | C — 무관리 |
|---|---|---|---|
| 종합 S1+S2+S3 | **91.7%** | 91.7% | 62.5% |

- **A−C = +29.2%p** — 사전 등록 기준(≥20%p) 충족. 프로그램 전체에서 발생한
  실제 오염 감염(3건)은 전부 약한 소비자 모델의 C 암에서 나왔습니다.
- 프로그램 전체 54개 응답에서 A 암은 오염된 주장에 단 1점도 잃지 않았습니다.
  혼합 조건(레포 접근 허용)에서 무라벨 암의 실패는 *전염*이었습니다: 거짓
  하나가 발견되면 참인 사실까지 연좌제로 배척 — 라벨은 `verified`의 신뢰를
  유지한 채 `⚠ conflict`만 선별 격리했습니다.
- 요점: 코드가 반증할 수 있는 거짓은 에이전트가 스스로 막습니다. **라벨이
  지키는 것은 코드 밖의 진실입니다** — 결정, 제약, 맥락. 소비자가 약할수록
  보호 효과는 커집니다.

전체 프로토콜·사전 등록·암 artifact·원문 응답 54개:
[docs/experiments/h4](docs/experiments/h4/). 채점은 구조적으로 블라인드된
심판의 에이전트 채점(사전 등록 루브릭)이며, 사람 채점 복제 킷은 포함되어
있으나 실행하지 않았습니다 — 한계는 숨기지 않고 명시합니다.

그리고 실제 운영 모노레포에서 (2026-07-06, 그 레포 자신의 장부 기준):

| 측정한 것 | 결과 |
|---|---|
| write-back 파일럿: capture → 장부 → compile → 신규 소비자 | 온보딩 답변 **5.5/10 → 9.5/10**, 비용 불변(1 도구 호출) |
| 신뢰 라벨의 소비자 도달 | 소비 에이전트가 시키지 않아도 자기 답을 `[generated]`로 명시 |
| Mode 3 백엔드 제안, 그라운드트루스 대비 채점 | goal C+ / risk D → 부트스트랩 채널로 재배치(위 참조) |

사전등록 전체 재실행은 달력이 아니라 볼륨 게이트(세션 10 + 판정된
write-back 10)에서 실행됩니다.

## 캡처: 훅 연결

라이브 캡처가 장부와 내장 측정 카운터를 먹입니다. Claude Code라면
프로젝트의 `.claude/settings.json`에 추가하세요 (훅은 프로젝트 디렉토리에서
실행되므로 `kervo`가 PATH에 있으면 됩니다):

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

다이제스트 자체도 신경 쓰지 않고 최신으로 — git post-commit 훅:

```bash
printf '#!/bin/sh\nkervo compile >/dev/null 2>&1 || true\n' > .git/hooks/post-commit
chmod +x .git/hooks/post-commit
```

훅은 밀리초 예산의 로컬 append입니다 — LLM 없음, 네트워크 없음, 세션을
절대 깨지 않음(쓰레기가 들어와도 exit 0). 커밋되는 장부에는 **이름·경로·
크기만** 저장됩니다: 프롬프트와 파일 내용은 머신을 떠나지도, git 히스토리에
들어가지도 않습니다.

```bash
kervo capture -type decision -body "JWT over sessions"   # 손으로 기록
kervo review                                             # 검토 큐: 제안을 하나씩 판정
kervo trust -id 01KWP -to verified -reason "team agreed" # ID로 판정 (스크립트용)
kervo status                                             # 한 화면 트러스트 뷰
kervo metrics                                            # 프롬프트 크기: artifact 유/무 비교
kervo import claude                                      # 과거 Claude Code 세션 백필
```

채팅에서 바로 판정하고 싶다면? MCP 서버를 등록하면 대화가 곧 검토
표면이 됩니다 — *"검토 큐 보여줘"* → *"2번 verify, 증거 확인했어"*:

```json
{ "mcpServers": { "kervo": { "command": "kervo", "args": ["mcp"] } } }
```

도구 4개: `read_context`(사실 내보내기), `kervo_capture`(write-back),
`review_queue` / `review_judge`(사람이 말한 판정의 중계 — 에이전트 자신의
판단은 금지). 대량 판정에는 `kervo review -web`이 일회성 로컬 페이지를
띄웁니다 — 명령이 살아 있는 동안만 존재하고 127.0.0.1에만 바인딩되며,
설계 보증(데몬 없음·계정 없음)을 전부 지킵니다.

### 전체 조망: `kervo dash`

`kervo compile`이 실행될 때마다 워크스페이스 **경로**(경로만, 머신 로컬,
커밋 안 됨)가 `~/.kervo/workspaces.json`에 등록됩니다. `kervo dash`는 그
전부를 한 페이지에 펼칩니다 — 레포별 대기 판정·28일 활동·트러스트
구성·프로젝트 개요(선언된 명령어, 최근 변경, 모듈)·커밋 이력으로 증명된
결합도·실제 연결된 어댑터 — 키보드 퍼스트 인라인 판정(`1`–`9` 레포 열기,
`j`/`k` 이동, `v`/`s`/`d` 판정, `?` 키 안내) 포함이며, 각 판정은 해당
레포 자신의 장부에 기록됩니다. 큐 아래 지식 뷰는 verified·observed
항목 전문을 — 주장 먼저, 증거 첨부 — 렌더하고, 퇴역 항목은 사유와 함께
남습니다.

<p align="center"><img src="assets/dash-detail.png" width="860" alt="kervo dash — 워크스페이스 상세: 판정, 개요, 결합도"></p> 진실은 레포별로 git에 남고, 대시보드는 저장소가 아니라
렌즈이며, 명령과 함께 사라집니다. UI는 사용자의 언어를 따릅니다 —
`$LANG`, `-lang en|ko|ja`, 또는 화면의 언어 스위처(선택은 다음 실행에도
유지). 관찰 본문은 제안자가 쓴 언어 그대로입니다.

## 명령어

| 명령 | 기능 |
|---|---|
| `kervo init` | 최초 1회: 스캔 → artifact → `CLAUDE.md` 주입 (멱등) |
| `kervo compile [-lang en\|ko\|ja] [-inject block\|import]` | 증분 재스캔 + 재컴파일; Mode 3 → 2 → 1 폴백 |
| `kervo capture -type <t> -body <내용>` | 관찰을 장부에 기록 |
| `kervo trust -id <접두> -to verified\|stale\|deprecated -reason <사유>` | ID로 관찰 판정 (스크립트용 프리미티브) |
| `kervo review [-web]` | 검토 큐 — 제안과 ⚠ 충돌을 하나씩 판정; `-web`은 일회성 로컬 페이지 |
| `kervo dash` | 전체 대시보드 — 등록된 모든 워크스페이스를 한 페이지에, 인라인 판정 |
| `kervo status` | 한 화면 장부 + 트러스트 뷰 |
| `kervo metrics` | artifact 유/무 프롬프트 크기 (내장 A/B 카운터) |
| `kervo import claude` | Claude Code 트랜스크립트에서 장부 백필 (크기만) |
| `kervo hook` | 소비자 훅 진입점 (stdin JSON, 밀리초 예산) |
| `kervo mcp` | stdio MCP 서버 — 컨텍스트 출력, write-back 수신, 채팅 판정 |
| `kervo version` | 버전 출력 |

## 설계 보증

- **결정적 스켈레톤** — 같은 워크스페이스, 같은 언어면 같은 바이트; CI의
  골든 파일로 고정. Fact 경로에 LLM은 절대 없음.
- **이벤트가 진실** — append-only JSONL 장부가 git에 커밋됨
  (`merge=union`); artifact와 인덱스는 파생물이며 재구축 가능. 레포를
  클론하면 컴파일된 기억이 함께 이동.
- **경계는 검사로** — 순수 코어는 어댑터를 import할 수 없음
  (`make arch-check`); 데이터에서 온 텍스트는 구조 마커를 사칭할 수 없음;
  프로바이더는 `generated` 위로 자기 승격 불가.
- **서버 없음, 데몬 없음, DB 없음, 계정 없음** — 모든 상태는 `.kervo/`와
  `CLAUDE.md`에.

## 현재 상태

v0, 빠르게 움직이는 중. 시맨틱 슬롯을 켠 Cold Start 검증 통과;
캡처 → 검증 → stale 루프가 이 저장소에서 실가동 중입니다. PRD / RFC /
실험 프로토콜은 안정화되는 대로 공개합니다.

---

kervo는 코딩 도구가 아닙니다. git 위에서 사는 모든 팀을 위한 기억
계층입니다 — 개발자는 이미 작업을 커밋으로 저장하고 있기에 첫 시장일
뿐입니다.

라이선스: [Apache-2.0](LICENSE).
