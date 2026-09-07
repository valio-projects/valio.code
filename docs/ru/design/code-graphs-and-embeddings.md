# Графы кода и эмбеддинги

Это проектная рекомендация, а не заявление, что предложенные графы уже существуют. Сначала документ фиксирует проверяемую текущую границу, затем выбирает графовые итерации, которые дают разработчику и агенту надёжные ответы с явными границами.

## Фактический аудит возможностей

Единица публикации — неизменяемый source view в рамках workspace. Он закрепляет снимки репозиториев, принадлежность файлов, ревизии проектов и синтаксический профиль. Исходный текст хранится в SurrealDB, а текстовый поиск проверяется по нему. При ingestion Go-файлы разбираются; JSON-отчёт на файл содержит синтаксические объявления, вхождения идентификаторов, импорты, контуры типов Go, сигнатуры функций и диагностику. Богатые дескрипторы переносятся в каталог с областью действия и доказательствами. Ссылка на символ бывает точной, кандидатной или неразрешённой: совпадение имени не превращается молча в точную связь.

| Область | Фактическое состояние | Граница |
|---|---|---|
| Исходный текст и поиск | Готовы: неизменяемая публикация файлов/view и проверяемый exact, substring и RE2 поиск | Поиск не делает семантического или графового вывода. |
| Контур Go AST | Создаётся при ingestion | Сохраняется JSON-отчёт парсера; постоянной проекции рёбер AST и общего AST-запроса нет. |
| Типы и символы | Частично: доступны объявления Go и богатый контракт типа | Сервер использует `go-ast-syntax/v1;syntax-default`; written type uses и ordinary uses не разрешаются. |
| Необязательный `go/types` | `AnalyzeWithOptions(CheckTypes: true)` проверяет один переданный Go-файл | Он не подгружает автоматически sibling files, build tags или module context; связывает только local definitions с local uses. При ingestion не включён. |
| Отпечаток синтаксиса | `structure.Fingerprint` хеширует форму корректного Go AST | При ingestion не вызывается и не сохраняется. Равные хеши — кандидаты с `equivalence=unknown`, не доказательство клона или поведения. |
| Ключи конфигурации | Agent хранит только безопасные позиции ключей JSON и `.env`, значения редактируются | Это захват метаданных, не разрешённый configuration/dependency graph. |
| Риск | `analytics.ScoreRisk` агрегирует предоставленные нормализованные факторы с явным порогом достаточности | Нет producer-а истории, покрытия, зависимостей или графа для этих входов. |
| Calls, CFG, DFG, PDG, CPG, aliasing, API/SQL/event, тестовые и исторические графы | Не реализованы | Public view корректно отмечает references, structural fingerprint и configuration graph как unsupported. |
| Vectors, embeddings, HNSW и context generation | Не реализованы | Нет vector store, model profile, embeddings или semantic endpoint. |

[`DefaultRegistry`](../../../src/back-end/internal/projections/registry.go) описывает десять семейств и проверяет порядок зависимостей, fingerprints и downstream invalidation. Это хороший контракт, но ingestion, persistence и read API его не вызывают; называть его работающим graph pipeline нельзя. Карта статусов ingestion отдельная и меньше.

SocratiCode — полезный сравнительный ориентир, но он не меняет этот аудит. Его resolver намеренно использует local name matching, imported-dependency name matching и один re-export hop. В реализации прямо сказано, что type inference нет, а method calls разрешаются по имени. Метки `local`, `unique`, `multiple-candidates`, `unresolved` — уровни уверенности для кандидатов, а не семантика вызова, подтверждённая компилятором. Это полезный шаблон UX, хранения candidate edges и отчёта о покрытии, но не доказательство, что такие рёбра уже есть в valio.code.

## Термины, которые нельзя смешивать

**AST** хранит синтаксическую вложенность. Он подходит для навигации, точных диапазонов, нормализации и syntax queries; он не разрешает имя и не доказывает call target.

**CFG** связывает исполнимые basic blocks возможными переходами управления. **CDG** (control dependence graph) выводится из CFG, обычно через post-dominance, и показывает, какой предикат управляет исполнением операции. CDG — не другое название CFG.

Под **DFG** здесь понимаются def-use/data-dependence рёбра при явно заданной модели анализа. **PDG** содержит *и data, и control dependences*; это не просто CFG и DFG рядом. Slice обязан объявлять seed, направление, виды dependence, interprocedural и alias policy, обработку exception/async paths. В исходном определении PDG оба вида зависимостей явны. [Ferrante, Ottenstein, Warren (1987)](https://bears.ece.ucsb.edu/class/ece253/papers/ferrante87.pdf)

**CPG** — property-graph представление, соединяющее AST, CFG и PDG на общих узлах программы. Это представление для запросов, не гарантия точного анализа, GNN или semantic equivalence. Качество ограничивают language front end, build configuration, dispatch/alias policy и полнота исходников. [Композиция у Yamaguchi et al.](https://www.sec.cs.tu-bs.de/pubs/2014-ieeesp.pdf); [спецификация Joern CPG](https://cpg.joern.io/) описывает CFG и control-dependence layers.

## Общие факты и производные проекции

Не следует создавать отдельную БД на каждый граф или делать CPG источником истины. SurrealDB остаётся единственным application store. Нужно хранить версионированные **факты** с доказательствами, затем материализовать узкие заменяемые проекции и adjacency/index records. У каждого узла/ребра должны быть immutable view, producer/profile/version, evidence IDs, status (`ready`, `partial`, `stale`, `unsupported` и т. п.) и счётчики полноты. Проекция ссылается на source span, а не дублирует исходник.

```text
immutable source view + project/build profile
  -> syntax facts, compiler/SCIP facts, external-contract facts
  -> symbol/reference/type/member facts
  -> call + CFG + def-use facts
  -> PDG/slices, effects, system/test/history projections
  -> retrieval candidates и цитируемый контекст агента
```

Поэтому symbol graph, call graph, CPG view, impact view и agent-context view — проекции общих фактов, а не конкурирующие хранилища. Каждый producer получает profile fingerprint: версию compiler/SCIP, language/build configuration, dependency-lock identity при её использовании, schema version и source fingerprint. Изменение профиля инвалидирует вывод и объявленных потомков. Существующий registry — начало механизма, но реальные зависимости надо уточнять: общее `data_flow -> call` слишком грубо для внутрипроцедурного def-use.

## Приоритетная карта развития

### 1. Symbols, references, types, members и calls, разрешённые сборкой

Это первый граф с наибольшей ценностью. Providers получают явный build profile и возвращают declarations, imports/modules, exact и candidate references, type/member selection и call sites. Для Go нужен package/module-aware front end, прежде чем называть cross-file use точным; существующий `go/types` показывает правильную основу object identity, но single-file запуск на это намеренно не способен. SCIP может быть import format/provider-ом при наличии stable identity и occurrence evidence, но не заменяет source/version provenance.

Сохранять `declares`, `refers_to`, `has_type`, `member_of`, `imports`, `calls`, `overrides/implements` только там, где это обосновано. Candidate calls держат все цели и причины; external/dynamic calls имеют отдельный unresolved/external state. Это даёт find-references, caller/callee exploration, rename preflight, API discovery и bounded impact traversal.

Нужны package loading, build tags/targets, lockfile/dependency resolution, language adapters, stable IDs и incremental per-package rebuilds. Инвалидировать изменённый пакет и reverse import/call consumers, не весь workspace. Fixtures включают shadowing, overloads, generics, import aliases, promoted members, generated-code policy, missing dependencies, dynamic dispatch и partial builds. Публиковать exact/candidate/unresolved counts по языку и view.

### 2. Внутрипроцедурные CFG, def-use, PDG и ограниченные slices

Для каждого resolved function body создавать basic blocks и явные edge kinds: `normal`, `true`, `false`, `return`, `panic/throw`, `defer/finally`, `exception`, если поддержано. CDG выводится из CFG, def-use — из явного SSA/IR или эквивалента. PDG образуется, только если есть оба входа.

Первыми продуктами будут backward slices («что влияет на эту запись, return или sink?»), forward slices («на что влияет этот input?»), кандидаты dead/unreachable с объяснением и control-path context для review. Каждый ответ содержит seed, направление, максимум nodes/edges/depth, timeout, включённые edge kinds и omitted/unknown behavior. Начать внутри функции. Позднему interprocedural summary engine нужны SCC/fixpoint/widening, context sensitivity и alias model; локальные slices нельзя молча превращать в whole-program выводы.

### 3. Points-to/alias, effects, resources и concurrency

Точность alias определяет, можно ли доверять reads/writes, side effects и interprocedural flow. До заявления о точном heap flow нужна настраиваемая points-to/escape model. Затем выводятся summaries эффектов: field/global read-write, I/O, transaction, network, filesystem, lock/channel/task creation и resource acquire/release. Async scheduling, await/join и lock order — отдельные факты; `happens-before unknown` допустим.

Кэшировать function/package summaries, использовать budgeted fixed points и показывать widening, timeout и unknown aliases. Это даёт более безопасный impact и ответы об effects, leak/deadlock candidates. Сначала проверить curated positive/negative fixtures.

### 4. System contracts и operational paths

Из parsers и framework adapters выводить API route/RPC/service, database schema/query, event producer/consumer, configuration-key/provider и infrastructure-resource relations. У каждого ребра contract/version, точный source/config span или imported artifact и resolution status. Совпадение строк не доказывает deployment reachability. Protocol/state-machine facts требуют declarations, annotations или verified patterns; heuristic transitions остаются кандидатами.

Это отвечает на «какие handlers достигают table?», «что публикует event?», «какие config keys влияют на endpoint?» и «какие resources использует job?». Так configuration capture становится полезным без хранения secret values.

### 5. Tests, coverage, history и risk

Импортировать test discovery и versioned coverage reports, привязать к тому же source view, затем вывести `tests`, `covers`, failure/stack и changed-symbol relations. Commit DAG, semantic diff, ownership и co-change добавлять только с явной history policy. Existing risk function получает только versioned normalized facts и сохраняет knownness/sufficiency. «Нет coverage» остаётся unknown, пока полный report для view не докажет отсутствие.

### Позже: детерминированная структура, затем оценка GNN

Текущий AST fingerprint остаётся labelled candidate filter. Сначала добавить deterministic features: normalized syntax, control/data motifs, symbol/type roles, path/context features и locality-sensitive candidate buckets. Они объяснимы и дёшевы в invalidation, но не доказывают semantic equivalence. GNN — необязательная ranking model над стабильным оценённым graph, не prerequisite CPG queries и не замена анализатора. Версионировать labels, graph snapshot/profile, repository/time split, cost, calibration и fallback.

## Контракт запросов агента

Graph operations должны возвращать доказательства раньше прозы.

| Операция | Обязательные границы | Обязательные доказательства |
|---|---|---|
| `symbol_resolve` / `references_find` | `viewId`, project/profile, symbol или span, candidate limit | Exact ID либо все candidates, причина resolution, spans, producer/profile и completeness. |
| `impact_trace` / `call_trace` | direction, edge kinds, depth, node/edge/time budget, external policy | Paths с edge evidence/confidence, truncated frontier, omitted dynamic/external paths. |
| `slice` | function/view seed span, forward/backward, edge kinds, interprocedural и alias policy, budget | Версия CFG/PDG producer, included nodes/edges, unknown aliases и cutoff reason. |
| `system_trace` | endpoint/event/table/config/resource seed, relation kinds и depth | Contract/source evidence, exact vs candidate edges, unresolved adapters и snapshot consistency. |
| `context_build` | task, token/edge/file budget, ranking profile и citation policy | Snippets с view ID, spans, relation paths, scores/features и exclusions. |

Каждая операция один раз разрешает `latest` в начале и возвращает concrete view ID. Mixed views отклоняются, если operation не поддерживает явно отмеченный composite view. Исчерпание budget — состояние результата, не пустой ответ.

## Эмбеддинги: выбирать по retrieval-задаче, не по ярлыку

Эмбеддинги дополняют lexical search и graph traversal. Они ранжируют кандидатов; не доказывают type, call, flow или equivalence. Использовать hybrid retrieval: exact identifier/text search для точности, vector retrieval для vocabulary mismatch, затем reranking и graph expansion в pinned view. Авторы BGE рекомендуют hybrid retrieval и reranking для BGE-M3. [Карточка BGE-M3](https://huggingface.co/BAAI/bge-m3)

| Ось/класс | Значение | Роль в valio.code |
|---|---|---|
| General multilingual text embedding | Training/task focus — обычный multilingual retrieval | Baseline для issues, документации и русско-/англоязычного query-to-code; может терять identifier/syntax signals. |
| Code-specialized embedding | Обучалась/оценивалась для code или text-to-code retrieval | Кандидат для code chunks и NL-to-code. Jina v2 base code заявляет 30 языков и Apache-2.0. [Model card](https://huggingface.co/jinaai/jina-embeddings-v2-base-code) |
| Deterministic structural features | Получены из syntax/semantic facts, не neural vector | Первый выбор для explainable structural similarity и graph-role ranking. |
| Graph/GNN representation | Модель получает заданный graph snapshot | Поздний ranking-вариант после надёжных graph facts и relevance labels. |
| Contrastive training | Learning objective сближает positives и раздвигает negatives | Может обучать general-text, code-specialized и graph models; это не конкурирующая storage/retrieval modality. Text/code contrastive pre-training применяется для code search. [Neelakantan et al.](https://cdn.openai.com/papers/Text_and_Code_Embeddings_by_Contrastive_Pre_Training.pdf) |

Ollama — runtime/inference выбор, не model family. Он годится для local baseline, только если закреплён и оценён точный model artifact. BGE-M3 разумно проверить как multilingual baseline: опубликованная карточка указывает 1 024 dimensions, вход 8 192 tokens, dense/sparse/multi-vector modes и MIT. Это **не** доказывает качество code search этого репозитория. Ollama сообщает, что ширина vector зависит от модели, поэтому schema collection надо probe-ить, а не предполагать. [Документация Ollama embeddings](https://docs.ollama.com/capabilities/embeddings)

Для каждой vector generation сохранять publisher/name и immutable revision/digest, license/terms review, runtime/quantization, tokenizer/context limit, native и stored dimensions, normalization, query/document prefixes, chunker и path/symbol context format, source-view/profile fingerprints, fusion/reranker version и access policy. Изменение model, dimension, chunking или prefix создаёт новую generation и требует reindex; vectors разных spaces не сравниваются. До remote provider применяются upload/redaction policy; raw source не попадает в telemetry.

## Evaluation gates

1. **Truth fixtures:** по языку positive, negative, ambiguous, incomplete-build и dynamic/framework cases; проверять edge evidence и candidate status, не только counts.
2. **Incrementality:** изменить один file/package/profile и проверить точно, какие generations стали stale и пересобрались; включить delete, rename и dependency-lock change.
3. **Query correctness:** precision exact/candidate, recall references/calls против compiler или curated oracle, agreement slices с вручную проверенными малыми программами и отсутствие cross-view joins.
4. **Retrieval:** held-out repository/time split; отдельные exact identifier, NL-to-code, bug/issue, cross-lingual и structural-similarity sets; публиковать Recall@K, MRR/nDCG, latency, token/context cost и citation coverage. Сравнивать lexical, vector, hybrid, reranked и graph-expanded варианты при одном budget.
5. **Operational bounds:** index/reindex time, storage на строку исходника, traversal fan-out, peak memory, timeout rate, unresolved/candidate rates и 10M-line benchmark до заявлений о масштабе.

Release gate — не «граф сохранён» и не «эмбеддинг возвращён». Результат называет pinned view, объясняет evidence path, показывает unknowns и остаётся в declared budget.
