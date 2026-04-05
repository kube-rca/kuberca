# Changelog

## [0.8.0](https://github.com/kube-rca/kuberca/compare/helm-chart-0.7.0...helm-chart-0.8.0) (2026-04-05)


### Features

* add built-in Redaction option [KAN-174] ([b1dfa5f](https://github.com/kube-rca/kuberca/commit/b1dfa5fdc1f60b5f9fdc05e1c0efe2de2084ce31))
* add cert manager and move ingress hosts to 2 domains ([ed8959c](https://github.com/kube-rca/kuberca/commit/ed8959cda3192fd944c2cd2ab60715ae8d432455))
* add gp2 values for stateful apps ([f34c8a5](https://github.com/kube-rca/kuberca/commit/f34c8a5175aea0c7a119c5d1154d2dd4f9acad43))
* alert flapping 환경변수 (KAN-155) ([e595663](https://github.com/kube-rca/kuberca/commit/e595663fef695c4d41f76f0583278b2a1d8fc451))
* flapping enable/disable (KAN-155) ([b80c570](https://github.com/kube-rca/kuberca/commit/b80c570cd32da23776a2e73d0c1f2f1d2f9ee3c1))
* **helm:** add inline apiKey value fields to values.yaml ([07ee269](https://github.com/kube-rca/kuberca/commit/07ee269972bd5529b367615af3dd5d85805dbbb1))
* **helm:** add postgresql and embedding secret name auto-resolve helpers ([424c2ac](https://github.com/kube-rca/kuberca/commit/424c2ac20651334610d127768bf78fe1134dc591))
* **helm:** populate agent secrets with inline apiKey values ([277c1b2](https://github.com/kube-rca/kuberca/commit/277c1b273d6f1bc093ed1d40b2a924fea9287e96))
* **helm:** use auto-resolve helper for agent session DB secret ([21d8e95](https://github.com/kube-rca/kuberca/commit/21d8e958c6ca325f821d1a802b3392826850564a))
* **helm:** use auto-resolve helpers for postgresql and embedding secrets ([ca9cb70](https://github.com/kube-rca/kuberca/commit/ca9cb702edc60b918f2c3dfa0f049d3e102f741d))
* **kube-rca:** add agent masking regex helm settings (KAN-146) ([#8](https://github.com/kube-rca/kuberca/issues/8)) ([9a009a3](https://github.com/kube-rca/kuberca/commit/9a009a35fd47f106c7f9c664509d727a0f2d9ec4))
* **kube-rca:** add LLM retry configuration for agent (KAN-159) ([5653ef2](https://github.com/kube-rca/kuberca/commit/5653ef2da593bb5a3a3173eb53bb55e8ceb50dc6))
* **kube-rca:** add Loki env vars for agent deployment ([cfe9361](https://github.com/kube-rca/kuberca/commit/cfe9361147a70da379620935ae963bf5f92f7f91))
* **kube-rca:** add MANUAL_ANALYZE_SEVERITIES env for analysis mode (KAN-172) ([e8ff427](https://github.com/kube-rca/kuberca/commit/e8ff427762886d6e92d0961455b83206788eaa15))
* **kube-rca:** add OIDC authentication helm values and deployment env ([40311b4](https://github.com/kube-rca/kuberca/commit/40311b458c990acba6da54aafba0275adaec5389))
* **kube-rca:** move postgresql to dependency and update configuration (KAN-110) ([d067e33](https://github.com/kube-rca/kuberca/commit/d067e33e044d73982b182db573ec3e74a43c8028))
* **kube-rca:** replace hook job with initContainer and add DB-aware probes ([d5d191f](https://github.com/kube-rca/kuberca/commit/d5d191f5c473c2fd9d04198a12b45fea5fe32936))
* **observability:** add tempo stack and wire agent tempo settings (KAN-73) ([#7](https://github.com/kube-rca/kuberca/issues/7)) ([c540029](https://github.com/kube-rca/kuberca/commit/c540029b10ff336e6d66588db974cf492609bc40))


### Bug Fixes

* **chart:** add agent startupProbe and sync eks values ([b2d5a41](https://github.com/kube-rca/kuberca/commit/b2d5a41d1c5c00bd7952eb6856abd4ebe909f355))
* **chart:** add anthropic max token env ([77310a8](https://github.com/kube-rca/kuberca/commit/77310a8217c56038ebd056df55fb8bf773d9cccd))
* **helm:** align postgres host and hooks (KAN-110) ([8bf326b](https://github.com/kube-rca/kuberca/commit/8bf326b9e63df7fda42054019903619370db0a30))
* **helm:** harden tempo defaults for timeout and load (KAN-145) ([c1d6f3b](https://github.com/kube-rca/kuberca/commit/c1d6f3b3862e1fcd85833c20008e59f0e57de482))
* **helm:** run wait-for-db after postgres (KAN-110) ([481212a](https://github.com/kube-rca/kuberca/commit/481212ad691d7f032450ba2e1cc2fb81a4338078))
* **kube-rca:** harden chart security, add probes, fix secret bug ([e15824c](https://github.com/kube-rca/kuberca/commit/e15824c2f853ca083c288f8d80ec1ac5d1275b45))
* **kube-rca:** increase postgresql CPU limits to resolve throttling ([d530d91](https://github.com/kube-rca/kuberca/commit/d530d9133c7b9d273a3290a39a6291ca704113aa))
* **kube-rca:** remove container securityContext defaults ([9884eda](https://github.com/kube-rca/kuberca/commit/9884edaf777691b2837978268d36ad9176682887))
* **kube-rca:** remove runAsNonRoot from default podSecurityContext ([86b3ef6](https://github.com/kube-rca/kuberca/commit/86b3ef64c282303928baa2df984d2c7718007233))
* publish chart and image paths under kube-rca-ecr ([31130cf](https://github.com/kube-rca/kuberca/commit/31130cf94c72749982cbdc8641c44a779ddfb2ea))


### Performance Improvements

* **kube-rca:** increase agent replicaCount default to 2 ([b2c6994](https://github.com/kube-rca/kuberca/commit/b2c69946dc903794bcfa55df31cc607114052a4e))
