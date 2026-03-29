# Changelog

## [0.4.0](https://github.com/kube-rca/frontend/compare/0.3.0...0.4.0) (2026-03-29)


### Features

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


### Bug Fixes

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

## [0.3.0](https://github.com/kube-rca/frontend/compare/0.2.0...0.3.0) (2026-02-27)


### Features

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


### Bug Fixes

* Comment 메뉴바 표시와 일부 단축어 오류 수정 (KAN-153) ([1364ff5](https://github.com/kube-rca/frontend/commit/1364ff5573aaeb2fcb910f172a535382b2bfc582))
* Comment 편집 및 삭제 오류 수정 (KAN-153) ([8efe3bc](https://github.com/kube-rca/frontend/commit/8efe3bc942856b95ea3dca393c6863a7083c9658))
* daymode 글자안보이던 현상 해결 (KAN-96) ([307ec02](https://github.com/kube-rca/frontend/commit/307ec0268f3a3f441f17d26b159de4f2a59dbe6c))
* import unimported modules ([871f192](https://github.com/kube-rca/frontend/commit/871f1922dba8d06233cfc7857a6c98f8fcc47a16))
* import unimported modules ([60151b7](https://github.com/kube-rca/frontend/commit/60151b79d97d0502c475917d9a6b1cd202ef8940))
* improve OIDC error display with prominent banner and URL cleanup ([a0f1b9b](https://github.com/kube-rca/frontend/commit/a0f1b9bf3cd50d0575cbb6dd2c44f842f3141e65))
* route frontend image pushes to component repository ([5741190](https://github.com/kube-rca/frontend/commit/5741190bae0cbdb5d3e394e61e9f487e7487022f))
* 자동 메세지 전송처리 삭제 ([040ce56](https://github.com/kube-rca/frontend/commit/040ce566dd7440e6f259a1b4e4c453432bc1d725))

## [0.2.0](https://github.com/kube-rca/frontend/compare/0.1.0...0.2.0) (2026-02-01)


### Features

* add Dockerfile ([0e2e755](https://github.com/kube-rca/frontend/commit/0e2e75578b4ffdccaaa8d8bb412c4f0f21812fa1))
* **auth:** add auth panel and token refresh [KAN-88] [build] ([4b46068](https://github.com/kube-rca/frontend/commit/4b4606801a3f4dd09dd83319ebbdfd8aa68bcfac))
* **auth:** add auth panel and token refresh [KAN-88] [build] ([598638d](https://github.com/kube-rca/frontend/commit/598638d7fa27514b0c94d1265aa126821597d2f2))
* Create ci.yml ([f2dbf7b](https://github.com/kube-rca/frontend/commit/f2dbf7b1550983849c4d6ca5a6a18311e4daab5d))
* implement alerts dashboard with filtering and pagination ([52f3eed](https://github.com/kube-rca/frontend/commit/52f3eed2609946b26b0e27e446c49a65c134d21c))
* RCA 편집 기능 구현 [KAN-56] ([416aff0](https://github.com/kube-rca/frontend/commit/416aff0af53b4a0ea5acd9b9a8d53fb5c97b758b))
* 다크모드 구현 [KAN-89] [build] ([1171e62](https://github.com/kube-rca/frontend/commit/1171e62daa5de727644af2db111f53800d650253))
* 특정 RCA 보고서 조회 [KAN-55] ([b219f5e](https://github.com/kube-rca/frontend/commit/b219f5eb0301d7560d7addd16b38c18feed729fe))


### Bug Fixes

* **auth:** improve dark mode input contrast [KAN-89] [build] ([77025c2](https://github.com/kube-rca/frontend/commit/77025c294c9891a478b528b597d23ce5340149bc))
* **docs:** use absolute GitHub URLs for related components ([1beda23](https://github.com/kube-rca/frontend/commit/1beda23f926df3334ab2ac35cd6b0c1a7dfb4d0b))
* **docs:** use absolute URL for logo image ([303c66b](https://github.com/kube-rca/frontend/commit/303c66b000bf49140dd9f9e8fa2ef150b26e6951))
* resolve npm audit vulnerabilities ([94fae8b](https://github.com/kube-rca/frontend/commit/94fae8bff91d8cc3ad8775d7e4528b6f41f30f1b))
* Update README.md [build] ([8db3788](https://github.com/kube-rca/frontend/commit/8db37884d449d60937c479d0832b327c676392a0))
