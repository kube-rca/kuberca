# Changelog

## [1.0.0](https://github.com/kube-rca/agent/compare/0.1.0...1.0.0) (2026-02-01)


### ⚠ BREAKING CHANGES

* **agent:** rename analyze endpoint to /analyze [build]

### Features

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


### Bug Fixes

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


### Documentation

* **agent:** add README and response schema ([cffe4ed](https://github.com/kube-rca/agent/commit/cffe4ede61f70169ac882c550ee7da621e175055))
* **agent:** translate README to English ([2e7f8f0](https://github.com/kube-rca/agent/commit/2e7f8f0fb52e01dc5f4311fcae42c32834ad967c))
* enhance README for open source project ([3fbd4da](https://github.com/kube-rca/agent/commit/3fbd4dad511896709075a9749ab5584ba95a5bfe))
