# Проектирование поиска: текущее состояние и возможное развитие

## Статус и границы

Это **проектное предложение**, а не реализованная функция. В нём отдельно
указано, что уже доказано текущим backend, а что только может появиться в
системе поиска. Документ не обещает универсального прироста качества, лучшей
для всех модели или неизменного количества результатов для каждого
репозитория.

Предложение ограничено действующей политикой исходных данных: индексировать
можно только очищенные неизменяемые входы представления; workspace,
неизменяемое представление, project, repository и build profile должны
оставаться фиксированными при поиске и сборке контекста. Результат поиска —
это свидетельство, а не утверждение, что модель установила факт о программе.

## Аудит текущего backend

| Область | Что есть сейчас | Существенное ограничение |
|---|---|---|
| Единица поиска | Пакет search перебирает целые файлы выбранного неизменяемого представления, проверяет каждое положительное совпадение по исходному содержимому и только затем делает пагинацию. | Это не поиск по пассажам и не семантический поиск. |
| Индекс кандидатов | `BlockBytes` равен 16 KiB; есть перекрытие в две руны и нормализованные триграммы. Блоки выровнены по байтам исходника и служат фильтром. | Блок 16 KiB **не** является семантическим фрагментом и не имеет смысла объявления, функции или родителя/потомка. |
| Точность | Режим `exact` сравнивает целое выбранное поле; по умолчанию content/path/symbol — проверяемый подстрочный поиск; regex использует RE2. | BM25 или векторы не могут заменить этот контракт точного/подстрочного совпадения. |
| Парсер | Для Go используются `go/parser` и `go/ast`; сохраняются синтаксические диапазоны, объявления, импорты, вхождения и диагностики. | Tree-sitter сейчас не используется. |
| Типовая информация | Необязательный `go/types` получает только один разобранный файл. Mapper создаёт точные идентичности только локальных объявлений. | Импорты, объявления из других файлов и разрешение compiler/package остаются неразрешёнными; пройти от `PaymentService` по репозиторию нельзя. |
| Границы проекта/версии | Манифесты и неизменяемые представления фиксируют содержимое и membership проекта; поиск типов сохраняет неоднозначность. | Нет глобальной карты import/export, графа символов или обхода графа. |
| Контракты проекций | Семейства `vector` и `context` объявлены с зависимостями. | Это только контракты: embeddings, индекс BM25, reranker и builder контекста отсутствуют. |

Свидетельства находятся в [README поиска](../../../src/back-end/internal/search/README.md),
[триграммном индексе](../../../src/back-end/internal/search/index.go),
[Go-парсере](../../../src/back-end/internal/analysis/go_parser.go),
[проверке одного файла](../../../src/back-end/internal/analysis/type_checker.go)
и [реестре проекций](../../../src/back-end/internal/projections/registry.go).

## Предлагаемые записи поиска

### 1. Контекстные синтаксические фрагменты

Нужен языконезависимый порт `SyntaxChunker` с адаптерами языков. Для Go первый
адаптер должен использовать существующее Go AST. Tree-sitter можно добавить
позже для языков без подходящего compiler parser либо для терпимого
инкрементального разбора. Tree-sitter — генератор парсеров и инкрементальный
парсер с concrete syntax tree и байтовыми диапазонами; его подключение будет
новой явной зависимостью и решением о версиях грамматик, а не описанием
текущей реализации. [Документация Tree-sitter](https://tree-sitter.github.io/)
описывает эти свойства.

Для каждого фрагмента следует хранить:

* неизменяемые идентификаторы `viewId`, workspace, project, repository, file,
  language и build profile;
* стабильные `chunkId`, при необходимости `parentId`, ordinal, kind и исходный
  полуоткрытый диапазон байтов UTF-8;
* digest содержимого и версии chunker/grammar/profile, которыми он создан;
* идентичность объявления, если она известна, либо явное состояние
  `unresolved` или `ambiguous`; и
* сгенерированное представление для поиска, отдельное от исходного текста.

Границы фрагментов: файл/package, type, function/method и выбранные вложенные
body. Небольшая функция помещается в один фрагмент. Функция больше бюджета
представления режется по границам AST body, получает общий `parentId`,
упорядоченную последовательность и только ограниченное перекрытие
signature/header. Потомок не должен неявно заявлять, что содержит весь
родитель. Размеры границ и перекрытия — параметры, калибруемые на fixtures, а
не инварианты продукта.

Header injection может добавлять к embedding или reranking представлению
сгенерированный заголовок: язык, qualified declaration, signature, контекст
package/path и идентичность родителя. Его нужно хранить с меткой
`generated-representation`, версией и исходным байтовым диапазоном. Это не
исходный текст, не меняет его координаты и не делает ответ модели достоверным.
Показываемый исходник, highlight и evidence по-прежнему используют исходные
неизменяемые байты и записанные диапазоны.

### 2. Восемь профилей представления, включая код

Не следует делать summary единственным входом векторного поиска. Идентификаторы
кода, разрешённые политикой литералы, написание API и редкий текст ошибок часто
и являются причиной запроса. Summary может их опустить, расходиться с версией
исходника и добавляет стоимость генерации и инвалидирования.

Нужны независимо версионируемые профили вместе с каноническим исходным кодом.
Следующие восемь — начальная схема, а не обязательство включать каждый профиль
в каждом развёртывании:

| Профиль | Вход | Назначение и ограничение |
|---|---|---|
| `code-body/v1` | Очищенный канонический код фрагмента | Сохраняет идентификаторы и синтаксис; действует политика capture. |
| `declaration/v1` | Имя, qualified name, signature, kind | Сильная поверхность для запросов символов и API. |
| `documentation/v1` | Привязанные comments/docstrings | Полезно для естественного языка; документация может отсутствовать или устареть. |
| `header-context/v1` | Ограниченный generated header и body | Добавляет scope, не выдавая заголовок за source. |
| `identifier-lexicon/v1` | Токенизированные identifiers и aliases | Помогает редким именам; не заменяет exact. |
| `parent-summary/v1` | Явно сгенерированное versioned summary | Необязательное вспомогательное средство, помечено generated и не является единственным evidence. |
| `neighborhood/v1` | Ограниченные resolved relations и выбранные signatures | Доступно только после готовности графа; не разворачивает целый класс. |
| `path-metadata/v1` | Путь в repository, package/module и язык | Помогает навигации; не расширяет access scope. |

Fingerprint каждого профиля включает source content, схему профиля, tokenizer,
модель/version embedding, prompt/header template при его наличии и версии
вышестоящего графа. Изменение source инвалидирует потомков; изменение
profile/model инвалидирует только этот профиль и зависимые векторы. Это
соответствует уже существующему принципу fingerprint проекций, а не считает
векторы вечными данными.

### 3. Граф import/export и символов

До обещания cross-file navigation следует построить версионируемую карту
import/export для конкретного build profile. Каждый языковой адаптер должен
выдавать declaration/export records, imports, references и исход разрешения:
`exact`, `candidate`, `unresolved` или `unsupported`. Разрешение хранит
выбранный scope и evidence; одинаковые имена в двух пакетах остаются разными
кандидатами.

Только после появления такого графа запрос «follow `PaymentService`» сможет
разрешить его declaration, imports/re-exports, references, implementations или
выбранные members. Текущие import strings и локальные Go type facts — полезные
входы, но не такой граф. Compiler-backed resolution требует корректных
package/module/build inputs и не должен выводиться из одного написания имени.

Расширение контекста начинается только с уже выбранного chunk либо resolved
symbol. Оно ограничено фиксированными scope, depth, bytes/tokens, числом nodes
и политикой selected members. Добавляются signature объявления, прямо
релевантная call/reference edge и только member/body, выбранный запросом или
reranker. Нельзя раскрывать целый большой class, package или repository только
из-за одного совпадения.

## Предлагаемый pipeline поиска

Нужно сохранить три независимо наблюдаемых канала:

1. **Проверяемый точный канал.** Существующий exact/substring/regex поиск
   остаётся авторитетным для literal identifiers, paths и source spans. Он
   возвращает проверенные совпадения, даже когда semantic retrieval недоступен.
2. **Лексический ранжированный канал.** Добавить persisted BM25-подобный индекс
   в scope представления по разрешённым полям chunk/profile. BM25 ранжирует
   лексические свидетельства, но не даёт нынешнюю гарантию точной подстроки.
3. **Dense ранжированный канал.** Добавить embeddings выбранных профилей с
   закреплёнными в записи model/tokenizer/dimension/distance/profile versions.
   Dense vectors полезны для paraphrase и intent, но не являются ни всегда
   плохими для identifiers, ни всегда хорошими для кода. Оба класса запросов
   надо измерять отдельно.

Лексический и dense ранги следует объединять, а не складывать несопоставимые
raw scores. Reciprocal Rank Fusion — разумный начальный baseline, потому что
объединяет ранжированные списки без предположения о калибровке score; его
константа и глубина каналов всё равно подбираются offline. Исходная работа RRF
исследует fusion ранжированных систем, но не этот продукт и не его corpus.
[Cormack, Clarke, Büttcher (SIGIR 2009)](https://doi.org/10.1145/1571941.1572114).

Cross-encoder reranker может оценивать query и ограниченное представление
кандидата после fusion. Cross encoder обычно качественнее bi-encoder, но имеет
стоимость на пару query/document; поэтому распространён retrieve-then-rerank.
[Sentence Transformers описывает этот компромисс](https://www.sbert.net/docs/cross_encoder/usage/usage.html).

Для экспериментальной калибровки можно начать не более чем с 50 fused
candidates и возвращать не более 5 reranked chunks только если это укладывается
в downstream context budget около 8k tokens. Это не постоянное правило
`top 50 → 5`: фактические пороги определяют тип запроса, recall кандидатов,
размер corpus, latency и бюджет контекста. API должен сообщать число
кандидатов, membership каналов, profile/model versions, truncation и пропуск
reranking.

Для RU/EN запросов и code context до выбора нужно проверить language coverage,
code-query quality, context limit, deployment constraints, cost и license.
Документация Cohere заявляет multilingual support, но также считает суммарные
query/document tokens в context limit модели; это capability поставщика, а не
benchmark поиска по коду. Полезны [детали Cohere rerank](https://docs.cohere.com/docs/rerank)
и [поддержка языков](https://docs.cohere.com/docs/rerank-overview). Локальный
кандидат — multilingual `bge-reranker-v2-m3` от BAAI; model card называет его
reranker и указывает Apache-2.0, что всё равно требует legal и deployment
review проекта. [Model card BAAI](https://huggingface.co/BAAI/bge-reranker-v2-m3).

Ни один provider не должен получать source за пределами действующей capture
policy или подтверждённой границы вызывающего. Недоступность provider,
отклонённый content, превышение context или несовместимая license оставляют
exact и lexical results доступными и сообщают, что semantic channel недоступен
либо partial.

Dense retrieval показал ценность в собственных условиях оценки, но это не
переносится как гарантированный результат для данного repository. DPR сообщил
улучшения над BM25 baseline в open-domain QA, а CodeSearchNet создан именно
потому, что natural-language-to-code retrieval имеет отдельный semantic gap и
требует code-specific evaluation. [DPR](https://arxiv.org/abs/2004.04906) и
[CodeSearchNet](https://arxiv.org/abs/1909.09436) поддерживают необходимость
измерять, а не предполагать результат этого дизайна.

## План реализации и оценки

| Приоритет | Необходимая возможность | Зависимости и инвалидирование | Acceptance fixtures и измерения |
|---|---|---|---|
| P0 | Запись retrieval, scope guard, registry профилей, provenance response | Immutable views, policy filter, projection fingerprints | Запрет cross-workspace; source range не меняется после header injection; в ответе видны profile/version. |
| P1 | AST chunker и source/header representations | Языковой адаптер; source digest; parent/child records | Пустой/error-tolerant файл, nested declarations, split большой функции, Unicode byte ranges, ограниченное signature overlap. |
| P2 | Persisted exact и BM25 каналы | View-scoped postings и deterministic analyzer | Recall редкого identifier/path в exact равен 100% внутри scanned scope; BM25 не заявляет exactness; измерения latency и index size. |
| P3 | Карта import/export и selected graph expansion | Per-language resolver, build profile, symbol/reference projection | Два имени `PaymentService`, aliases, re-exports, unresolved import, version change; нет whole-class expansion. |
| P4 | Dense profiles и vector index | P0–P3, выбранные model/tokenizer, operational и license review | RU/EN natural-language и identifier sets, code-only queries, stale-summary detection, recall@K/NDCG@K, embedding cost и rebuild time. |
| P5 | Fusion и reranking | Candidate telemetry, bounded representations, provider/local adapter | Перебор параметров RRF; recall до rerank; MRR/NDCG@5, p50/p95 latency, rejected/oversize documents, per-query cost. |

Нужен reviewed relevance set из реальных очищенных вопросов по repository,
стратифицированный по языку (RU/EN), типу запроса (identifier, exact phrase,
natural language, path, symbol-following), размеру проекта и ambiguous names.
Для каждого judgement фиксируются view и ожидаемое evidence. Следует сообщать
Recall@K до reranking, MRR и NDCG@5 после ранжирования, coverage exact channel,
coverage/ambiguity resolution, p50/p95 latency, bytes/tokens у каждого
provider, размер индекса, длительность rebuild и денежную стоимость на
индексированный byte и query. Сравнение ведётся с текущим verified search
baseline; нельзя публиковать общее утверждение «в X раз лучше» без fixture,
corpus, version и confidence interval, которые его подтверждают.

## Итог решения

Следующий безопасный шаг реализации — P0/P1: durable scoped syntax chunks с
исходными координатами и versioned representations. BM25, vectors, global
symbol resolution и reranking зависят от этих записей и evaluation fixtures.
Существующий verified search остаётся обязательным путём точного evidence на
всех этапах.
