# Changelog

All notable changes to the UteamUP CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.92.0](https://github.com/UteamUP/cli/compare/1.91.0...1.92.0) (2026-07-24)


### Features

* **marketplace:** add the seller browse filters to the CLI registry ([cd2f89c](https://github.com/UteamUP/cli/commit/cd2f89c8abfe5620b24a831f90cd11a765c7101c))
* **registry:** add workforce capacity readiness + scenario CLI domains ([fc8769e](https://github.com/UteamUP/cli/commit/fc8769e9391e630bd536ae26680d3526faf0ebc0))
* **registry:** expose SLA milestone search and overlap acknowledgement ([0a9d626](https://github.com/UteamUP/cli/commit/0a9d6266d925c8fdbb66ecf52143050ac334ded8))
* **schedule:** add policy restore and the archive listing flag ([5ab1165](https://github.com/UteamUP/cli/commit/5ab11656c25f877fb797325297f7eecb9ac4f2c0))
* **service-price-list:** expose the lifecycle and rule preview from the CLI ([2ca3df4](https://github.com/UteamUP/cli/commit/2ca3df4704c7394dcd32fda95b4ea372b3eef6be))
* **service-price-list:** expose the replacement route as a CLI action ([e99274f](https://github.com/UteamUP/cli/commit/e99274f7e95a27bd9eb004d5303fa7f02624bcf9))


### Bug Fixes

* **schedule:** send the optimization policy frequency as a string ([a5079bd](https://github.com/UteamUP/cli/commit/a5079bd3afa0bd65f6913226e44f235794cb650b))


### Documentation

* **operations-brain:** mark site-guids as not yet supported ([5de05a8](https://github.com/UteamUP/cli/commit/5de05a8a39dc794021bb29ce6bf02f5c00e6db79))

## [1.91.0](https://github.com/UteamUP/cli/compare/1.90.0...1.91.0) (2026-07-23)


### Features

* **cli:** expose run recollection and the new invoice evidence flags ([ec4c5a2](https://github.com/UteamUP/cli/commit/ec4c5a24a83206d9a1d08479f965eca1113d0bc1))
* **inventory:** expose the scope-enforced bulk delete on the CLI ([d410622](https://github.com/UteamUP/cli/commit/d41062260372576ad7d49c1da50729cc11021493))
* **upmate:** portal-request-classify-cost action for authoritative cost disclosure ([4b6c48e](https://github.com/UteamUP/cli/commit/4b6c48ee621a431b2d6e2ddfb38d75a8819b8a43))


### Bug Fixes

* avoid attachment output flag collision ([e8a2b40](https://github.com/UteamUP/cli/commit/e8a2b40426c177b454c47366468e1cddde8b51e1))
* compose REST query strings safely ([d1cdfc4](https://github.com/UteamUP/cli/commit/d1cdfc430ee7a11aa88f9e79c6216d6443572324))
* honor runtime connection flags on Windows ([d57045d](https://github.com/UteamUP/cli/commit/d57045dddd56bfb92f9a7d0b2bb7e4b9e4c56c56))
* **registry:** correct supplier-invoice match-status docs ([b5b7f90](https://github.com/UteamUP/cli/commit/b5b7f9074af8674732d3be418e1396934203a896))
* route workorder commands through GUID endpoints ([f523a03](https://github.com/UteamUP/cli/commit/f523a03f805972032ce10546756f6cdef0232260))
* send attachment upload media types ([03fd848](https://github.com/UteamUP/cli/commit/03fd848f42011c9c271aa4a3ed8dea52b37c9858))
* send workorder searches as query parameters ([de1f9b2](https://github.com/UteamUP/cli/commit/de1f9b2d5d9acd445845214e84030b838048d001))
* stream bug attachments to Windows files ([daf9c12](https://github.com/UteamUP/cli/commit/daf9c12687b5e47a01ed98574f25270ea346eb9d))


### Tests

* **billing:** cover the run recollect action and invoice evidence flags ([c5f35cc](https://github.com/UteamUP/cli/commit/c5f35cc3406f489264938428ccfc097e706680be))

## [1.90.0](https://github.com/UteamUP/cli/compare/1.89.0...1.90.0) (2026-07-20)


### Features

* **cli:** add governed IoT command workflows ([cd9b758](https://github.com/UteamUP/cli/commit/cd9b7585afe1892dfd71388ab0398b79b8eff145))
* **cli:** add UPMate field service tools ([574ab97](https://github.com/UteamUP/cli/commit/574ab975ccc80594bd6df0454b887ee423365514))

## [1.89.0](https://github.com/UteamUP/cli/compare/1.88.0...1.89.0) (2026-07-19)


### Features

* add retry-safe field service CLI actions ([559f4c3](https://github.com/UteamUP/cli/commit/559f4c3d0518952ddc9d1db2023d70fc317fc548))
* **cli:** add approved active template read ([93bbb1f](https://github.com/UteamUP/cli/commit/93bbb1f1e487933c6b7c61befc8807774b7bcad5))
* **cli:** add return case lifecycle commands ([ef10684](https://github.com/UteamUP/cli/commit/ef10684e106a2b5affd66a065ffa26230b160ac5))
* **cli:** manage bookable scheduling resources ([e0fea4c](https://github.com/UteamUP/cli/commit/e0fea4c18339b81b2210d1508e541baa37f89b91))
* **cli:** mirror approved stock reads ([f68f307](https://github.com/UteamUP/cli/commit/f68f3079b6281f2307cb7a84c79dfce8549f7373))
* **codecatalog:** update code-catalog registry domain ([e55ef82](https://github.com/UteamUP/cli/commit/e55ef826da9c7d4d476fcf3afe31f1da589b51c3))
* **fleet:** add intelligence CLI domain ([636d42e](https://github.com/UteamUP/cli/commit/636d42e998e2172d8eca8b4667a948cd8fc11b90))
* **operations:** add governed planning commands ([a5b593a](https://github.com/UteamUP/cli/commit/a5b593a80e0f50d3d2f5b0b68acdcef39036e2d7))
* **procurement:** add supplier invoice and report CLI domains ([2df96a8](https://github.com/UteamUP/cli/commit/2df96a8d82bd5ac295e98d9c0bb7784662e6c6d8))
* **registry:** add duration-kind and currency flags to promotion actions ([7935a9f](https://github.com/UteamUP/cli/commit/7935a9fc22e3fcef8271685c0d8cc4d128b55d39))
* **registry:** attendance-station domain, timesheet actions, planning + oncall parity ([9ee2b40](https://github.com/UteamUP/cli/commit/9ee2b4048b9f3cb65882c5ea41219e69627de184))
* **reliability:** add evidence CLI commands ([d83db37](https://github.com/UteamUP/cli/commit/d83db37ee2c1ed637e48cc1ec9739de951ea379f))
* **schedule:** add publication and capacity commands ([27fb314](https://github.com/UteamUP/cli/commit/27fb3143f917a72dd484bf23f660f0adc09b1c47))
* **schedule:** add supporting resource commands ([9a3697e](https://github.com/UteamUP/cli/commit/9a3697e3400a9c0cd68d129a300e20ecd6adc197))


### Bug Fixes

* **auth:** fall back to visible input when stdin is not a console ([db993aa](https://github.com/UteamUP/cli/commit/db993aad2f46df9389999f2b52ea137a8ce194c5))
* **checkpoint:** correct process liveness check on Windows ([0e57dd7](https://github.com/UteamUP/cli/commit/0e57dd7f0bd29a87cc28a0162cca6c49941c78b2))
* **cli:** route approved report analytics ([8f6e9fa](https://github.com/UteamUP/cli/commit/8f6e9fa12d2142eef04c68037c001677b9ba9375))


### Documentation

* **cli:** record the Windows build and install path ([ea68869](https://github.com/UteamUP/cli/commit/ea68869505059b8f7766068b1c0e432a1b50d293))


### Tests

* **cli:** isolate Windows filesystem contracts ([1693c58](https://github.com/UteamUP/cli/commit/1693c58e372123befc842045a2c78262cb1b00c3))

## [1.88.0](https://github.com/UteamUP/cli/compare/1.87.0...1.88.0) (2026-07-17)


### Features

* **cli:** expose procurement intelligence evidence ([54137d2](https://github.com/UteamUP/cli/commit/54137d28cc5b3d93cf9e9ea498ae88d93ac0350e))
* **cli:** read project field context ([223e986](https://github.com/UteamUP/cli/commit/223e986caf0fc6799002922fe00ecb93ee2394cb))

## [1.87.0](https://github.com/UteamUP/cli/compare/1.86.0...1.87.0) (2026-07-17)


### Features

* **cli:** manage accounting exports ([b2ea77b](https://github.com/UteamUP/cli/commit/b2ea77b70c2237571e43c3edba4ef89ab5945892))
* **stock:** add optimization run preparation command ([d77720b](https://github.com/UteamUP/cli/commit/d77720bb01384dee471b192e09a88eb905a44022))


### Miscellaneous

* **cli:** migrate lint config and clear quality debt ([e06a9a4](https://github.com/UteamUP/cli/commit/e06a9a49a48cf3e6c866ab97b8c87eeb5f852a91))

## [1.86.0](https://github.com/UteamUP/cli/compare/1.85.1...1.86.0) (2026-07-17)


### Features

* **cli:** manage location access schedules ([30beec9](https://github.com/UteamUP/cli/commit/30beec97f966e80fda438297ea26df9560c2cdde))


### Bug Fixes

* **cli:** validate schedule team GUID ([7c63ef2](https://github.com/UteamUP/cli/commit/7c63ef2c043d9cc658e77ad2afb5162ec53518bc))

## [1.85.1](https://github.com/UteamUP/cli/compare/1.85.0...1.85.1) (2026-07-17)


### Bug Fixes

* **cli:** align kaizen and suggestion GUID routes ([580322d](https://github.com/UteamUP/cli/commit/580322ddb25e0ebd562b8bb50db1a732756d4142))
* **cli:** expose RCA GUID workflows ([a36b043](https://github.com/UteamUP/cli/commit/a36b043020c4902d651912b798a6ad25e36da2ec))
* **meters:** expose GUID asset filter on compliance summary ([118ca98](https://github.com/UteamUP/cli/commit/118ca98a8b9ce769a7ae51c25c34db8a42592d68))


### Documentation

* **meters:** clarify GUID-only schedule routes ([b048467](https://github.com/UteamUP/cli/commit/b0484672dac40cb8b345975096561e418a0898a4))

## [1.85.0](https://github.com/UteamUP/cli/compare/1.84.0...1.85.0) (2026-07-17)


### Features

* **asset-rental:** expose active rental reads ([61b90db](https://github.com/UteamUP/cli/commit/61b90db61aede6f081ed35071e72dfba469e7501))
* **asset-rental:** expose revenue reads ([8070851](https://github.com/UteamUP/cli/commit/8070851c0226fde2641feed7e44742ed7dd36934))
* **fleet:** add dashboard read parity ([b02a6da](https://github.com/UteamUP/cli/commit/b02a6daed5e130281caff4613ea8745e37c19e0d))
* **fleet:** expose corrective inspection confirmation ([11e7507](https://github.com/UteamUP/cli/commit/11e75071f8daaebd54cfb8f5fcfcaf85856263ed))
* **fleet:** expose guid driver assignment commands ([efc5518](https://github.com/UteamUP/cli/commit/efc5518f5d4e9c6592060b8ac91beff14f412b8d))
* **fleet:** expose guid driver commands ([f021913](https://github.com/UteamUP/cli/commit/f021913c7fc2dcacb8fcfe08df071ef0080be970))
* **fleet:** expose guid inspection commands ([5deb267](https://github.com/UteamUP/cli/commit/5deb2678f185b362e23a98ff2b3d746033915974))
* **fleet:** expose route optimization evidence ([8e5399e](https://github.com/UteamUP/cli/commit/8e5399ecedc41ec4406acae7739bbb5256cb873a))
* **fleet:** make fuel commands guid-first ([8848123](https://github.com/UteamUP/cli/commit/88481230204add0e7f583580272d68118a1dcd6c))
* **reliability:** expose risk intelligence in cli ([e7cab13](https://github.com/UteamUP/cli/commit/e7cab13d3c070a6433d3e59ced5d51014e25c946))
* **reliability:** expose strategy proposals in cli ([80b986c](https://github.com/UteamUP/cli/commit/80b986ce232dcb3e7f143de59423955b34753558))
* **service-billing:** expose operational billing commands ([c093185](https://github.com/UteamUP/cli/commit/c0931851b5094655843fef47c1743902afd7b680))
* **vendors:** add guid-first performance commands ([8d35f70](https://github.com/UteamUP/cli/commit/8d35f70313318fc2958bbd92914d47afdd7af102))


### Bug Fixes

* **cli:** expose condition assessment GUID flags ([5bda29c](https://github.com/UteamUP/cli/commit/5bda29cfae0995c0594a31ec5bd42cf4f6992f73))
* **cli:** expose criticality GUID flags ([9e230f0](https://github.com/UteamUP/cli/commit/9e230f0baafbe75df19bc8beaf141c39547e9a3a))
* **cli:** identify vendor performance actions ([17c77f0](https://github.com/UteamUP/cli/commit/17c77f07909eebb312235ff45f71c7eab4ae836a))
* **cli:** make improvement workflows GUID-first ([c9de6e6](https://github.com/UteamUP/cli/commit/c9de6e634eec7abbb902ddba62782e8c2978d9ab))
* **failures:** make CLI actions GUID-first ([b9a9a0a](https://github.com/UteamUP/cli/commit/b9a9a0a8f2c391b9df0f3d5714e4b76c01cd4257))
* **fleet:** repair registry contract tests ([0595797](https://github.com/UteamUP/cli/commit/0595797af63ab047e36b019d259928dd84bd0787))
* **meters:** use GUID workorder reading flags ([c82ed41](https://github.com/UteamUP/cli/commit/c82ed4174bb423ca21926339813f88056c5255c5))
* **route:** expose guid operational actions ([c111d99](https://github.com/UteamUP/cli/commit/c111d99eb68d9e173d29568a5f7da7062d89862c))

## [1.84.0](https://github.com/UteamUP/cli/compare/1.83.0...1.84.0) (2026-07-17)


### Features

* **asset-rental:** expose available reads ([11de7dc](https://github.com/UteamUP/cli/commit/11de7dc3f5f8bedc3a9237f0e65eaa0558ab5d2a))

## [1.83.0](https://github.com/UteamUP/cli/compare/1.82.0...1.83.0) (2026-07-17)


### Features

* **asset-failure:** expose filtered reads by severity ([b4a870e](https://github.com/UteamUP/cli/commit/b4a870e3c0ffdb08537d990c9c739fa7ab6d91c3))

## [1.82.0](https://github.com/UteamUP/cli/compare/1.81.0...1.82.0) (2026-07-17)


### Features

* **asset-calibration:** expose due-soon reads ([c1b06e8](https://github.com/UteamUP/cli/commit/c1b06e868fd6f4195b6fd13f5b5746ecaeadecb6))
* **asset-calibration:** expose overdue reads ([b18c93a](https://github.com/UteamUP/cli/commit/b18c93aade7b7f8407214372afbe293f8b59756e))
* **asset-certification:** expose status reads ([49f723a](https://github.com/UteamUP/cli/commit/49f723a0965f1316e3a0a5b6a75ec74247458317))
* **asset-failure:** expose open failure reads ([01fd21c](https://github.com/UteamUP/cli/commit/01fd21cfba9382a9c6b3a5d42c7e74db3b785547))
* **asset-lifecycle:** expose filtered reads ([4e8cc7c](https://github.com/UteamUP/cli/commit/4e8cc7c3473b0e09c657268c2012fac4c5c2dd97))

## [1.81.0](https://github.com/UteamUP/cli/compare/1.80.1...1.81.0) (2026-07-17)


### Features

* add service SLA milestone commands ([6c48f75](https://github.com/UteamUP/cli/commit/6c48f75a0ac6038dcfc40a88d95a8393f1fae0c5))
* **fleet:** expose overdue inspection reads ([33c8642](https://github.com/UteamUP/cli/commit/33c86429d62188c0d2543f4be96b2cbad8d15221))
* **schedule:** expose personal schedule reads ([220fa69](https://github.com/UteamUP/cli/commit/220fa69714f0a72cd6c4c026de61a18a99074edf))

## [1.80.1](https://github.com/UteamUP/cli/compare/1.80.0...1.80.1) (2026-07-16)


### Bug Fixes

* require emergency insertion idempotency ([938fec6](https://github.com/UteamUP/cli/commit/938fec6dab3ef4126d7b98e4113834f02d74e8a6))

## [1.80.0](https://github.com/UteamUP/cli/compare/1.79.0...1.80.0) (2026-07-16)


### Features

* **agreements:** add service price list CLI domain ([3da7413](https://github.com/UteamUP/cli/commit/3da74138e893f05f95bc104a59cd115d8a40af98))


### Bug Fixes

* **fleet:** mirror GUID-first contact tools in CLI ([78d5080](https://github.com/UteamUP/cli/commit/78d5080666c5e31e591a0a9a3a1f4f17118b3d97))

## [1.79.0](https://github.com/UteamUP/cli/compare/1.78.0...1.79.0) (2026-07-16)


### Features

* **agreements:** add entitlement CLI domain ([187cf74](https://github.com/UteamUP/cli/commit/187cf749e3a0503bd480e72f7036495d3c7d43b1))
* **agreements:** add service agreement CLI domain ([5b9962a](https://github.com/UteamUP/cli/commit/5b9962a3dc279d29b47d7681d1f4e0f6e38b89c4))
* **ai:** expose measured outcomes in CLI ([c49e9e0](https://github.com/UteamUP/cli/commit/c49e9e08b448052d0ddc57d281c893d74a6e2f9c))
* **stock:** expose replenishment proposal command ([225f11e](https://github.com/UteamUP/cli/commit/225f11ec2c04af4afadbb31ce367713a266fb529))


### Bug Fixes

* **fleet:** mirror GUID calendar tools in CLI ([a5a173e](https://github.com/UteamUP/cli/commit/a5a173e4c3e22e450a9303b287594b14156c7958))
* keep AI provider routing server-owned ([c936d49](https://github.com/UteamUP/cli/commit/c936d495fd5ad6ce804b77e5d979abd3759ea142))
* **security:** harden CLI transport and retry jitter ([c89fa1e](https://github.com/UteamUP/cli/commit/c89fa1efbeec5bd50adb30cce4d5ec02b1121af1))


### CI/CD

* **security:** add dependency release cooldown ([568a25b](https://github.com/UteamUP/cli/commit/568a25b1a55182c5d67bb598eaf00824924e179a))

## [1.78.0](https://github.com/UteamUP/cli/compare/1.77.0...1.78.0) (2026-07-16)


### Features

* **ai:** expose data readiness in CLI ([614f1b3](https://github.com/UteamUP/cli/commit/614f1b3b53269ada0ac647472f984d8c6479067b))
* **ai:** expose governance snapshot in CLI ([16e8f35](https://github.com/UteamUP/cli/commit/16e8f359ea1ae9cb1a49df8b37480e3d0e8840f1))
* **ai:** expose Knowledge tutorial trust in CLI ([6858270](https://github.com/UteamUP/cli/commit/68582707f5f3fff3abcd5c1d1116dd79a84b3772))
* route CLI media analysis through tenant AI governance ([0fdaffc](https://github.com/UteamUP/cli/commit/0fdaffcb05e381e17b9586cc5f65a9eeb4e2c478))

## [1.77.0](https://github.com/UteamUP/cli/compare/1.76.0...1.77.0) (2026-07-16)


### Features

* **registry:** add maintenance plan templates and due-projection domain ([bd65c30](https://github.com/UteamUP/cli/commit/bd65c30fe70a84ddbad749b171dda82c1cae2bbd))

## [1.76.0](https://github.com/UteamUP/cli/compare/1.75.0...1.76.0) (2026-07-16)


### Features

* **journals:** expose CLI field note context ([fefd688](https://github.com/UteamUP/cli/commit/fefd68808444c1dd869957a27283c281c8004302))

## [1.75.0](https://github.com/UteamUP/cli/compare/1.74.0...1.75.0) (2026-07-16)


### Features

* add fleet maintenance proposal CLI command ([ca555e5](https://github.com/UteamUP/cli/commit/ca555e55b88c1cf9cd4b68d5a48ef80b0e3aaa15))
* **schedule:** add guid booking CLI actions ([eb99659](https://github.com/UteamUP/cli/commit/eb99659d47ef794f379d21562c2db253fe4be498))
* **schedule:** add optimization policy commands ([970b6d1](https://github.com/UteamUP/cli/commit/970b6d1009626e2dc37db2b8b5267a4bfbcd8938))
* **schedule:** list optimization run history ([acb7f5b](https://github.com/UteamUP/cli/commit/acb7f5b283405f486abab62e2b5d819f299fbef6))
* **schedule:** preview emergency insertions ([4885287](https://github.com/UteamUP/cli/commit/4885287201c9b045528b25f0b5edfb94f016565f))


### Bug Fixes

* route fleet maintenance proposals safely ([016c936](https://github.com/UteamUP/cli/commit/016c9362b3a10af089ac1cf61c28889460952b70))


### Documentation

* **fleet:** clarify maintenance package GUID ([ee2ad63](https://github.com/UteamUP/cli/commit/ee2ad63c003405329e260614a444f6e369e542ae))


### Tests

* **schedule:** restore shift GUID route coverage ([88114f9](https://github.com/UteamUP/cli/commit/88114f9822ecb33b6d6ccae8d67d080591286736))

## [1.74.0](https://github.com/UteamUP/cli/compare/1.73.0...1.74.0) (2026-07-15)


### Features

* **schedule:** add optimization apply and revert commands ([f75c199](https://github.com/UteamUP/cli/commit/f75c199b18aacc8eff55142a4bd765ea9ef97238))

## [1.73.0](https://github.com/UteamUP/cli/compare/1.72.0...1.73.0) (2026-07-15)


### Features

* **schedule:** add optimization run CLI domain ([1e7ac63](https://github.com/UteamUP/cli/commit/1e7ac63bafbf739da5034893b713e1fd008f70ec))

## [1.72.0](https://github.com/UteamUP/cli/compare/1.71.0...1.72.0) (2026-07-15)


### Features

* **registry:** add customer portal message and rating domains ([b3eb26d](https://github.com/UteamUP/cli/commit/b3eb26d4aaa01b0e348cbc4b8466db83cde1c26d))


### Bug Fixes

* **registry:** rename UteamupProjectGetBudget to UteamupProjectBudgetGet ([70ffbb1](https://github.com/UteamUP/cli/commit/70ffbb182d60896d2241648656f4716328ca6a1b))

## [1.71.0](https://github.com/UteamUP/cli/compare/1.70.0...1.71.0) (2026-07-15)


### Features

* add GUID portal work request commands ([44ac55a](https://github.com/UteamUP/cli/commit/44ac55a02cc4bcadfc6a6eb3ab501031b74e2d73))


### Tests

* **registry:** keep UteamupProjectGetHealth in UPMate parity list ([aca77bf](https://github.com/UteamUP/cli/commit/aca77bf3f1f0c92d234c9390a6c7bb813cc8fa1f))

## [1.70.0](https://github.com/UteamUP/cli/compare/1.69.0...1.70.0) (2026-07-15)


### Features

* **registry:** add customer-portal-user domain, drop legacy customer-portal ([414081b](https://github.com/UteamUP/cli/commit/414081b3946e48624a6c2d6f9f347f90c195f31b))


### Tests

* mirror UPMate IoT rules read ([a10e112](https://github.com/UteamUP/cli/commit/a10e1120a83e76566d16947b63983932214c50c5))
* **registry:** keep UteamupWorkorderTemplateAnalyzePreview in UPMate parity list ([78bf0ba](https://github.com/UteamUP/cli/commit/78bf0ba139ef34c1d7ab43e9b5a6c1646abebe18))

## [1.69.0](https://github.com/UteamUP/cli/compare/1.68.0...1.69.0) (2026-07-15)


### Features

* mirror enabled UPMate tools in CLI registry ([9bafbbb](https://github.com/UteamUP/cli/commit/9bafbbbcbace6398a9f2a7851b719b5de71dff6a))


### Bug Fixes

* use GUID maintenance plan routes ([da5cf47](https://github.com/UteamUP/cli/commit/da5cf478cf31b595204cc105629213c5d28b4fc4))


### Miscellaneous

* rename IoT domain label from Beta to Preview ([255ba69](https://github.com/UteamUP/cli/commit/255ba69e98e4f9ff3edc0b7f0c0ca6be3cd1a05e))


### Documentation

* update stale API guideline references ([bb6266e](https://github.com/UteamUP/cli/commit/bb6266e903413e3cdf17c38fd140818d37ecb81d))


### Tests

* expand UPMate stock CLI parity ([1cdf4f3](https://github.com/UteamUP/cli/commit/1cdf4f3fbf74dbe354504a74a9f2bec0f2c79472))
* mirror UPMate asset detail in CLI ([ad3def6](https://github.com/UteamUP/cli/commit/ad3def63644bea8b86b90bfe5ad3812d91f30455))
* mirror UPMate document and vendor reads ([07326dd](https://github.com/UteamUP/cli/commit/07326dd2d39d09cf3e6cf6630df863a9808ff118))
* mirror UPMate fleet dashboard reads ([4b9ef2d](https://github.com/UteamUP/cli/commit/4b9ef2d3c75c85c36e6bac6f8480d8f078e15aad))
* mirror UPMate project and shift details ([b7ec249](https://github.com/UteamUP/cli/commit/b7ec24978e8c7a485609be13957d6f29b678a995))

## [1.68.0](https://github.com/UteamUP/cli/compare/1.67.1...1.68.0) (2026-07-15)


### Features

* add on-call calendar cli action ([a17aba4](https://github.com/UteamUP/cli/commit/a17aba46f36e4552ec0b081c22f89b87dee2ce74))
* add on-call subscription cli actions ([3fd7f1a](https://github.com/UteamUP/cli/commit/3fd7f1abb926ffadfc155ca14139812b92aacb3b))
* add on-call summary cli action ([5f2eb11](https://github.com/UteamUP/cli/commit/5f2eb11821be23960ef6de391627bf5b83c2257f))

## [1.67.1](https://github.com/UteamUP/cli/compare/1.67.0...1.67.1) (2026-07-14)


### Code Refactoring

* retire shift-template CLI domain ([9998bf1](https://github.com/UteamUP/cli/commit/9998bf1534107748dba042859d022fc9d288ae67))

## [1.67.0](https://github.com/UteamUP/cli/compare/1.66.1...1.67.0) (2026-07-14)


### Features

* add GUID previous handover CLI route ([24eb697](https://github.com/UteamUP/cli/commit/24eb6978bbd7d221ad56dfbbc10b18cc9cf417b4))
* **handover:** add secure numeric redemption command ([0540498](https://github.com/UteamUP/cli/commit/0540498436a44dd620aca4a77a3cd78866e35d16))
* use GUID routes for shift CLI ([a96760a](https://github.com/UteamUP/cli/commit/a96760aee005770a0808e23f295f0dc4e90c25d3))
* use GUID routes for shift instance CLI ([00bc101](https://github.com/UteamUP/cli/commit/00bc101bc513d56008f9183c98d5812a08e62294))
* use GUID routes for shift request CLI ([d07949a](https://github.com/UteamUP/cli/commit/d07949a60873fde2e48f6075b38281d7d6b82061))
* use GUID routes for tenant holiday CLI ([e57579b](https://github.com/UteamUP/cli/commit/e57579b7519c48560775710044f4db1dd64dfbef))
* use GUID routes for workforce group CLI ([a2005b3](https://github.com/UteamUP/cli/commit/a2005b35ede9c9f7384ac51bb7f70386910ec92f))
* use GUID routes for workforce training CLI ([40c94e3](https://github.com/UteamUP/cli/commit/40c94e347ed9d09c349cc4738e194078c27fe5c9))


### Bug Fixes

* **cli:** use GUIDs for geofence zones ([28807dd](https://github.com/UteamUP/cli/commit/28807dd9715cb4e5bad5f258c96335e4306e6cd9))


### Miscellaneous

* format CLI sources ([05ea4b0](https://github.com/UteamUP/cli/commit/05ea4b066642e0fa67551a99d9958ea279b86cb7))

## [1.66.1](https://github.com/UteamUP/cli/compare/1.66.0...1.66.1) (2026-07-13)


### Bug Fixes

* **shift-handover:** make CLI challenge redemption explicit ([b4f28cb](https://github.com/UteamUP/cli/commit/b4f28cb3b292c601d7a628c0955e81974d80c8e4))

## [1.66.0](https://github.com/UteamUP/cli/compare/1.65.0...1.66.0) (2026-07-13)


### Features

* **shift-handover:** expose operational metrics in CLI ([e1c7e7a](https://github.com/UteamUP/cli/commit/e1c7e7a78ab032e1f40e4bbc9d487b7b2ac8d678))
* **upmate:** generalize tutorial discovery ([9454c7d](https://github.com/UteamUP/cli/commit/9454c7da25a37c88199b2fff9f7a92b3aa670a6c))

## [1.65.0](https://github.com/UteamUP/cli/compare/1.64.0...1.65.0) (2026-07-13)


### Features

* **cli:** acknowledge critical handover items ([3c58781](https://github.com/UteamUP/cli/commit/3c5878118559373fd66ffa99168dcf705184dc8e))
* **cli:** add authoritative offer sharing ([3adaf86](https://github.com/UteamUP/cli/commit/3adaf86a173cd8cc2e3e00d9f6da3d6adb46cae4))
* **cli:** add contractor application AI drafts ([0757980](https://github.com/UteamUP/cli/commit/0757980215b4009f0f8d52bf6db1cc3138eeff51))
* **cli:** add marketplace contact sharing ([5fe9956](https://github.com/UteamUP/cli/commit/5fe99569e86d8277528efacaa0321c9c0341497a))
* **cli:** add marketplace meeting commands ([bacc62c](https://github.com/UteamUP/cli/commit/bacc62c7e1d11887fb2321d5fc4a6a7b95ba851b))
* **cli:** add marketplace messages with mentions ([3d4491b](https://github.com/UteamUP/cli/commit/3d4491b2f46905c3faa9bf92ff8b57f9dc7cad41))
* **cli:** add shift handover operational link actions ([99e7066](https://github.com/UteamUP/cli/commit/99e706637d7bf774bc6b67641579f0d2c79b468f))
* describe IoT operational monitoring ([1073122](https://github.com/UteamUP/cli/commit/1073122c921aae4ed2c43b71ae28291ff76d975a))
* expose shift handover audit commands ([3787348](https://github.com/UteamUP/cli/commit/378734894bbf056436470e061b6ca4a358e23fdb))
* expose shift handover operational baton ([c86d029](https://github.com/UteamUP/cli/commit/c86d02997a7eb6d104f3f4b6bcd55a9abd0e7973))


### Bug Fixes

* **cli:** make labour rate commands GUID-first ([6acfeb8](https://github.com/UteamUP/cli/commit/6acfeb8b7e5d2ca200a0240c6bbdc1396c32d072))

## [1.64.0](https://github.com/UteamUP/cli/compare/1.63.0...1.64.0) (2026-07-12)


### Features

* **marketplace:** add CLI conversation domain ([e7803bb](https://github.com/UteamUP/cli/commit/e7803bb6e5b28b8bd879257cc23bea8611fb1c9e))
* **scheduling:** add GUID-first handover carry-over actions ([76cdf1d](https://github.com/UteamUP/cli/commit/76cdf1d4510508964224ac07caa8c5d60cec1c30))

## [1.63.0](https://github.com/UteamUP/cli/compare/1.62.0...1.63.0) (2026-07-12)


### Features

* **iot:** add CLI monitoring domain ([d0323d1](https://github.com/UteamUP/cli/commit/d0323d182760ec0414dd5fe6a5b6c3b5ef6f0ec6))
* **scheduling:** add shift-handover section actions ([6732c4f](https://github.com/UteamUP/cli/commit/6732c4fa71d831be428d55efa697517359bd8956))

## [1.62.0](https://github.com/UteamUP/cli/compare/1.61.0...1.62.0) (2026-07-12)


### Features

* **marketplace:** add labour worker replacement command ([c3f66d1](https://github.com/UteamUP/cli/commit/c3f66d1e48b6a857ca4e4c826f5837b072b37ebe))
* **marketplace:** add labour workspace commands ([1e09f9a](https://github.com/UteamUP/cli/commit/1e09f9a10cda043d7ea75e8dfc4e118f4615fcfd))
* **marketplace:** list labour agreement timesheets ([b803481](https://github.com/UteamUP/cli/commit/b80348127af3ec60dbd9e9d0d37394061e313c77))
* **scheduling:** add shift-handover submit and mutation guard flags ([d648572](https://github.com/UteamUP/cli/commit/d64857263860df479ebf6aa24b0c4605b1b9de29))

## [1.61.0](https://github.com/UteamUP/cli/compare/1.60.0...1.61.0) (2026-07-12)


### Features

* **marketplace:** add contractor analytics commands ([65a9506](https://github.com/UteamUP/cli/commit/65a9506720e53e6f78c89cb3037a5fd00c21edfc))

## [1.60.0](https://github.com/UteamUP/cli/compare/1.59.0...1.60.0) (2026-07-11)


### Features

* **marketplace:** add participant-scoped conversation AI summary domain ([633af60](https://github.com/UteamUP/cli/commit/633af60419f085df666f9ff4f5ee47a5b835d55b))

## [1.59.0](https://github.com/UteamUP/cli/compare/1.58.0...1.59.0) (2026-07-11)


### Features

* add shift handover acceptance commands ([eac5797](https://github.com/UteamUP/cli/commit/eac5797f47d2efddbc05fb80f8b797dbc64e5df3))
* add shift handover start-review and complete commands ([8beeb07](https://github.com/UteamUP/cli/commit/8beeb07c0394a15e818df108383ae7c3e19471e5))
* **marketplace:** add labour AI CLI commands ([34c7fea](https://github.com/UteamUP/cli/commit/34c7feae758ec8c79c538da91014c6fe61077b70))
* **marketplace:** add labour AI offer-comparison commands ([af7713b](https://github.com/UteamUP/cli/commit/af7713bfa85db2fdfd7073a932b7f4ac3ec78995))

## [1.58.0](https://github.com/UteamUP/cli/compare/1.57.0...1.58.0) (2026-07-11)


### Features

* **registry:** add handoverattestation domain ([ef0522a](https://github.com/UteamUP/cli/commit/ef0522a275d385e8d12a5d0622c4ca7eb7a993bf))

## [1.57.0](https://github.com/UteamUP/cli/compare/1.56.0...1.57.0) (2026-07-11)


### Features

* **registry:** add rotationtemplate domain, oncall classify-standby, and workingtime holidays ([4c502d5](https://github.com/UteamUP/cli/commit/4c502d5d31ca5de113c4a7797fcf2ac943bd5b26))

## [1.56.0](https://github.com/UteamUP/cli/compare/1.55.0...1.56.0) (2026-07-11)


### Features

* **registry:** add workingtime domain and oncall override-add action ([b969395](https://github.com/UteamUP/cli/commit/b96939572d7248aa8d0e5648f27b309a0b68a25a))

## [1.55.0](https://github.com/UteamUP/cli/compare/1.54.0...1.55.0) (2026-07-11)


### Features

* **oncall:** add schedule-list, schedule-create, and layer-add actions ([aa9bf98](https://github.com/UteamUP/cli/commit/aa9bf981f3e403264d8695eb86af0954d2e0d5bd))

## [1.54.0](https://github.com/UteamUP/cli/compare/1.53.0...1.54.0) (2026-07-11)


### Features

* **registry:** add oncall domain with 'who' action ([b237083](https://github.com/UteamUP/cli/commit/b23708382967ab830d4be3f628c37b00b00f0f87))


### Miscellaneous

* ignore MSBuild binary logs ([4d6f92a](https://github.com/UteamUP/cli/commit/4d6f92a91382f0377dcf07fdc9041aad9d1b53ec))

## [1.53.0](https://github.com/UteamUP/cli/compare/1.52.0...1.53.0) (2026-07-10)


### Features

* **registry:** add tutorial domain for published UPMate tutorials ([2cbe30c](https://github.com/UteamUP/cli/commit/2cbe30cdcfe7295446bfc8987d17dd57bcf45c99))

## [1.52.0](https://github.com/UteamUP/cli/compare/1.51.0...1.52.0) (2026-07-10)


### Features

* **registry:** add condition grade filter to stock search ([940c1b8](https://github.com/UteamUP/cli/commit/940c1b86a493355adbf143df50a5606844e96594))

## [1.51.0](https://github.com/UteamUP/cli/compare/1.50.0...1.51.0) (2026-07-10)


### Features

* **registry:** add stock-reseller-catalog CLI bindings ([ae1fc8f](https://github.com/UteamUP/cli/commit/ae1fc8f071eff1e259b1d09e111d8d7a91368552))

## [1.50.0](https://github.com/UteamUP/cli/compare/1.49.0...1.50.0) (2026-07-10)


### Features

* **registry:** add stock ai-count-session review actions ([dc19891](https://github.com/UteamUP/cli/commit/dc198912be6bd14d5104ed70c21afdd99400a15c))

## [1.49.0](https://github.com/UteamUP/cli/compare/1.48.1...1.49.0) (2026-07-10)


### Features

* **registry:** add stock seasonal intelligence (8 actions) ([fa3562c](https://github.com/UteamUP/cli/commit/fa3562ceb06460fa718ececcfb71a7a0c2ef032b))

## [1.48.1](https://github.com/UteamUP/cli/compare/1.48.0...1.48.1) (2026-07-10)


### Code Refactoring

* **registry:** rename project-governance ToolNames to PascalCase ([e441887](https://github.com/UteamUP/cli/commit/e4418872ffa7a6a64df18506b8cd5aebe831b4f2))

## [1.48.0](https://github.com/UteamUP/cli/compare/1.47.0...1.48.0) (2026-07-10)


### Features

* **registry:** add project governance + stock activity feeds and make set-owner GUID-first ([d83662d](https://github.com/UteamUP/cli/commit/d83662d5f66bacb4d764d651fe1e0245f0c201ca))

## [1.47.0](https://github.com/UteamUP/cli/compare/1.46.0...1.47.0) (2026-07-09)


### Features

* **registry:** add asset ask + shift-handover generate-summary ([e16112e](https://github.com/UteamUP/cli/commit/e16112e554db93eed8ec13b9ba2c7395753bf65d))

## [1.46.0](https://github.com/UteamUP/cli/compare/1.45.0...1.46.0) (2026-07-09)


### Features

* **registry:** make chemical domain GUID-first (get/update/delete take externalGuid) ([2449a07](https://github.com/UteamUP/cli/commit/2449a07d873fc89c60cc76c490d7748c91bd89b6))

## [1.45.0](https://github.com/UteamUP/cli/compare/1.44.0...1.45.0) (2026-07-09)


### Features

* **registry:** add currentLatitude/currentLongitude flags to workforce-ai daily-brief ([47ec17d](https://github.com/UteamUP/cli/commit/47ec17ded21355ad5db44c208b6cbefca80917cd))

## [1.44.0](https://github.com/UteamUP/cli/compare/1.43.1...1.44.0) (2026-07-09)


### Features

* **registry:** restore AI tier CLI parity with workforce-ai and work-permit-ai domains ([1ecaf87](https://github.com/UteamUP/cli/commit/1ecaf872825848a54440ed996e3207617ea01a4d))

## [1.43.1](https://github.com/UteamUP/cli/compare/1.43.0...1.43.1) (2026-07-08)


### Bug Fixes

* **deps:** pin cloud.google.com/go/ai v0.8.0 for generative-ai-go compat ([407d0bb](https://github.com/UteamUP/cli/commit/407d0bbb17ef12fa58f732c9ba3fb42ddaa55ac9))


### Build System

* **deps:** update go dependencies ([71889eb](https://github.com/UteamUP/cli/commit/71889eb4a446c1f62e95e2bd3b143f03ab473463))

## [1.43.0](https://github.com/UteamUP/cli/compare/1.42.0...1.43.0) (2026-07-08)


### Features

* **bug-and-feature:** update attachments-* CLI descriptions to reflect documents and videos ([360f1fa](https://github.com/UteamUP/cli/commit/360f1faffee17fa751f172d97e1b68885b15ec59))
* **cli:** add health command to report auth env and backend status ([871e271](https://github.com/UteamUP/cli/commit/871e27154bb870330b41d45a8de31480e4a770d7))
* **marketplace:** buyer-reputation CLI action (H13) ([450997a](https://github.com/UteamUP/cli/commit/450997ac2640f637490363af0b3a70d7eec1f732))
* **registry:** add aicreditrequest CLI domain ([c0e63cf](https://github.com/UteamUP/cli/commit/c0e63cf718d0bb0c3a63bd0bc8083f2332f1edf9))
* **registry:** make project CRUD GUID-first (get/update/delete take externalGuid) ([b0c7515](https://github.com/UteamUP/cli/commit/b0c751500f3cb9cbb453c524973c05aa92c6bcc6))


### Documentation

* **guidelines:** update reference to ApiGuidelines in GUID-first domains section ([d22a6be](https://github.com/UteamUP/cli/commit/d22a6bee9592a28ed9e6c57660ceba78f00cf5bf))

## [1.42.0](https://github.com/UteamUP/cli/compare/1.41.0...1.42.0) (2026-07-05)


### Features

* **marketplace:** add facets CLI action ([3d5a6e1](https://github.com/UteamUP/cli/commit/3d5a6e176685157e6d469e39668d5b95904e4519))

## [1.41.0](https://github.com/UteamUP/cli/compare/1.40.0...1.41.0) (2026-07-05)


### Features

* **marketplace:** add seller-scorecard CLI action ([35e1fe5](https://github.com/UteamUP/cli/commit/35e1fe5ff20c9ede9b59e32af2780cce726d953d))

## [1.40.0](https://github.com/UteamUP/cli/compare/1.39.0...1.40.0) (2026-07-05)


### Features

* **marketplace:** add saved-search CLI actions ([6b431bc](https://github.com/UteamUP/cli/commit/6b431bc68e5b23d1b86776b0b4aa55e1e4d3871a))


### Miscellaneous

* **partner:** update wholesaler comment to reference partner program ([b00995b](https://github.com/UteamUP/cli/commit/b00995b4517d8c94c2e3451bd2eb9875f97107a5))

## [1.39.0](https://github.com/UteamUP/cli/compare/1.38.0...1.39.0) (2026-07-04)


### Features

* **registry:** add duplicate actions for code catalog subtrees and stock items ([d62304d](https://github.com/UteamUP/cli/commit/d62304d36baeaba21cd31d65a92ac772afeb9b09))
* **registry:** add reseller CLI surface with actions and tests ([a25a51f](https://github.com/UteamUP/cli/commit/a25a51f1aaa14944b89e5d27dd1d2c41a48bad74))
* **registry:** update partner CLI surface and add new actions for listing and messaging ([5da676d](https://github.com/UteamUP/cli/commit/5da676d4eb6d4aed231560629d7de608708dccb1))

## [1.38.0](https://github.com/UteamUP/cli/compare/1.37.0...1.38.0) (2026-07-03)


### Features

* **projectplanning:** add work-order link commands for stages and risks ([e1232c0](https://github.com/UteamUP/cli/commit/e1232c037dbbb16282d5c5571d123b58837c5de3))

## [1.37.0](https://github.com/UteamUP/cli/compare/1.36.0...1.37.0) (2026-07-03)


### Features

* **cli:** add UPMate estimate + estimate-apply project-copilot actions ([b5f16b0](https://github.com/UteamUP/cli/commit/b5f16b0ed9f7dccaf6b900ae9803de5beded093e))
* **marketplace:** add near-me lat/long flags to browse ([4b9f6a6](https://github.com/UteamUP/cli/commit/4b9f6a6e059e7ad5130afdd75b2bf4d84d0ee64b))
* **plan:** add ai-credit-packages CLI action ([b63ac49](https://github.com/UteamUP/cli/commit/b63ac4960e156bacd0c9d27e62413a5df9894e09))

## [1.36.0](https://github.com/UteamUP/cli/compare/1.35.0...1.36.0) (2026-07-03)


### Features

* **registry:** Part 2 plan-management domains ([99f41e8](https://github.com/UteamUP/cli/commit/99f41e8a69037526dccc673e1236dab5f2e24ea8))


### Bug Fixes

* **registry:** key knowledge article CRUD by GUID ([b2b2036](https://github.com/UteamUP/cli/commit/b2b203682bc42f1b9b74121bde081c90cd83f7f5))

## [1.35.0](https://github.com/UteamUP/cli/compare/1.34.0...1.35.0) (2026-07-02)


### Features

* **cli:** add plan-audit domain for plan change-history and export ([013ee1f](https://github.com/UteamUP/cli/commit/013ee1f534e0454aa2d0febdbf1dcd67c4799f5b))

## [1.34.0](https://github.com/UteamUP/cli/compare/1.33.0...1.34.0) (2026-07-02)


### Features

* **cli:** add promotion bulk grants and tenant extend-trial action ([8f3e5cb](https://github.com/UteamUP/cli/commit/8f3e5cb22f5f7481012fe347e805ac6aea00cf90))

## [1.33.0](https://github.com/UteamUP/cli/compare/1.32.0...1.33.0) (2026-07-02)


### Features

* **cli:** add marketplace and wholesaler read domains ([8deae80](https://github.com/UteamUP/cli/commit/8deae80eda485f70d579140f876586cc4551649a))
* **cli:** add project PM domains (risk register, insights, cost budget thresholds) and AI planning suite ([4a5a80a](https://github.com/UteamUP/cli/commit/4a5a80a81823a299dbdf72fead42bd92c6362e52))

## [1.32.0](https://github.com/UteamUP/cli/compare/1.31.0...1.32.0) (2026-07-02)


### Features

* **cli:** add knowledge entity-link actions (linked/link/unlink) and fix APIPath ([660db3d](https://github.com/UteamUP/cli/commit/660db3d7e7ac0759c713f392613f39f934b8e737))
* **cli:** add project-planning domains (stage/output/budget) and project setters ([ddabc21](https://github.com/UteamUP/cli/commit/ddabc21727f53e539fe0868d05fd76e9da237dbd))

## [1.31.0](https://github.com/UteamUP/cli/compare/1.30.0...1.31.0) (2026-07-01)


### Features

* **cli:** add aicreditgrant domain (issue/mine/claim/revoke, GUID-first) ([2b62bb6](https://github.com/UteamUP/cli/commit/2b62bb637e0f2b2643c181451e0334758fe581b9))

## [1.30.0](https://github.com/UteamUP/cli/compare/1.29.0...1.30.0) (2026-06-30)


### Features

* **cli:** add knowledgespace and knowledgeai (UPMate) domains ([e4ad6fa](https://github.com/UteamUP/cli/commit/e4ad6fa543e2fa18bb468f8d7163c1b75faec02c))

## [1.29.0](https://github.com/UteamUP/cli/compare/1.28.0...1.29.0) (2026-06-30)


### Features

* **cli:** add promotion domain, project-copilot image-report, GUID-keyed plan get ([15bb3f7](https://github.com/UteamUP/cli/commit/15bb3f77d670944c650a4092bd9e8741282a39f6))

## [1.28.0](https://github.com/UteamUP/cli/compare/1.27.0...1.28.0) (2026-06-30)


### Features

* **cli:** add 'cleanup' command to detect unused code via the Usage Verifier ([0f5fb19](https://github.com/UteamUP/cli/commit/0f5fb19d569fe9a0b1f1a8a0787d4d2272c98e7e))
* **workorder-template:** add AI analyze actions and language translate domain ([fa12731](https://github.com/UteamUP/cli/commit/fa12731983e2d4b05f42c40a77ece21ba6ec5bed))


### Miscellaneous

* **gitignore:** add rules to ignore temporary files while keeping .tmpl templates ([996f4d8](https://github.com/UteamUP/cli/commit/996f4d8946c5ecabafc63f07a5db34f146bbb3c1))

## [1.27.0](https://github.com/UteamUP/cli/compare/1.26.0...1.27.0) (2026-06-25)


### Features

* **apikey:** add `ut apikey` domain to mint + manage tenant API keys from the terminal ([d2712f1](https://github.com/UteamUP/cli/commit/d2712f12554c496cb087e05207b2c9cfd1135f1a))
* **cli:** add notification-preference domain (get/set) ([37c4fe8](https://github.com/UteamUP/cli/commit/37c4fe897076abafe3f1dc7c89ccd1c956bbe421))
* **reseller:** add the application meetings CLI action ([4c75ffc](https://github.com/UteamUP/cli/commit/4c75ffc664b058776966ce4587f79827d1386886))


### Bug Fixes

* **bugs:** wire up attachments-upload to POST multipart to {bugExternalGuid}/attachments ([13d0463](https://github.com/UteamUP/cli/commit/13d0463c412cc2f70e4c61c8f2557105b1d6a954))

## [1.26.0](https://github.com/UteamUP/cli/compare/1.25.0...1.26.0) (2026-06-19)


### Features

* **reseller:** add self-serve application, checklist, referral-codes, and tenant-manager CLI actions ([2be0e8e](https://github.com/UteamUP/cli/commit/2be0e8e1c7c9dd01359e0b0910f501e4cefbe3d8))

## [1.25.0](https://github.com/UteamUP/cli/compare/1.24.0...1.25.0) (2026-06-18)


### Features

* **reseller:** CLI program-defaults command mirrors the new MCP tool ([a1c4f13](https://github.com/UteamUP/cli/commit/a1c4f13ae3f37a828466b491951d53969e04cdc4))

## [1.24.0](https://github.com/UteamUP/cli/compare/1.23.0...1.24.0) (2026-06-17)


### Features

* add reseller domain to the CLI registry ([a7f6290](https://github.com/UteamUP/cli/commit/a7f62909b6214d2fc0650286996ba0fca120b2ff))

## [1.23.0](https://github.com/UteamUP/cli/compare/1.22.0...1.23.0) (2026-06-17)


### Features

* **stock:** add count-from-photo action to inventory CLI domain ([2a9f7c2](https://github.com/UteamUP/cli/commit/2a9f7c2ed4d4db823efecfd736536884e86d20fd))

## [1.22.0](https://github.com/UteamUP/cli/compare/1.21.0...1.22.0) (2026-06-14)


### Features

* **workorder:** add --asset-guid filter to list action ([f978a4b](https://github.com/UteamUP/cli/commit/f978a4b2bd91d65a8bf65de09051416b216ee09d))

## [1.21.0](https://github.com/UteamUP/cli/compare/1.20.0...1.21.0) (2026-06-11)


### Features

* **stock:** forecast actions and a project-copilot domain (batch 5) ([4874813](https://github.com/UteamUP/cli/commit/4874813b3641cffdd1855142f06ee9fe581cfb41))
* **stock:** marketplace, warranty, rental and intelligence actions (batch 7) ([13cc37e](https://github.com/UteamUP/cli/commit/13cc37e1e9149a92623bb7d16236f41e0b8ee644))
* **stock:** quality, approvals, reports and settings actions (batch 4) ([6f4fa0a](https://github.com/UteamUP/cli/commit/6f4fa0a0b13c03cec0b8ef78d86fd7db269bf8a1))
* **stock:** scan resolve, offline ops batch and device-token actions (batch 6) ([c3825f1](https://github.com/UteamUP/cli/commit/c3825f151806ab587d52f328a5f99c2bc0192476))
* **stock:** units, reservations and ATP actions (phase 3) ([6e8530d](https://github.com/UteamUP/cli/commit/6e8530d8970807efdbe209b5090b671f54910647))

## [1.20.0](https://github.com/UteamUP/cli/compare/1.19.0...1.20.0) (2026-06-10)


### Features

* add inventory domain with full test coverage ([cef413b](https://github.com/UteamUP/cli/commit/cef413b9e8a7060f3daf590908f14411e95ef408))
* **inventory:** alerts, ack, item search, PO lifecycle, reorder-policy commands ([2c5cd8e](https://github.com/UteamUP/cli/commit/2c5cd8e56283b36d72c9ac1c0f69b3eeef2ae0a3))

## [1.19.0](https://github.com/UteamUP/cli/compare/1.18.0...1.19.0) (2026-06-10)


### Features

* **asset:** add edit-code-assignment action to edit a coded asset by GUID ([d4a28fb](https://github.com/UteamUP/cli/commit/d4a28fb25f12b1e9113caffc91a0d9557275ce0f))

## [1.18.0](https://github.com/UteamUP/cli/compare/1.17.0...1.18.0) (2026-06-09)


### Features

* **asset:** add duplicate action to deep-copy a coded asset by GUID ([c471596](https://github.com/UteamUP/cli/commit/c471596f827af8aeb498d9167fa44c67325d51a2))

## [1.17.0](https://github.com/UteamUP/cli/compare/1.16.1...1.17.0) (2026-06-07)


### Features

* **codecatalog:** add move action to reparent a code catalog entry by GUID ([acbb5cd](https://github.com/UteamUP/cli/commit/acbb5cdc4f5fc03253f097c9bb107f6203d44d75))

## [1.16.1](https://github.com/UteamUP/cli/compare/1.16.0...1.16.1) (2026-06-04)


### Bug Fixes

* **bugsandfeatures:** list WaitList status in list filter and update-status ([768da1e](https://github.com/UteamUP/cli/commit/768da1e2441771be95960377440b3b684bc3dedf))

## [1.16.0](https://github.com/UteamUP/cli/compare/1.15.0...1.16.0) (2026-06-03)


### Features

* **bugsandfeatures:** add `mine` action listing the caller's own reports ([69dc88b](https://github.com/UteamUP/cli/commit/69dc88bd4af5aa4b626aeee8817f2b559189b300))

## [1.15.0](https://github.com/UteamUP/cli/compare/1.14.1...1.15.0) (2026-06-03)


### Features

* **bugsandfeatures:** add reporter conversation read action + share-with-reporter comment flag ([51161b2](https://github.com/UteamUP/cli/commit/51161b26fa71407b192cd99baa4d8198d07071f5))

## [1.14.1](https://github.com/UteamUP/cli/compare/1.14.0...1.14.1) (2026-06-02)


### Bug Fixes

* **codingsystem:** list RDS-PP and RDS-PS in coding-system description ([99f8eb1](https://github.com/UteamUP/cli/commit/99f8eb15d15e50475f8e153ecb4116bf4ee09164))

## [1.14.0](https://github.com/UteamUP/cli/compare/1.13.0...1.14.0) (2026-06-02)


### Features

* **meter-schedule:** add open-workorders action listing all open meter-reading workorders by asset GUID ([8eef372](https://github.com/UteamUP/cli/commit/8eef37263b0ce1fde5a9373d84c3a5a8aca86e45))
* **registry:** add GUID-first responsible-owners, meter record-workorder, and document lifecycle CLI actions ([d118907](https://github.com/UteamUP/cli/commit/d11890729121c4009b68282807243b38ab42fea2))

## [1.13.0](https://github.com/UteamUP/cli/compare/1.12.0...1.13.0) (2026-06-01)


### Features

* **asset:** add get-documents-aggregated CLI action ([64d55d9](https://github.com/UteamUP/cli/commit/64d55d9def4ac6f1374b98736f0d18658754f9ca))


### CI/CD

* remove dead Homebrew tap repository_dispatch step from release workflow ([5bfbbcb](https://github.com/UteamUP/cli/commit/5bfbbcb8c946135b8c5f745c6c8730b22696da6b))

## [1.12.0](https://github.com/UteamUP/cli/compare/1.11.0...1.12.0) (2026-05-31)


### Features

* **meter-schedule:** add calendar-recurrence flags to create and update ([7e1a68c](https://github.com/UteamUP/cli/commit/7e1a68cec64333d507d2c2649005514ed4875bfa))

## [1.11.0](https://github.com/UteamUP/cli/compare/1.10.0...1.11.0) (2026-05-29)


### Features

* **meter-schedule:** add create-workorder action with template flags ([360b573](https://github.com/UteamUP/cli/commit/360b57358f74c484e37dcef3198b0dd2f9d5807e))

## [1.10.0](https://github.com/UteamUP/cli/compare/1.9.0...1.10.0) (2026-05-27)


### Features

* **codecatalog:** add --aspect flag to update-by-guid and assign-asset ([a387b5f](https://github.com/UteamUP/cli/commit/a387b5fd6f2afd72ae2932571543cfec887b75d1))

## [1.9.0](https://github.com/UteamUP/cli/compare/1.8.0...1.9.0) (2026-05-27)


### Features

* **cli:** migrate meter-schedule domain to Guid-first ([65c6e16](https://github.com/UteamUP/cli/commit/65c6e16b354504002eab140864d5d045e70b5cca))
* **codecatalog:** add designation and identification-letters actions ([3803da8](https://github.com/UteamUP/cli/commit/3803da8e950e064f4b1348310635a301b150fe64))
* **workorder-template:** add run-schedule-now verb to wot domain ([74dfb78](https://github.com/UteamUP/cli/commit/74dfb7835787830ac1168aa5a70a4a2f3f665452))

## [1.8.0](https://github.com/UteamUP/cli/compare/1.7.1...1.8.0) (2026-05-26)


### Features

* **admin-users:** add global-admin user management domain ([8b9930e](https://github.com/UteamUP/cli/commit/8b9930ef5b2329203257f7809e01313c963775ba))

## [1.7.1](https://github.com/UteamUP/cli/compare/1.7.0...1.7.1) (2026-05-23)


### Code Refactoring

* **asset-type:** migrate get/update/delete to externalGuid arg ([be59ee4](https://github.com/UteamUP/cli/commit/be59ee4b59cc18e3400edfe817147859108d9fd0))

## [1.7.0](https://github.com/UteamUP/cli/compare/1.6.0...1.7.0) (2026-05-21)


### Features

* **bugsandfeatures:** add increment-hit verb for manual occurrence recording ([ed565c2](https://github.com/UteamUP/cli/commit/ed565c29a87280711d44d7f8d1ae87bbbde1cb1d))

## [1.6.0](https://github.com/UteamUP/cli/compare/1.5.0...1.6.0) (2026-05-20)


### Features

* **asset:** add get-assigned-stock action to asset domain registry ([fa412a3](https://github.com/UteamUP/cli/commit/fa412a301c11a6277b79fee3fdd82d8f0455c7d1))

## [1.5.0](https://github.com/UteamUP/cli/compare/1.4.1...1.5.0) (2026-05-19)


### Features

* **workorder-template:** add create-workorder action ([90ecc68](https://github.com/UteamUP/cli/commit/90ecc6809a1dd63f2b1345498e41b92a14a3619d))


### Miscellaneous

* commit pending changes ([9508b79](https://github.com/UteamUP/cli/commit/9508b796dabab20c7e8cb7f4318b89defcdaa10c))

## [1.4.1](https://github.com/UteamUP/cli/compare/1.4.0...1.4.1) (2026-05-14)


### CI/CD

* **release:** pass HOMEBREW_TAP_GITHUB_TOKEN through to GoReleaser ([a13a6ac](https://github.com/UteamUP/cli/commit/a13a6ac572b2eba73a18c7df62f028213041244a))

## [1.4.0](https://github.com/UteamUP/cli/compare/1.3.2...1.4.0) (2026-05-14)


### Features

* **document:** add get-metadata and get-timeline actions to document domain ([8bc8dc7](https://github.com/UteamUP/cli/commit/8bc8dc708e435e77e5c5f87f8a2361a2e7f7e210))


### Bug Fixes

* **auth:** enforce minimum TLS version 1.2 in HTTP clients ([579ba21](https://github.com/UteamUP/cli/commit/579ba2140fa1a9f372e0b9f5d9879b1be6efa387))


### Miscellaneous

* **deps:** bump google.golang.org/api v0.278.0 -&gt; v0.279.0 ([96ece81](https://github.com/UteamUP/cli/commit/96ece8161182415f9f2c5ed4f1be1c2012eb8adf))

## [1.3.2](https://github.com/UteamUP/cli/compare/1.3.1...1.3.2) (2026-05-10)


### Bug Fixes

* **config:** improve export JSON prompt handling in config init command ([c0b6bbd](https://github.com/UteamUP/cli/commit/c0b6bbdba00eec2752a5a9ee8f157f95857d6fc2))

## [1.3.1](https://github.com/UteamUP/cli/compare/1.3.0...1.3.1) (2026-05-09)


### Bug Fixes

* **cli:** keep GUIDs at the boundary on bug create + login output ([53fc231](https://github.com/UteamUP/cli/commit/53fc231f6abf365c68283242c068a29f7613ce76))


### Miscellaneous

* **deps:** bump 3 direct go module minors ([476c378](https://github.com/UteamUP/cli/commit/476c378251a045eae7adeb07895f29db7c3f176e))


### Documentation

* **cli:** comments-list now returns top-level comments newest-first ([ea9f7ec](https://github.com/UteamUP/cli/commit/ea9f7ec611c5595379795684d2ee92f8a4ea1f64))

## [1.3.0](https://github.com/UteamUP/cli/compare/1.2.0...1.3.0) (2026-05-06)


### Features

* **attachments:** add commands for managing bug attachments (list, upload, download, delete) ([8a518fe](https://github.com/UteamUP/cli/commit/8a518fed9d3b6b68692ae7f9e28358468a75a1d8))
* **comments:** add commands for listing and adding comments on bugs ([ff7878b](https://github.com/UteamUP/cli/commit/ff7878bc00de9e2dcce6650ca3bb1cd6f6d4bda3))


### Bug Fixes

* **registry:** route sub-resource verbs through path templates and rename body fields ([73777af](https://github.com/UteamUP/cli/commit/73777aff0492720766dccf37595a939d7caabb24))


### Documentation

* update release process guidelines for automation and remove manual steps ([cefd11a](https://github.com/UteamUP/cli/commit/cefd11a3d2ac57ef5da8818d959fddc8e3e98ef4))

## [1.2.0](https://github.com/UteamUP/cli/compare/1.1.0...1.2.0) (2026-04-27)


### Features

* **bugs:** add performance auto-monitoring to validated --source flag values ([57d6b9b](https://github.com/UteamUP/cli/commit/57d6b9b18d493e461e60f4e095dfe3babc80ea25))
* **industry-coding:** CLI domain for hotspot CRUD (Task 7.2) ([3d9959d](https://github.com/UteamUP/cli/commit/3d9959d01f88a8fcb3b887efdaa71e2f888e6e49))
* **registry:** add assign-asset command for code-catalog entry assignment with audit log preservation ([198e8fb](https://github.com/UteamUP/cli/commit/198e8fb75359e3c6c5a70ea476257500129aeedb))
* **registry:** add search parameter for free-text search in bugs and features ([e2a8a4f](https://github.com/UteamUP/cli/commit/e2a8a4fd53897ac5d51c19ee15249948f69108b0))
* **registry:** add source filter for bug and feature queries ([862e9b3](https://github.com/UteamUP/cli/commit/862e9b335e87c484c888d0111b02b8966fa97164))
* **registry:** add update-notes command for admin notes management and enhance REST path handling for update sub-routes ([6f61c0c](https://github.com/UteamUP/cli/commit/6f61c0c16e9ff3fcf76bc5c6d8d3a7c566fb1700))
* **registry:** add update-type command for converting submissions between Bug and Feature with audit history ([274050f](https://github.com/UteamUP/cli/commit/274050fe47dd6a4ecf3a7c854659e37667e7c0f5))


### Code Refactoring

* **registry:** remove unused helper functions ([f75a71b](https://github.com/UteamUP/cli/commit/f75a71b0241e20ba2132846fb4ee682d6784433a))

## [1.1.0](https://github.com/UteamUP/cli/compare/1.0.0...1.1.0) (2026-04-24)


### Features

* Add all MCP domains, install to /usr/local/bin, .zshrc PATH setup ([1f4958c](https://github.com/UteamUP/cli/commit/1f4958ca81f5f930703a4b03fc5bedb72ca9fdcc))
* add asset-type-meter domain with actions for managing meter definitions ([6d832c5](https://github.com/UteamUP/cli/commit/6d832c5297627e4de76fbd6918b20b78edf7b008))
* add auth, plan validation, and tenant override to video analyzer ([52e58d4](https://github.com/UteamUP/cli/commit/52e58d499df0acc51b0d6eeaa6fbc88c737d4460))
* add domain for managing subscription plans with list and get actions ([43c7e52](https://github.com/UteamUP/cli/commit/43c7e521eeace57d0d8a820ab5d85999eb116de2))
* add dual progress bars to image and video analyzers ([6788145](https://github.com/UteamUP/cli/commit/678814576e5cee335fa0f27b7d7c4de9671c613b))
* Add JSON export config option for CLI responses ([1e59cf0](https://github.com/UteamUP/cli/commit/1e59cf04ac656600884ca4950d3d5ab9ed8faeff))
* add multi-tenant selector, plan validation, and auth status tenant info ([736f562](https://github.com/UteamUP/cli/commit/736f562c3488388b199d1b6a79b0fc2d4aaf8d74))
* add new domains for condition, criticality, geofence, improvement, meter schedule, and sales booking management ([3277fa0](https://github.com/UteamUP/cli/commit/3277fa0ac873f4a03d4acd5808b5aa160a8865c4))
* add report-analytics and asset-reports domains ([1adb97f](https://github.com/UteamUP/cli/commit/1adb97f6cbf637d6c0a07beb7345e54edd859dd7))
* Add REST API support for email/password login auth ([1203c37](https://github.com/UteamUP/cli/commit/1203c3768ae94dcd3f7063e134ff2a4edac765df))
* add tenant show and tenant select commands ([876ef8f](https://github.com/UteamUP/cli/commit/876ef8f0831e4006d18d3c5f94e7b60ff6c9d82f))
* add video analyzer command for CMMS inventory extraction from MP4/MOV videos ([26b6eed](https://github.com/UteamUP/cli/commit/26b6eedf523a0ad3a7fed885c538eb4047f0dea0))
* **ai:** Add BYOK AI provider CLI domain registry ([d15e90c](https://github.com/UteamUP/cli/commit/d15e90cacea79730aec5177d151a5c52c63c0351))
* **asset:** add `ut asset get-by-guid <guid>` subcommand ([6a8426d](https://github.com/UteamUP/cli/commit/6a8426d40b2b13c1fd00cca2b4e7fedec00e4035))
* **asset:** Multi-type flags and get-specs subcommand for ut asset ([544cc79](https://github.com/UteamUP/cli/commit/544cc79fb422d86a897f36969325eb790d1a4967))
* **bank-transfer:** add CLI domain registry for bank transfer billing ([f669649](https://github.com/UteamUP/cli/commit/f66964928e44b67aaf52c5e4c8d325630a0fdcce))
* **bugs-and-features:** add delete action for global-admin to permanently remove submissions ([2333f73](https://github.com/UteamUP/cli/commit/2333f73a00d82350e50fbbf85a50f64256627b34))
* **bugsandfeatures:** add bugsandfeatures CLI domain ([fd4727a](https://github.com/UteamUP/cli/commit/fd4727ad29165fe62b7d53bc03e412c7d08b98cf))
* **cli:** add codecatalog update-by-guid / deactivate-by-guid / remove-asset-assignment ([b3d18ed](https://github.com/UteamUP/cli/commit/b3d18ed19bd041d5cda2bfecd61c4474bbee608d))
* **cli:** add config apikey and config model shortcut commands ([6fc6cb2](https://github.com/UteamUP/cli/commit/6fc6cb27c78a12481a9e589a79a8ce595cce85dc))
* **cli:** add document-import, logbook-import, document-review, ai-usage domains ([b5aa10d](https://github.com/UteamUP/cli/commit/b5aa10d129d616ba343b02dfd58f6a48498b2218))
* **cli:** add image analyze command with Gemini config integration ([3337106](https://github.com/UteamUP/cli/commit/33371066014fe9a881b1b1d06fbe3ec64cf915d1))
* **cli:** add user-ui-state domain registry ([74119d2](https://github.com/UteamUP/cli/commit/74119d23cd8afa430db69e49c71e6c11d699a8c6))
* **dedicated-instance:** Add dedicated instance integration with domain registry and client URL resolution ([b6fb5ef](https://github.com/UteamUP/cli/commit/b6fb5ef03c8b6c73ba46e596de87b6e182b78910))
* **document:** Add version and archive actions to document domain ([a6f7237](https://github.com/UteamUP/cli/commit/a6f7237b60c4ef7a82d71f1e0e6306b9c5fff07f))
* Initial UteamUP CLI project — Go CLI mirroring MCP server ([8f280cb](https://github.com/UteamUP/cli/commit/8f280cbdf560ad3fa37be6d6ce9533cd8da4fb53))
* **journal-code-linking:** Add journal and codecatalog CLI domains ([79e607b](https://github.com/UteamUP/cli/commit/79e607bf81cb8c3ed38e8d4c8531dfa88b46fa33))
* **journal:** add import, create-from-image, and mention search domains ([4e39d05](https://github.com/UteamUP/cli/commit/4e39d057d4bae556469053288946ff3f7a1e955f))
* **meter-reading:** add CLI domain for GUID-based meter-reading commands ([c0ac999](https://github.com/UteamUP/cli/commit/c0ac999453794bbfdc87f9cecc08a29260624a60))
* native Go image analyzer — remove Python dependency entirely ([0d539bd](https://github.com/UteamUP/cli/commit/0d539bdcfc2c80c98120f7f6e3b4100e82029d2a))
* **output:** enhance `bugs get` command to display full status history in a dedicated block ([981c52f](https://github.com/UteamUP/cli/commit/981c52fb0f6e7b925e6ecddb8508066ab2d8b6af))
* **project:** add my-projects subcommand ([ed79e47](https://github.com/UteamUP/cli/commit/ed79e47d176c019470f8709d6edb33b0080d9a02))
* **registry:** add support for update-status action and GUID-based identifier handling ([fb83a56](https://github.com/UteamUP/cli/commit/fb83a560b2517d9a321a4c95b99376523250bd6f))
* **registry:** implement admin-billing-gateway commands for tenant billing management ([1a60625](https://github.com/UteamUP/cli/commit/1a606252f5f3f4c270a2713df478b7c7e87ad207))
* **tenant:** add invite-defaults get/set commands ([2d1d0e8](https://github.com/UteamUP/cli/commit/2d1d0e829ab76542e96b5b1ace5971f7fbf2032c))
* **ux-simplification:** Add quick-report CLI domain ([69a5407](https://github.com/UteamUP/cli/commit/69a5407f35fb5b380f10a6d2baf4d1ab2f99bfc3))
* v0.3.0 — add CLIGuidelines.md and update changelog ([0decf59](https://github.com/UteamUP/cli/commit/0decf591f35e54bbf94545be7d3448c34d78502f))
* vendor/location enrichment with GPS geocoding and online lookup ([0714680](https://github.com/UteamUP/cli/commit/0714680b5668e03e0e16af871d0c4d1c128981e3))
* **workorder:** add `ut workorder quick-close` action with required and optional flags ([3ca51ab](https://github.com/UteamUP/cli/commit/3ca51abdeb85d9477623b3e40ee616a922a09a67))
* **workorder:** add quick-close action with required and optional flags ([490633e](https://github.com/UteamUP/cli/commit/490633ee815d92c6b893547a0af716cafe301e35))


### Bug Fixes

* Add tenant headers (X-Tenant-ID, X-Tenant-Guid) to all API requests ([3256adc](https://github.com/UteamUP/cli/commit/3256adc0799cd4f145571931eb8f119132f48b95))
* **ci:** correct release-please-action SHA to v4.4.0 ([520ff24](https://github.com/UteamUP/cli/commit/520ff24a684607706e746f803b2a1c83c1fed0d0))
* **cli:** honor UTEAMUP_API_BASE_URL when no config file exists ([86a9f3e](https://github.com/UteamUP/cli/commit/86a9f3ec47b6486bfd84342ad97c6a82d9830f14))
* **cli:** print image analyzer status banner to stdout for clean display ([d7b332b](https://github.com/UteamUP/cli/commit/d7b332b502a9e46a9b4d8ab3399e89b0d1b68fe1))
* **cli:** show errors instead of silent exit, expand analyzer search paths ([67ef32d](https://github.com/UteamUP/cli/commit/67ef32dd74f0feb2a669adf1b1b59aa4eae577e2))
* correct Gemini MIME type — ImageData() prepends 'image/' automatically ([fd24466](https://github.com/UteamUP/cli/commit/fd244664845d4e9764917c18c2e19b4d35a2d6a3))
* move checkpoint to ~/.uteamup/, add superpowers to .gitignore ([c52f4b9](https://github.com/UteamUP/cli/commit/c52f4b90511062a7291d2c04278f4065eca880dc))
* resolve panic on float flag with int default in domain registry ([10405e1](https://github.com/UteamUP/cli/commit/10405e1bbb80933c1ba7dc1221bc7ea1d2bac92d))


### Miscellaneous

* add firebase-debug.log to gitignore ([71d78fd](https://github.com/UteamUP/cli/commit/71d78fd9a2f0ad3eb02ea8ed5d5462e3efe969c1))
* Add MIT LICENSE file ([1f52848](https://github.com/UteamUP/cli/commit/1f5284866163c26098fc178a87add4becdcc0615))
* update go.mod/go.sum, ignore Images/ directory ([5f92665](https://github.com/UteamUP/cli/commit/5f9266547f33d68ea002de9286be98a490386453))
* Update GoReleaser GitHub owner to UteamUP ([cc9f490](https://github.com/UteamUP/cli/commit/cc9f4907ddd24dd8c1cebcb45f831a36aabea14d))


### Documentation

* add detailed flag documentation with before/after examples ([6bda188](https://github.com/UteamUP/cli/commit/6bda1886de865efff662e50645e370ff50cbb6d6))
* Add version management section (upgrade, downgrade, pin) ([204fb08](https://github.com/UteamUP/cli/commit/204fb083912cd6aac40de627f954920b4bfbc8dd))
* comprehensive release process documentation in CLIGuidelines.md ([f23c0f7](https://github.com/UteamUP/cli/commit/f23c0f78aa498973585bf21f4ef8aaad42ef00f4))
* **guidelines:** document REST-routing and CSRF/auth rules for domain commands ([8902781](https://github.com/UteamUP/cli/commit/8902781304c3e19c78d0302b0160c4eaf207860d))
* update changelog for v0.10.0 ([2d2d251](https://github.com/UteamUP/cli/commit/2d2d251e47cdeb7fd30d0cfa9e8786d35cebb61f))
* update changelog version to v0.6.0 ([808438a](https://github.com/UteamUP/cli/commit/808438a162140c89637df65b5d95a743c707c557))
* update changelog with image analyzer CLI features ([0a33c31](https://github.com/UteamUP/cli/commit/0a33c3160d888a75e015dadacdf882e35c930795))
* update README with image analysis, Gemini config, and v0.3.0 references ([bc2adc6](https://github.com/UteamUP/cli/commit/bc2adc6a7428d1b3c4e7761db437f17543b8e264))
* update README with video analysis, tenant management, vendor/location enrichment ([bb19e8f](https://github.com/UteamUP/cli/commit/bb19e8f724e66eab045c5e670436276061a3aad6))
* **workorder:** document canonical priority tiers on CLI flags ([8479de5](https://github.com/UteamUP/cli/commit/8479de535c8bb861e5ff092ba1e429c2de754a8e))


### Tests

* **project:** add domains_project_test.go and verify CLI pipeline ([2b1695c](https://github.com/UteamUP/cli/commit/2b1695c584bd1812097569779674a672cdb03cfa))


### CI/CD

* add CI ([db51ec9](https://github.com/UteamUP/cli/commit/db51ec95244359ab714641f1fc8e075195acfc75))
* add CODEOWNERS ([76d395d](https://github.com/UteamUP/cli/commit/76d395d73f5e268dad50befedf73cb7b51eb7ac7))
* add Dependabot ([0c8dbaa](https://github.com/UteamUP/cli/commit/0c8dbaaa5b20e9f932e4fe6a5a8d4a1b59b5f88c))
* add release ([90772f2](https://github.com/UteamUP/cli/commit/90772f277673fdbff4b6994b60bb43f5c9370de8))
* add Release Please automated versioning ([06f7362](https://github.com/UteamUP/cli/commit/06f736281a7103163b6ecd8c6be04d88eefe30ac))

## [Unreleased]

### Added
- **`uteamup workorder list --asset-guid <guid>` filter.** Added an optional `asset-guid` string flag to the `workorder list` action in `domains_workorder.go` (kebab → camelCase `assetGuid`) so the list can be scoped to a single asset's work orders — the watch NFC → asset → its workorders flow. New contract test `TestWorkorderListHasAssetGuidFlag` in `domains_workorder_test.go` asserts the flag is present, typed `string`, and optional.
- **Performance Auto-Monitoring (CLI).** Added `performance-auto` to the validated `--source` flag values for the `uteamup bugs list` command in `domains_bugs.go`. Added corresponding test coverage.

### Changed
- **`uteamup bugs get <externalGuid>` default human output now surfaces the full `statusHistory[]` as a `History:` block.** `-o json` and `-o yaml` were already pass-through and already carried the array; the gap was the default table renderer, which flattened the nested array into a 60-char-truncated JSON blob. `internal/output/table.go::printObjectTable` now skips the `statusHistory` key from the key/value section and, after the main object, prints a chronological `History:` block — one line per entry with ISO-8601 timestamp, `from -> to` transition, author (`changedByUserEmail` falling back to `changedByUserId` so the `system:auto-ingest` sentinel remains visible), and the note truncated to fit the terminal width (`$COLUMNS` with a 160-col fallback). `bugs list` output is intentionally unchanged — per-row history would turn a one-screen list into screens of noise. Six new tests in `internal/output/table_test.go` cover: multi-entry chronological ordering, `system:auto-ingest` author visibility, long `[auto-reopen]` note truncation with `...`, empty history renders `History: (none)` rather than crashing, list output does NOT expand history per row, `fromStatus -> toStatus` arrow marker. `go vet ./... && go test ./... -race && make build` all clean.

## [0.10.0] — 2026-04-22

### Fixed
- **`cmd/root.go` + `cmd/login.go` — honor `UTEAMUP_API_BASE_URL` when no `~/.uteamup/config.json` exists.** Both entry points previously applied the env var override only *inside* `config.Load()` (which returns an error when there is no config file), then silently fell back to the hardcoded production URL `https://api.uteamup.com`. The uteamup-debug Claude skill, CI, and anyone else following the README's "set `UTEAMUP_API_BASE_URL`" instructions were unknowingly hitting prod. `runLogin` and `registerDomainCommands` now consult the env var before the hardcoded fallback. An active profile's `BaseURL` still wins when a config file is present, so existing workflows are unchanged.
- **`internal/client/client.go::CallREST` — unconditionally set `X-Requested-With: XMLHttpRequest`.** Mutating endpoints (POST/PUT/PATCH/DELETE) on routes like `/api/bugsandfeatures` reject requests without the marker with HTTP 400 `"Missing required X-Requested-With header."`. The frontend `apiCall()` has always sent it; the CLI did not, which made `uteamup bugs update-status … Fixed` fail on its PATCH step.
- **`internal/registry/registry.go::buildRESTPath` — route `update-status` and GUID-keyed actions correctly.** The `HTTPMethod` map gained `"update-status": "PATCH"`, and `buildRESTPath` now accepts `args["externalGuid"]` as the positional identifier in addition to `args["id"]`, so GUID-first domains get `PATCH /api/<domain>/{guid}/status` and `GET /api/<domain>/{guid}` instead of the list endpoint as a fallback. Unblocks `uteamup bugs update-status <externalGuid> Fixed --resolution-reference <sha>`.

### Docs
- **`CLIGuidelines.md`** — added three subsections under "Architecture Quick Reference" documenting the above changes: (1) REST routing table (action → HTTP method → URL pattern), (2) CSRF header requirement on mutating calls, and (3) the `LocalOrAzureAdPolicy` pattern that new CLI-facing backend controllers should prefer over the legacy stacked `[Authorize(Policy="AzureAdPolicy"), Authorize(Policy="LocalPolicy")]` (a third scheme like Google listed in a policy triggers a Google JWT Bearer challenge that short-circuits to 401 even when Local validates).

### Added
- **`tenant` domain.** New `internal/registry/domains_tenant.go` registers two actions mirroring the new backend MCP tools: `invite-defaults-get <tenantGuid>` and `invite-defaults-set <tenantGuid>` with flags `--auto-license`, `--license-type 0|1` (0=Regular, 1=Helpdesk), `--auto-role`, `--role-id <guid>`. These let operators configure a tenant so every new invite automatically assigns a license + role. `go vet` clean, registry tests pass under `-race`, `make build` produces `bin/uteamup` and `bin/ut`.

- **`ut workorder quick-close`** — new action on the existing `workorder` domain that mirrors the backend `UteamupWorkorderQuickClose` MCP tool (atomic create + close from a pre-approved template). Required flags: `--template <guid>`, `--asset <guid>`, `--note <text>`. Optional flags: `--idempotency-key <guid>` (CLI generates one per invocation when omitted), `--industry-code <guid>` (informational), `--performed-at <ISO-8601>` (clamped to ±0/−30 days server-side). Falls under the stricter automation rate-limit tier (5/min, 50/day). Test file `internal/registry/domains_workorder_test.go` verifies the action is registered with `ToolName: UteamupWorkorderQuickClose`, that the three required flags are marked `Required: true`, that the three optional flags stay optional, and that the action takes zero positional args (all identifiers are GUIDs that would be painful to position-order). 3/3 passing under `go test -race`.
- **`bugsandfeatures` domain.** New `internal/registry/domains_bugsandfeatures.go` registers four actions that mirror the new MCP tools: `list` (global-admin; filters by type/status/severity/tenant/submitter, pagination, default hides Rejected/Confirmed), `get <externalGuid>` (global-admin), `create` (any authenticated user; requires `--title`, `--description`, `--idempotency-key`), and `update-status <externalGuid> <toStatus>` (global-admin; `--note` required on reject/reopen, `--resolution-reference` required on Fixed — enforced server-side). Aliases: `bugs`, `features`, `baf`.

### Changed
- `ut workorder list` / `create` / `update` priority flag descriptions now enumerate the canonical tiers `1=Low, 2=Medium, 3=High, 4=Urgent, 5=Critical` to match the backend `WorkorderPriority` enum and the new Critical dropdown option in the frontend. Registry metadata only — no behavioral change; `go vet` clean.

### Added
- `ut asset get-by-guid <guid>` (also `uteamup asset get-by-guid ...`) — new subcommand mirroring the backend `uteamup_asset_get_by_guid` MCP tool / `GET /api/asset/by-guid/{guid}`. Fetches an asset using its stable `ExternalGuid` (survives migrations and reseeds) rather than the integer id, making CLI invocations safe to copy between environments. Registered in `internal/registry/domains_asset.go` with ToolName `UteamupAssetGetByGuid`, and `domains_asset` test expects the new `get-by-guid` action (alongside `list`, `get`, `create`, `update`, `delete`, `search`).
- Four new domain registries mirroring the new image/document import MCP tools: `domains_document_import.go` (`document-import get`), `domains_logbook_import.go` (`logbook-import get`), `domains_document_review.go` (`document-review queue`, `document-review acknowledge`), `domains_ai_usage.go` (`ai-usage summary`). Read-only + acknowledge surface only; multipart upload and batch commit stay HTTP-only by design.
- `ut project my-projects` (also `uteamup project my-projects`) — new subcommand mirroring the backend `GET /api/project/my-projects` endpoint. Lists projects that contain workorders assigned to the current user (primary or secondary). Registered in `internal/registry/domains_project.go` with ToolName `UteamupProjectMyProjects`; the backend MediatR handler was added in the same PR (MCP `UteamupProjectMyProjects` tool).
- `internal/registry/domains_project_test.go` — unit tests covering the project domain registration, `projects` alias, `search` / `my-projects` action/ToolName mapping, and a regression guard asserting `my-projects` takes zero args and zero flags (user identity must come from the API key). Follows the existing `domains_journal_test.go` pattern.

### Tests
- Verified `go fmt`, `go vet`, `go test ./... -race`, and `make build` all pass. Note: `make lint` (golangci-lint) continues to fail on pre-existing issues in `internal/imageanalyzer/analyzer/json_fix.go` and `internal/videoanalyzer/gps/mp4meta.go` (commits `0d539bd` / `26b6eed`, March 2026) — unrelated to this change.

## [0.7.1] — 2026-03-28

### Fixed
- Fixed panic when running geofence commands: `interface conversion: interface {} is int, not float64`
- Made float flag default handling in domain registry defensive against int/float64 type mismatch

## [0.7.0] — 2026-03-27

### Added
- `report-analytics` domain (alias: `report-stats`) — view aggregated report analytics with cost trends, top assets by maintenance cost, and completion metrics
- `asset-reports` domain — view paginated reports for a specific asset with summary statistics
- Enriched `report` domain description with cost breakdown, checklists, meter readings, labour, and tool usage details

## [0.6.3] — 2026-03-23

### Added
- Dual progress bars for image and video analysis: per-item steps (load/upload/analyze/save) + overall 0%-100% progress
- File size display in video analysis per-video headers
- Per-item entity count summary after each image/video is analyzed

## [0.6.2] — 2026-03-23

### Added
- `ut tenant show` command (aliases: `list`, `ls`) — lists all tenants with name, GUID, plan, and status
- `ut tenant select` command — interactive tenant picker that saves selection to config and updates active token
- Tenant selection updates both config profile (`tenantGuid`) and cached token so `ut auth status` reflects the change immediately

## [0.6.1] — 2026-03-23

### Added
- Video analyzer requires UteamUP authentication (login) and active tenant subscription plan
- Interactive multi-tenant selector when user has access to multiple tenants and no `tenantGuid` is configured
- `tenantGuid` field in CLI profile config (`~/.uteamup/config.json`) for tenant override
- `UTEAMUP_TENANT_GUID` environment variable override
- Tenant mismatch detection: re-authentication required when config tenant differs from logged-in tenant
- Plan name and tenant name displayed in video analyzer banner

## [0.6.0] — 2026-03-23

### Added
- `uteamup video analyze <path>` command for AI-powered CMMS video analysis
- Video file validation via magic byte MIME detection (MP4, MOV supported; GIF routed to image analyzer)
- Gemini File API integration with async upload, processing poll with terminal spinner, and automatic cleanup
- Video-specific CMMS entity extraction prompt with timestamp detection (MM:SS format)
- GPS coordinate extraction from MP4/MOV container metadata (©xyz and ISO 6709 atoms)
- Vendor enrichment via follow-up Gemini lookup (website, full name, business category)
- Temporal deduplication to merge same-entity detections across video frames
- Cross-video deduplication using existing grouping algorithm
- Consistent CSV output (assets, tools, parts, chemicals, vendors, locations) matching image analyzer format
- Dry-run mode for video cost estimation
- Video Analysis section in CLIGuidelines.md

## [0.3.0] — 2026-03-22

### Added
- `uteamup image analyze <path>` command for AI-powered CMMS image analysis
- Gemini AI configuration in CLI profiles (`geminiApiKey`, `geminiModel`)
- `ut config apikey [key]` shortcut to get/set Gemini API key
- `ut config model [name]` shortcut to get/set default Gemini model
- `ut config model list` to display all available Gemini models
- Pre-processing status banner showing image count, model, and output path
- Config init prompts for Gemini settings with model selection
- Support for `=` syntax in config commands (`ut config apikey=xyz`)
- Image analyze requires authentication (login required)
- CLIGuidelines.md with full release, packaging, and Homebrew documentation

## [0.1.0] — 2026-03-22

### Added
- Initial project scaffold with Cobra CLI framework
- Dual authentication: interactive login (email/password) and API key auth (OAuth 2.0 + PKCE)
- `ut` shortname alias for `uteamup` binary
- JSON config file with multi-profile support (~/.uteamup/config.json)
- Domain registry pattern for declarative command definitions
- Starter domains: Asset, WorkOrder, User
- HTTP client with exponential backoff retry and SSE parsing
- Output formatters: table (default), JSON, YAML
- Auth gate requiring login before any command
- Cross-platform installers: MSI (Windows), .pkg + Homebrew (macOS), .deb + .rpm (Linux)
- Shell completions for bash, zsh, fish, powershell
- Structured logging with sensitive data redaction
