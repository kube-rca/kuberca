# Changelog

## [1.2.0](https://github.com/kube-rca/agent/compare/1.1.0...1.2.0) (2026-04-02)


### Features

* add built-in redaction masking-chaining [KAN-174] ([fb7777e](https://github.com/kube-rca/agent/commit/fb7777e50ed7c7d768e2eefff511e86c26bfdb72))
* add settings option ([241e12c](https://github.com/kube-rca/agent/commit/241e12c4ba80449bd9a86e4fb57fadf6ef44799b))
* add total timeout stop condition to LLM retry logic ([41ab472](https://github.com/kube-rca/agent/commit/41ab472e05dc6632505c60b8b65c85db15b513ad))
* **agent:** add Loki log aggregation client and Strands tools ([fc4788b](https://github.com/kube-rca/agent/commit/fc4788b8e0716b36278f596699976e9eb7824f90))
* **agent:** add tool timing logs for analysis ([6777f6e](https://github.com/kube-rca/agent/commit/6777f6ee854a2f72a5c9857a41b9b73290aa225a))
* **agent:** async 엔드포인트 전환으로 분석 병렬 처리 지원 ([97d1b0d](https://github.com/kube-rca/agent/commit/97d1b0dbd4a8e2363c88d3a4701604782ae2c35f))
* **agent:** portable K8s baseline context and capability-aware prompt ([8f111d0](https://github.com/kube-rca/agent/commit/8f111d0ec76b5b65050c2a87fd3d277632e7e1d5))
* **agent:** tune LLM retry to 3-min exponential backoff ([89dc3d2](https://github.com/kube-rca/agent/commit/89dc3d214f8b02367c791899a447db03f069bdb1))
* **analyze:** differentiate resolved alert analysis from firing ([da6ed90](https://github.com/kube-rca/agent/commit/da6ed904a9c157f8c2979385e24b0610c683af78))


### Bug Fixes

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


### Documentation

* **agent:** note manual resolve as analysis trigger source ([d915336](https://github.com/kube-rca/agent/commit/d915336ef2de1f977458d66cb2300e5f7f35221d))

## [1.1.0](https://github.com/kube-rca/agent/compare/1.0.0...1.1.0) (2026-02-22)


### Features

* add tenacity-based LLM API retry for transient errors (KAN-159) ([3ed6790](https://github.com/kube-rca/agent/commit/3ed6790129266a7c084e59bc11d272610c14a78a))
* **agent:** add generic manifest read tools ([ebe49e7](https://github.com/kube-rca/agent/commit/ebe49e724aa19a4e1a55269de33cdb146078077b))
* **agent:** add regex masking for llm and storage flows (KAN-146) ([#10](https://github.com/kube-rca/agent/issues/10)) ([e1aad94](https://github.com/kube-rca/agent/commit/e1aad94c94db2882d1805ec2b34cb1e4865c2e37))
* **agent:** enrich daemonset and events context (KAN-79) ([8c678bc](https://github.com/kube-rca/agent/commit/8c678bc51bc04059abe6a1b9c3e7a7b18eac7c29))
* **agent:** enrich daemonset and events context (KAN-79) ([ace5c00](https://github.com/kube-rca/agent/commit/ace5c003ac9c4877256d6088a69e39e66ecff4ea))
* **agent:** integrate tempo trace context into analysis (KAN-145) ([#9](https://github.com/kube-rca/agent/issues/9)) ([e17cd5e](https://github.com/kube-rca/agent/commit/e17cd5efe834e0680393a8fbf62dd0f2bdf47365))
* 채팅 기능 추가 (KAN-158) ([4fd3051](https://github.com/kube-rca/agent/commit/4fd3051219d31d5d1e07a8cb5e4f62497e975d4d))


### Bug Fixes

* **agent:** align tempo timeout default and docs (KAN-145) ([a19bd6e](https://github.com/kube-rca/agent/commit/a19bd6eb34134133264a7af23e9958885f112b09))
* build error ([4751f41](https://github.com/kube-rca/agent/commit/4751f4104f96f6699f322e7576029f7a55f6efd7))
* bump cryptography to 46.0.5 for CVE-2026-26007 ([9036346](https://github.com/kube-rca/agent/commit/90363461a6fba2ab38af8a6d00af21ce4c23c6d0))
* Gemini 400 function-call turn order error in multi-turn chat ([4d87b9a](https://github.com/kube-rca/agent/commit/4d87b9a18979949b9be8500b0d2bb83bc36de9d9))
* recover chat session state and standardize english chat errors ([2d08669](https://github.com/kube-rca/agent/commit/2d08669128de6d092e720be09a07045e696c1b43))
* route agent image pushes to component repository ([35c6b84](https://github.com/kube-rca/agent/commit/35c6b84f922962fedcf4d3bad8227bb1ddb41f35))
* **tempo:** normalize time window and surface query failures ([f1e4cf8](https://github.com/kube-rca/agent/commit/f1e4cf8753593665c0a1511fb8bde2ac02e9ccfd))


### Documentation

* add LLM retry environment variables to README (KAN-159) ([ad4faa5](https://github.com/kube-rca/agent/commit/ad4faa5846add0aaff4ed2ab7de4c66e8ac84364))
* add release-please merge policy ([#8](https://github.com/kube-rca/agent/issues/8)) ([52c8e91](https://github.com/kube-rca/agent/commit/52c8e91193d3df3d96a1ec311365ffeb63bb8a95))
* **agent:** align README with current API schema ([b3755af](https://github.com/kube-rca/agent/commit/b3755afa50d9b3ca5e5fef286c84689b060810d4))
* **agent:** update README for generic manifest tools ([974a17f](https://github.com/kube-rca/agent/commit/974a17f752cb89c18d7da8308d883042ea467eb5))
* **license:** align with MIT ([14719ed](https://github.com/kube-rca/agent/commit/14719ed1eb9010b7759b145952680e9cb03a3728))

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
