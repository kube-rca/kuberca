# Changelog

## Unreleased

### BREAKING

- Relicensed from MIT to Apache License 2.0. Existing forks must update their LICENSE references. Contributor consent for this change tracked in `.github/RELICENSE_CONSENT.md`. (#<TBD-PR-number>)

## [1.3.2](https://github.com/kube-rca/kuberca/compare/1.3.1...1.3.2) (2026-05-10)


### Bug Fixes

* **agent/deps:** bump python-multipart to 0.0.28 (CVE-2026-42561) ([#87](https://github.com/kube-rca/kuberca/issues/87)) ([1b30cb2](https://github.com/kube-rca/kuberca/commit/1b30cb24809e41f5306dd08e264820caa534ea32))
* **ci:** bump trivy-action to v0.36.0 in release-please workflow ([#91](https://github.com/kube-rca/kuberca/issues/91)) ([f3872f4](https://github.com/kube-rca/kuberca/commit/f3872f42be0b6ffac154733f72fa0801461d0425))

## [1.3.1](https://github.com/kube-rca/kuberca/compare/v1.3.0...1.3.1) (2026-05-10)


### Bug Fixes

* **agent:** use TraceQL char class for namespace fallback regex ([#88](https://github.com/kube-rca/kuberca/issues/88)) ([48599ee](https://github.com/kube-rca/kuberca/commit/48599ee334cb4702cea7a13157ab1b96fdcf9259))
* **release-please:** sync values.yaml/appVersion/CHANGELOG with manifest version ([#85](https://github.com/kube-rca/kuberca/issues/85)) ([e7bc445](https://github.com/kube-rca/kuberca/commit/e7bc4452e7ea8b3bd1f7bf5893c97de62ab000dc))

## [1.3.0](https://github.com/kube-rca/kuberca/compare/v1.2.0...v1.3.0) (2026-05-10)


### Features

* **chaos,charts,agent:** HashMode redaction demo + align AI model defaults ([#79](https://github.com/kube-rca/kuberca/issues/79)) ([c7dc068](https://github.com/kube-rca/kuberca/commit/c7dc068c04f1c18b2c352cc757bff6d2e6f2fb04))
* **frontend:** make header logo clickable to navigate home ([#82](https://github.com/kube-rca/kuberca/issues/82)) ([ea80328](https://github.com/kube-rca/kuberca/commit/ea803288ada258ae28577c46d837f0e59153db1f))


### Bug Fixes

* **agent:** redactor coverage for safe-prefix annotations and arbitrary args ([#84](https://github.com/kube-rca/kuberca/issues/84)) ([b54d5e4](https://github.com/kube-rca/kuberca/commit/b54d5e454ad774a050e8dc887f18c8598fd22f81))
* **chaos:** add make clean-redaction-demo for orphaned-resource cleanup ([#83](https://github.com/kube-rca/kuberca/issues/83)) ([3009128](https://github.com/kube-rca/kuberca/commit/30091280bfaa98574d9f550368633bd8eccb0976))
* **chaos:** normalize JWT inputs in redaction-demo so 5 positions hash to same digest ([#81](https://github.com/kube-rca/kuberca/issues/81)) ([9d13815](https://github.com/kube-rca/kuberca/commit/9d13815519831769f8189712e5665eb7f6a2a98f))

## [1.2.0](https://github.com/kube-rca/kuberca/compare/v1.1.0...v1.2.0) (2026-05-05)


### Features

* **agent:** expand client coverage with monkeypatch fakes + testcontainers ([#76](https://github.com/kube-rca/kuberca/issues/76)) ([#78](https://github.com/kube-rca/kuberca/issues/78)) ([21fe00e](https://github.com/kube-rca/kuberca/commit/21fe00ec8ae961fb5199d52f15f961fb0d8029c9))
* **agent:** expose session window size + align helm defaults to python config ([#62](https://github.com/kube-rca/kuberca/issues/62)) ([118a68c](https://github.com/kube-rca/kuberca/commit/118a68c8e39fa4b0d32db4c718834a7277d7c1c9))
* **agent:** gate istio crd tools via agent.istio.enabled helm flag ([#65](https://github.com/kube-rca/kuberca/issues/65)) ([24c8c5f](https://github.com/kube-rca/kuberca/commit/24c8c5f638c18460de877a7306259e9266b7cca2))
* **backend:** integration tests for internal/db with testcontainers-go and bump gate to 30% ([#76](https://github.com/kube-rca/kuberca/issues/76)) ([#77](https://github.com/kube-rca/kuberca/issues/77)) ([d6d197f](https://github.com/kube-rca/kuberca/commit/d6d197fd14b85ff89239ce44409a942708146065))
* **log:** add Sanitize helper to strip control characters ([e2ba3a1](https://github.com/kube-rca/kuberca/commit/e2ba3a1768a61b3cfc5ace771da5611f560fac3b))
* **log:** add sanitize_log helper to strip control characters ([08c9eaf](https://github.com/kube-rca/kuberca/commit/08c9eaf36438cc9f6b9239c93eaffeb5b94146a6))


### Bug Fixes

* **agent:** clean up bandit findings + drop continue-on-error gate ([#64](https://github.com/kube-rca/kuberca/issues/64)) ([af02bfd](https://github.com/kube-rca/kuberca/commit/af02bfd7ad268da07abdcd4d151078270d6a1dad))
* **agent:** sanitize user-controlled values in logger calls (CodeQL py/log-injection) ([5f2b01b](https://github.com/kube-rca/kuberca/commit/5f2b01b48ec2a88376d74e8c18a685f1554cf094))
* **agent:** validate http/https scheme + silence bandit B310 ([#41](https://github.com/kube-rca/kuberca/issues/41)) ([0254bbc](https://github.com/kube-rca/kuberca/commit/0254bbcdfddd5b88161443b9a0f1ab288863e1d3))
* **api:** sanitize user-controlled values in logger calls ([b3f0f41](https://github.com/kube-rca/kuberca/commit/b3f0f416b9e7dc208dc0b3a4dc36b12095ac906b))
* **backend:** close residual go/log-injection (CodeQL barrier model + coverage) ([f92a41d](https://github.com/kube-rca/kuberca/commit/f92a41dfaf84b388dda22c4c999ad587fd6cec43))
* **backend:** sanitize remaining log-injection sites in handler and client ([c1586b3](https://github.com/kube-rca/kuberca/commit/c1586b304fbd33fd79ba028f3d70259f3ca78fd8))
* **backend:** sanitize user-controlled values in log calls (50 CodeQL findings) ([e7dc25a](https://github.com/kube-rca/kuberca/commit/e7dc25a4c2a122e767fae308612f45d433c644b3))
* **ci:** allow golang.org/x/arch license (compound BSD-3-Clause + Google patent) ([#39](https://github.com/kube-rca/kuberca/issues/39)) ([ca55a39](https://github.com/kube-rca/kuberca/commit/ca55a3975825d28f95ea25d23449133f2f0a86f0))
* **ci:** bump trivy-action to v0.36.0 (setup-trivy v0.2.1 missing) ([6cec18a](https://github.com/kube-rca/kuberca/commit/6cec18a14a9bfaf8528fa91b1821538e9d8a3cfc))
* **ci:** bump trivy-action to v0.36.0 to fix missing setup-trivy tag ([5e65568](https://github.com/kube-rca/kuberca/commit/5e65568b29bb36e22e62bbac3aea156f069d1b4f))
* **ci:** correct helm/kind-action SHA in chaos-weekly workflow ([#75](https://github.com/kube-rca/kuberca/issues/75)) ([d630d65](https://github.com/kube-rca/kuberca/commit/d630d65e5e99430de46777d9e3133128826f8e3f))
* **deps:** bump postcss to &gt;=8.5.10 (CVE) ([70b794d](https://github.com/kube-rca/kuberca/commit/70b794da894fa27c15c1ad38cc04e2f87f575276))
* **deps:** bump postcss to &gt;=8.5.10 (CVE) ([12e90b9](https://github.com/kube-rca/kuberca/commit/12e90b9a35b5e274894aec421ee54da01b1c563c))
* **deps:** bump python-dotenv to &gt;=1.2.2 (CVE) ([b0f2392](https://github.com/kube-rca/kuberca/commit/b0f2392307419b259ce9401fbf135b8957a6c10f))
* **deps:** bump python-dotenv to &gt;=1.2.2 (CVE) ([d3b2a85](https://github.com/kube-rca/kuberca/commit/d3b2a8577a61aad84ffae79c60fc49f1fa314f65))
* **logutil:** drop early returns so Sanitize is recognized as a barrier ([a007ac7](https://github.com/kube-rca/kuberca/commit/a007ac7bbb2fcf6718f2e422dee5d4ab74965f28))
* **logutil:** drop early returns so Sanitize is recognized as a barrier ([f0c6c74](https://github.com/kube-rca/kuberca/commit/f0c6c744e0e51dd1f836c2a109214ef15a20a57c))
* **security:** close residual 3 log-injection alerts ([19d35b1](https://github.com/kube-rca/kuberca/commit/19d35b1e7da168e6326a3ba139d7c69d7e2ab468))
* **security:** close residual 3 log-injection alerts ([5486c09](https://github.com/kube-rca/kuberca/commit/5486c09740abf0c615bc37a6efbccf407b6fc910))
* **service:** sanitize user-controlled inputs in log calls ([d6831f1](https://github.com/kube-rca/kuberca/commit/d6831f183fa7c4a94e18e1be668ea8e0e7cc5d6c))
* **test:** bump vitest to v4 for vite 7 compatibility, add coverage gitignore ([2a87300](https://github.com/kube-rca/kuberca/commit/2a87300646c02fd8b78e647da0e83ac6610db453))

## [1.1.0](https://github.com/kube-rca/kuberca/compare/v1.0.0...v1.1.0) (2026-04-26)


### ⚠ BREAKING CHANGES

* **license:** KubeRCA is now distributed under Apache License 2.0. Existing forks must update their LICENSE references when picking up changes from main after this commit lands. Contributor consent for the relicense is tracked in .github/RELICENSE_CONSENT.md.

### Features

* add ENG ver. ([d7521b6](https://github.com/kube-rca/kuberca/commit/d7521b600d121352a404466b6ab2c499f1e51cd8))
* **chaos:** harden teardown to prevent residual pod-netns state ([#9](https://github.com/kube-rca/kuberca/issues/9)) ([75d4a59](https://github.com/kube-rca/kuberca/commit/75d4a59fd2148e82e4c0a911332ae26f02de684b))


### Bug Fixes

* **agent:** reduce incident summary prompt size and raise anthropic max tokens ([#8](https://github.com/kube-rca/kuberca/issues/8)) ([39025a3](https://github.com/kube-rca/kuberca/commit/39025a3a9e6ec50b9e6f3f2659cb9a7cc65815d7))
* **deps:** patch security vulnerability alerts ([#10](https://github.com/kube-rca/kuberca/issues/10)) ([990c93a](https://github.com/kube-rca/kuberca/commit/990c93a56d1c5b3c75c7067c19ec09f7cb8c40a3))
* **frontend:** bump vite to 7.3.2 for security patches ([e224e59](https://github.com/kube-rca/kuberca/commit/e224e593e92f4d2dd2141867f50ee1e16f2cd4a5))
* lint error ([9977099](https://github.com/kube-rca/kuberca/commit/997709930432dc6758aac0e9bec4d20caf4ca629))
* test error ([fdcd0ce](https://github.com/kube-rca/kuberca/commit/fdcd0ce19aacbb121fee4e52a30ec4eb44c7e537))


### Miscellaneous Chores

* **license:** relicense to Apache-2.0 ([383d239](https://github.com/kube-rca/kuberca/commit/383d2398628d7cbcf1fafdd21b011c0a5f956dd8))
* pin next release to 1.1.0 ([bf23eaa](https://github.com/kube-rca/kuberca/commit/bf23eaa65017520575b7f7ee2162da7afcec0ea9))

## 1.0.0 (2026-04-05)


### Features

* Add a required `name` field to webhook configurations, includin… ([52c767c](https://github.com/kube-rca/kuberca/commit/52c767cf8432b0c1a80b547159eed0093bb36b95))
* Add a required `name` field to webhook configurations, including database schema, API requests, and validation logic. ([951eff1](https://github.com/kube-rca/kuberca/commit/951eff11a5f9841aee05c24cba444b5ce2f9202f))
* Add alert variables and refactor webhook body variable insertion with a reusable component and improved UI. ([dd60fd9](https://github.com/kube-rca/kuberca/commit/dd60fd95b705a161dc0208e9796b72fcc4de9150))
* Add alert variables and refactor webhook body variables ([f10417f](https://github.com/kube-rca/kuberca/commit/f10417fb78f4df7ca291c61bdba0ba88edd1a093))
* Add alert variables and refactor webhook body variables ([f10417f](https://github.com/kube-rca/kuberca/commit/f10417fb78f4df7ca291c61bdba0ba88edd1a093))
* add analysis mode indicator to sidebar (KAN-172) ([2e60ff0](https://github.com/kube-rca/kuberca/commit/2e60ff025953593c9a2fa31a5d17d0bd65a7ac0d))
* add analysis mode UI with manual analyze buttons (KAN-172) ([f896b86](https://github.com/kube-rca/kuberca/commit/f896b8629cb8d3292f53e95bf853fcc54717dcba))
* add auto/manual analysis mode with severity-based control (KAN-172) ([053ed52](https://github.com/kube-rca/kuberca/commit/053ed529d77bc00ac08e1e2f5537819f6101743f))
* add bg color to reserved field ([3de4862](https://github.com/kube-rca/kuberca/commit/3de48622a7dd07e17b08fb2ba5b5da5d06a22784))
* add bg color to reserved field ([1874aec](https://github.com/kube-rca/kuberca/commit/1874aec6ed84c3c4d76f0ffb348d7c514638c24b))
* add built-in redaction masking-chaining [KAN-174] ([ba19e50](https://github.com/kube-rca/kuberca/commit/ba19e50578c9e29b4f359f2b30651cc2325ebf6d))
* add built-in Redaction option [KAN-174] ([b1dfa5f](https://github.com/kube-rca/kuberca/commit/b1dfa5fdc1f60b5f9fdc05e1c0efe2de2084ce31))
* add cert manager and move ingress hosts to 2 domains ([ed8959c](https://github.com/kube-rca/kuberca/commit/ed8959cda3192fd944c2cd2ab60715ae8d432455))
* add firing/resolved analysis split to alert detail API ([504a745](https://github.com/kube-rca/kuberca/commit/504a7459bbdd655fbb9cf14d5d62213f844621d9))
* add gp2 values for stateful apps ([f34c8a5](https://github.com/kube-rca/kuberca/commit/f34c8a5175aea0c7a119c5d1154d2dd4f9acad43))
* add info severity to AnalysisSettings and RCADetailView ([24f3935](https://github.com/kube-rca/kuberca/commit/24f39355c88ec1791fa2d39a70427b875bad7f36))
* add monorepo GitHub Actions workflows (CI + release-please) ([0911d33](https://github.com/kube-rca/kuberca/commit/0911d3329aacaac40baa99270db55d6735acedbc))
* add multi-provider OIDC support with auto-detection ([d330906](https://github.com/kube-rca/kuberca/commit/d33090677a923fffe49561a78847b75b69ba12fc))
* add notifier events and webhook-based routing ([5d9164a](https://github.com/kube-rca/kuberca/commit/5d9164a82b7042738967f0a8941afd5af7e39534))
* add notifier events and webhook-based routing ([5d9164a](https://github.com/kube-rca/kuberca/commit/5d9164a82b7042738967f0a8941afd5af7e39534))
* add notifier events and webhook-based routing ([7d14423](https://github.com/kube-rca/kuberca/commit/7d14423de38dbbbf53b1cd17b4662147a099f01a))
* add OIDC authentication with security hardening ([ed19018](https://github.com/kube-rca/kuberca/commit/ed190180fdba1493b9c084502e97851a62a67f67))
* add OIDC login button and callback error handling ([5b375ac](https://github.com/kube-rca/kuberca/commit/5b375acc8985b0726a9daa5dec9670c348af43fe))
* add settings option ([adbb4c2](https://github.com/kube-rca/kuberca/commit/adbb4c27a3c402da32b7dd44c0b2ab50f7563147))
* add settings option ([a03a65c](https://github.com/kube-rca/kuberca/commit/a03a65c86e13a51a48df0b00aac83ba1653ffe5b))
* add settings option ([dfc4bca](https://github.com/kube-rca/kuberca/commit/dfc4bca41fd9f34be51b23843e1714718ab0cf03))
* add settings page ([7871de3](https://github.com/kube-rca/kuberca/commit/7871de33523295a1a22cc05f2b786a9b1448b974))
* add settings page ([5747f98](https://github.com/kube-rca/kuberca/commit/5747f9899e09b1d2fd72a03141008cdd21f09da8))
* add severity-channel mapping rule ([562a9bb](https://github.com/kube-rca/kuberca/commit/562a9bba2e0695cea3ef6df96762fc375e2ce380))
* add severity-channel mapping rule ([9b39f93](https://github.com/kube-rca/kuberca/commit/9b39f93db2d8801e59001d45266b5e798508640f))
* Add show/hide toggle for token input fields and remove an infor… ([c2357ee](https://github.com/kube-rca/kuberca/commit/c2357ee1819788e67ec67263386f4c3c8333eb54))
* Add show/hide toggle for token input fields and remove an informational section in webhook settings. ([44c593f](https://github.com/kube-rca/kuberca/commit/44c593f4f662d08c94706f5a6178636dcd3f423d))
* add SSE support for real-time event notifications (KAN-165) ([a80928e](https://github.com/kube-rca/kuberca/commit/a80928e424670441745eaeb64ad4da21e2a845ce))
* add tenacity-based LLM API retry for transient errors (KAN-159) ([d28cf83](https://github.com/kube-rca/kuberca/commit/d28cf8312c0655db37586c91d5d5cd54ec6cdbfe))
* add total timeout stop condition to LLM retry logic ([99e341f](https://github.com/kube-rca/kuberca/commit/99e341f2ed55c89dfba2fd86dcdf52914f4efda6))
* add webhook config api ([6564238](https://github.com/kube-rca/kuberca/commit/65642388bcc40e85670b0a04e493ce35ebdde039))
* add webhook config api ([6564238](https://github.com/kube-rca/kuberca/commit/65642388bcc40e85670b0a04e493ce35ebdde039))
* add webhook config api ([ad05fbb](https://github.com/kube-rca/kuberca/commit/ad05fbba446a27d5a7bced618f5e79702ce138b5))
* add webhook list page ([7032e61](https://github.com/kube-rca/kuberca/commit/7032e61283e63176dc5d899512362a974476047a))
* add webhook list page ([da59fe4](https://github.com/kube-rca/kuberca/commit/da59fe4df14e4261a235bc3c100b08679e480354))
* **agent:** add generic manifest read tools ([4d4da60](https://github.com/kube-rca/kuberca/commit/4d4da60120f4368ecef24a213a3f55b9ece9c6d9))
* **agent:** add Loki log aggregation client and Strands tools ([6a30f9e](https://github.com/kube-rca/kuberca/commit/6a30f9ee9e2e3786fa16ef2f1f3f9fcb729016f7))
* **agent:** add node-level alert analysis support ([9a22db2](https://github.com/kube-rca/kuberca/commit/9a22db2179b10d031455abec6c827cac1fc90b5b))
* **agent:** add regex masking for llm and storage flows (KAN-146) ([#10](https://github.com/kube-rca/kuberca/issues/10)) ([ae71166](https://github.com/kube-rca/kuberca/commit/ae71166b651fb9074235b9cbc833589810f3b9f0))
* **agent:** add tool timing logs for analysis ([d7c8f39](https://github.com/kube-rca/kuberca/commit/d7c8f3965ccc29319053ecb4aaa457f9f9fe579d))
* **agent:** async 엔드포인트 전환으로 분석 병렬 처리 지원 ([fa74bf1](https://github.com/kube-rca/kuberca/commit/fa74bf1ac520b1d6c4477e78ea5a7e70b4c0cb34))
* **agent:** enrich daemonset and events context (KAN-79) ([bc6ae22](https://github.com/kube-rca/kuberca/commit/bc6ae22ae86ac03e6209a6a4a11e3ad1332137c5))
* **agent:** enrich daemonset and events context (KAN-79) ([f8b7c62](https://github.com/kube-rca/kuberca/commit/f8b7c62a9dde6efe9002f4158ccc98931140cdb6))
* **agent:** integrate tempo trace context into analysis (KAN-145) ([#9](https://github.com/kube-rca/kuberca/issues/9)) ([db45189](https://github.com/kube-rca/kuberca/commit/db45189bf50d4dc97745c7056ee3d568e2136333))
* **agent:** portable K8s baseline context and capability-aware prompt ([f96e4c4](https://github.com/kube-rca/kuberca/commit/f96e4c4dc2c81779d8042df39f15c5d3db431861))
* **agent:** tune LLM retry to 3-min exponential backoff ([d0c9b59](https://github.com/kube-rca/kuberca/commit/d0c9b59a1459c0f9478b93745fb3b0e25aec4d80))
* Alanlysis Dashboard 추가 (KAN-171) ([9097016](https://github.com/kube-rca/kuberca/commit/909701636b91e64897027a941d16446a1fb838d2))
* Alanlysis 기능 추가 (KAN-170) ([bebecbe](https://github.com/kube-rca/kuberca/commit/bebecbec1aa21d3b1c20bf49319b7fd1aa04c39e))
* alert flapping (KAN-155) ([40be145](https://github.com/kube-rca/kuberca/commit/40be1452552a6eb9f0942eb57e4593e8e4673f1b))
* alert flapping 환경변수 (KAN-155) ([e595663](https://github.com/kube-rca/kuberca/commit/e595663fef695c4d41f76f0583278b2a1d8fc451))
* allow info severity payload ([6f71009](https://github.com/kube-rca/kuberca/commit/6f710098004fd589713f1c89f67cb3a10e100ee5))
* analysis export 및 꺾은선 value 추가 ([70e3eb1](https://github.com/kube-rca/kuberca/commit/70e3eb117ef7c5696d8fe5e2586f60aa5442b653))
* **analyze:** add analysis SSE events and is_analyzing API field (KAN-173) ([0d58547](https://github.com/kube-rca/kuberca/commit/0d5854751ce0cee6f89e74c54d92925b99d71c82))
* **analyze:** add in-flight dedup guard and configurable agent timeout ([4c02ea4](https://github.com/kube-rca/kuberca/commit/4c02ea4cad1b2e4f057d711c95f47b8dec9a2a43))
* **analyze:** add incident analysis polling and remove banner auto-dismiss ([8654004](https://github.com/kube-rca/kuberca/commit/8654004d997b41a52935d1425fbe102a14638705))
* **analyze:** differentiate resolved alert analysis from firing ([f2c73a8](https://github.com/kube-rca/kuberca/commit/f2c73a81570eb57a1cd757a1a605083b95ae4ce3))
* **analyze:** differentiate resolved alert analysis from firing ([e9fadb7](https://github.com/kube-rca/kuberca/commit/e9fadb735946d6176a561c5f9b63133bc0ec668e))
* apply same readability improvements to AlertTable ([96a9915](https://github.com/kube-rca/kuberca/commit/96a9915c0a413501abe61c9db488e189ceb88b20))
* **backend:** add IsManual flag to AlertStatusChangedEvent for manual resolve Slack prefix ([a0cdc23](https://github.com/kube-rca/kuberca/commit/a0cdc2338b5c0f64785c98b88d5e103885dd9c5d))
* **backend:** add manual alert resolve request/response models ([19831b5](https://github.com/kube-rca/kuberca/commit/19831b5dd8af5040c6a5915dd035d987f3259019))
* **backend:** add ManualResolveAlert DB function ([eb78a15](https://github.com/kube-rca/kuberca/commit/eb78a15c1a9a721c969bd65cf9b51a7ee6485365))
* **backend:** add ResolveAlert and BulkResolveAlerts handlers with alertService dependency ([940ff56](https://github.com/kube-rca/kuberca/commit/940ff562eaa308fe9d4e318d247f277f387fa59a))
* **backend:** add ResolveAlert and BulkResolveAlerts service methods with tests ([1d2b0be](https://github.com/kube-rca/kuberca/commit/1d2b0bef7b20255957fa3a7ef25a3aaf2096f710))
* **backend:** DB-backed notification delivery tracking with strict thread routing ([d991807](https://github.com/kube-rca/kuberca/commit/d991807dcc4fd0e776ad7476049fa94079003783))
* **backend:** register alert resolve routes and update OpenAPI docs ([eb60835](https://github.com/kube-rca/kuberca/commit/eb6083573a0fd544044cf521cbe0f60f75e4cc76))
* Comment 수정 및 삭제 기능 (KAN-151) ([287a7d6](https://github.com/kube-rca/kuberca/commit/287a7d681544eb24b4eefcb227c21f738566d374))
* Comment 편집 및 삭제 추가 (KAN-153) ([33362a3](https://github.com/kube-rca/kuberca/commit/33362a3320655de47149885d68711f2e16b86c1e))
* **db:** add connection retry with exponential backoff and DB-aware health endpoints ([c6eb1a3](https://github.com/kube-rca/kuberca/commit/c6eb1a3b05cb3b7f1e2b1b9583a0dc9b150b2ea3))
* display firing/resolved analyses in stacked layout ([04a050a](https://github.com/kube-rca/kuberca/commit/04a050ac543c194ae063d521c0050d2470b7daba))
* display OIDC login_id as 'Name (email)' format ([25a47d9](https://github.com/kube-rca/kuberca/commit/25a47d9f793424817bcc9d55ed2fd2d8e2efbe36))
* export 기능 추가 (KAN-169) ([5e16726](https://github.com/kube-rca/kuberca/commit/5e16726e88e54a5c73cd7937af5878933aa8c0c8))
* flapping enable/disable (KAN-155) ([e2c16e4](https://github.com/kube-rca/kuberca/commit/e2c16e476251c0e903c661f1b7fd76b04e706991))
* flapping enable/disable (KAN-155) ([b80c570](https://github.com/kube-rca/kuberca/commit/b80c570cd32da23776a2e73d0c1f2f1d2f9ee3c1))
* **frontend:** add bulk resolve UI with checkboxes to AlertTable ([96a5d22](https://github.com/kube-rca/kuberca/commit/96a5d222285a5266e5d045f77e6ddd9c37edbce1))
* **frontend:** add manual Resolve button to AlertDetailView ([a73d70e](https://github.com/kube-rca/kuberca/commit/a73d70ecf46a0b7c7c8e8f8229e0e62068587084))
* **frontend:** add resolveAlert and bulkResolveAlerts API functions ([40daa6d](https://github.com/kube-rca/kuberca/commit/40daa6d282a151bdfa96666e8224b80127e27aab))
* **helm:** add inline apiKey value fields to values.yaml ([07ee269](https://github.com/kube-rca/kuberca/commit/07ee269972bd5529b367615af3dd5d85805dbbb1))
* **helm:** add postgresql and embedding secret name auto-resolve helpers ([424c2ac](https://github.com/kube-rca/kuberca/commit/424c2ac20651334610d127768bf78fe1134dc591))
* **helm:** populate agent secrets with inline apiKey values ([277c1b2](https://github.com/kube-rca/kuberca/commit/277c1b273d6f1bc093ed1d40b2a924fea9287e96))
* **helm:** use auto-resolve helper for agent session DB secret ([21d8e95](https://github.com/kube-rca/kuberca/commit/21d8e958c6ca325f821d1a802b3392826850564a))
* **helm:** use auto-resolve helpers for postgresql and embedding secrets ([ca9cb70](https://github.com/kube-rca/kuberca/commit/ca9cb702edc60b918f2c3dfa0f049d3e102f741d))
* Implement webhook name support across frontend components, API … ([874f8ad](https://github.com/kube-rca/kuberca/commit/874f8ade43a9e0aceb31136cc6effc60d8fb7ca9))
* Implement webhook name support across frontend components, API types, and add a feature plan. ([8d0d1be](https://github.com/kube-rca/kuberca/commit/8d0d1be7a7c1cc6df9aec50f69adb12329713d7c))
* improve analysis UX, fix settings load bug, add logo (KAN-173, KAN-172) ([dbe0ab5](https://github.com/kube-rca/kuberca/commit/dbe0ab5673bd42b107f4fb65a2c230d286a0a157))
* improve incident table readability with larger text and column dividers ([943f8fa](https://github.com/kube-rca/kuberca/commit/943f8fa77a2bb989cd1af1c1488c15667ece2199))
* Incident & Alarm에 Comment 기능 (KAN-153) ([98b5588](https://github.com/kube-rca/kuberca/commit/98b55883c08190e147dbeb15bdf31d264ecacb77))
* Incident & Alarm에 Comment 추가 (KAN-151) ([c72a110](https://github.com/kube-rca/kuberca/commit/c72a1104156f6c6fda033c41432b78068d6baf95))
* **incidents:** expose is_analyzing field in incident detail API ([50e9085](https://github.com/kube-rca/kuberca/commit/50e90856451e44433b3c785f4fe4b39095a63339))
* **kube-rca:** add agent masking regex helm settings (KAN-146) ([#8](https://github.com/kube-rca/kuberca/issues/8)) ([9a009a3](https://github.com/kube-rca/kuberca/commit/9a009a35fd47f106c7f9c664509d727a0f2d9ec4))
* **kube-rca:** add LLM retry configuration for agent (KAN-159) ([5653ef2](https://github.com/kube-rca/kuberca/commit/5653ef2da593bb5a3a3173eb53bb55e8ceb50dc6))
* **kube-rca:** add Loki env vars for agent deployment ([cfe9361](https://github.com/kube-rca/kuberca/commit/cfe9361147a70da379620935ae963bf5f92f7f91))
* **kube-rca:** add MANUAL_ANALYZE_SEVERITIES env for analysis mode (KAN-172) ([e8ff427](https://github.com/kube-rca/kuberca/commit/e8ff427762886d6e92d0961455b83206788eaa15))
* **kube-rca:** add OIDC authentication helm values and deployment env ([40311b4](https://github.com/kube-rca/kuberca/commit/40311b458c990acba6da54aafba0275adaec5389))
* **kube-rca:** replace hook job with initContainer and add DB-aware probes ([d5d191f](https://github.com/kube-rca/kuberca/commit/d5d191f5c473c2fd9d04198a12b45fea5fe32936))
* new webhook routing navagation ([26796b7](https://github.com/kube-rca/kuberca/commit/26796b70689dce87f8c746768516707c068f7887))
* **observability:** add tempo stack and wire agent tempo settings (KAN-73) ([#7](https://github.com/kube-rca/kuberca/issues/7)) ([c540029](https://github.com/kube-rca/kuberca/commit/c540029b10ff336e6d66588db974cf492609bc40))
* redesign UI to Industrial Ops style ([483f724](https://github.com/kube-rca/kuberca/commit/483f72467a4cced25c3a19f4aea5c4cd478e3eeb))
* remove alert owner hint message ([71f9cfc](https://github.com/kube-rca/kuberca/commit/71f9cfccd6da0218b37c394a0bd52d8250d2b67c))
* remove duplicated setting ([2a077ea](https://github.com/kube-rca/kuberca/commit/2a077ea9f87fb0a386f7ff215b772ffa9bd7a2f2))
* remove duplicated setting ([e685d50](https://github.com/kube-rca/kuberca/commit/e685d500ca2536d5a7f9ba12e3277eb4e05a784a))
* render provider-specific OIDC login button ([2341ccb](https://github.com/kube-rca/kuberca/commit/2341ccb0549a3dccfe9496eac3676330d50290b7))
* replace 1s polling with SSE + 30s fallback polling (KAN-165) ([26dc8de](https://github.com/kube-rca/kuberca/commit/26dc8de7ea9f8bc71fd7ad15fae3d0e98cee7ce4))
* **scenario:** add ratings-multi fault scenario ([94b6fe7](https://github.com/kube-rca/kuberca/commit/94b6fe72b203b74bad77cfe5c24dc2410fd6ac96))
* shrink Time column and expand Title column in tables ([b31a274](https://github.com/kube-rca/kuberca/commit/b31a27430a559aa577cb4652e1753110875ca926))
* support slack bot token and channel webhook settings ([8f3a7e4](https://github.com/kube-rca/kuberca/commit/8f3a7e47bc3042c99e1923581f234fe10a3560f7))
* support slack bot token and channel webhook settings ([f86f0b7](https://github.com/kube-rca/kuberca/commit/f86f0b72c0fe056cde6221dc328b30314e834653))
* sync UI/Helm setting ([4d480d1](https://github.com/kube-rca/kuberca/commit/4d480d12daba27bc7481cc7d2e3f6ea2e930329a))
* unify table styles across Alert, Incident, and Archive dashboards ([c04d040](https://github.com/kube-rca/kuberca/commit/c04d040a39c8c0c3b7caf456c66c603805b0c152))
* 검색 기능 추가 (KAN-143) ([ba20807](https://github.com/kube-rca/kuberca/commit/ba20807e825fad8dbcaf72c231e2ad4207278c3f))
* 라벨 기반 검색을 위해 목록 반환시 라벨 추가(KAN-143) ([26e4c69](https://github.com/kube-rca/kuberca/commit/26e4c69447ed79e690be1369db9089999e6902c3))
* 채팅 기능 (KAN-157) ([c5b9bda](https://github.com/kube-rca/kuberca/commit/c5b9bdab9f5029e5de656c6a1da1045a9da9a297))
* 채팅 기능 (KAN-86) ([46b40b9](https://github.com/kube-rca/kuberca/commit/46b40b9b922f03a25ceffb0697afb33a8e9e121c))
* 채팅 기능 추가 (KAN-158) ([0efd63a](https://github.com/kube-rca/kuberca/commit/0efd63aab4a5cf7ff09034506cd2bbf2e42a3b69))


### Bug Fixes

* add prompt=select_account to OIDC auth URL ([80a99b1](https://github.com/kube-rca/kuberca/commit/80a99b10eea13a58be81cd0bdf1347108568f930))
* **agent:** add httpx transport errors to LLM retry and sanitize messages ([2e230c5](https://github.com/kube-rca/kuberca/commit/2e230c57eaf5a0d98d7c224f07e7272d66d5fba0))
* **agent:** align tempo timeout default and docs (KAN-145) ([3712dec](https://github.com/kube-rca/kuberca/commit/3712dece3c7f00ca976310a375645c1b84563b6b))
* **agent:** bypass kubernetes client event_time validation via raw JSON ([7bc93e2](https://github.com/kube-rca/kuberca/commit/7bc93e29fdea4852fdce5e3918be9600953cddc4))
* **agent:** context-aware missing data and stronger tool usage directive ([e273cd5](https://github.com/kube-rca/kuberca/commit/e273cd5a88f8d01247d19e984d79db3f90a26c15))
* **agent:** disable default PrintingCallbackHandler in Strands Agent ([daa1394](https://github.com/kube-rca/kuberca/commit/daa1394f989e70862ae0a8a35e1cc8798990dd8a))
* **agent:** enforce completed-findings tone and active tool usage in prompt ([6d32be2](https://github.com/kube-rca/kuberca/commit/6d32be2482b65b2c5120bb054c2e055385f70a30))
* **agent:** extract detail section only and strip markdown bold from title ([5bce3b6](https://github.com/kube-rca/kuberca/commit/5bce3b660819fe9cbc7c87a2ed0b028ef007291e))
* **agent:** handle Gemini turn-order violations from sliding-window truncation ([e18fc61](https://github.com/kube-rca/kuberca/commit/e18fc61957c54d5e06a5f3611a582af4f97b1666))
* **agent:** handle PostgreSQL race condition in session schema initialization ([8699050](https://github.com/kube-rca/kuberca/commit/86990504dd5310e3830648b2939f94f0710a88c6))
* **agent:** improve analysis fallback for missing pod labels and empty exceptions ([f27d4ad](https://github.com/kube-rca/kuberca/commit/f27d4adc1bb1657ca21a01dcb1651cb383f18139))
* **agent:** improve analysis reliability and resource management ([29ca1e4](https://github.com/kube-rca/kuberca/commit/29ca1e47ebffb71c739e812cabf1d473d4fadb45))
* **agent:** improve incident summary parsing for bold-without-colon LLM responses ([b83aa4f](https://github.com/kube-rca/kuberca/commit/b83aa4f6ee791f97f6dc1dd16bcdc1819c2bf045))
* **agent:** prevent entire LLM response from leaking into incident title and summary ([f9848c4](https://github.com/kube-rca/kuberca/commit/f9848c4f0bb4c4831daf56a188066dad61a43209))
* **agent:** remove incident summary length limit to prevent truncation ([f3176ea](https://github.com/kube-rca/kuberca/commit/f3176ea16f29c8a911c9c136d877a16cf483eff5))
* **agent:** set anthropic max token defaults ([4e96a6a](https://github.com/kube-rca/kuberca/commit/4e96a6aab87f732c7be5894deb6e03ee2ef0347a))
* **agent:** upgrade strands-agents 1.23.0 → 1.30.0, remove thought signature monkey-patch ([498632e](https://github.com/kube-rca/kuberca/commit/498632e58227ddf6eb2ea5fe51871e69d6592c27))
* **agent:** use FQDN service.name for Tempo TraceQL queries ([adf4b7d](https://github.com/kube-rca/kuberca/commit/adf4b7d7978bd9153b5bae4f7d3d9536d0fc1d2a))
* **alerts:** resolve fingerprint dedup causing alert history loss ([#25](https://github.com/kube-rca/kuberca/issues/25)) ([9c5d340](https://github.com/kube-rca/kuberca/commit/9c5d340e9edcad69a6242c94804df5ef15f1a84a))
* analysis-channel mapping rule ([64023b2](https://github.com/kube-rca/kuberca/commit/64023b211f8a2b969ee02ac5af1b9efc707ff01e))
* **analysis:** remove hardcoded summary length limits ([644bd4d](https://github.com/kube-rca/kuberca/commit/644bd4d9e9f89192e6e4c35ab19cb4986adddb15))
* **analyze:** always show Re-Analyze button, add incident analyze, remove Analyze All ([9c7b876](https://github.com/kube-rca/kuberca/commit/9c7b8764d74905127371c4758ed61750bb299981))
* **analyze:** correct fingerprint lookup and add incident-level analysis ([5c27602](https://github.com/kube-rca/kuberca/commit/5c276023e11ed8291b7e80dcc970f6d7794dc676))
* **analyze:** improve polling logic with grace period and seenAnalyzing tracking ([fd3c520](https://github.com/kube-rca/kuberca/commit/fd3c52066f493910765d5a316ca35e8a62da5f82))
* **backend:** add retry with exponential backoff for agent analysis requests ([427b3e5](https://github.com/kube-rca/kuberca/commit/427b3e565e344a7783ee97b6d58342ab22c490cf))
* **backend:** allow resolved analysis to wait for in-flight firing analysis ([850c89f](https://github.com/kube-rca/kuberca/commit/850c89f2c0afda00f45b578c4e8434886894d8da))
* **backend:** enable markdown rendering for Slack attachments ([f5bd92b](https://github.com/kube-rca/kuberca/commit/f5bd92b9ce44c0ee56c4855c18c50e434661d09d))
* **backend:** handle triple asterisk in Slack markdown conversion ([17e0127](https://github.com/kube-rca/kuberca/commit/17e0127311dca638a5831daaeb9f32f20d419c9a))
* **backend:** preserve thread_ts on resolved alerts for async analysis routing ([562c514](https://github.com/kube-rca/kuberca/commit/562c514a10ebf634679d480f668544834785171b))
* **backend:** set StartsAt and prefer direct ThreadTS in manual resolve ([4d0e8d1](https://github.com/kube-rca/kuberca/commit/4d0e8d1aaa2c6181ad8cc55c8cedc4f1b68346cf))
* **backend:** strip markdown bold markers inside heading for Slack conversion ([41d0bc7](https://github.com/kube-rca/kuberca/commit/41d0bc7006658e1111fd37622443dc133d0cd03d))
* **backend:** upgrade go-jose/v4 to 4.1.4 for JWE panic fix (GHSA high) ([76435e5](https://github.com/kube-rca/kuberca/commit/76435e5daa8e47366b2b06cf69fd96640cee5efe))
* **backend:** use semantic JSON comparison for SyncEnvDefaults ([c801ab9](https://github.com/kube-rca/kuberca/commit/c801ab958a372bbc843634d68e23b874557247a2))
* build error ([e82a2a6](https://github.com/kube-rca/kuberca/commit/e82a2a6d400d05133a2f03cf113814af075d0978))
* bump cryptography to 46.0.5 for CVE-2026-26007 ([147e415](https://github.com/kube-rca/kuberca/commit/147e4150d62077b8c015bcf99cd505b3036cd9c5))
* **chart:** add agent startupProbe and sync eks values ([b2d5a41](https://github.com/kube-rca/kuberca/commit/b2d5a41d1c5c00bd7952eb6856abd4ebe909f355))
* **chart:** add anthropic max token env ([77310a8](https://github.com/kube-rca/kuberca/commit/77310a8217c56038ebd056df55fb8bf773d9cccd))
* **ci:** add id-token permission to release-please for reusable workflow calls ([e06b000](https://github.com/kube-rca/kuberca/commit/e06b0006dcc0aa3459ebf69c6df624ed7b0c25b1))
* **ci:** add retry logic for concurrent helm values push ([e062669](https://github.com/kube-rca/kuberca/commit/e06266971039521c557edeb6249cfd266fd232a1))
* Comment 메뉴바 표시와 일부 단축어 오류 수정 (KAN-153) ([0cdc6d9](https://github.com/kube-rca/kuberca/commit/0cdc6d9a8053a2ab22dc77b9c5f67cf34d47818a))
* Comment 편집 및 삭제 오류 수정 (KAN-153) ([af06234](https://github.com/kube-rca/kuberca/commit/af06234be9a3f6310eafd33ffd6678bbd5de76b3))
* daymode 글자안보이던 현상 해결 (KAN-96) ([a820262](https://github.com/kube-rca/kuberca/commit/a820262e59a34e983e18a47bab66779e4786c9c6))
* **deps:** upgrade cryptography 46.0.6, requests 2.33.0 ([90100a4](https://github.com/kube-rca/kuberca/commit/90100a4f5415f105744316f63b2483da0ac3e36d))
* **deps:** upgrade flatted to 3.4.2 ([784b3c0](https://github.com/kube-rca/kuberca/commit/784b3c0680cb8537cbfb19b8d2244bd8fe88533d))
* **deps:** upgrade google.golang.org/grpc to v1.79.3 ([0bf5c17](https://github.com/kube-rca/kuberca/commit/0bf5c177b517fcb4cfe33c322c89510d7f34b76f))
* **deps:** upgrade picomatch to fix CVE-2026-33672 (GHSA-3v7f-55p6-f55p) ([5d82afe](https://github.com/kube-rca/kuberca/commit/5d82afebda93425f5e7aee1441d972ba1fee8efb))
* **deps:** upgrade pyasn1 to 0.6.3 ([60798e9](https://github.com/kube-rca/kuberca/commit/60798e998277743b40f81f1eeedb8d2be3066fb9))
* **deps:** upgrade Pygments 2.19.2 to 2.20.0 ([9029d79](https://github.com/kube-rca/kuberca/commit/9029d792c5cfa92c2dea593c702540ebf718a152))
* **deps:** upgrade PyJWT 2.10.1 → 2.12.1 (CVE-2026-32597) ([940d05b](https://github.com/kube-rca/kuberca/commit/940d05bcafc567e4781ee40d4ad1c1221f01b6ef))
* edit thread_ts analysis mapping rule ([141bdd3](https://github.com/kube-rca/kuberca/commit/141bdd3b90cdc709bdcf9e2842c7161ad0bbf81c))
* **frontend:** add line-clamp and tooltip to title columns in all tables ([ee55b5c](https://github.com/kube-rca/kuberca/commit/ee55b5cf5615efde8dd9ad0ef7db6d69826ce812))
* **frontend:** hide checkbox column when no firing alerts exist ([8318f62](https://github.com/kube-rca/kuberca/commit/8318f623ed7e70b43ee9bf69c9b2e1fb306e0c5f))
* **frontend:** remove duplicate severity badge in webhook routing ([d6da244](https://github.com/kube-rca/kuberca/commit/d6da244c27d5acbe39dbc2d4c6a0ead8a7d0eb45))
* Gemini 400 function-call turn order error in multi-turn chat ([32e1595](https://github.com/kube-rca/kuberca/commit/32e159520941e54bd876aa6f0bc55badb1ad47d1))
* **helm:** harden tempo defaults for timeout and load (KAN-145) ([c1d6f3b](https://github.com/kube-rca/kuberca/commit/c1d6f3b3862e1fcd85833c20008e59f0e57de482))
* **i18n:** localize remaining Korean alert message to English ([30ec6d6](https://github.com/kube-rca/kuberca/commit/30ec6d690e4e719c79021bee5bfdac5c8ede4196))
* import unimported modules ([5e10f3a](https://github.com/kube-rca/kuberca/commit/5e10f3a81e24ef106a9652f95c44f4a3ae2e9cc9))
* import unimported modules ([b0b23de](https://github.com/kube-rca/kuberca/commit/b0b23de07567b3732faad4f480e45bdafd62d338))
* improve OIDC error display with prominent banner and URL cleanup ([54855b4](https://github.com/kube-rca/kuberca/commit/54855b406ec9c2ebcae83b93c4be0b9c42e6349c))
* **incidents:** prevent duplicate ongoing incidents with atomic upsert ([1fb7590](https://github.com/kube-rca/kuberca/commit/1fb759095a9ed7482eda2deac0da7f73c66cb1b2))
* **kube-rca:** harden chart security, add probes, fix secret bug ([e15824c](https://github.com/kube-rca/kuberca/commit/e15824c2f853ca083c288f8d80ec1ac5d1275b45))
* **kube-rca:** increase postgresql CPU limits to resolve throttling ([d530d91](https://github.com/kube-rca/kuberca/commit/d530d9133c7b9d273a3290a39a6291ca704113aa))
* **kube-rca:** remove container securityContext defaults ([9884eda](https://github.com/kube-rca/kuberca/commit/9884edaf777691b2837978268d36ad9176682887))
* **kube-rca:** remove runAsNonRoot from default podSecurityContext ([86b3ef6](https://github.com/kube-rca/kuberca/commit/86b3ef64c282303928baa2df984d2c7718007233))
* lint check ([57d95d9](https://github.com/kube-rca/kuberca/commit/57d95d967858a75e27ad35ec09917a6d2c8297e1))
* normalize webhook config mapping for list and edit ([f1d79d6](https://github.com/kube-rca/kuberca/commit/f1d79d6fe189ef21617a101a08c58cf627d49b4f))
* normalize webhook config mapping for list and edit ([2875698](https://github.com/kube-rca/kuberca/commit/287569869c76f6b6ecfb85575b0caffe41b7a8e3))
* prevent ID column text wrapping in all tables ([d7de3f7](https://github.com/kube-rca/kuberca/commit/d7de3f7527d520082b6ff275cb951649fba3edbc))
* prevent URL sync from clearing OIDC error params before AuthPanel reads them ([5d89ef1](https://github.com/kube-rca/kuberca/commit/5d89ef122595d003c76ed3f0462cc62415fe7d7e))
* **prompt:** enforce bold markdown headers for analysis subsections ([4d8b9a1](https://github.com/kube-rca/kuberca/commit/4d8b9a1b80a8840df0521aa8aa713a39ad33bbaa))
* publish chart and image paths under kube-rca-ecr ([31130cf](https://github.com/kube-rca/kuberca/commit/31130cf94c72749982cbdc8641c44a779ddfb2ea))
* recover chat session state and standardize english chat errors ([6e9c0fd](https://github.com/kube-rca/kuberca/commit/6e9c0fd6f1944eb0b0ea43f28dae02b316705e80))
* replace login screen icon with actual logo and fix favicon ([1ba14c9](https://github.com/kube-rca/kuberca/commit/1ba14c9ea525b8c7007278de56ed712b73f90d4b))
* resolve all npm audit vulnerabilities ([6e93ee2](https://github.com/kube-rca/kuberca/commit/6e93ee2de8152fdf3e5b9e0f3c09e1c1cd9cb23b))
* robustly map slack webhook data in list and edit ([2f42417](https://github.com/kube-rca/kuberca/commit/2f4241795eca68c0f8e050c5321750f1b9c9cb36))
* robustly map slack webhook data in list and edit ([cac6f3a](https://github.com/kube-rca/kuberca/commit/cac6f3ad6ada99fcce460b9f0ec4750904730914))
* route agent image pushes to component repository ([4600466](https://github.com/kube-rca/kuberca/commit/4600466ed52d151f615abfa67d19e6233ed64eca))
* route backend image pushes to component repository ([aed51c5](https://github.com/kube-rca/kuberca/commit/aed51c5cc068310214b34c716d35845e77ee4b92))
* route frontend image pushes to component repository ([294f43a](https://github.com/kube-rca/kuberca/commit/294f43a0eda89f2cfba47ff7da3f31081d79953e))
* **scenario:** prevent signal re-entrance during cleanup ([e327ccc](https://github.com/kube-rca/kuberca/commit/e327cccddc6e1295a5681a76ddf0e56b8adce35b))
* **scenario:** use original manifest for cleanup instead of temp file ([8070895](https://github.com/kube-rca/kuberca/commit/80708957e266b995512377ba834d69e4efad001a))
* skip URL param sync when not authenticated. ([5d89ef1](https://github.com/kube-rca/kuberca/commit/5d89ef122595d003c76ed3f0462cc62415fe7d7e))
* **tempo:** normalize time window and surface query failures ([e2cdbb3](https://github.com/kube-rca/kuberca/commit/e2cdbb3503a4937a56f624804dd2236be854caef))
* **ui:** resolve dashboard toolbar overflow and table responsive issues ([5cf556c](https://github.com/kube-rca/kuberca/commit/5cf556cd5ee45a8140b6a6a2195b6d125f8491a6))
* use email as login_id for OIDC users instead of oidc_{sub} ([7878524](https://github.com/kube-rca/kuberca/commit/7878524809220e99a2edb681d9874eeebbf2d39c))
* 자동 메세지 전송처리 삭제 ([511125a](https://github.com/kube-rca/kuberca/commit/511125a6d294080a952fd955881842e237e99150))
* 채팅 기능 404 에러 (KAN-157) ([0a87c78](https://github.com/kube-rca/kuberca/commit/0a87c783133e87994115dace7108ebbbc68ef2d5))


### Performance Improvements

* **agent:** add analysis pipeline timing instrumentation ([74ad719](https://github.com/kube-rca/kuberca/commit/74ad7193705bc415190c14edf8f8d3a08999c445))
* **kube-rca:** increase agent replicaCount default to 2 ([b2c6994](https://github.com/kube-rca/kuberca/commit/b2c69946dc903794bcfa55df31cc607114052a4e))

## 1.0.0 (2026-04-05)


### Features

* Add a required `name` field to webhook configurations, includin… ([52c767c](https://github.com/kube-rca/kuberca/commit/52c767cf8432b0c1a80b547159eed0093bb36b95))
* Add a required `name` field to webhook configurations, including database schema, API requests, and validation logic. ([951eff1](https://github.com/kube-rca/kuberca/commit/951eff11a5f9841aee05c24cba444b5ce2f9202f))
* Add alert variables and refactor webhook body variable insertion with a reusable component and improved UI. ([dd60fd9](https://github.com/kube-rca/kuberca/commit/dd60fd95b705a161dc0208e9796b72fcc4de9150))
* Add alert variables and refactor webhook body variables ([f10417f](https://github.com/kube-rca/kuberca/commit/f10417fb78f4df7ca291c61bdba0ba88edd1a093))
* Add alert variables and refactor webhook body variables ([f10417f](https://github.com/kube-rca/kuberca/commit/f10417fb78f4df7ca291c61bdba0ba88edd1a093))
* add analysis mode indicator to sidebar (KAN-172) ([2e60ff0](https://github.com/kube-rca/kuberca/commit/2e60ff025953593c9a2fa31a5d17d0bd65a7ac0d))
* add analysis mode UI with manual analyze buttons (KAN-172) ([f896b86](https://github.com/kube-rca/kuberca/commit/f896b8629cb8d3292f53e95bf853fcc54717dcba))
* add auto/manual analysis mode with severity-based control (KAN-172) ([053ed52](https://github.com/kube-rca/kuberca/commit/053ed529d77bc00ac08e1e2f5537819f6101743f))
* add bg color to reserved field ([3de4862](https://github.com/kube-rca/kuberca/commit/3de48622a7dd07e17b08fb2ba5b5da5d06a22784))
* add bg color to reserved field ([1874aec](https://github.com/kube-rca/kuberca/commit/1874aec6ed84c3c4d76f0ffb348d7c514638c24b))
* add built-in redaction masking-chaining [KAN-174] ([ba19e50](https://github.com/kube-rca/kuberca/commit/ba19e50578c9e29b4f359f2b30651cc2325ebf6d))
* add built-in Redaction option [KAN-174] ([b1dfa5f](https://github.com/kube-rca/kuberca/commit/b1dfa5fdc1f60b5f9fdc05e1c0efe2de2084ce31))
* add cert manager and move ingress hosts to 2 domains ([ed8959c](https://github.com/kube-rca/kuberca/commit/ed8959cda3192fd944c2cd2ab60715ae8d432455))
* add firing/resolved analysis split to alert detail API ([504a745](https://github.com/kube-rca/kuberca/commit/504a7459bbdd655fbb9cf14d5d62213f844621d9))
* add gp2 values for stateful apps ([f34c8a5](https://github.com/kube-rca/kuberca/commit/f34c8a5175aea0c7a119c5d1154d2dd4f9acad43))
* add info severity to AnalysisSettings and RCADetailView ([24f3935](https://github.com/kube-rca/kuberca/commit/24f39355c88ec1791fa2d39a70427b875bad7f36))
* add monorepo GitHub Actions workflows (CI + release-please) ([0911d33](https://github.com/kube-rca/kuberca/commit/0911d3329aacaac40baa99270db55d6735acedbc))
* add multi-provider OIDC support with auto-detection ([d330906](https://github.com/kube-rca/kuberca/commit/d33090677a923fffe49561a78847b75b69ba12fc))
* add notifier events and webhook-based routing ([5d9164a](https://github.com/kube-rca/kuberca/commit/5d9164a82b7042738967f0a8941afd5af7e39534))
* add notifier events and webhook-based routing ([5d9164a](https://github.com/kube-rca/kuberca/commit/5d9164a82b7042738967f0a8941afd5af7e39534))
* add notifier events and webhook-based routing ([7d14423](https://github.com/kube-rca/kuberca/commit/7d14423de38dbbbf53b1cd17b4662147a099f01a))
* add OIDC authentication with security hardening ([ed19018](https://github.com/kube-rca/kuberca/commit/ed190180fdba1493b9c084502e97851a62a67f67))
* add OIDC login button and callback error handling ([5b375ac](https://github.com/kube-rca/kuberca/commit/5b375acc8985b0726a9daa5dec9670c348af43fe))
* add settings option ([adbb4c2](https://github.com/kube-rca/kuberca/commit/adbb4c27a3c402da32b7dd44c0b2ab50f7563147))
* add settings option ([a03a65c](https://github.com/kube-rca/kuberca/commit/a03a65c86e13a51a48df0b00aac83ba1653ffe5b))
* add settings option ([dfc4bca](https://github.com/kube-rca/kuberca/commit/dfc4bca41fd9f34be51b23843e1714718ab0cf03))
* add settings page ([7871de3](https://github.com/kube-rca/kuberca/commit/7871de33523295a1a22cc05f2b786a9b1448b974))
* add settings page ([5747f98](https://github.com/kube-rca/kuberca/commit/5747f9899e09b1d2fd72a03141008cdd21f09da8))
* add severity-channel mapping rule ([562a9bb](https://github.com/kube-rca/kuberca/commit/562a9bba2e0695cea3ef6df96762fc375e2ce380))
* add severity-channel mapping rule ([9b39f93](https://github.com/kube-rca/kuberca/commit/9b39f93db2d8801e59001d45266b5e798508640f))
* Add show/hide toggle for token input fields and remove an infor… ([c2357ee](https://github.com/kube-rca/kuberca/commit/c2357ee1819788e67ec67263386f4c3c8333eb54))
* Add show/hide toggle for token input fields and remove an informational section in webhook settings. ([44c593f](https://github.com/kube-rca/kuberca/commit/44c593f4f662d08c94706f5a6178636dcd3f423d))
* add SSE support for real-time event notifications (KAN-165) ([a80928e](https://github.com/kube-rca/kuberca/commit/a80928e424670441745eaeb64ad4da21e2a845ce))
* add tenacity-based LLM API retry for transient errors (KAN-159) ([d28cf83](https://github.com/kube-rca/kuberca/commit/d28cf8312c0655db37586c91d5d5cd54ec6cdbfe))
* add total timeout stop condition to LLM retry logic ([99e341f](https://github.com/kube-rca/kuberca/commit/99e341f2ed55c89dfba2fd86dcdf52914f4efda6))
* add webhook config api ([6564238](https://github.com/kube-rca/kuberca/commit/65642388bcc40e85670b0a04e493ce35ebdde039))
* add webhook config api ([6564238](https://github.com/kube-rca/kuberca/commit/65642388bcc40e85670b0a04e493ce35ebdde039))
* add webhook config api ([ad05fbb](https://github.com/kube-rca/kuberca/commit/ad05fbba446a27d5a7bced618f5e79702ce138b5))
* add webhook list page ([7032e61](https://github.com/kube-rca/kuberca/commit/7032e61283e63176dc5d899512362a974476047a))
* add webhook list page ([da59fe4](https://github.com/kube-rca/kuberca/commit/da59fe4df14e4261a235bc3c100b08679e480354))
* **agent:** add generic manifest read tools ([4d4da60](https://github.com/kube-rca/kuberca/commit/4d4da60120f4368ecef24a213a3f55b9ece9c6d9))
* **agent:** add Loki log aggregation client and Strands tools ([6a30f9e](https://github.com/kube-rca/kuberca/commit/6a30f9ee9e2e3786fa16ef2f1f3f9fcb729016f7))
* **agent:** add node-level alert analysis support ([9a22db2](https://github.com/kube-rca/kuberca/commit/9a22db2179b10d031455abec6c827cac1fc90b5b))
* **agent:** add regex masking for llm and storage flows (KAN-146) ([#10](https://github.com/kube-rca/kuberca/issues/10)) ([ae71166](https://github.com/kube-rca/kuberca/commit/ae71166b651fb9074235b9cbc833589810f3b9f0))
* **agent:** add tool timing logs for analysis ([d7c8f39](https://github.com/kube-rca/kuberca/commit/d7c8f3965ccc29319053ecb4aaa457f9f9fe579d))
* **agent:** async 엔드포인트 전환으로 분석 병렬 처리 지원 ([fa74bf1](https://github.com/kube-rca/kuberca/commit/fa74bf1ac520b1d6c4477e78ea5a7e70b4c0cb34))
* **agent:** enrich daemonset and events context (KAN-79) ([bc6ae22](https://github.com/kube-rca/kuberca/commit/bc6ae22ae86ac03e6209a6a4a11e3ad1332137c5))
* **agent:** enrich daemonset and events context (KAN-79) ([f8b7c62](https://github.com/kube-rca/kuberca/commit/f8b7c62a9dde6efe9002f4158ccc98931140cdb6))
* **agent:** integrate tempo trace context into analysis (KAN-145) ([#9](https://github.com/kube-rca/kuberca/issues/9)) ([db45189](https://github.com/kube-rca/kuberca/commit/db45189bf50d4dc97745c7056ee3d568e2136333))
* **agent:** portable K8s baseline context and capability-aware prompt ([f96e4c4](https://github.com/kube-rca/kuberca/commit/f96e4c4dc2c81779d8042df39f15c5d3db431861))
* **agent:** tune LLM retry to 3-min exponential backoff ([d0c9b59](https://github.com/kube-rca/kuberca/commit/d0c9b59a1459c0f9478b93745fb3b0e25aec4d80))
* Alanlysis Dashboard 추가 (KAN-171) ([9097016](https://github.com/kube-rca/kuberca/commit/909701636b91e64897027a941d16446a1fb838d2))
* Alanlysis 기능 추가 (KAN-170) ([bebecbe](https://github.com/kube-rca/kuberca/commit/bebecbec1aa21d3b1c20bf49319b7fd1aa04c39e))
* alert flapping (KAN-155) ([40be145](https://github.com/kube-rca/kuberca/commit/40be1452552a6eb9f0942eb57e4593e8e4673f1b))
* alert flapping 환경변수 (KAN-155) ([e595663](https://github.com/kube-rca/kuberca/commit/e595663fef695c4d41f76f0583278b2a1d8fc451))
* allow info severity payload ([6f71009](https://github.com/kube-rca/kuberca/commit/6f710098004fd589713f1c89f67cb3a10e100ee5))
* analysis export 및 꺾은선 value 추가 ([70e3eb1](https://github.com/kube-rca/kuberca/commit/70e3eb117ef7c5696d8fe5e2586f60aa5442b653))
* **analyze:** add analysis SSE events and is_analyzing API field (KAN-173) ([0d58547](https://github.com/kube-rca/kuberca/commit/0d5854751ce0cee6f89e74c54d92925b99d71c82))
* **analyze:** add in-flight dedup guard and configurable agent timeout ([4c02ea4](https://github.com/kube-rca/kuberca/commit/4c02ea4cad1b2e4f057d711c95f47b8dec9a2a43))
* **analyze:** add incident analysis polling and remove banner auto-dismiss ([8654004](https://github.com/kube-rca/kuberca/commit/8654004d997b41a52935d1425fbe102a14638705))
* **analyze:** differentiate resolved alert analysis from firing ([f2c73a8](https://github.com/kube-rca/kuberca/commit/f2c73a81570eb57a1cd757a1a605083b95ae4ce3))
* **analyze:** differentiate resolved alert analysis from firing ([e9fadb7](https://github.com/kube-rca/kuberca/commit/e9fadb735946d6176a561c5f9b63133bc0ec668e))
* apply same readability improvements to AlertTable ([96a9915](https://github.com/kube-rca/kuberca/commit/96a9915c0a413501abe61c9db488e189ceb88b20))
* **backend:** add IsManual flag to AlertStatusChangedEvent for manual resolve Slack prefix ([a0cdc23](https://github.com/kube-rca/kuberca/commit/a0cdc2338b5c0f64785c98b88d5e103885dd9c5d))
* **backend:** add manual alert resolve request/response models ([19831b5](https://github.com/kube-rca/kuberca/commit/19831b5dd8af5040c6a5915dd035d987f3259019))
* **backend:** add ManualResolveAlert DB function ([eb78a15](https://github.com/kube-rca/kuberca/commit/eb78a15c1a9a721c969bd65cf9b51a7ee6485365))
* **backend:** add ResolveAlert and BulkResolveAlerts handlers with alertService dependency ([940ff56](https://github.com/kube-rca/kuberca/commit/940ff562eaa308fe9d4e318d247f277f387fa59a))
* **backend:** add ResolveAlert and BulkResolveAlerts service methods with tests ([1d2b0be](https://github.com/kube-rca/kuberca/commit/1d2b0bef7b20255957fa3a7ef25a3aaf2096f710))
* **backend:** DB-backed notification delivery tracking with strict thread routing ([d991807](https://github.com/kube-rca/kuberca/commit/d991807dcc4fd0e776ad7476049fa94079003783))
* **backend:** register alert resolve routes and update OpenAPI docs ([eb60835](https://github.com/kube-rca/kuberca/commit/eb6083573a0fd544044cf521cbe0f60f75e4cc76))
* Comment 수정 및 삭제 기능 (KAN-151) ([287a7d6](https://github.com/kube-rca/kuberca/commit/287a7d681544eb24b4eefcb227c21f738566d374))
* Comment 편집 및 삭제 추가 (KAN-153) ([33362a3](https://github.com/kube-rca/kuberca/commit/33362a3320655de47149885d68711f2e16b86c1e))
* **db:** add connection retry with exponential backoff and DB-aware health endpoints ([c6eb1a3](https://github.com/kube-rca/kuberca/commit/c6eb1a3b05cb3b7f1e2b1b9583a0dc9b150b2ea3))
* display firing/resolved analyses in stacked layout ([04a050a](https://github.com/kube-rca/kuberca/commit/04a050ac543c194ae063d521c0050d2470b7daba))
* display OIDC login_id as 'Name (email)' format ([25a47d9](https://github.com/kube-rca/kuberca/commit/25a47d9f793424817bcc9d55ed2fd2d8e2efbe36))
* export 기능 추가 (KAN-169) ([5e16726](https://github.com/kube-rca/kuberca/commit/5e16726e88e54a5c73cd7937af5878933aa8c0c8))
* flapping enable/disable (KAN-155) ([e2c16e4](https://github.com/kube-rca/kuberca/commit/e2c16e476251c0e903c661f1b7fd76b04e706991))
* flapping enable/disable (KAN-155) ([b80c570](https://github.com/kube-rca/kuberca/commit/b80c570cd32da23776a2e73d0c1f2f1d2f9ee3c1))
* **frontend:** add bulk resolve UI with checkboxes to AlertTable ([96a5d22](https://github.com/kube-rca/kuberca/commit/96a5d222285a5266e5d045f77e6ddd9c37edbce1))
* **frontend:** add manual Resolve button to AlertDetailView ([a73d70e](https://github.com/kube-rca/kuberca/commit/a73d70ecf46a0b7c7c8e8f8229e0e62068587084))
* **frontend:** add resolveAlert and bulkResolveAlerts API functions ([40daa6d](https://github.com/kube-rca/kuberca/commit/40daa6d282a151bdfa96666e8224b80127e27aab))
* **helm:** add inline apiKey value fields to values.yaml ([07ee269](https://github.com/kube-rca/kuberca/commit/07ee269972bd5529b367615af3dd5d85805dbbb1))
* **helm:** add postgresql and embedding secret name auto-resolve helpers ([424c2ac](https://github.com/kube-rca/kuberca/commit/424c2ac20651334610d127768bf78fe1134dc591))
* **helm:** populate agent secrets with inline apiKey values ([277c1b2](https://github.com/kube-rca/kuberca/commit/277c1b273d6f1bc093ed1d40b2a924fea9287e96))
* **helm:** use auto-resolve helper for agent session DB secret ([21d8e95](https://github.com/kube-rca/kuberca/commit/21d8e958c6ca325f821d1a802b3392826850564a))
* **helm:** use auto-resolve helpers for postgresql and embedding secrets ([ca9cb70](https://github.com/kube-rca/kuberca/commit/ca9cb702edc60b918f2c3dfa0f049d3e102f741d))
* Implement webhook name support across frontend components, API … ([874f8ad](https://github.com/kube-rca/kuberca/commit/874f8ade43a9e0aceb31136cc6effc60d8fb7ca9))
* Implement webhook name support across frontend components, API types, and add a feature plan. ([8d0d1be](https://github.com/kube-rca/kuberca/commit/8d0d1be7a7c1cc6df9aec50f69adb12329713d7c))
* improve analysis UX, fix settings load bug, add logo (KAN-173, KAN-172) ([dbe0ab5](https://github.com/kube-rca/kuberca/commit/dbe0ab5673bd42b107f4fb65a2c230d286a0a157))
* improve incident table readability with larger text and column dividers ([943f8fa](https://github.com/kube-rca/kuberca/commit/943f8fa77a2bb989cd1af1c1488c15667ece2199))
* Incident & Alarm에 Comment 기능 (KAN-153) ([98b5588](https://github.com/kube-rca/kuberca/commit/98b55883c08190e147dbeb15bdf31d264ecacb77))
* Incident & Alarm에 Comment 추가 (KAN-151) ([c72a110](https://github.com/kube-rca/kuberca/commit/c72a1104156f6c6fda033c41432b78068d6baf95))
* **incidents:** expose is_analyzing field in incident detail API ([50e9085](https://github.com/kube-rca/kuberca/commit/50e90856451e44433b3c785f4fe4b39095a63339))
* **kube-rca:** add agent masking regex helm settings (KAN-146) ([#8](https://github.com/kube-rca/kuberca/issues/8)) ([9a009a3](https://github.com/kube-rca/kuberca/commit/9a009a35fd47f106c7f9c664509d727a0f2d9ec4))
* **kube-rca:** add LLM retry configuration for agent (KAN-159) ([5653ef2](https://github.com/kube-rca/kuberca/commit/5653ef2da593bb5a3a3173eb53bb55e8ceb50dc6))
* **kube-rca:** add Loki env vars for agent deployment ([cfe9361](https://github.com/kube-rca/kuberca/commit/cfe9361147a70da379620935ae963bf5f92f7f91))
* **kube-rca:** add MANUAL_ANALYZE_SEVERITIES env for analysis mode (KAN-172) ([e8ff427](https://github.com/kube-rca/kuberca/commit/e8ff427762886d6e92d0961455b83206788eaa15))
* **kube-rca:** add OIDC authentication helm values and deployment env ([40311b4](https://github.com/kube-rca/kuberca/commit/40311b458c990acba6da54aafba0275adaec5389))
* **kube-rca:** replace hook job with initContainer and add DB-aware probes ([d5d191f](https://github.com/kube-rca/kuberca/commit/d5d191f5c473c2fd9d04198a12b45fea5fe32936))
* new webhook routing navagation ([26796b7](https://github.com/kube-rca/kuberca/commit/26796b70689dce87f8c746768516707c068f7887))
* **observability:** add tempo stack and wire agent tempo settings (KAN-73) ([#7](https://github.com/kube-rca/kuberca/issues/7)) ([c540029](https://github.com/kube-rca/kuberca/commit/c540029b10ff336e6d66588db974cf492609bc40))
* redesign UI to Industrial Ops style ([483f724](https://github.com/kube-rca/kuberca/commit/483f72467a4cced25c3a19f4aea5c4cd478e3eeb))
* remove alert owner hint message ([71f9cfc](https://github.com/kube-rca/kuberca/commit/71f9cfccd6da0218b37c394a0bd52d8250d2b67c))
* remove duplicated setting ([2a077ea](https://github.com/kube-rca/kuberca/commit/2a077ea9f87fb0a386f7ff215b772ffa9bd7a2f2))
* remove duplicated setting ([e685d50](https://github.com/kube-rca/kuberca/commit/e685d500ca2536d5a7f9ba12e3277eb4e05a784a))
* render provider-specific OIDC login button ([2341ccb](https://github.com/kube-rca/kuberca/commit/2341ccb0549a3dccfe9496eac3676330d50290b7))
* replace 1s polling with SSE + 30s fallback polling (KAN-165) ([26dc8de](https://github.com/kube-rca/kuberca/commit/26dc8de7ea9f8bc71fd7ad15fae3d0e98cee7ce4))
* **scenario:** add ratings-multi fault scenario ([94b6fe7](https://github.com/kube-rca/kuberca/commit/94b6fe72b203b74bad77cfe5c24dc2410fd6ac96))
* shrink Time column and expand Title column in tables ([b31a274](https://github.com/kube-rca/kuberca/commit/b31a27430a559aa577cb4652e1753110875ca926))
* support slack bot token and channel webhook settings ([8f3a7e4](https://github.com/kube-rca/kuberca/commit/8f3a7e47bc3042c99e1923581f234fe10a3560f7))
* support slack bot token and channel webhook settings ([f86f0b7](https://github.com/kube-rca/kuberca/commit/f86f0b72c0fe056cde6221dc328b30314e834653))
* sync UI/Helm setting ([4d480d1](https://github.com/kube-rca/kuberca/commit/4d480d12daba27bc7481cc7d2e3f6ea2e930329a))
* unify table styles across Alert, Incident, and Archive dashboards ([c04d040](https://github.com/kube-rca/kuberca/commit/c04d040a39c8c0c3b7caf456c66c603805b0c152))
* 검색 기능 추가 (KAN-143) ([ba20807](https://github.com/kube-rca/kuberca/commit/ba20807e825fad8dbcaf72c231e2ad4207278c3f))
* 라벨 기반 검색을 위해 목록 반환시 라벨 추가(KAN-143) ([26e4c69](https://github.com/kube-rca/kuberca/commit/26e4c69447ed79e690be1369db9089999e6902c3))
* 채팅 기능 (KAN-157) ([c5b9bda](https://github.com/kube-rca/kuberca/commit/c5b9bdab9f5029e5de656c6a1da1045a9da9a297))
* 채팅 기능 (KAN-86) ([46b40b9](https://github.com/kube-rca/kuberca/commit/46b40b9b922f03a25ceffb0697afb33a8e9e121c))
* 채팅 기능 추가 (KAN-158) ([0efd63a](https://github.com/kube-rca/kuberca/commit/0efd63aab4a5cf7ff09034506cd2bbf2e42a3b69))


### Bug Fixes

* add prompt=select_account to OIDC auth URL ([80a99b1](https://github.com/kube-rca/kuberca/commit/80a99b10eea13a58be81cd0bdf1347108568f930))
* **agent:** add httpx transport errors to LLM retry and sanitize messages ([2e230c5](https://github.com/kube-rca/kuberca/commit/2e230c57eaf5a0d98d7c224f07e7272d66d5fba0))
* **agent:** align tempo timeout default and docs (KAN-145) ([3712dec](https://github.com/kube-rca/kuberca/commit/3712dece3c7f00ca976310a375645c1b84563b6b))
* **agent:** bypass kubernetes client event_time validation via raw JSON ([7bc93e2](https://github.com/kube-rca/kuberca/commit/7bc93e29fdea4852fdce5e3918be9600953cddc4))
* **agent:** context-aware missing data and stronger tool usage directive ([e273cd5](https://github.com/kube-rca/kuberca/commit/e273cd5a88f8d01247d19e984d79db3f90a26c15))
* **agent:** disable default PrintingCallbackHandler in Strands Agent ([daa1394](https://github.com/kube-rca/kuberca/commit/daa1394f989e70862ae0a8a35e1cc8798990dd8a))
* **agent:** enforce completed-findings tone and active tool usage in prompt ([6d32be2](https://github.com/kube-rca/kuberca/commit/6d32be2482b65b2c5120bb054c2e055385f70a30))
* **agent:** extract detail section only and strip markdown bold from title ([5bce3b6](https://github.com/kube-rca/kuberca/commit/5bce3b660819fe9cbc7c87a2ed0b028ef007291e))
* **agent:** handle Gemini turn-order violations from sliding-window truncation ([e18fc61](https://github.com/kube-rca/kuberca/commit/e18fc61957c54d5e06a5f3611a582af4f97b1666))
* **agent:** handle PostgreSQL race condition in session schema initialization ([8699050](https://github.com/kube-rca/kuberca/commit/86990504dd5310e3830648b2939f94f0710a88c6))
* **agent:** improve analysis fallback for missing pod labels and empty exceptions ([f27d4ad](https://github.com/kube-rca/kuberca/commit/f27d4adc1bb1657ca21a01dcb1651cb383f18139))
* **agent:** improve analysis reliability and resource management ([29ca1e4](https://github.com/kube-rca/kuberca/commit/29ca1e47ebffb71c739e812cabf1d473d4fadb45))
* **agent:** improve incident summary parsing for bold-without-colon LLM responses ([b83aa4f](https://github.com/kube-rca/kuberca/commit/b83aa4f6ee791f97f6dc1dd16bcdc1819c2bf045))
* **agent:** prevent entire LLM response from leaking into incident title and summary ([f9848c4](https://github.com/kube-rca/kuberca/commit/f9848c4f0bb4c4831daf56a188066dad61a43209))
* **agent:** remove incident summary length limit to prevent truncation ([f3176ea](https://github.com/kube-rca/kuberca/commit/f3176ea16f29c8a911c9c136d877a16cf483eff5))
* **agent:** set anthropic max token defaults ([4e96a6a](https://github.com/kube-rca/kuberca/commit/4e96a6aab87f732c7be5894deb6e03ee2ef0347a))
* **agent:** upgrade strands-agents 1.23.0 → 1.30.0, remove thought signature monkey-patch ([498632e](https://github.com/kube-rca/kuberca/commit/498632e58227ddf6eb2ea5fe51871e69d6592c27))
* **agent:** use FQDN service.name for Tempo TraceQL queries ([adf4b7d](https://github.com/kube-rca/kuberca/commit/adf4b7d7978bd9153b5bae4f7d3d9536d0fc1d2a))
* **alerts:** resolve fingerprint dedup causing alert history loss ([#25](https://github.com/kube-rca/kuberca/issues/25)) ([9c5d340](https://github.com/kube-rca/kuberca/commit/9c5d340e9edcad69a6242c94804df5ef15f1a84a))
* analysis-channel mapping rule ([64023b2](https://github.com/kube-rca/kuberca/commit/64023b211f8a2b969ee02ac5af1b9efc707ff01e))
* **analysis:** remove hardcoded summary length limits ([644bd4d](https://github.com/kube-rca/kuberca/commit/644bd4d9e9f89192e6e4c35ab19cb4986adddb15))
* **analyze:** always show Re-Analyze button, add incident analyze, remove Analyze All ([9c7b876](https://github.com/kube-rca/kuberca/commit/9c7b8764d74905127371c4758ed61750bb299981))
* **analyze:** correct fingerprint lookup and add incident-level analysis ([5c27602](https://github.com/kube-rca/kuberca/commit/5c276023e11ed8291b7e80dcc970f6d7794dc676))
* **analyze:** improve polling logic with grace period and seenAnalyzing tracking ([fd3c520](https://github.com/kube-rca/kuberca/commit/fd3c52066f493910765d5a316ca35e8a62da5f82))
* **backend:** add retry with exponential backoff for agent analysis requests ([427b3e5](https://github.com/kube-rca/kuberca/commit/427b3e565e344a7783ee97b6d58342ab22c490cf))
* **backend:** allow resolved analysis to wait for in-flight firing analysis ([850c89f](https://github.com/kube-rca/kuberca/commit/850c89f2c0afda00f45b578c4e8434886894d8da))
* **backend:** enable markdown rendering for Slack attachments ([f5bd92b](https://github.com/kube-rca/kuberca/commit/f5bd92b9ce44c0ee56c4855c18c50e434661d09d))
* **backend:** handle triple asterisk in Slack markdown conversion ([17e0127](https://github.com/kube-rca/kuberca/commit/17e0127311dca638a5831daaeb9f32f20d419c9a))
* **backend:** preserve thread_ts on resolved alerts for async analysis routing ([562c514](https://github.com/kube-rca/kuberca/commit/562c514a10ebf634679d480f668544834785171b))
* **backend:** set StartsAt and prefer direct ThreadTS in manual resolve ([4d0e8d1](https://github.com/kube-rca/kuberca/commit/4d0e8d1aaa2c6181ad8cc55c8cedc4f1b68346cf))
* **backend:** strip markdown bold markers inside heading for Slack conversion ([41d0bc7](https://github.com/kube-rca/kuberca/commit/41d0bc7006658e1111fd37622443dc133d0cd03d))
* **backend:** upgrade go-jose/v4 to 4.1.4 for JWE panic fix (GHSA high) ([76435e5](https://github.com/kube-rca/kuberca/commit/76435e5daa8e47366b2b06cf69fd96640cee5efe))
* **backend:** use semantic JSON comparison for SyncEnvDefaults ([c801ab9](https://github.com/kube-rca/kuberca/commit/c801ab958a372bbc843634d68e23b874557247a2))
* build error ([e82a2a6](https://github.com/kube-rca/kuberca/commit/e82a2a6d400d05133a2f03cf113814af075d0978))
* bump cryptography to 46.0.5 for CVE-2026-26007 ([147e415](https://github.com/kube-rca/kuberca/commit/147e4150d62077b8c015bcf99cd505b3036cd9c5))
* **chart:** add agent startupProbe and sync eks values ([b2d5a41](https://github.com/kube-rca/kuberca/commit/b2d5a41d1c5c00bd7952eb6856abd4ebe909f355))
* **chart:** add anthropic max token env ([77310a8](https://github.com/kube-rca/kuberca/commit/77310a8217c56038ebd056df55fb8bf773d9cccd))
* **ci:** add id-token permission to release-please for reusable workflow calls ([e06b000](https://github.com/kube-rca/kuberca/commit/e06b0006dcc0aa3459ebf69c6df624ed7b0c25b1))
* **ci:** add retry logic for concurrent helm values push ([e062669](https://github.com/kube-rca/kuberca/commit/e06266971039521c557edeb6249cfd266fd232a1))
* Comment 메뉴바 표시와 일부 단축어 오류 수정 (KAN-153) ([0cdc6d9](https://github.com/kube-rca/kuberca/commit/0cdc6d9a8053a2ab22dc77b9c5f67cf34d47818a))
* Comment 편집 및 삭제 오류 수정 (KAN-153) ([af06234](https://github.com/kube-rca/kuberca/commit/af06234be9a3f6310eafd33ffd6678bbd5de76b3))
* daymode 글자안보이던 현상 해결 (KAN-96) ([a820262](https://github.com/kube-rca/kuberca/commit/a820262e59a34e983e18a47bab66779e4786c9c6))
* **deps:** upgrade cryptography 46.0.6, requests 2.33.0 ([90100a4](https://github.com/kube-rca/kuberca/commit/90100a4f5415f105744316f63b2483da0ac3e36d))
* **deps:** upgrade flatted to 3.4.2 ([784b3c0](https://github.com/kube-rca/kuberca/commit/784b3c0680cb8537cbfb19b8d2244bd8fe88533d))
* **deps:** upgrade google.golang.org/grpc to v1.79.3 ([0bf5c17](https://github.com/kube-rca/kuberca/commit/0bf5c177b517fcb4cfe33c322c89510d7f34b76f))
* **deps:** upgrade picomatch to fix CVE-2026-33672 (GHSA-3v7f-55p6-f55p) ([5d82afe](https://github.com/kube-rca/kuberca/commit/5d82afebda93425f5e7aee1441d972ba1fee8efb))
* **deps:** upgrade pyasn1 to 0.6.3 ([60798e9](https://github.com/kube-rca/kuberca/commit/60798e998277743b40f81f1eeedb8d2be3066fb9))
* **deps:** upgrade Pygments 2.19.2 to 2.20.0 ([9029d79](https://github.com/kube-rca/kuberca/commit/9029d792c5cfa92c2dea593c702540ebf718a152))
* **deps:** upgrade PyJWT 2.10.1 → 2.12.1 (CVE-2026-32597) ([940d05b](https://github.com/kube-rca/kuberca/commit/940d05bcafc567e4781ee40d4ad1c1221f01b6ef))
* edit thread_ts analysis mapping rule ([141bdd3](https://github.com/kube-rca/kuberca/commit/141bdd3b90cdc709bdcf9e2842c7161ad0bbf81c))
* **frontend:** add line-clamp and tooltip to title columns in all tables ([ee55b5c](https://github.com/kube-rca/kuberca/commit/ee55b5cf5615efde8dd9ad0ef7db6d69826ce812))
* **frontend:** hide checkbox column when no firing alerts exist ([8318f62](https://github.com/kube-rca/kuberca/commit/8318f623ed7e70b43ee9bf69c9b2e1fb306e0c5f))
* **frontend:** remove duplicate severity badge in webhook routing ([d6da244](https://github.com/kube-rca/kuberca/commit/d6da244c27d5acbe39dbc2d4c6a0ead8a7d0eb45))
* Gemini 400 function-call turn order error in multi-turn chat ([32e1595](https://github.com/kube-rca/kuberca/commit/32e159520941e54bd876aa6f0bc55badb1ad47d1))
* **helm:** harden tempo defaults for timeout and load (KAN-145) ([c1d6f3b](https://github.com/kube-rca/kuberca/commit/c1d6f3b3862e1fcd85833c20008e59f0e57de482))
* **i18n:** localize remaining Korean alert message to English ([30ec6d6](https://github.com/kube-rca/kuberca/commit/30ec6d690e4e719c79021bee5bfdac5c8ede4196))
* import unimported modules ([5e10f3a](https://github.com/kube-rca/kuberca/commit/5e10f3a81e24ef106a9652f95c44f4a3ae2e9cc9))
* import unimported modules ([b0b23de](https://github.com/kube-rca/kuberca/commit/b0b23de07567b3732faad4f480e45bdafd62d338))
* improve OIDC error display with prominent banner and URL cleanup ([54855b4](https://github.com/kube-rca/kuberca/commit/54855b406ec9c2ebcae83b93c4be0b9c42e6349c))
* **incidents:** prevent duplicate ongoing incidents with atomic upsert ([1fb7590](https://github.com/kube-rca/kuberca/commit/1fb759095a9ed7482eda2deac0da7f73c66cb1b2))
* **kube-rca:** harden chart security, add probes, fix secret bug ([e15824c](https://github.com/kube-rca/kuberca/commit/e15824c2f853ca083c288f8d80ec1ac5d1275b45))
* **kube-rca:** increase postgresql CPU limits to resolve throttling ([d530d91](https://github.com/kube-rca/kuberca/commit/d530d9133c7b9d273a3290a39a6291ca704113aa))
* **kube-rca:** remove container securityContext defaults ([9884eda](https://github.com/kube-rca/kuberca/commit/9884edaf777691b2837978268d36ad9176682887))
* **kube-rca:** remove runAsNonRoot from default podSecurityContext ([86b3ef6](https://github.com/kube-rca/kuberca/commit/86b3ef64c282303928baa2df984d2c7718007233))
* lint check ([57d95d9](https://github.com/kube-rca/kuberca/commit/57d95d967858a75e27ad35ec09917a6d2c8297e1))
* normalize webhook config mapping for list and edit ([f1d79d6](https://github.com/kube-rca/kuberca/commit/f1d79d6fe189ef21617a101a08c58cf627d49b4f))
* normalize webhook config mapping for list and edit ([2875698](https://github.com/kube-rca/kuberca/commit/287569869c76f6b6ecfb85575b0caffe41b7a8e3))
* prevent ID column text wrapping in all tables ([d7de3f7](https://github.com/kube-rca/kuberca/commit/d7de3f7527d520082b6ff275cb951649fba3edbc))
* prevent URL sync from clearing OIDC error params before AuthPanel reads them ([5d89ef1](https://github.com/kube-rca/kuberca/commit/5d89ef122595d003c76ed3f0462cc62415fe7d7e))
* **prompt:** enforce bold markdown headers for analysis subsections ([4d8b9a1](https://github.com/kube-rca/kuberca/commit/4d8b9a1b80a8840df0521aa8aa713a39ad33bbaa))
* publish chart and image paths under kube-rca-ecr ([31130cf](https://github.com/kube-rca/kuberca/commit/31130cf94c72749982cbdc8641c44a779ddfb2ea))
* recover chat session state and standardize english chat errors ([6e9c0fd](https://github.com/kube-rca/kuberca/commit/6e9c0fd6f1944eb0b0ea43f28dae02b316705e80))
* replace login screen icon with actual logo and fix favicon ([1ba14c9](https://github.com/kube-rca/kuberca/commit/1ba14c9ea525b8c7007278de56ed712b73f90d4b))
* resolve all npm audit vulnerabilities ([6e93ee2](https://github.com/kube-rca/kuberca/commit/6e93ee2de8152fdf3e5b9e0f3c09e1c1cd9cb23b))
* robustly map slack webhook data in list and edit ([2f42417](https://github.com/kube-rca/kuberca/commit/2f4241795eca68c0f8e050c5321750f1b9c9cb36))
* robustly map slack webhook data in list and edit ([cac6f3a](https://github.com/kube-rca/kuberca/commit/cac6f3ad6ada99fcce460b9f0ec4750904730914))
* route agent image pushes to component repository ([4600466](https://github.com/kube-rca/kuberca/commit/4600466ed52d151f615abfa67d19e6233ed64eca))
* route backend image pushes to component repository ([aed51c5](https://github.com/kube-rca/kuberca/commit/aed51c5cc068310214b34c716d35845e77ee4b92))
* route frontend image pushes to component repository ([294f43a](https://github.com/kube-rca/kuberca/commit/294f43a0eda89f2cfba47ff7da3f31081d79953e))
* **scenario:** prevent signal re-entrance during cleanup ([e327ccc](https://github.com/kube-rca/kuberca/commit/e327cccddc6e1295a5681a76ddf0e56b8adce35b))
* **scenario:** use original manifest for cleanup instead of temp file ([8070895](https://github.com/kube-rca/kuberca/commit/80708957e266b995512377ba834d69e4efad001a))
* skip URL param sync when not authenticated. ([5d89ef1](https://github.com/kube-rca/kuberca/commit/5d89ef122595d003c76ed3f0462cc62415fe7d7e))
* **tempo:** normalize time window and surface query failures ([e2cdbb3](https://github.com/kube-rca/kuberca/commit/e2cdbb3503a4937a56f624804dd2236be854caef))
* **ui:** resolve dashboard toolbar overflow and table responsive issues ([5cf556c](https://github.com/kube-rca/kuberca/commit/5cf556cd5ee45a8140b6a6a2195b6d125f8491a6))
* use email as login_id for OIDC users instead of oidc_{sub} ([7878524](https://github.com/kube-rca/kuberca/commit/7878524809220e99a2edb681d9874eeebbf2d39c))
* 자동 메세지 전송처리 삭제 ([511125a](https://github.com/kube-rca/kuberca/commit/511125a6d294080a952fd955881842e237e99150))
* 채팅 기능 404 에러 (KAN-157) ([0a87c78](https://github.com/kube-rca/kuberca/commit/0a87c783133e87994115dace7108ebbbc68ef2d5))


### Performance Improvements

* **agent:** add analysis pipeline timing instrumentation ([74ad719](https://github.com/kube-rca/kuberca/commit/74ad7193705bc415190c14edf8f8d3a08999c445))
* **kube-rca:** increase agent replicaCount default to 2 ([b2c6994](https://github.com/kube-rca/kuberca/commit/b2c69946dc903794bcfa55df31cc607114052a4e))

## Pre-monorepo Component History

The four components below lived in separate repositories before being
merged into this monorepo on 2026-04-28. The histories are preserved
verbatim from `kube-rca/{backend,frontend,agent,helm-charts}` for
audit/archaeology purposes. Commit hashes link to the archived repos,
which remain read-only on GitHub.

### Backend (`kube-rca/backend`, archived)

#### [0.5.1](https://github.com/kube-rca/backend/compare/0.5.0...0.5.1) (2026-04-03)


##### Bug Fixes

* **backend:** add retry with exponential backoff for agent analysis requests ([9b34ed3](https://github.com/kube-rca/backend/commit/9b34ed36efb5ca336a0201b749d11fa7cf899f75))
* **backend:** upgrade go-jose/v4 to 4.1.4 for JWE panic fix (GHSA high) ([2d349e5](https://github.com/kube-rca/backend/commit/2d349e567e2dd428779e9fb133af80928f35e792))

#### [0.5.0](https://github.com/kube-rca/backend/compare/0.4.0...0.5.0) (2026-04-02)


##### Features

* Add a required `name` field to webhook configurations, includin… ([cb483de](https://github.com/kube-rca/backend/commit/cb483de3d0c06c324f84ed714421ad8bbc4e79c3))
* Add a required `name` field to webhook configurations, including database schema, API requests, and validation logic. ([fac832d](https://github.com/kube-rca/backend/commit/fac832d4cf5a7ae2853b100f023e176f311e3180))
* add auto/manual analysis mode with severity-based control (KAN-172) ([48ecf14](https://github.com/kube-rca/backend/commit/48ecf14fdab1f6fa8a8d6b7cc2f228dd27935e41))
* add firing/resolved analysis split to alert detail API ([4bcf251](https://github.com/kube-rca/backend/commit/4bcf251fa58b57d17210db9c58d2f22ac2ccd882))
* add multi-provider OIDC support with auto-detection ([04ee09b](https://github.com/kube-rca/backend/commit/04ee09b776b4ea1534e48d3ba1f39a18222c2e77))
* add notifier events and webhook-based routing ([e1aceba](https://github.com/kube-rca/backend/commit/e1aceba598a4398fd2bed710dd28bbbe2b412657))
* add notifier events and webhook-based routing ([e1aceba](https://github.com/kube-rca/backend/commit/e1aceba598a4398fd2bed710dd28bbbe2b412657))
* add notifier events and webhook-based routing ([c77a6f4](https://github.com/kube-rca/backend/commit/c77a6f4893cbc0060f1d929373066cc34577c448))
* add settings option ([b6c9231](https://github.com/kube-rca/backend/commit/b6c92315f00df16f7ed5e6cca572d4314844e1ea))
* add severity-channel mapping rule ([d8f8638](https://github.com/kube-rca/backend/commit/d8f863866c97c5982c3360547d9ca3e9e9214a20))
* add SSE support for real-time event notifications (KAN-165) ([d26e895](https://github.com/kube-rca/backend/commit/d26e895380870a3f0fa17d681bcf02faba2486cd))
* Alanlysis 기능 추가 (KAN-170) ([1d9f4a8](https://github.com/kube-rca/backend/commit/1d9f4a8e9723577156d3ef166abb42b833be17fe))
* allow info severity payload ([ad77cf9](https://github.com/kube-rca/backend/commit/ad77cf925b56a1cddeb3bd8f031b9b727d156b3f))
* **analyze:** add analysis SSE events and is_analyzing API field (KAN-173) ([b2ebbe4](https://github.com/kube-rca/backend/commit/b2ebbe437d95fdf59f2134586d8761b968d02ebc))
* **analyze:** add in-flight dedup guard and configurable agent timeout ([f421628](https://github.com/kube-rca/backend/commit/f4216280d8043d800dd598ee72e5f794ced9fdd4))
* **analyze:** differentiate resolved alert analysis from firing ([f5545f0](https://github.com/kube-rca/backend/commit/f5545f053cebc0d054860ea9d6e53f319ea3dab3))
* **backend:** add IsManual flag to AlertStatusChangedEvent for manual resolve Slack prefix ([69a09bc](https://github.com/kube-rca/backend/commit/69a09bcf853308dedfd23bf41a1b8613251183ef))
* **backend:** add manual alert resolve request/response models ([038b545](https://github.com/kube-rca/backend/commit/038b5455821f0ba8969a42f37bc433cb2217e837))
* **backend:** add ManualResolveAlert DB function ([b03d56d](https://github.com/kube-rca/backend/commit/b03d56d77e95e66de425e4b6ba8a0585a45d8c38))
* **backend:** add ResolveAlert and BulkResolveAlerts handlers with alertService dependency ([652266d](https://github.com/kube-rca/backend/commit/652266d8b971df2e79f8377d0951e30412a7a48a))
* **backend:** add ResolveAlert and BulkResolveAlerts service methods with tests ([8aa926c](https://github.com/kube-rca/backend/commit/8aa926c1668147531307d894a3a909a3c1084679))
* **backend:** DB-backed notification delivery tracking with strict thread routing ([433faa7](https://github.com/kube-rca/backend/commit/433faa7f25da4e634ec7bc5e5cba967d7980bf95))
* **backend:** register alert resolve routes and update OpenAPI docs ([6efb61b](https://github.com/kube-rca/backend/commit/6efb61b3afab4674265a9519a59928ff229d5c3c))
* **db:** add connection retry with exponential backoff and DB-aware health endpoints ([bc1c23a](https://github.com/kube-rca/backend/commit/bc1c23a26b6f424a8f9d6b89816bdb2ea5944a9b))
* display OIDC login_id as 'Name (email)' format ([493873f](https://github.com/kube-rca/backend/commit/493873f85d6cefec8d669969320f9c5e417d8015))
* **incidents:** expose is_analyzing field in incident detail API ([358c316](https://github.com/kube-rca/backend/commit/358c31680ce5b940208fcfd0f858b26b5904213e))
* remove duplicated setting ([add8fa3](https://github.com/kube-rca/backend/commit/add8fa3f2bc906a689d4c951c83d030b64d51874))
* sync UI/Helm setting ([a7dab54](https://github.com/kube-rca/backend/commit/a7dab54ed5393e69343c91d46d47133cb3ddb6d1))


##### Bug Fixes

* add prompt=select_account to OIDC auth URL ([50ce8f3](https://github.com/kube-rca/backend/commit/50ce8f39ee2b589d7f6cda73f342791b024c79e6))
* **alerts:** resolve fingerprint dedup causing alert history loss ([#25](https://github.com/kube-rca/backend/issues/25)) ([d1fc66e](https://github.com/kube-rca/backend/commit/d1fc66ec413e2cf528cd2bb5c5048feea1ac7c83))
* analysis-channel mapping rule ([22af24a](https://github.com/kube-rca/backend/commit/22af24a25cf9fdcc2ad741a56144d629296e6d6e))
* **analyze:** correct fingerprint lookup and add incident-level analysis ([7dbd88a](https://github.com/kube-rca/backend/commit/7dbd88a801d983b3f30c94d552c8a2e6cabd821f))
* **backend:** allow resolved analysis to wait for in-flight firing analysis ([8b83561](https://github.com/kube-rca/backend/commit/8b83561902355691ca562e4fac8a755d820d6017))
* **backend:** enable markdown rendering for Slack attachments ([d45c938](https://github.com/kube-rca/backend/commit/d45c93808e3e96a17b6761e609dd762252d78c89))
* **backend:** handle triple asterisk in Slack markdown conversion ([56eb03e](https://github.com/kube-rca/backend/commit/56eb03e3ea49aa905008789bcd4d638af1fb630f))
* **backend:** preserve thread_ts on resolved alerts for async analysis routing ([da9ed5c](https://github.com/kube-rca/backend/commit/da9ed5c63db41e90af787af4caaa7cb7ed55f63e))
* **backend:** set StartsAt and prefer direct ThreadTS in manual resolve ([77ce68e](https://github.com/kube-rca/backend/commit/77ce68e5b967ddec195403384a9f1c8969a7cd9a))
* **backend:** strip markdown bold markers inside heading for Slack conversion ([58b58bd](https://github.com/kube-rca/backend/commit/58b58bd51f4af33b8435e5914cd8ec8cc7dd52b9))
* **backend:** use semantic JSON comparison for SyncEnvDefaults ([af34aae](https://github.com/kube-rca/backend/commit/af34aae8e1d9794ed9ebcdc83ca8cfedd377e098))
* **deps:** upgrade google.golang.org/grpc to v1.79.3 ([ad6fda1](https://github.com/kube-rca/backend/commit/ad6fda1ddfa8985d9a2d8aaeb83549afffbb5a6e))
* edit thread_ts analysis mapping rule ([f8b309f](https://github.com/kube-rca/backend/commit/f8b309f76293a7301a63cbabea580a96055f4737))
* **incidents:** prevent duplicate ongoing incidents with atomic upsert ([09b391a](https://github.com/kube-rca/backend/commit/09b391a02948165b5b45f563344a2467d3dd3d36))
* use email as login_id for OIDC users instead of oidc_{sub} ([2efb25b](https://github.com/kube-rca/backend/commit/2efb25bf82a8bbaed3ca49125e6d2128feb5e05a))

#### [0.4.0](https://github.com/kube-rca/backend/compare/0.3.0...0.4.0) (2026-02-27)


##### Features

* add OIDC authentication with security hardening ([b182d24](https://github.com/kube-rca/backend/commit/b182d24714f6b57ad1b4cb5cacf7a20dbe3fbf18))
* add webhook config api ([65e788d](https://github.com/kube-rca/backend/commit/65e788d5adab46a39c57e25ca2922a0285effa6f))
* add webhook config api ([65e788d](https://github.com/kube-rca/backend/commit/65e788d5adab46a39c57e25ca2922a0285effa6f))
* add webhook config api ([15170f9](https://github.com/kube-rca/backend/commit/15170f9be16d74f0d18762304a342c934fb44fb2))
* alert flapping (KAN-155) ([0a55a99](https://github.com/kube-rca/backend/commit/0a55a9921d1e2d0354dbe195c4d90651b5d96853))
* Comment 수정 및 삭제 기능 (KAN-151) ([8f79c44](https://github.com/kube-rca/backend/commit/8f79c4473d7e4764a910160b16f5e8212e146e57))
* flapping enable/disable (KAN-155) ([804316b](https://github.com/kube-rca/backend/commit/804316bbaab7393f20096d412028ea860f5b997f))
* Incident & Alarm에 Comment 추가 (KAN-151) ([751570e](https://github.com/kube-rca/backend/commit/751570e6d61ae946a6366c6db503407491024c95))
* 라벨 기반 검색을 위해 목록 반환시 라벨 추가(KAN-143) ([7bdeb97](https://github.com/kube-rca/backend/commit/7bdeb979fd63aaa17eb6c67dc4ff3e0e43def080))
* 채팅 기능 (KAN-157) ([feaadae](https://github.com/kube-rca/backend/commit/feaadae2710a064ffdcb4081c3a4a0c70fc79ac9))


##### Bug Fixes

* route backend image pushes to component repository ([b0274da](https://github.com/kube-rca/backend/commit/b0274da089e1f99e40d80965a82ea2454fe686bd))
* 채팅 기능 404 에러 (KAN-157) ([9dbde7e](https://github.com/kube-rca/backend/commit/9dbde7e71d9fcbd37cf170003dac904575a9aa2a))

#### [0.3.0](https://github.com/kube-rca/backend/compare/0.2.0...0.3.0) (2026-02-01)


##### Features

* add ci pipeline ([ac63d12](https://github.com/kube-rca/backend/commit/ac63d12073eed1dd6b161a2a0071648a84b1e41f))
* add ci pipeline ([4d7d485](https://github.com/kube-rca/backend/commit/4d7d485426f4d26da05b0eec0c149ae541a6713d))
* add embedding request/response models ([8db38ef](https://github.com/kube-rca/backend/commit/8db38efa57113fa51e714dfe214e6abdb43d3a7d))
* add embeddings handler and route ([49747e8](https://github.com/kube-rca/backend/commit/49747e8a3f749d7fc12c27bfa2bc5557c1345f24))
* add embeddings insert query ([844c903](https://github.com/kube-rca/backend/commit/844c903ace23c630e78d3ecea9fc645bcb382b24))
* add embeddings service ([a26a398](https://github.com/kube-rca/backend/commit/a26a398acc80d66e1e02a67bf3e8f7439d58f1e6))
* add genai embedding client ([bcd0c2f](https://github.com/kube-rca/backend/commit/bcd0c2f965f207cc1798938f8bf722a41a5c120f))
* add pg connection logic ([1af289f](https://github.com/kube-rca/backend/commit/1af289fef2cd009e4e8934027e16d61114519b62))
* add pg connection logic ([fa06783](https://github.com/kube-rca/backend/commit/fa06783fdaf13c2723487a9e1d0b6dee5e9e487f))
* **agent:** include incident_id in analysis request [build] ([6486662](https://github.com/kube-rca/backend/commit/64866621ef43aab906746f8b76c58066cb37e5dd))
* Alert Manager 웹훅 DTO 및 handler 추가 ([461a524](https://github.com/kube-rca/backend/commit/461a52410e30c4d09ba21d562b72d4b26b5e5c8f))
* **backend:** add JWT auth [KAN-87] ([e1db648](https://github.com/kube-rca/backend/commit/e1db6483b1f0e9254066ef37b9b1b4239a298094))
* **backend:** add JWT auth [KAN-87] ([f965e75](https://github.com/kube-rca/backend/commit/f965e75ae1ca88a1a9744272103353e7006ffea2))
* **backend:** add swagger docs [KAN-92] [build] ([a20e9bf](https://github.com/kube-rca/backend/commit/a20e9bfcc4ec75f4b7057d221d97d5c91f633188))
* **backend:** persist alert analysis artifacts (KAN-102) ([92dfb33](https://github.com/kube-rca/backend/commit/92dfb338436575136144176fb5532e0a33f6b4ce))
* **backend:** store alert summary/detail separately (KAN-124) ([ae857b4](https://github.com/kube-rca/backend/commit/ae857b4a8bf27bfd84a31bd4d426f7755d6f9975))
* **backend:** support local .env [KAN-87] ([3f2b7a0](https://github.com/kube-rca/backend/commit/3f2b7a0a2a758673824cb82ad3cc38acd34e6129))
* **client:** agent client 구현 ([c6bddee](https://github.com/kube-rca/backend/commit/c6bddee980ba8a1268af48e93deb6b9890d29e32))
* **client:** Slack API 연동 로직 추가 ([416f03d](https://github.com/kube-rca/backend/commit/416f03d6cde1a2f0c6a5981a64c94445f815536e))
* **client:** Slack 메시지 포맷, 필터링, 전송 로직 추가 ([73a2b76](https://github.com/kube-rca/backend/commit/73a2b7643af3b61b4a7f069a88d1be5d7a33bd3d))
* DB is_enabled 컬럼 추가 [build] ([597f924](https://github.com/kube-rca/backend/commit/597f924d784f71f2377b2fbe1d6d65df04c248cd))
* DB is_enabled 컬럼 추가 [build] ([015aaf1](https://github.com/kube-rca/backend/commit/015aaf1d3926d0f1ad2ee4b51225f9ff2f762c61))
* **embedding:** add vector search API endpoint ([63310b0](https://github.com/kube-rca/backend/commit/63310b092783539b81eed5d8d128911a67aa7769))
* **embedding:** add vector search API endpoint ([63310b0](https://github.com/kube-rca/backend/commit/63310b092783539b81eed5d8d128911a67aa7769))
* **embedding:** add vector search API endpoint ([2eff8dc](https://github.com/kube-rca/backend/commit/2eff8dcca2177c4750f0467c7dcf605961fdd188))
* enable EnsureEmbeddingSchema on startup ([c4576b7](https://github.com/kube-rca/backend/commit/c4576b774600ec4e9311297554f0108a829cc332))
* enable EnsureEmbeddingSchema on startup ([c4576b7](https://github.com/kube-rca/backend/commit/c4576b774600ec4e9311297554f0108a829cc332))
* enable EnsureEmbeddingSchema on startup ([03b3388](https://github.com/kube-rca/backend/commit/03b338882941857b10b8ecdb1c188ee601fbd1c8))
* **handler:** Service 레이어 분리, main 로직 변경 ([21c7759](https://github.com/kube-rca/backend/commit/21c77599c988220ad11cad75b81225fa0cba9beb))
* **handler:** 디버깅용 payload 로깅 ([8f052e6](https://github.com/kube-rca/backend/commit/8f052e6bd91a71afbf69df0cf6163262ad6ff769))
* RCA 보드 목록 반환 API [KAN-54] ([89c3710](https://github.com/kube-rca/backend/commit/89c37109c0cf6bfe09ca22f48576512479a31e8c))
* **service:** agent 비동기 호출/콜백 로직 추가 ([2a1c119](https://github.com/kube-rca/backend/commit/2a1c11986c9703846cda6c61ba4d2f350918102f))
* 특정 RCA 정보 반환 [KAN-80] ([401f3e7](https://github.com/kube-rca/backend/commit/401f3e79829d2e2a2bc06fd876de5ff709883a6c))
* 특정 RCA 편집 [KAN-81] [build] ([70731d0](https://github.com/kube-rca/backend/commit/70731d07a07fea8a810672307224fdf3894ed181))


##### Bug Fixes

* add pg go module ([a2e304c](https://github.com/kube-rca/backend/commit/a2e304cad2416b3f73669c5dcacdb4c0a972d48d))
* **backend:** convert headings for slack [KAN-132] ([0d90067](https://github.com/kube-rca/backend/commit/0d90067b4b220220de8012375359006dfa597536))
* **backend:** slack markdown conversion [KAN-132] ([ecbf12c](https://github.com/kube-rca/backend/commit/ecbf12cfa582111d2c35b5ebb54947320fc2353d))
* change ai secret env var name [build] ([25113dd](https://github.com/kube-rca/backend/commit/25113ddb71e9e3e4ca81fb94e74d155ad8f9b1e6))
* change file name and position ([411c9db](https://github.com/kube-rca/backend/commit/411c9db2c1584d9c984dd76b26c08a5d5c149969))
* change go version from 1.22 to 1.23 ([173fe03](https://github.com/kube-rca/backend/commit/173fe032442f7fc2b2c6ede21aed05043ebb8a84))
* change go version to 1.24 [build] ([0a97136](https://github.com/kube-rca/backend/commit/0a97136887f188f852aedf76ca46df26f90b632c))
* **ci:** avoid double colon in tag [build] ([d09cd6e](https://github.com/kube-rca/backend/commit/d09cd6e5e35b470f7a446907563cd097843b3835))
* **ci:** CA 인증서 복사 ([a0ac26b](https://github.com/kube-rca/backend/commit/a0ac26b705e2b06f2aa940ba6b4bdc88ed701211))
* **ci:** write image tag output [build] ([8736efb](https://github.com/kube-rca/backend/commit/8736efbe22cc4dacf0139508e7c89863c7b2c8a3))
* correct embedding_test.go to use config.EmbeddingConfig ([cff1013](https://github.com/kube-rca/backend/commit/cff10138f2ce96d0aa6e0c5319cf90ebc1b01fb8))
* correct short SHA extraction and image tag setting in CI workflow [build] ([8165c8a](https://github.com/kube-rca/backend/commit/8165c8a0386ce6f90cf18293b8f1c491d24fb8f7))
* debug 위해 embedding api key 로그에 출력 ([fa15415](https://github.com/kube-rca/backend/commit/fa15415a595253b5d3cd23eb97492ea41957b261))
* **docs:** use absolute GitHub URLs for related components ([813555e](https://github.com/kube-rca/backend/commit/813555ef3e6dadfbcb406bfe045be763670eab9d))
* **docs:** use absolute URL for logo image ([a759fc7](https://github.com/kube-rca/backend/commit/a759fc70193532ec421020dbb4a82c312aa0079d))
* embedding schema 초기화 롤백 ([1011b1a](https://github.com/kube-rca/backend/commit/1011b1a44ac6d9810b6fdece831198556a8d6e1f))
* embedding text 전달방식 수정 ([0100e41](https://github.com/kube-rca/backend/commit/0100e41f4686e03427014911ff2fd2b85334a280))
* embedding 관련 db schema 초기화하도록 설정 ([edea791](https://github.com/kube-rca/backend/commit/edea7911d93e3e05d6dab3d134516cf258d949c8))
* embedding 설정 관련 수정: [build] ([12a755b](https://github.com/kube-rca/backend/commit/12a755b2573ea3bda9fda987d9d78449fa086c78))
* embedding 설정 관련 수정: [build] ([89c5024](https://github.com/kube-rca/backend/commit/89c5024613905f135afff5d89ebbfa3d8027fbb4))
* enable ci pipeline ([09a255a](https://github.com/kube-rca/backend/commit/09a255af841be9d3ddc5ec3c1e355e8efaa79616))
* gemini client backend endpoint 명시적으로 지정 [build] ([1f4e051](https://github.com/kube-rca/backend/commit/1f4e0512b727da7d9b5adf91105c246e7de7fd9a))
* improve duplicate resolved alert detection [KAN-139] ([23b7a2d](https://github.com/kube-rca/backend/commit/23b7a2d068c7569a1597837d41a790275dfc168a))
* **logging:** skip health and openapi access logs ([8934f05](https://github.com/kube-rca/backend/commit/8934f0513fbf454732c19f681630ac1767bf16b6))
* login to public ecr [build] ([ed2b44d](https://github.com/kube-rca/backend/commit/ed2b44d87b8a231a7a8cc6a5091231685c35c0f9))
* update CI pipeline to correctly set image tags and environment variables ([2165203](https://github.com/kube-rca/backend/commit/21652038c16e18a5bc7bf4395775b041f31e2e9c))
* update CI workflow to correctly set short SHA and image tag outputs [build] ([e8a2d9c](https://github.com/kube-rca/backend/commit/e8a2d9c0f10aa1617a69a6c2477c42f2008ea6bc))
* update image tag format to avoid invalid reference formats in CI workflow [build] ([c85122a](https://github.com/kube-rca/backend/commit/c85122aff0c29cd4339ec9d660687989a60c0228))
* update image tag generation to use repository name instead of environment variable [build] ([6915c47](https://github.com/kube-rca/backend/commit/6915c47353890476ebead86fef43232621be6a5a))

#### [0.2.0](https://github.com/kube-rca/backend/compare/kube-rca-backend-v0.1.0...kube-rca-backend-v0.2.0) (2026-02-01)


##### Features

* add ci pipeline ([ac63d12](https://github.com/kube-rca/backend/commit/ac63d12073eed1dd6b161a2a0071648a84b1e41f))
* add ci pipeline ([4d7d485](https://github.com/kube-rca/backend/commit/4d7d485426f4d26da05b0eec0c149ae541a6713d))
* add embedding request/response models ([8db38ef](https://github.com/kube-rca/backend/commit/8db38efa57113fa51e714dfe214e6abdb43d3a7d))
* add embeddings handler and route ([49747e8](https://github.com/kube-rca/backend/commit/49747e8a3f749d7fc12c27bfa2bc5557c1345f24))
* add embeddings insert query ([844c903](https://github.com/kube-rca/backend/commit/844c903ace23c630e78d3ecea9fc645bcb382b24))
* add embeddings service ([a26a398](https://github.com/kube-rca/backend/commit/a26a398acc80d66e1e02a67bf3e8f7439d58f1e6))
* add genai embedding client ([bcd0c2f](https://github.com/kube-rca/backend/commit/bcd0c2f965f207cc1798938f8bf722a41a5c120f))
* add pg connection logic ([1af289f](https://github.com/kube-rca/backend/commit/1af289fef2cd009e4e8934027e16d61114519b62))
* add pg connection logic ([fa06783](https://github.com/kube-rca/backend/commit/fa06783fdaf13c2723487a9e1d0b6dee5e9e487f))
* **agent:** include incident_id in analysis request [build] ([6486662](https://github.com/kube-rca/backend/commit/64866621ef43aab906746f8b76c58066cb37e5dd))
* Alert Manager 웹훅 DTO 및 handler 추가 ([461a524](https://github.com/kube-rca/backend/commit/461a52410e30c4d09ba21d562b72d4b26b5e5c8f))
* **backend:** add JWT auth [KAN-87] ([e1db648](https://github.com/kube-rca/backend/commit/e1db6483b1f0e9254066ef37b9b1b4239a298094))
* **backend:** add JWT auth [KAN-87] ([f965e75](https://github.com/kube-rca/backend/commit/f965e75ae1ca88a1a9744272103353e7006ffea2))
* **backend:** add swagger docs [KAN-92] [build] ([a20e9bf](https://github.com/kube-rca/backend/commit/a20e9bfcc4ec75f4b7057d221d97d5c91f633188))
* **backend:** persist alert analysis artifacts (KAN-102) ([92dfb33](https://github.com/kube-rca/backend/commit/92dfb338436575136144176fb5532e0a33f6b4ce))
* **backend:** store alert summary/detail separately (KAN-124) ([ae857b4](https://github.com/kube-rca/backend/commit/ae857b4a8bf27bfd84a31bd4d426f7755d6f9975))
* **backend:** support local .env [KAN-87] ([3f2b7a0](https://github.com/kube-rca/backend/commit/3f2b7a0a2a758673824cb82ad3cc38acd34e6129))
* **client:** agent client 구현 ([c6bddee](https://github.com/kube-rca/backend/commit/c6bddee980ba8a1268af48e93deb6b9890d29e32))
* **client:** Slack API 연동 로직 추가 ([416f03d](https://github.com/kube-rca/backend/commit/416f03d6cde1a2f0c6a5981a64c94445f815536e))
* **client:** Slack 메시지 포맷, 필터링, 전송 로직 추가 ([73a2b76](https://github.com/kube-rca/backend/commit/73a2b7643af3b61b4a7f069a88d1be5d7a33bd3d))
* DB is_enabled 컬럼 추가 [build] ([597f924](https://github.com/kube-rca/backend/commit/597f924d784f71f2377b2fbe1d6d65df04c248cd))
* DB is_enabled 컬럼 추가 [build] ([015aaf1](https://github.com/kube-rca/backend/commit/015aaf1d3926d0f1ad2ee4b51225f9ff2f762c61))
* **embedding:** add vector search API endpoint ([63310b0](https://github.com/kube-rca/backend/commit/63310b092783539b81eed5d8d128911a67aa7769))
* **embedding:** add vector search API endpoint ([63310b0](https://github.com/kube-rca/backend/commit/63310b092783539b81eed5d8d128911a67aa7769))
* **embedding:** add vector search API endpoint ([2eff8dc](https://github.com/kube-rca/backend/commit/2eff8dcca2177c4750f0467c7dcf605961fdd188))
* enable EnsureEmbeddingSchema on startup ([c4576b7](https://github.com/kube-rca/backend/commit/c4576b774600ec4e9311297554f0108a829cc332))
* enable EnsureEmbeddingSchema on startup ([c4576b7](https://github.com/kube-rca/backend/commit/c4576b774600ec4e9311297554f0108a829cc332))
* enable EnsureEmbeddingSchema on startup ([03b3388](https://github.com/kube-rca/backend/commit/03b338882941857b10b8ecdb1c188ee601fbd1c8))
* **handler:** Service 레이어 분리, main 로직 변경 ([21c7759](https://github.com/kube-rca/backend/commit/21c77599c988220ad11cad75b81225fa0cba9beb))
* **handler:** 디버깅용 payload 로깅 ([8f052e6](https://github.com/kube-rca/backend/commit/8f052e6bd91a71afbf69df0cf6163262ad6ff769))
* RCA 보드 목록 반환 API [KAN-54] ([89c3710](https://github.com/kube-rca/backend/commit/89c37109c0cf6bfe09ca22f48576512479a31e8c))
* **service:** agent 비동기 호출/콜백 로직 추가 ([2a1c119](https://github.com/kube-rca/backend/commit/2a1c11986c9703846cda6c61ba4d2f350918102f))
* 특정 RCA 정보 반환 [KAN-80] ([401f3e7](https://github.com/kube-rca/backend/commit/401f3e79829d2e2a2bc06fd876de5ff709883a6c))
* 특정 RCA 편집 [KAN-81] [build] ([70731d0](https://github.com/kube-rca/backend/commit/70731d07a07fea8a810672307224fdf3894ed181))


##### Bug Fixes

* add pg go module ([a2e304c](https://github.com/kube-rca/backend/commit/a2e304cad2416b3f73669c5dcacdb4c0a972d48d))
* **backend:** convert headings for slack [KAN-132] ([0d90067](https://github.com/kube-rca/backend/commit/0d90067b4b220220de8012375359006dfa597536))
* **backend:** slack markdown conversion [KAN-132] ([ecbf12c](https://github.com/kube-rca/backend/commit/ecbf12cfa582111d2c35b5ebb54947320fc2353d))
* change ai secret env var name [build] ([25113dd](https://github.com/kube-rca/backend/commit/25113ddb71e9e3e4ca81fb94e74d155ad8f9b1e6))
* change file name and position ([411c9db](https://github.com/kube-rca/backend/commit/411c9db2c1584d9c984dd76b26c08a5d5c149969))
* change go version from 1.22 to 1.23 ([173fe03](https://github.com/kube-rca/backend/commit/173fe032442f7fc2b2c6ede21aed05043ebb8a84))
* change go version to 1.24 [build] ([0a97136](https://github.com/kube-rca/backend/commit/0a97136887f188f852aedf76ca46df26f90b632c))
* **ci:** avoid double colon in tag [build] ([d09cd6e](https://github.com/kube-rca/backend/commit/d09cd6e5e35b470f7a446907563cd097843b3835))
* **ci:** CA 인증서 복사 ([a0ac26b](https://github.com/kube-rca/backend/commit/a0ac26b705e2b06f2aa940ba6b4bdc88ed701211))
* **ci:** write image tag output [build] ([8736efb](https://github.com/kube-rca/backend/commit/8736efbe22cc4dacf0139508e7c89863c7b2c8a3))
* correct embedding_test.go to use config.EmbeddingConfig ([cff1013](https://github.com/kube-rca/backend/commit/cff10138f2ce96d0aa6e0c5319cf90ebc1b01fb8))
* correct short SHA extraction and image tag setting in CI workflow [build] ([8165c8a](https://github.com/kube-rca/backend/commit/8165c8a0386ce6f90cf18293b8f1c491d24fb8f7))
* debug 위해 embedding api key 로그에 출력 ([fa15415](https://github.com/kube-rca/backend/commit/fa15415a595253b5d3cd23eb97492ea41957b261))
* **docs:** use absolute GitHub URLs for related components ([813555e](https://github.com/kube-rca/backend/commit/813555ef3e6dadfbcb406bfe045be763670eab9d))
* **docs:** use absolute URL for logo image ([a759fc7](https://github.com/kube-rca/backend/commit/a759fc70193532ec421020dbb4a82c312aa0079d))
* embedding schema 초기화 롤백 ([1011b1a](https://github.com/kube-rca/backend/commit/1011b1a44ac6d9810b6fdece831198556a8d6e1f))
* embedding text 전달방식 수정 ([0100e41](https://github.com/kube-rca/backend/commit/0100e41f4686e03427014911ff2fd2b85334a280))
* embedding 관련 db schema 초기화하도록 설정 ([edea791](https://github.com/kube-rca/backend/commit/edea7911d93e3e05d6dab3d134516cf258d949c8))
* embedding 설정 관련 수정: [build] ([12a755b](https://github.com/kube-rca/backend/commit/12a755b2573ea3bda9fda987d9d78449fa086c78))
* embedding 설정 관련 수정: [build] ([89c5024](https://github.com/kube-rca/backend/commit/89c5024613905f135afff5d89ebbfa3d8027fbb4))
* enable ci pipeline ([09a255a](https://github.com/kube-rca/backend/commit/09a255af841be9d3ddc5ec3c1e355e8efaa79616))
* gemini client backend endpoint 명시적으로 지정 [build] ([1f4e051](https://github.com/kube-rca/backend/commit/1f4e0512b727da7d9b5adf91105c246e7de7fd9a))
* improve duplicate resolved alert detection [KAN-139] ([23b7a2d](https://github.com/kube-rca/backend/commit/23b7a2d068c7569a1597837d41a790275dfc168a))
* **logging:** skip health and openapi access logs ([8934f05](https://github.com/kube-rca/backend/commit/8934f0513fbf454732c19f681630ac1767bf16b6))
* login to public ecr [build] ([ed2b44d](https://github.com/kube-rca/backend/commit/ed2b44d87b8a231a7a8cc6a5091231685c35c0f9))
* update CI pipeline to correctly set image tags and environment variables ([2165203](https://github.com/kube-rca/backend/commit/21652038c16e18a5bc7bf4395775b041f31e2e9c))
* update CI workflow to correctly set short SHA and image tag outputs [build] ([e8a2d9c](https://github.com/kube-rca/backend/commit/e8a2d9c0f10aa1617a69a6c2477c42f2008ea6bc))
* update image tag format to avoid invalid reference formats in CI workflow [build] ([c85122a](https://github.com/kube-rca/backend/commit/c85122aff0c29cd4339ec9d660687989a60c0228))
* update image tag generation to use repository name instead of environment variable [build] ([6915c47](https://github.com/kube-rca/backend/commit/6915c47353890476ebead86fef43232621be6a5a))

### Frontend (`kube-rca/frontend`, archived)

#### [0.4.1](https://github.com/kube-rca/frontend/compare/0.4.0...0.4.1) (2026-04-03)


##### Bug Fixes

* **frontend:** remove duplicate severity badge in webhook routing ([c44ab8e](https://github.com/kube-rca/frontend/commit/c44ab8e98dffa427e31734a9b1b8993b43b7c416))

#### [0.4.0](https://github.com/kube-rca/frontend/compare/0.3.0...0.4.0) (2026-03-29)


##### Features

* add analysis mode indicator to sidebar (KAN-172) ([e743255](https://github.com/kube-rca/frontend/commit/e7432552699c80a33b1f868d6103e23718eaf6e2))
* add analysis mode UI with manual analyze buttons (KAN-172) ([32d8c2c](https://github.com/kube-rca/frontend/commit/32d8c2cf4010f10b71319ae0b5ad3f1bf743466f))
* add info severity to AnalysisSettings and RCADetailView ([1f537c8](https://github.com/kube-rca/frontend/commit/1f537c8c0e935102579863193c6004cb6e5fa23c))
* add settings option ([0808105](https://github.com/kube-rca/frontend/commit/080810515f476c24f0eecc8e6f83985a51896a91))
* add severity-channel mapping rule ([3a80c86](https://github.com/kube-rca/frontend/commit/3a80c86baeac2b00bfaf57cf0eeb937711669250))
* Add show/hide toggle for token input fields and remove an infor… ([29436d6](https://github.com/kube-rca/frontend/commit/29436d6ace2df4521e7573d05288eedf7359f5f2))
* Add show/hide toggle for token input fields and remove an informational section in webhook settings. ([f25ba75](https://github.com/kube-rca/frontend/commit/f25ba752fbc391ec8cf67065960f85c875bfdd76))
* Alanlysis Dashboard 추가 (KAN-171) ([2ff848d](https://github.com/kube-rca/frontend/commit/2ff848daf866d52d84a122a0e36d785d6983fd60))
* analysis export 및 꺾은선 value 추가 ([cf42400](https://github.com/kube-rca/frontend/commit/cf42400e93c865fb7c86188d961f245d9b49a0bb))
* **analyze:** add incident analysis polling and remove banner auto-dismiss ([d814e82](https://github.com/kube-rca/frontend/commit/d814e8234983d34c1e3c8aeae2662a66591fac2e))
* apply same readability improvements to AlertTable ([0db9042](https://github.com/kube-rca/frontend/commit/0db90426a4b5c121f81b649dce1084dc868f8a37))
* display firing/resolved analyses in stacked layout ([d55b6b9](https://github.com/kube-rca/frontend/commit/d55b6b9212e60ca82dc6b802ef3042d2fb9e623f))
* export 기능 추가 (KAN-169) ([acfb2ec](https://github.com/kube-rca/frontend/commit/acfb2ecaaa8b86e146e062cf701bb7f5b0579a02))
* **frontend:** add bulk resolve UI with checkboxes to AlertTable ([922dd76](https://github.com/kube-rca/frontend/commit/922dd763aa2efba747b945934cf83a6a9a48aa53))
* **frontend:** add manual Resolve button to AlertDetailView ([a8e3671](https://github.com/kube-rca/frontend/commit/a8e3671260a412b34fa8034729de6cbea7232649))
* **frontend:** add resolveAlert and bulkResolveAlerts API functions ([bfc400f](https://github.com/kube-rca/frontend/commit/bfc400f034d16220757eb9a89100859015cc2ba7))
* Implement webhook name support across frontend components, API … ([d233fb4](https://github.com/kube-rca/frontend/commit/d233fb4fc32a156b985f7f09a7c0ea1f0bcaf40b))
* Implement webhook name support across frontend components, API types, and add a feature plan. ([9dbd26c](https://github.com/kube-rca/frontend/commit/9dbd26cb3c2eb06e608a7f1964518bb6a28ec691))
* improve analysis UX, fix settings load bug, add logo (KAN-173, KAN-172) ([92fd158](https://github.com/kube-rca/frontend/commit/92fd1587e822a492c69af8bc8b6ae68764c8fdb8))
* improve incident table readability with larger text and column dividers ([3dcebab](https://github.com/kube-rca/frontend/commit/3dcebabeda59f9e7ea1e6b5a578a3cde03fa31ea))
* new webhook routing navagation ([0465f22](https://github.com/kube-rca/frontend/commit/0465f22d40d1c5c209676082e5acaf522afb3e89))
* redesign UI to Industrial Ops style ([4d1bbc5](https://github.com/kube-rca/frontend/commit/4d1bbc55d03f7d2f44c43d8db7e8c1ef41c62616))
* remove alert owner hint message ([538724d](https://github.com/kube-rca/frontend/commit/538724d2f6a332c8701b86e84279c2b2d22e4765))
* remove duplicated setting ([9e2fe94](https://github.com/kube-rca/frontend/commit/9e2fe945378ce6813e8ff5e56f9ae76dfd1d4d18))
* render provider-specific OIDC login button ([4f0aa53](https://github.com/kube-rca/frontend/commit/4f0aa531b497a6ac9277f6b26b86671d53b3b485))
* replace 1s polling with SSE + 30s fallback polling (KAN-165) ([b8fe313](https://github.com/kube-rca/frontend/commit/b8fe3133b826eaa75296c956072192b96114b510))
* shrink Time column and expand Title column in tables ([06c3de0](https://github.com/kube-rca/frontend/commit/06c3de06d5aeba986b989ce07ee13062930f0a81))
* support slack bot token and channel webhook settings ([dbd8fee](https://github.com/kube-rca/frontend/commit/dbd8fee7168e9d3c23983c9251aa12a68725ce32))
* support slack bot token and channel webhook settings ([538cef6](https://github.com/kube-rca/frontend/commit/538cef6ada9314bf899de131474c57f1e28052a1))
* unify table styles across Alert, Incident, and Archive dashboards ([af8cb78](https://github.com/kube-rca/frontend/commit/af8cb78567418bd845ad782a5e90ea1fac782eb4))


##### Bug Fixes

* **analyze:** always show Re-Analyze button, add incident analyze, remove Analyze All ([27e990a](https://github.com/kube-rca/frontend/commit/27e990ad0a64a879c6870154dc648087ef7e3e54))
* **analyze:** improve polling logic with grace period and seenAnalyzing tracking ([4c35f23](https://github.com/kube-rca/frontend/commit/4c35f23f0c317cf0c5d817906cc3ac40882c9285))
* **deps:** upgrade flatted to 3.4.2 ([86c02e2](https://github.com/kube-rca/frontend/commit/86c02e2305ab8d824b4ff2f08993304075c93d00))
* **deps:** upgrade picomatch to fix CVE-2026-33672 (GHSA-3v7f-55p6-f55p) ([bf44b53](https://github.com/kube-rca/frontend/commit/bf44b534460264ed4ed3264bb687db5397ca23b2))
* **frontend:** add line-clamp and tooltip to title columns in all tables ([efce8ad](https://github.com/kube-rca/frontend/commit/efce8adc16238b5a95dc1e6e86d378d784a33fac))
* **frontend:** hide checkbox column when no firing alerts exist ([600ea74](https://github.com/kube-rca/frontend/commit/600ea74a1efb30e158f7e24c7969222bd894bcef))
* **i18n:** localize remaining Korean alert message to English ([0b56d07](https://github.com/kube-rca/frontend/commit/0b56d07a8f7f10c2222525507d243d21095fd9a8))
* normalize webhook config mapping for list and edit ([d21c5ba](https://github.com/kube-rca/frontend/commit/d21c5ba82473c1125ae729762b146869a8b4c91d))
* normalize webhook config mapping for list and edit ([affd5e9](https://github.com/kube-rca/frontend/commit/affd5e98d5d561f8c567a3a5cf28c4317711b0f8))
* prevent ID column text wrapping in all tables ([03e812a](https://github.com/kube-rca/frontend/commit/03e812a60068c0a4786a47b8120a0142c01c5c85))
* prevent URL sync from clearing OIDC error params before AuthPanel reads them ([fe07231](https://github.com/kube-rca/frontend/commit/fe07231bd549674abe99fc67e2386dfe17310e0a))
* replace login screen icon with actual logo and fix favicon ([165380d](https://github.com/kube-rca/frontend/commit/165380dc52b26c3c90cd135ae8a1ae29ca5126c4))
* resolve all npm audit vulnerabilities ([9122750](https://github.com/kube-rca/frontend/commit/9122750a5ea1e7f9e6a3208cf2c9579cc90c1863))
* robustly map slack webhook data in list and edit ([5b75d63](https://github.com/kube-rca/frontend/commit/5b75d6352740025fdf2dbe6ff204dbc1aed62c18))
* robustly map slack webhook data in list and edit ([55140a0](https://github.com/kube-rca/frontend/commit/55140a0e318c6b75ccdd896154bd54b31c62989e))
* skip URL param sync when not authenticated. ([fe07231](https://github.com/kube-rca/frontend/commit/fe07231bd549674abe99fc67e2386dfe17310e0a))
* **ui:** resolve dashboard toolbar overflow and table responsive issues ([3145448](https://github.com/kube-rca/frontend/commit/31454486e0182e30cb02f6d169d4a49e84224a19))

#### [0.3.0](https://github.com/kube-rca/frontend/compare/0.2.0...0.3.0) (2026-02-27)


##### Features

* Add alert variables and refactor webhook body variable insertion with a reusable component and improved UI. ([5124fb8](https://github.com/kube-rca/frontend/commit/5124fb8af35771320fef7b41ddb98180663af6b4))
* Add alert variables and refactor webhook body variables ([7a630da](https://github.com/kube-rca/frontend/commit/7a630da84c1a2fdc30187f2a85a71c215b9176d3))
* Add alert variables and refactor webhook body variables ([7a630da](https://github.com/kube-rca/frontend/commit/7a630da84c1a2fdc30187f2a85a71c215b9176d3))
* add bg color to reserved field ([903a945](https://github.com/kube-rca/frontend/commit/903a945f7cb14e417f4aa2826671fd227240fcb8))
* add bg color to reserved field ([9f13b53](https://github.com/kube-rca/frontend/commit/9f13b53b42d4f96943b74ca0116e626706c63653))
* add OIDC login button and callback error handling ([7189ec5](https://github.com/kube-rca/frontend/commit/7189ec530657847319bc804e95c805a96801af01))
* add settings page ([c03b3ac](https://github.com/kube-rca/frontend/commit/c03b3ac9ec949d4ed813d59e8df93a13d9c28505))
* add settings page ([a2a3e8c](https://github.com/kube-rca/frontend/commit/a2a3e8cb7c89cb9c1578a5440d558c5f38265834))
* add webhook list page ([77d774a](https://github.com/kube-rca/frontend/commit/77d774afd31c120ffb48913af3169027a1f91c7f))
* add webhook list page ([b91a692](https://github.com/kube-rca/frontend/commit/b91a6924d843f90fb1a18eee838d2f1ac494ce2e))
* Comment 편집 및 삭제 추가 (KAN-153) ([ddf70cf](https://github.com/kube-rca/frontend/commit/ddf70cff8782e77b0f4f4ddc0303986b7284c8ee))
* Incident & Alarm에 Comment 기능 (KAN-153) ([6b9668c](https://github.com/kube-rca/frontend/commit/6b9668c03194e954e1779e4cc266aac602cc7135))
* 검색 기능 추가 (KAN-143) ([032a93f](https://github.com/kube-rca/frontend/commit/032a93fed7b16f8af8304fe74235ddd5f093b614))
* 채팅 기능 (KAN-86) ([e2242b9](https://github.com/kube-rca/frontend/commit/e2242b9b982fb5077cee42cc94e0ae8ebfc26898))


##### Bug Fixes

* Comment 메뉴바 표시와 일부 단축어 오류 수정 (KAN-153) ([1364ff5](https://github.com/kube-rca/frontend/commit/1364ff5573aaeb2fcb910f172a535382b2bfc582))
* Comment 편집 및 삭제 오류 수정 (KAN-153) ([8efe3bc](https://github.com/kube-rca/frontend/commit/8efe3bc942856b95ea3dca393c6863a7083c9658))
* daymode 글자안보이던 현상 해결 (KAN-96) ([307ec02](https://github.com/kube-rca/frontend/commit/307ec0268f3a3f441f17d26b159de4f2a59dbe6c))
* import unimported modules ([871f192](https://github.com/kube-rca/frontend/commit/871f1922dba8d06233cfc7857a6c98f8fcc47a16))
* import unimported modules ([60151b7](https://github.com/kube-rca/frontend/commit/60151b79d97d0502c475917d9a6b1cd202ef8940))
* improve OIDC error display with prominent banner and URL cleanup ([a0f1b9b](https://github.com/kube-rca/frontend/commit/a0f1b9bf3cd50d0575cbb6dd2c44f842f3141e65))
* route frontend image pushes to component repository ([5741190](https://github.com/kube-rca/frontend/commit/5741190bae0cbdb5d3e394e61e9f487e7487022f))
* 자동 메세지 전송처리 삭제 ([040ce56](https://github.com/kube-rca/frontend/commit/040ce566dd7440e6f259a1b4e4c453432bc1d725))

#### [0.2.0](https://github.com/kube-rca/frontend/compare/0.1.0...0.2.0) (2026-02-01)


##### Features

* add Dockerfile ([0e2e755](https://github.com/kube-rca/frontend/commit/0e2e75578b4ffdccaaa8d8bb412c4f0f21812fa1))
* **auth:** add auth panel and token refresh [KAN-88] [build] ([4b46068](https://github.com/kube-rca/frontend/commit/4b4606801a3f4dd09dd83319ebbdfd8aa68bcfac))
* **auth:** add auth panel and token refresh [KAN-88] [build] ([598638d](https://github.com/kube-rca/frontend/commit/598638d7fa27514b0c94d1265aa126821597d2f2))
* Create ci.yml ([f2dbf7b](https://github.com/kube-rca/frontend/commit/f2dbf7b1550983849c4d6ca5a6a18311e4daab5d))
* implement alerts dashboard with filtering and pagination ([52f3eed](https://github.com/kube-rca/frontend/commit/52f3eed2609946b26b0e27e446c49a65c134d21c))
* RCA 편집 기능 구현 [KAN-56] ([416aff0](https://github.com/kube-rca/frontend/commit/416aff0af53b4a0ea5acd9b9a8d53fb5c97b758b))
* 다크모드 구현 [KAN-89] [build] ([1171e62](https://github.com/kube-rca/frontend/commit/1171e62daa5de727644af2db111f53800d650253))
* 특정 RCA 보고서 조회 [KAN-55] ([b219f5e](https://github.com/kube-rca/frontend/commit/b219f5eb0301d7560d7addd16b38c18feed729fe))


##### Bug Fixes

* **auth:** improve dark mode input contrast [KAN-89] [build] ([77025c2](https://github.com/kube-rca/frontend/commit/77025c294c9891a478b528b597d23ce5340149bc))
* **docs:** use absolute GitHub URLs for related components ([1beda23](https://github.com/kube-rca/frontend/commit/1beda23f926df3334ab2ac35cd6b0c1a7dfb4d0b))
* **docs:** use absolute URL for logo image ([303c66b](https://github.com/kube-rca/frontend/commit/303c66b000bf49140dd9f9e8fa2ef150b26e6951))
* resolve npm audit vulnerabilities ([94fae8b](https://github.com/kube-rca/frontend/commit/94fae8bff91d8cc3ad8775d7e4528b6f41f30f1b))
* Update README.md [build] ([8db3788](https://github.com/kube-rca/frontend/commit/8db37884d449d60937c479d0832b327c676392a0))

### Agent (`kube-rca/agent`, archived)

#### [1.2.1](https://github.com/kube-rca/agent/compare/1.2.0...1.2.1) (2026-04-03)


##### Bug Fixes

* **agent:** bypass kubernetes client event_time validation via raw JSON ([e692ab8](https://github.com/kube-rca/agent/commit/e692ab8a3c90f23cc3e674b19fe607c95375d7e8))
* **agent:** handle PostgreSQL race condition in session schema initialization ([93e0dda](https://github.com/kube-rca/agent/commit/93e0dda04ceabd2f7189550f80b04f1025bdf737))
* **agent:** improve analysis reliability and resource management ([88bdf8b](https://github.com/kube-rca/agent/commit/88bdf8bb10ba601a53b6aca4e8ce903488d6dae0))


##### Performance Improvements

* **agent:** add analysis pipeline timing instrumentation ([f71b7a8](https://github.com/kube-rca/agent/commit/f71b7a8cb180297633e857df7243ed35b04f70de))

#### [1.2.0](https://github.com/kube-rca/agent/compare/1.1.0...1.2.0) (2026-04-02)


##### Features

* add built-in redaction masking-chaining [KAN-174] ([fb7777e](https://github.com/kube-rca/agent/commit/fb7777e50ed7c7d768e2eefff511e86c26bfdb72))
* add settings option ([241e12c](https://github.com/kube-rca/agent/commit/241e12c4ba80449bd9a86e4fb57fadf6ef44799b))
* add total timeout stop condition to LLM retry logic ([41ab472](https://github.com/kube-rca/agent/commit/41ab472e05dc6632505c60b8b65c85db15b513ad))
* **agent:** add Loki log aggregation client and Strands tools ([fc4788b](https://github.com/kube-rca/agent/commit/fc4788b8e0716b36278f596699976e9eb7824f90))
* **agent:** add tool timing logs for analysis ([6777f6e](https://github.com/kube-rca/agent/commit/6777f6ee854a2f72a5c9857a41b9b73290aa225a))
* **agent:** async 엔드포인트 전환으로 분석 병렬 처리 지원 ([97d1b0d](https://github.com/kube-rca/agent/commit/97d1b0dbd4a8e2363c88d3a4701604782ae2c35f))
* **agent:** portable K8s baseline context and capability-aware prompt ([8f111d0](https://github.com/kube-rca/agent/commit/8f111d0ec76b5b65050c2a87fd3d277632e7e1d5))
* **agent:** tune LLM retry to 3-min exponential backoff ([89dc3d2](https://github.com/kube-rca/agent/commit/89dc3d214f8b02367c791899a447db03f069bdb1))
* **analyze:** differentiate resolved alert analysis from firing ([da6ed90](https://github.com/kube-rca/agent/commit/da6ed904a9c157f8c2979385e24b0610c683af78))


##### Bug Fixes

* **agent:** add httpx transport errors to LLM retry and sanitize messages ([8ca98d7](https://github.com/kube-rca/agent/commit/8ca98d76c6bbdea81d8ec2f5380c021cc95cfb09))
* **agent:** context-aware missing data and stronger tool usage directive ([175d5f0](https://github.com/kube-rca/agent/commit/175d5f0b31f8d5401fc2f0e8f234f21a2b541885))
* **agent:** disable default PrintingCallbackHandler in Strands Agent ([6f4620c](https://github.com/kube-rca/agent/commit/6f4620ce55ca76ad232646ab5e6b2eefeea6327b))
* **agent:** enforce completed-findings tone and active tool usage in prompt ([b92db2a](https://github.com/kube-rca/agent/commit/b92db2a602d54d6fb05a6ec81faf6e4fb6a89238))
* **agent:** extract detail section only and strip markdown bold from title ([d421e11](https://github.com/kube-rca/agent/commit/d421e11b18d7d41aad31d14a69b9d9b908859ba6))
* **agent:** handle Gemini turn-order violations from sliding-window truncation ([c715989](https://github.com/kube-rca/agent/commit/c71598913181299c83d2e9c2d95cf270d48cc7d6))
* **agent:** improve analysis fallback for missing pod labels and empty exceptions ([79c5440](https://github.com/kube-rca/agent/commit/79c544059de0e25e2ac097d64afaeaa8aa15ae02))
* **agent:** improve incident summary parsing for bold-without-colon LLM responses ([22fd82b](https://github.com/kube-rca/agent/commit/22fd82b32e17e60afdddc78fe50e1f176d24a17d))
* **agent:** prevent entire LLM response from leaking into incident title and summary ([3fb72ec](https://github.com/kube-rca/agent/commit/3fb72eccaf060ebf7de8a8d9607214d80f0ef595))
* **agent:** remove incident summary length limit to prevent truncation ([8c06521](https://github.com/kube-rca/agent/commit/8c06521e6357fe4895780eec8a1397ffe494dbb6))
* **agent:** set anthropic max token defaults ([55a0abd](https://github.com/kube-rca/agent/commit/55a0abd399b02e334a560219533675b30d0f9cb0))
* **agent:** upgrade strands-agents 1.23.0 → 1.30.0, remove thought signature monkey-patch ([dbbcd22](https://github.com/kube-rca/agent/commit/dbbcd226d35e81bc7a3873f4c744b73953e8ee2e))
* **agent:** use FQDN service.name for Tempo TraceQL queries ([6f1d453](https://github.com/kube-rca/agent/commit/6f1d453a8691616255d3ea7c8105cb0538d344e2))
* **analysis:** remove hardcoded summary length limits ([c4ea258](https://github.com/kube-rca/agent/commit/c4ea258b6b099bef1f5254553ab9afd4e9d19b5d))
* **deps:** upgrade cryptography 46.0.6, requests 2.33.0 ([20bcf65](https://github.com/kube-rca/agent/commit/20bcf6554657aab48decf2269571b572ac307410))
* **deps:** upgrade pyasn1 to 0.6.3 ([00374f3](https://github.com/kube-rca/agent/commit/00374f35310714126a19bd4e350cb82e03f1c6e9))
* **deps:** upgrade Pygments 2.19.2 to 2.20.0 ([eff7999](https://github.com/kube-rca/agent/commit/eff79991cbb4aa5bfd008a0d2a843750bdade7b4))
* **deps:** upgrade PyJWT 2.10.1 → 2.12.1 (CVE-2026-32597) ([f4badf7](https://github.com/kube-rca/agent/commit/f4badf7f057c9716ed67e2f2f31777137114536b))
* lint check ([fe64a5b](https://github.com/kube-rca/agent/commit/fe64a5b3702ebbb1cfac82fc68fb2c19a6fb6a8f))
* **prompt:** enforce bold markdown headers for analysis subsections ([37c5215](https://github.com/kube-rca/agent/commit/37c521511c656af8cca84476c805f458558a7746))


##### Documentation

* **agent:** note manual resolve as analysis trigger source ([d915336](https://github.com/kube-rca/agent/commit/d915336ef2de1f977458d66cb2300e5f7f35221d))

#### [1.1.0](https://github.com/kube-rca/agent/compare/1.0.0...1.1.0) (2026-02-22)


##### Features

* add tenacity-based LLM API retry for transient errors (KAN-159) ([3ed6790](https://github.com/kube-rca/agent/commit/3ed6790129266a7c084e59bc11d272610c14a78a))
* **agent:** add generic manifest read tools ([ebe49e7](https://github.com/kube-rca/agent/commit/ebe49e724aa19a4e1a55269de33cdb146078077b))
* **agent:** add regex masking for llm and storage flows (KAN-146) ([#10](https://github.com/kube-rca/agent/issues/10)) ([e1aad94](https://github.com/kube-rca/agent/commit/e1aad94c94db2882d1805ec2b34cb1e4865c2e37))
* **agent:** enrich daemonset and events context (KAN-79) ([8c678bc](https://github.com/kube-rca/agent/commit/8c678bc51bc04059abe6a1b9c3e7a7b18eac7c29))
* **agent:** enrich daemonset and events context (KAN-79) ([ace5c00](https://github.com/kube-rca/agent/commit/ace5c003ac9c4877256d6088a69e39e66ecff4ea))
* **agent:** integrate tempo trace context into analysis (KAN-145) ([#9](https://github.com/kube-rca/agent/issues/9)) ([e17cd5e](https://github.com/kube-rca/agent/commit/e17cd5efe834e0680393a8fbf62dd0f2bdf47365))
* 채팅 기능 추가 (KAN-158) ([4fd3051](https://github.com/kube-rca/agent/commit/4fd3051219d31d5d1e07a8cb5e4f62497e975d4d))


##### Bug Fixes

* **agent:** align tempo timeout default and docs (KAN-145) ([a19bd6e](https://github.com/kube-rca/agent/commit/a19bd6eb34134133264a7af23e9958885f112b09))
* build error ([4751f41](https://github.com/kube-rca/agent/commit/4751f4104f96f6699f322e7576029f7a55f6efd7))
* bump cryptography to 46.0.5 for CVE-2026-26007 ([9036346](https://github.com/kube-rca/agent/commit/90363461a6fba2ab38af8a6d00af21ce4c23c6d0))
* Gemini 400 function-call turn order error in multi-turn chat ([4d87b9a](https://github.com/kube-rca/agent/commit/4d87b9a18979949b9be8500b0d2bb83bc36de9d9))
* recover chat session state and standardize english chat errors ([2d08669](https://github.com/kube-rca/agent/commit/2d08669128de6d092e720be09a07045e696c1b43))
* route agent image pushes to component repository ([35c6b84](https://github.com/kube-rca/agent/commit/35c6b84f922962fedcf4d3bad8227bb1ddb41f35))
* **tempo:** normalize time window and surface query failures ([f1e4cf8](https://github.com/kube-rca/agent/commit/f1e4cf8753593665c0a1511fb8bde2ac02e9ccfd))


##### Documentation

* add LLM retry environment variables to README (KAN-159) ([ad4faa5](https://github.com/kube-rca/agent/commit/ad4faa5846add0aaff4ed2ab7de4c66e8ac84364))
* add release-please merge policy ([#8](https://github.com/kube-rca/agent/issues/8)) ([52c8e91](https://github.com/kube-rca/agent/commit/52c8e91193d3df3d96a1ec311365ffeb63bb8a95))
* **agent:** align README with current API schema ([b3755af](https://github.com/kube-rca/agent/commit/b3755afa50d9b3ca5e5fef286c84689b060810d4))
* **agent:** update README for generic manifest tools ([974a17f](https://github.com/kube-rca/agent/commit/974a17f752cb89c18d7da8308d883042ea467eb5))
* **license:** align with MIT ([14719ed](https://github.com/kube-rca/agent/commit/14719ed1eb9010b7759b145952680e9cb03a3728))

#### [1.0.0](https://github.com/kube-rca/agent/compare/0.1.0...1.0.0) (2026-02-01)


##### ⚠ BREAKING CHANGES

* **agent:** rename analyze endpoint to /analyze [build]

##### Features

* **agent:** accept alert analysis request payload ([d42b2c0](https://github.com/kube-rca/agent/commit/d42b2c0772b5d38c932f8fe957ea3c8ec3680b43))
* **agent:** add alert analysis endpoint ([212d3db](https://github.com/kube-rca/agent/commit/212d3db392173a403badadea21ab5900f3ef810c))
* **agent:** add alert context artifacts (KAN-102) ([f624861](https://github.com/kube-rca/agent/commit/f6248612e3d9c580d7b53271b091ace870e635ce))
* **agent:** add CrashLoop/ImagePull test targets and refactor startup ([43e3715](https://github.com/kube-rca/agent/commit/43e37150487669d216d032471e70f0143f378d49))
* **agent:** add openapi export [KAN-92] [build] ([38cb58c](https://github.com/kube-rca/agent/commit/38cb58c7133f3e387701dc47c8b46b24641d1d76))
* **agent:** add pod discovery and prometheus metrics listing tools ([70b139f](https://github.com/kube-rca/agent/commit/70b139f1a06431b7966f5074a1e84a29ebb708b1))
* **agent:** add prometheus client and expand analysis tools [build] ([f63a01e](https://github.com/kube-rca/agent/commit/f63a01e608a44ec0e654b9342ca622ac9e95490a))
* **agent:** add prompt budget and session summaries ([1150336](https://github.com/kube-rca/agent/commit/11503360c6c420de6e5751585d6267ee87345a05))
* **agent:** improve OOMKilled testing scripts and docs [build] ([a8fc161](https://github.com/kube-rca/agent/commit/a8fc161e2443bce0dc294526a650395174dfdd3c))
* **agent:** migrate agent to FastAPI [KAN-66] [build] ([084a5ad](https://github.com/kube-rca/agent/commit/084a5adb1b652f5006e9d2e206b694a30f9734a3))
* **agent:** rename analyze endpoint to /analyze [build] ([b0cef8c](https://github.com/kube-rca/agent/commit/b0cef8cc4078de1ce59558857f5fc11e69eb5f1e))
* **agent:** require PROMETHEUS_URL for Prometheus queries (KAN-134) ([c7f99b4](https://github.com/kube-rca/agent/commit/c7f99b4ee7a144d9e9c4f888db5c539157997746))
* **agent:** route docs via proxy [KAN-92] [build] ([dd717ec](https://github.com/kube-rca/agent/commit/dd717ec7e029e0f62b3dcbe5dc8380951bb0c17c))
* **agent:** split alert summary/detail (KAN-124) ([11f6665](https://github.com/kube-rca/agent/commit/11f6665c7078269c0bde83e8ccf1e4a6c3695949))
* **agent:** support provider model ids (KAN-144) ([218b0c5](https://github.com/kube-rca/agent/commit/218b0c503dc08d1cc746d25e8e3a5163da2ffb15))
* **agent:** support WEB_CONCURRENCY workers [KAN-77] ([1ee94d1](https://github.com/kube-rca/agent/commit/1ee94d1e2c35776ede731765babe44726600d4e3))
* **api:** pass incident_id to analysis engine ([68b29f8](https://github.com/kube-rca/agent/commit/68b29f85575aec14ac6e19de6e681c95ed982bc8))
* **clients:** add PostgreSQL session repository ([ca5e78e](https://github.com/kube-rca/agent/commit/ca5e78e2857fe5784037bbeeafe69bd5fe7e2c1b))
* **clients:** add session-based agent caching ([524bc3d](https://github.com/kube-rca/agent/commit/524bc3d7fbc51822f6721437beebc8846f3e9178))
* **config:** add session DB and cache settings ([8f39c37](https://github.com/kube-rca/agent/commit/8f39c37d08f00f453fbdfb7164859378105dd3b3))
* **llm:** support multi-provider (Gemini, OpenAI, Anthropic) (KAN-144) ([1aeef99](https://github.com/kube-rca/agent/commit/1aeef994702cd992d2d249620c64f5931056d477))
* **prometheus:** add query_range for time-series history data [KAN-134] ([cb0130f](https://github.com/kube-rca/agent/commit/cb0130f2f85b507f03620514fffe12eb113c0db2))


##### Bug Fixes

* **agent:** bump kubernetes for urllib3 security fixes ([527ca45](https://github.com/kube-rca/agent/commit/527ca45ae79494808c7f5d0de9ed02fb6fdc7423))
* **agent:** bump kubernetes for urllib3 security fixes ([459406e](https://github.com/kube-rca/agent/commit/459406e596babb620e3e377dda11a0ccc82fc587))
* **agent:** enforce list dash in prompt [KAN-132] ([7166a06](https://github.com/kube-rca/agent/commit/7166a06bc22cbf03c1ea6a938f9fca9059b99330))
* **agent:** guard empty analysis response ([78d02be](https://github.com/kube-rca/agent/commit/78d02be74ce867c6e340a5fcc484a3b3ad455b33))
* **agent:** handle non-utf8 thought signatures [build] ([22551d0](https://github.com/kube-rca/agent/commit/22551d039c5836d69d47a0a20c3c6f73978a9556))
* **agent:** isolate alert and summary sessions ([d28a72b](https://github.com/kube-rca/agent/commit/d28a72bcb85f8f93b7b2107fcec0bf4c46e75639))
* **agent:** make session writes idempotent ([c70a142](https://github.com/kube-rca/agent/commit/c70a14283f17d2155fd1f668e894eff08e726697))
* **agent:** refine output formatting [KAN-132] ([98dabda](https://github.com/kube-rca/agent/commit/98dabda4f1a7de6f9b45a21b527d14005b12200d))
* **agent:** resolve ruff lint errors [build] ([8d4dde9](https://github.com/kube-rca/agent/commit/8d4dde9a107542ba6e453bce3198d8a89377a372))
* **agent:** serialize strands session writes [KAN-118] ([dc55e5e](https://github.com/kube-rca/agent/commit/dc55e5e136b3a86bd8112b3369df01773076d7d6))
* **agent:** tolerate gemini formatter signature changes ([921c742](https://github.com/kube-rca/agent/commit/921c7427486c2cd8719620035f7f00db29ec5dbd))
* **config:** revert gemini model to gemini-3-flash-preview ([7ed05c2](https://github.com/kube-rca/agent/commit/7ed05c2b4c1e72aaf53f10264b6e3456c3abf6e1))
* **docs:** use absolute GitHub URLs for related components ([a33bd99](https://github.com/kube-rca/agent/commit/a33bd99bdc8308c0bc8b82c4bc1538e91bd16531))
* **docs:** use absolute URL for logo image ([05347ae](https://github.com/kube-rca/agent/commit/05347ae799170155b82cc2fafecf23ac91b5f4b1))
* **logging:** filter health and openapi access logs ([4c10348](https://github.com/kube-rca/agent/commit/4c10348ff779b5960bd741e634a5e9ea433c7d3b))
* serialize alert payload and add local OOM test target [build] ([6ac5bf6](https://github.com/kube-rca/agent/commit/6ac5bf6ad5782790b1bb7c488e5dced12c1c18de))


##### Documentation

* **agent:** add README and response schema ([cffe4ed](https://github.com/kube-rca/agent/commit/cffe4ede61f70169ac882c550ee7da621e175055))
* **agent:** translate README to English ([2e7f8f0](https://github.com/kube-rca/agent/commit/2e7f8f0fb52e01dc5f4311fcae42c32834ad967c))
* enhance README for open source project ([3fbd4da](https://github.com/kube-rca/agent/commit/3fbd4dad511896709075a9749ab5584ba95a5bfe))

### Helm Chart (`kube-rca/helm-charts`, archived)

#### [0.7.0](https://github.com/kube-rca/helm-charts/compare/0.6.0...0.7.0) (2026-04-03)


##### Features

* **helm:** add inline apiKey value fields to values.yaml ([c8b1630](https://github.com/kube-rca/helm-charts/commit/c8b163096f0bce1a0c5a511bb90733670343930a))
* **helm:** add postgresql and embedding secret name auto-resolve helpers ([a209880](https://github.com/kube-rca/helm-charts/commit/a209880dcf45e69bc919383ab9ce81a07059d4df))
* **helm:** populate agent secrets with inline apiKey values ([e207f3f](https://github.com/kube-rca/helm-charts/commit/e207f3f7dd7083f81a415d468f706b67ddf86b3e))
* **helm:** use auto-resolve helper for agent session DB secret ([b29727e](https://github.com/kube-rca/helm-charts/commit/b29727e804687b4a912df6d8015638829c9d9e86))
* **helm:** use auto-resolve helpers for postgresql and embedding secrets ([d0e1ae9](https://github.com/kube-rca/helm-charts/commit/d0e1ae9e12b54bf8fc9ee903be31d75ca8c27798))


##### Performance Improvements

* **kube-rca:** increase agent replicaCount default to 2 ([9ceefb2](https://github.com/kube-rca/helm-charts/commit/9ceefb244b8442d1c7eb155a13433d4839057c3c))

#### [0.6.0](https://github.com/kube-rca/helm-charts/compare/0.5.0...0.6.0) (2026-04-02)


##### Features

* add built-in Redaction option [KAN-174] ([f1a0915](https://github.com/kube-rca/helm-charts/commit/f1a09154aa4ca730fb91ca7346d2c9ea1d01aff8))
* add cert manager and move ingress hosts to 2 domains ([413da8e](https://github.com/kube-rca/helm-charts/commit/413da8e82378538c5f403f4a092a0160f487936d))
* add gp2 values for stateful apps ([5fcd369](https://github.com/kube-rca/helm-charts/commit/5fcd369002967cd18a5a32a5cb4df7ba0d5d942a))
* alert flapping 환경변수 (KAN-155) ([7532d90](https://github.com/kube-rca/helm-charts/commit/7532d90f6b4485644795601e4081cc3356d97575))
* flapping enable/disable (KAN-155) ([09f8058](https://github.com/kube-rca/helm-charts/commit/09f8058bafe6f3f7092d6967bb5a1f5bd64e6399))
* **kube-rca:** add LLM retry configuration for agent (KAN-159) ([605d9dd](https://github.com/kube-rca/helm-charts/commit/605d9dd713d21f074ffe76f41bbcab33db5caf1d))
* **kube-rca:** add Loki env vars for agent deployment ([1ecbe44](https://github.com/kube-rca/helm-charts/commit/1ecbe440f00c9068172456afaed0ec055982113e))
* **kube-rca:** add MANUAL_ANALYZE_SEVERITIES env for analysis mode (KAN-172) ([d2091d6](https://github.com/kube-rca/helm-charts/commit/d2091d65460457ed20bf3cc71f49d90c62c6de2f))
* **kube-rca:** add OIDC authentication helm values and deployment env ([7a96d50](https://github.com/kube-rca/helm-charts/commit/7a96d50748531a36842ee9cf7b42a39567ddbed2))
* **kube-rca:** replace hook job with initContainer and add DB-aware probes ([e708991](https://github.com/kube-rca/helm-charts/commit/e708991ffb6d6511e559bf010581336964d13488))


##### Bug Fixes

* **chart:** add agent startupProbe and sync eks values ([c94dcda](https://github.com/kube-rca/helm-charts/commit/c94dcdaab9a1dad843ca0f1d8510a3227db80dcd))
* **chart:** add anthropic max token env ([a5c8090](https://github.com/kube-rca/helm-charts/commit/a5c8090b2192a9e6e8d9987e65634bb97a58596f))
* **kube-rca:** harden chart security, add probes, fix secret bug ([03c8560](https://github.com/kube-rca/helm-charts/commit/03c8560fc4ffa2e0fa40619a24ca7ca3c3e08dfb))
* **kube-rca:** increase postgresql CPU limits to resolve throttling ([147216d](https://github.com/kube-rca/helm-charts/commit/147216d267ba33b3128a2abb7850adb697955e5f))
* **kube-rca:** remove container securityContext defaults ([cf439c6](https://github.com/kube-rca/helm-charts/commit/cf439c6dc0437c6f74e29ee5a3db79bc6c2b9cdc))
* **kube-rca:** remove runAsNonRoot from default podSecurityContext ([4e8dc20](https://github.com/kube-rca/helm-charts/commit/4e8dc20133e4ccdc6d06387d6cb20c8369c6fddb))
* publish chart and image paths under kube-rca-ecr ([128aaa6](https://github.com/kube-rca/helm-charts/commit/128aaa6c043e8caa2a50fbdcac62e4ae754f0293))

#### [0.5.0](https://github.com/kube-rca/helm-charts/compare/0.4.0...0.5.0) (2026-02-09)


##### Features

* **kube-rca:** add agent masking regex helm settings (KAN-146) ([#8](https://github.com/kube-rca/helm-charts/issues/8)) ([3fec5b5](https://github.com/kube-rca/helm-charts/commit/3fec5b52ccc6c276cb5329fe46f9ff399e633a8e))
* **observability:** add tempo stack and wire agent tempo settings (KAN-73) ([#7](https://github.com/kube-rca/helm-charts/issues/7)) ([3a1642f](https://github.com/kube-rca/helm-charts/commit/3a1642f474bf2c72faf99fcad2efccc332acfc13))


##### Bug Fixes

* **helm:** harden tempo defaults for timeout and load (KAN-145) ([3459fde](https://github.com/kube-rca/helm-charts/commit/3459fde4c0a1560f022682f6cf97f6eaff06e260))

#### [0.4.0](https://github.com/kube-rca/helm-charts/compare/0.3.0...0.4.0) (2026-02-05)


##### Features

* **kube-rca:** move postgresql to dependency and update configuration (KAN-110) ([c8229e2](https://github.com/kube-rca/helm-charts/commit/c8229e2396c86e5891f54514dbb030cc39bc59d9))


##### Bug Fixes

* **helm:** align postgres host and hooks (KAN-110) ([a71a963](https://github.com/kube-rca/helm-charts/commit/a71a963b43a22b22a6014f276a39d02fd3961b99))
* **helm:** run wait-for-db after postgres (KAN-110) ([4de3f14](https://github.com/kube-rca/helm-charts/commit/4de3f143cde7664914eaa547c09fd090a6eb666d))

#### [0.3.0](https://github.com/kube-rca/helm-charts/compare/0.2.0...0.3.0) (2026-02-01)


##### Features

* add db config env vars ([5f6e8af](https://github.com/kube-rca/helm-charts/commit/5f6e8af22207f46285ed7e20576c179be5e0e2a0))
* get ai api key from env ([afc9d15](https://github.com/kube-rca/helm-charts/commit/afc9d15d1fc14d53f1e4fc08cbed76bdb9653704))
* **helm:** add backend auth secrets [KAN-87] ([f8a4f7f](https://github.com/kube-rca/helm-charts/commit/f8a4f7f5acebb1cb4eaf253393bd8c290c18a4b0))
* **helm:** add openapi docs service [KAN-92] [build] ([8783f56](https://github.com/kube-rca/helm-charts/commit/8783f566673a38f78a43846d655c210d14c323be))
* **helm:** add provider model ids (KAN-144) ([5a15249](https://github.com/kube-rca/helm-charts/commit/5a15249bb7f6901e18b2a127837c5e4077c853f5))
* **helm:** configure agent worker count [KAN-77] ([5eb9d50](https://github.com/kube-rca/helm-charts/commit/5eb9d5045bd9a07ba5bfffa6942dcb0cce168820))
* **helm:** enforce deploy ordering https://kube-rca.atlassian.net/browse/KAN-109 [KAN-109] ([c9cebc6](https://github.com/kube-rca/helm-charts/commit/c9cebc66f73702fb8fec316d0beac9816cdd9fea))
* **helm:** expose prompt settings for agent ([d7ef0b2](https://github.com/kube-rca/helm-charts/commit/d7ef0b2f1f215aa58c8b9893740e313923ee36d9))
* **helm:** wire PROMETHEUS_URL for agent (KAN-134) ([a8aa0e8](https://github.com/kube-rca/helm-charts/commit/a8aa0e8cb31e5a7868210476802de059f7fadb9c))
* **kube-rca:** add agent component and update docs ([85cf826](https://github.com/kube-rca/helm-charts/commit/85cf82652a8c256b1dd6a181e3925e21f6d548f8))
* **kube-rca:** add agent K8s/Prometheus env vars to Helm chart ([696820d](https://github.com/kube-rca/helm-charts/commit/696820d9320596e9dff1f1fb83e94fbcdaaae7ae))
* **kube-rca:** add frontend/backend ingress ([c74a0ca](https://github.com/kube-rca/helm-charts/commit/c74a0ca76c2e533306b7d67dab3cb8263d208879))
* **kube-rca:** add helm chart and Argo CD application ([d2949c6](https://github.com/kube-rca/helm-charts/commit/d2949c60d1113f94c4fca653b8123cf03ecea936))
* **kube-rca:** add Helm hooks for deployment ordering [KAN-109] ([86fe597](https://github.com/kube-rca/helm-charts/commit/86fe597759bc6e5e85f19041561596a662cf4a65))
* **kube-rca:** make PostgreSQL secret configurable via values ([c0def5c](https://github.com/kube-rca/helm-charts/commit/c0def5cfad1888edf370bd9c5faca0ecf34162d5))
* **kube-rca:** update gemini model to gemini-2.5-flash in values ([f800806](https://github.com/kube-rca/helm-charts/commit/f8008065f7ba3034164b4a928a6a59cbea1eb932))
* update backend image tag ([b821eb2](https://github.com/kube-rca/helm-charts/commit/b821eb254e65c5f15ab08bb0216b43c593456484))
* Update Gemini embedding API key secret name and key ([1a01be6](https://github.com/kube-rca/helm-charts/commit/1a01be6f2304a20729cb2a4f7979066fae9a76ff))


##### Bug Fixes

* **chart:** revert gemini model to gemini-3-flash-preview ([b17326c](https://github.com/kube-rca/helm-charts/commit/b17326c805e457eaaf07f3cf80784405aee2bbb9))
* embedding model 정보 주입 방식 수정 ([1298d78](https://github.com/kube-rca/helm-charts/commit/1298d788b6c72ac6ed978680a351c31ea8060d06))
* fix postgres host name ([123fad5](https://github.com/kube-rca/helm-charts/commit/123fad5016186a33b2918c7c174195e977046e05))
* **helm:** align agent port to 8000 ([78e8bd6](https://github.com/kube-rca/helm-charts/commit/78e8bd6c85c6cff1d9670f2930ccfb2895a19013))
* **helm:** disable backend ingress override [KAN-92] ([54191d6](https://github.com/kube-rca/helm-charts/commit/54191d6400a555dfb205b50a9a994f65346008af))
* **helm:** proxy docs api routes [KAN-92] [build] ([9a09028](https://github.com/kube-rca/helm-charts/commit/9a090282dad67ac1e0055ee2e54e888c3ad2712b))
* **helm:** rollout on config change [KAN-92] ([02d180c](https://github.com/kube-rca/helm-charts/commit/02d180c66304b462774b8fdd68b3e0c099dfe274))
* **kube-rca:** align ai secret and API key env ([893b3c9](https://github.com/kube-rca/helm-charts/commit/893b3c969a5fcfb6d62b9ede3f837ab6e8a92ac5))
* **kube-rca:** remove openapi ingress basic auth ([c5c1ec9](https://github.com/kube-rca/helm-charts/commit/c5c1ec99d53430b66ae131c2aa196a9a26171048))
* **kube-rca:** route /api to backend from frontend ingress ([6ead897](https://github.com/kube-rca/helm-charts/commit/6ead897ea374182ca7fcf26b2b1af9b773e1e5f4))

#### [0.2.0](https://github.com/kube-rca/helm-charts/compare/kube-rca-v0.1.0...kube-rca-v0.2.0) (2026-02-01)


##### Features

* add db config env vars ([5f6e8af](https://github.com/kube-rca/helm-charts/commit/5f6e8af22207f46285ed7e20576c179be5e0e2a0))
* get ai api key from env ([afc9d15](https://github.com/kube-rca/helm-charts/commit/afc9d15d1fc14d53f1e4fc08cbed76bdb9653704))
* **helm:** add backend auth secrets [KAN-87] ([f8a4f7f](https://github.com/kube-rca/helm-charts/commit/f8a4f7f5acebb1cb4eaf253393bd8c290c18a4b0))
* **helm:** add openapi docs service [KAN-92] [build] ([8783f56](https://github.com/kube-rca/helm-charts/commit/8783f566673a38f78a43846d655c210d14c323be))
* **helm:** add provider model ids (KAN-144) ([5a15249](https://github.com/kube-rca/helm-charts/commit/5a15249bb7f6901e18b2a127837c5e4077c853f5))
* **helm:** configure agent worker count [KAN-77] ([5eb9d50](https://github.com/kube-rca/helm-charts/commit/5eb9d5045bd9a07ba5bfffa6942dcb0cce168820))
* **helm:** enforce deploy ordering https://kube-rca.atlassian.net/browse/KAN-109 [KAN-109] ([c9cebc6](https://github.com/kube-rca/helm-charts/commit/c9cebc66f73702fb8fec316d0beac9816cdd9fea))
* **helm:** expose prompt settings for agent ([d7ef0b2](https://github.com/kube-rca/helm-charts/commit/d7ef0b2f1f215aa58c8b9893740e313923ee36d9))
* **helm:** wire PROMETHEUS_URL for agent (KAN-134) ([a8aa0e8](https://github.com/kube-rca/helm-charts/commit/a8aa0e8cb31e5a7868210476802de059f7fadb9c))
* **kube-rca:** add agent component and update docs ([85cf826](https://github.com/kube-rca/helm-charts/commit/85cf82652a8c256b1dd6a181e3925e21f6d548f8))
* **kube-rca:** add agent K8s/Prometheus env vars to Helm chart ([696820d](https://github.com/kube-rca/helm-charts/commit/696820d9320596e9dff1f1fb83e94fbcdaaae7ae))
* **kube-rca:** add frontend/backend ingress ([c74a0ca](https://github.com/kube-rca/helm-charts/commit/c74a0ca76c2e533306b7d67dab3cb8263d208879))
* **kube-rca:** add helm chart and Argo CD application ([d2949c6](https://github.com/kube-rca/helm-charts/commit/d2949c60d1113f94c4fca653b8123cf03ecea936))
* **kube-rca:** add Helm hooks for deployment ordering [KAN-109] ([86fe597](https://github.com/kube-rca/helm-charts/commit/86fe597759bc6e5e85f19041561596a662cf4a65))
* **kube-rca:** make PostgreSQL secret configurable via values ([c0def5c](https://github.com/kube-rca/helm-charts/commit/c0def5cfad1888edf370bd9c5faca0ecf34162d5))
* **kube-rca:** update gemini model to gemini-2.5-flash in values ([f800806](https://github.com/kube-rca/helm-charts/commit/f8008065f7ba3034164b4a928a6a59cbea1eb932))
* update backend image tag ([b821eb2](https://github.com/kube-rca/helm-charts/commit/b821eb254e65c5f15ab08bb0216b43c593456484))
* Update Gemini embedding API key secret name and key ([1a01be6](https://github.com/kube-rca/helm-charts/commit/1a01be6f2304a20729cb2a4f7979066fae9a76ff))


##### Bug Fixes

* **chart:** revert gemini model to gemini-3-flash-preview ([b17326c](https://github.com/kube-rca/helm-charts/commit/b17326c805e457eaaf07f3cf80784405aee2bbb9))
* embedding model 정보 주입 방식 수정 ([1298d78](https://github.com/kube-rca/helm-charts/commit/1298d788b6c72ac6ed978680a351c31ea8060d06))
* fix postgres host name ([123fad5](https://github.com/kube-rca/helm-charts/commit/123fad5016186a33b2918c7c174195e977046e05))
* **helm:** align agent port to 8000 ([78e8bd6](https://github.com/kube-rca/helm-charts/commit/78e8bd6c85c6cff1d9670f2930ccfb2895a19013))
* **helm:** disable backend ingress override [KAN-92] ([54191d6](https://github.com/kube-rca/helm-charts/commit/54191d6400a555dfb205b50a9a994f65346008af))
* **helm:** proxy docs api routes [KAN-92] [build] ([9a09028](https://github.com/kube-rca/helm-charts/commit/9a090282dad67ac1e0055ee2e54e888c3ad2712b))
* **helm:** rollout on config change [KAN-92] ([02d180c](https://github.com/kube-rca/helm-charts/commit/02d180c66304b462774b8fdd68b3e0c099dfe274))
* **kube-rca:** align ai secret and API key env ([893b3c9](https://github.com/kube-rca/helm-charts/commit/893b3c969a5fcfb6d62b9ede3f837ab6e8a92ac5))
* **kube-rca:** remove openapi ingress basic auth ([c5c1ec9](https://github.com/kube-rca/helm-charts/commit/c5c1ec99d53430b66ae131c2aa196a9a26171048))
* **kube-rca:** route /api to backend from frontend ingress ([6ead897](https://github.com/kube-rca/helm-charts/commit/6ead897ea374182ca7fcf26b2b1af9b773e1e5f4))
