# Changelog

## [0.7.0](https://github.com/kube-rca/helm-charts/compare/0.6.0...0.7.0) (2026-04-03)


### Features

* **helm:** add inline apiKey value fields to values.yaml ([c8b1630](https://github.com/kube-rca/helm-charts/commit/c8b163096f0bce1a0c5a511bb90733670343930a))
* **helm:** add postgresql and embedding secret name auto-resolve helpers ([a209880](https://github.com/kube-rca/helm-charts/commit/a209880dcf45e69bc919383ab9ce81a07059d4df))
* **helm:** populate agent secrets with inline apiKey values ([e207f3f](https://github.com/kube-rca/helm-charts/commit/e207f3f7dd7083f81a415d468f706b67ddf86b3e))
* **helm:** use auto-resolve helper for agent session DB secret ([b29727e](https://github.com/kube-rca/helm-charts/commit/b29727e804687b4a912df6d8015638829c9d9e86))
* **helm:** use auto-resolve helpers for postgresql and embedding secrets ([d0e1ae9](https://github.com/kube-rca/helm-charts/commit/d0e1ae9e12b54bf8fc9ee903be31d75ca8c27798))


### Performance Improvements

* **kube-rca:** increase agent replicaCount default to 2 ([9ceefb2](https://github.com/kube-rca/helm-charts/commit/9ceefb244b8442d1c7eb155a13433d4839057c3c))

## [0.6.0](https://github.com/kube-rca/helm-charts/compare/0.5.0...0.6.0) (2026-04-02)


### Features

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


### Bug Fixes

* **chart:** add agent startupProbe and sync eks values ([c94dcda](https://github.com/kube-rca/helm-charts/commit/c94dcdaab9a1dad843ca0f1d8510a3227db80dcd))
* **chart:** add anthropic max token env ([a5c8090](https://github.com/kube-rca/helm-charts/commit/a5c8090b2192a9e6e8d9987e65634bb97a58596f))
* **kube-rca:** harden chart security, add probes, fix secret bug ([03c8560](https://github.com/kube-rca/helm-charts/commit/03c8560fc4ffa2e0fa40619a24ca7ca3c3e08dfb))
* **kube-rca:** increase postgresql CPU limits to resolve throttling ([147216d](https://github.com/kube-rca/helm-charts/commit/147216d267ba33b3128a2abb7850adb697955e5f))
* **kube-rca:** remove container securityContext defaults ([cf439c6](https://github.com/kube-rca/helm-charts/commit/cf439c6dc0437c6f74e29ee5a3db79bc6c2b9cdc))
* **kube-rca:** remove runAsNonRoot from default podSecurityContext ([4e8dc20](https://github.com/kube-rca/helm-charts/commit/4e8dc20133e4ccdc6d06387d6cb20c8369c6fddb))
* publish chart and image paths under kube-rca-ecr ([128aaa6](https://github.com/kube-rca/helm-charts/commit/128aaa6c043e8caa2a50fbdcac62e4ae754f0293))

## [0.5.0](https://github.com/kube-rca/helm-charts/compare/0.4.0...0.5.0) (2026-02-09)


### Features

* **kube-rca:** add agent masking regex helm settings (KAN-146) ([#8](https://github.com/kube-rca/helm-charts/issues/8)) ([3fec5b5](https://github.com/kube-rca/helm-charts/commit/3fec5b52ccc6c276cb5329fe46f9ff399e633a8e))
* **observability:** add tempo stack and wire agent tempo settings (KAN-73) ([#7](https://github.com/kube-rca/helm-charts/issues/7)) ([3a1642f](https://github.com/kube-rca/helm-charts/commit/3a1642f474bf2c72faf99fcad2efccc332acfc13))


### Bug Fixes

* **helm:** harden tempo defaults for timeout and load (KAN-145) ([3459fde](https://github.com/kube-rca/helm-charts/commit/3459fde4c0a1560f022682f6cf97f6eaff06e260))

## [0.4.0](https://github.com/kube-rca/helm-charts/compare/0.3.0...0.4.0) (2026-02-05)


### Features

* **kube-rca:** move postgresql to dependency and update configuration (KAN-110) ([c8229e2](https://github.com/kube-rca/helm-charts/commit/c8229e2396c86e5891f54514dbb030cc39bc59d9))


### Bug Fixes

* **helm:** align postgres host and hooks (KAN-110) ([a71a963](https://github.com/kube-rca/helm-charts/commit/a71a963b43a22b22a6014f276a39d02fd3961b99))
* **helm:** run wait-for-db after postgres (KAN-110) ([4de3f14](https://github.com/kube-rca/helm-charts/commit/4de3f143cde7664914eaa547c09fd090a6eb666d))

## [0.3.0](https://github.com/kube-rca/helm-charts/compare/0.2.0...0.3.0) (2026-02-01)


### Features

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


### Bug Fixes

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

## [0.2.0](https://github.com/kube-rca/helm-charts/compare/kube-rca-v0.1.0...kube-rca-v0.2.0) (2026-02-01)


### Features

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


### Bug Fixes

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
