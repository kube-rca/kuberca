# Changelog

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
