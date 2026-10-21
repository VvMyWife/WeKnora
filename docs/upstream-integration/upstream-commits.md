# 本轮纳入的全部官方独有提交

固定目标 eecfe4b3c8c34c94e9b559e5dd02187b4ae447cd；相对定制基线 a418adc45f383b9accb0d51bdcf226c468e656fb，共644个官方独有提交。采用正式merge，纳入目标最终状态（包括撤回），不是逐条重放。原始提交主题保留，未按主观价值筛选；不含目标之后或未合并PR。

| 序号 | 日期 | 官方提交 | 原始主题 |
|---|---|---|---|
| 1 | 2026-09-04 | [2d6b84cb23](https://github.com/Tencent/WeKnora/commit/2d6b84cb23ce6eac25b420773429194d7faf8520) | feat(session): show owner badge for API sessions in sidebar |
| 2 | 2026-09-04 | [5b057af467](https://github.com/Tencent/WeKnora/commit/5b057af4674a8c7114936cb3ed09ed20998d6a2d) | fix(agent): web_fetch 兼容 LLM 双重序列化的 items 参数 |
| 3 | 2026-09-03 | [7c4b5f1bfe](https://github.com/Tencent/WeKnora/commit/7c4b5f1bfef709ca27c5d36d252f498786659543) | feat(miniprogram): add zh/en i18n, tab icons, and base library compatibility |
| 4 | 2026-09-05 | [2834e6dcf6](https://github.com/Tencent/WeKnora/commit/2834e6dcf61ce5d1f221f4bbb312e90ff0c2ebe8) | fix: widen audit log drawer (#3001) |
| 5 | 2026-09-05 | [d5cb842580](https://github.com/Tencent/WeKnora/commit/d5cb842580d461050cce1f0ce684a1fe6667b352) | feat(agent): unify sandbox tools and drop host preloaded skills (#3033) |
| 6 | 2026-09-05 | [e847114d19](https://github.com/Tencent/WeKnora/commit/e847114d19b5419a92050c1b59f1c0c8e09d6164) | fix(feishu): preserve images embedded in post messages |
| 7 | 2026-09-05 | [df7e84eb30](https://github.com/Tencent/WeKnora/commit/df7e84eb302ba51897863eec340b51b259ee00f4) | fix(message): log why chat history indexing was skipped |
| 8 | 2026-09-05 | [1e36e071ec](https://github.com/Tencent/WeKnora/commit/1e36e071ec156d98a970b497f1a228516200e60e) | feat(mcp): add per-tool enable toggle (#2989) |
| 9 | 2026-09-05 | [816bb3c982](https://github.com/Tencent/WeKnora/commit/816bb3c982355ef95c1f7a06040b78189eb16e40) | feat(session): show in-progress status in the sidebar |
| 10 | 2026-09-06 | [dcc89d4b2d](https://github.com/Tencent/WeKnora/commit/dcc89d4b2df97e7cd21b40b58fea7566701d0943) | feat(system): add admin UI for creating users (#3034) |
| 11 | 2026-09-06 | [3d3bb7f6d1](https://github.com/Tencent/WeKnora/commit/3d3bb7f6d1acca8caa84bb73b189fe46c30967a9) | fix(system): keep one-time create-user passwords and copy full account details (#3039) |
| 12 | 2026-09-07 | [316db899d8](https://github.com/Tencent/WeKnora/commit/316db899d8249db10b4292557dfc1cc71bc6f5bb) | fix(security): reject traversal in local prefixes, task IDs, and wiki sort |
| 13 | 2026-09-07 | [b037e26260](https://github.com/Tencent/WeKnora/commit/b037e2626071e5e259ebc371be7e26e8905dd7a2) | fix(chat): keep OCR/caption text of scanned-page image hits in merge (#3064) |
| 14 | 2026-09-07 | [4fb0b08719](https://github.com/Tencent/WeKnora/commit/4fb0b0871957f4f19937929b437982832d304576) | fix(initialization): pass stored ExtraConfig to model connection test |
| 15 | 2026-09-06 | [801005f736](https://github.com/Tencent/WeKnora/commit/801005f736b5e76c514588f3ac9abe5212ebb219) | fix(knowledge): keep long non-ASCII folder names valid UTF-8 |
| 16 | 2026-09-07 | [160c806b94](https://github.com/Tencent/WeKnora/commit/160c806b946a4dc3c628c2db7a7f80777cb456d2) | feat(datasource): keep the source document's last-modified time through ingestion |
| 17 | 2026-09-06 | [84f2345641](https://github.com/Tencent/WeKnora/commit/84f234564154fc474482b6825f84a5cacacb1c74) | fix(frontend): finish knowledge base tour without creating a knowledge base |
| 18 | 2026-09-07 | [524deeaad5](https://github.com/Tencent/WeKnora/commit/524deeaad53397eee7adb59a90e7cff94ed97ebe) | feat(sandbox): run sandboxed execs as root by default (#3067) |
| 19 | 2026-09-07 | [7d5c13ba8e](https://github.com/Tencent/WeKnora/commit/7d5c13ba8e63da52986411dc7bd05e4ffe4b189d) | fix(organization): 修正成员权限说明中编辑角色的共享知识库权限 |
| 20 | 2026-09-06 | [6f5a9c8c31](https://github.com/Tencent/WeKnora/commit/6f5a9c8c3124453cd713693631696c3b64911dde) | feat(frontend): expose HTTP status on API responses via $httpStatus |
| 21 | 2026-09-07 | [fee252e936](https://github.com/Tencent/WeKnora/commit/fee252e936e6081046bff3f4619fea63aad62b22) | refactor(system): extract reset-password popup into ResetPasswordDialog (#3047) |
| 22 | 2026-09-07 | [91a24fc376](https://github.com/Tencent/WeKnora/commit/91a24fc37657052d6d11a405548593360d3ebc2e) | fix: sanitize qdrant string payloads (#2997) |
| 23 | 2026-09-06 | [01377b686c](https://github.com/Tencent/WeKnora/commit/01377b686c7fa320acb162531c1fbcd786e8da6c) | fix: persist cleared IM knowledge bases |
| 24 | 2026-09-06 | [25a930b9a0](https://github.com/Tencent/WeKnora/commit/25a930b9a09251cc2f057176c4c2acab64351c0f) | fix(preview): stop routing legacy .doc/.ppt to the OOXML previewers |
| 25 | 2026-09-01 | [f7cc5969e3](https://github.com/Tencent/WeKnora/commit/f7cc5969e3bf355307f616331f98c1307bb73ef0) | feat(mcp): add manual knowledge update tool |
| 26 | 2026-08-31 | [2544fedcd3](https://github.com/Tencent/WeKnora/commit/2544fedcd356527b4a8b93c179de9b79fd1b2e9b) | fix: report embedding dimension mismatches |
| 27 | 2026-09-01 | [25fcbb9401](https://github.com/Tencent/WeKnora/commit/25fcbb9401ec1ecb79212d9374f44ba6240c2295) | fix(wiki): treat rate-limit-shaped 403 responses as transient |
| 28 | 2026-09-02 | [5fa50711e2](https://github.com/Tencent/WeKnora/commit/5fa50711e27a378d93427e44984384bcee30d40b) | fix(ui): prevent invite registration form header overlap |
| 29 | 2026-09-07 | [86136b25e3](https://github.com/Tencent/WeKnora/commit/86136b25e3eb5a905f769870d7216dfdf3abd955) | fix(session): serve org-shared KB images through the message file proxy (#3069) |
| 30 | 2026-09-07 | [02c1759044](https://github.com/Tencent/WeKnora/commit/02c1759044fbc906914dfdaf63d26ed94aaadf0c) | fix(knowledge): confirm document deletion before reporting success |
| 31 | 2026-09-07 | [5f98eaadaa](https://github.com/Tencent/WeKnora/commit/5f98eaadaa3576538c4a4fd0eef382063fb1a303) | fix(preview): lazy-load markdown images without dropping data URIs |
| 32 | 2026-09-07 | [5be8d13fc6](https://github.com/Tencent/WeKnora/commit/5be8d13fc6252d5f1b0a653625b56cf2a046de68) | feat(integrations): add CLI setup page (#3081) |
| 33 | 2026-09-07 | [9e8673816e](https://github.com/Tencent/WeKnora/commit/9e8673816e940c7737d49ffd85ccc571ba6b52ce) | fix(wiki): keep folder names that look like page-type labels when placing pages |
| 34 | 2026-09-07 | [384fbcf7ed](https://github.com/Tencent/WeKnora/commit/384fbcf7ede8568d5038d2edbd6de1fd7fb02ec6) | fix(frontend): project create-user HTTP status onto a 'created' flag |
| 35 | 2026-09-07 | [c591f16d53](https://github.com/Tencent/WeKnora/commit/c591f16d539a3e458b65b1c65f504ac0ddb9c495) | refactor(system): drop idempotent flag from create-user response body |
| 36 | 2026-09-07 | [1c448bf72c](https://github.com/Tencent/WeKnora/commit/1c448bf72c6bf30fedb380e2dff3b1178478b217) | fix(chat): send one completion-token field per provider (#3082) |
| 37 | 2026-09-07 | [6f0ed94a47](https://github.com/Tencent/WeKnora/commit/6f0ed94a47c587a98ff363ca3f5f3042ae8f4bd6) | feat(i18n): add Japanese (ja-JP) locale support (#3080) |
| 38 | 2026-09-07 | [071a87d5f7](https://github.com/Tencent/WeKnora/commit/071a87d5f767cf9223f417745cd72cd0d7ce51f0) | fix(i18n): proofread ja-JP locale strings |
| 39 | 2026-09-07 | [6a55860431](https://github.com/Tencent/WeKnora/commit/6a55860431629ad84c77802e5fd8b93f524a86a8) | feat(ui): multi-stage Docker frontend build (no host npm) (#3074) |
| 40 | 2026-09-07 | [061df8efdc](https://github.com/Tencent/WeKnora/commit/061df8efdc25f7dde2e95a5d14f03efb8d5d8db5) | feat(datasource): persist source created time; Feishu uses content edit time for source_updated_at (#3083) |
| 41 | 2026-09-07 | [bd6c4a9b96](https://github.com/Tencent/WeKnora/commit/bd6c4a9b96c3a3c2bcba738269e5a891db302454) | fix(frontend): recover the chat stream from an expired access token (#3028) |
| 42 | 2026-09-07 | [bd7aa40119](https://github.com/Tencent/WeKnora/commit/bd7aa4011914ba263cb4042f6a7a064daf6defc6) | fix(wiki): make the category_path directory filter match on SQLite |
| 43 | 2026-09-07 | [15727591a7](https://github.com/Tencent/WeKnora/commit/15727591a7c7ec336d14bedc522e38bd659d3a6e) | feat(web-search): add Bocha AI Search provider (#3036) |
| 44 | 2026-09-07 | [df7d871af0](https://github.com/Tencent/WeKnora/commit/df7d871af08d1c85723108a8c8b3989e67b5a9ad) | fix(i18n): restore natural Japanese in proofread strings |
| 45 | 2026-09-07 | [647848f395](https://github.com/Tencent/WeKnora/commit/647848f3954dae34473b8a8d0e0eef5e0fb3a58e) | fix(agent): improve web search and persistent page reading (#3086) |
| 46 | 2026-09-08 | [ef9cbf456d](https://github.com/Tencent/WeKnora/commit/ef9cbf456d042de4c442674e5b940dd93d15a2d2) | docs: update website guides and improve writing style (#3095) |
| 47 | 2026-09-08 | [694f0413dd](https://github.com/Tencent/WeKnora/commit/694f0413ddc395a93673ebef629b5dcabea5af90) | refactor(agent): discover and invoke MCP tools on demand |
| 48 | 2026-09-08 | [a57e97c4e1](https://github.com/Tencent/WeKnora/commit/a57e97c4e1f85147b7499332ddae00600e498b3e) | fix(frontend): keep agent editor dependency loading alive for contributors |
| 49 | 2026-09-08 | [e18cf138a3](https://github.com/Tencent/WeKnora/commit/e18cf138a3e96edbb2cc3e706ceb89ae40e24e21) | fix: remove GPL converter and bundle third-party license sources |
| 50 | 2026-09-08 | [3d81b977f1](https://github.com/Tencent/WeKnora/commit/3d81b977f12f7c19bb560d2118a7651f7fa00ba8) | refactor(resource-access): unify scoped authorization and transfer recovery (#3097) |
| 51 | 2026-09-08 | [277d6ee851](https://github.com/Tencent/WeKnora/commit/277d6ee851c1891d00c0c6596669d16e443bcde1) | feat(sandbox): add an interactive terminal and files tab beside chat (#3110) |
| 52 | 2026-09-08 | [4d95cd4581](https://github.com/Tencent/WeKnora/commit/4d95cd4581a82fa148e5972d8335792d96b49a60) | fix(mcp): require schema discovery before calls and render tool results (#3113) |
| 53 | 2026-09-08 | [8d7298fb5d](https://github.com/Tencent/WeKnora/commit/8d7298fb5d759973cb1e481cadc5ecdf16dca599) | fix(graph): feed OCR chunks to graph extraction instead of image-link placeholders (#3092) |
| 54 | 2026-09-08 | [3e6010e7cd](https://github.com/Tencent/WeKnora/commit/3e6010e7cd3937f289cc1dbadc829e71eb1163f4) | docs: add English and bilingual product tours to READMEs (#3117) |
| 55 | 2026-09-09 | [85b76d1a30](https://github.com/Tencent/WeKnora/commit/85b76d1a3012df37377065aaad21eff1578b6506) | feat(mcp)!: persist tool catalogs and add guided configuration (#3118) |
| 56 | 2026-09-09 | [918c20342c](https://github.com/Tencent/WeKnora/commit/918c20342cb28812df46414d7ee7455c65a6230d) | fix(skills): accept UTF-8 BOM in skill manifests |
| 57 | 2026-09-09 | [73462de222](https://github.com/Tencent/WeKnora/commit/73462de222d6ca0b2db65bfacb37b66f839c7740) | feat(mcp)!: streamline setup and use short routing handles |
| 58 | 2026-09-09 | [4b04cd83ee](https://github.com/Tencent/WeKnora/commit/4b04cd83eef47072811976410ea9de43c6049113) | fix(mcp): normalize JSON-encoded argument envelopes |
| 59 | 2026-09-09 | [fbaff900a0](https://github.com/Tencent/WeKnora/commit/fbaff900a09b2543c291fb0545ecfaf361bfaafc) | feat(chat): append messages to a running agent turn (#3123) |
| 60 | 2026-09-09 | [3ab6d3940d](https://github.com/Tencent/WeKnora/commit/3ab6d3940d1564f4db69d011941041e366370de7) | feat(sandbox): unify the interactive PTY prompt and reconnect running sandboxes (#3128) |
| 61 | 2026-09-09 | [5cf093706e](https://github.com/Tencent/WeKnora/commit/5cf093706ebecdfe8bc4eca80886e80c01805289) | fix(chat): preserve steering continuity and composer focus |
| 62 | 2026-09-09 | [e0ee15807c](https://github.com/Tencent/WeKnora/commit/e0ee15807c1388f84d1f91fa0516319ff3137c1c) | fix(embed): add missing conversationTime and referencesDrawer i18n keys for all locales |
| 63 | 2026-09-09 | [0f3ffbe492](https://github.com/Tencent/WeKnora/commit/0f3ffbe4924dd001dab9a1460110d1b0518ee3d8) | fix(chat): refine sandbox artifact panel interactions |
| 64 | 2026-09-09 | [b01abb6374](https://github.com/Tencent/WeKnora/commit/b01abb6374e8f1f13723d95c6bb64c4cc4521314) | fix(chat): reconcile steered messages and artifact ownership |
| 65 | 2026-09-09 | [79f6d4e269](https://github.com/Tencent/WeKnora/commit/79f6d4e269cedb9b1787b990d062f3ed1f955930) | fix(chat): prevent pending steer messages from jittering during streaming |
| 66 | 2026-09-09 | [a826a1f12c](https://github.com/Tencent/WeKnora/commit/a826a1f12c47f219780fc056ceec695e8866187a) | fix(skills): verify runtime prerequisites and support installer steering |
| 67 | 2026-09-09 | [2d7df096ca](https://github.com/Tencent/WeKnora/commit/2d7df096ca30a4a2edc531c64286703049014186) | fix(skills): pin install guidance and generalize runtime prompts |
| 68 | 2026-09-10 | [b60351f861](https://github.com/Tencent/WeKnora/commit/b60351f86128d22bfa998b13989530917d32ef2b) | fix(agent): prevent MCP answers from citing unrelated KB sources (#3144) |
| 69 | 2026-09-10 | [1c16db3d26](https://github.com/Tencent/WeKnora/commit/1c16db3d26b658f99a533bed97f2a495ce571351) | fix(preview): improve keyboard scrolling and fullscreen controls |
| 70 | 2026-09-10 | [7b36b4315e](https://github.com/Tencent/WeKnora/commit/7b36b4315efa1f97fe03b4357ebb6ef7f4f26d38) | fix(preview): align header controls and artifact cards |
| 71 | 2026-09-10 | [80c7b2f2bd](https://github.com/Tencent/WeKnora/commit/80c7b2f2bd36db31dc81edfdcabc86d7b5cd9987) | fix(preview): support storage in sandboxed HTML artifacts |
| 72 | 2026-09-10 | [5db13a131e](https://github.com/Tencent/WeKnora/commit/5db13a131e10e8ee2105211f665412ebc13bd98e) | feat(faq): filter entries by enabled status (#3157) |
| 73 | 2026-09-11 | [01205b3d56](https://github.com/Tencent/WeKnora/commit/01205b3d56c1f007457c76be223f7aad568d9593) | feat(sandbox): add safe live file management to session sandboxes (#3181) |
| 74 | 2026-09-11 | [72581cb462](https://github.com/Tencent/WeKnora/commit/72581cb462611341c3fa3f1a75762d76bc56600c) | Revert "feat(sandbox): add safe live file management to session sandboxes (#3…" |
| 75 | 2026-09-11 | [8c2c976127](https://github.com/Tencent/WeKnora/commit/8c2c97612784b7ef4fde5a23e31884b60b6b1116) | fix(agent): honor MCP service advanced_config.timeout for CallTool window |
| 76 | 2026-09-11 | [7e53c31671](https://github.com/Tencent/WeKnora/commit/7e53c3167164d68cfb6c2e897ddd3e755b4b5a45) | feat(browser): connect local Chrome with task controls and display modes (#3180) |
| 77 | 2026-09-11 | [7a98a8e3de](https://github.com/Tencent/WeKnora/commit/7a98a8e3de5a9aa85c8f1fa6b02772986cb435d0) | feat(browser): polish connection settings, open tasks in a new window, and streamline prompts (#3193) |
| 78 | 2026-09-11 | [081851c6b8](https://github.com/Tencent/WeKnora/commit/081851c6b8802225ae968ba67d8de444e258c67b) | fix(browser): support human handoffs and resume interrupted tasks |
| 79 | 2026-09-12 | [3910cfb117](https://github.com/Tencent/WeKnora/commit/3910cfb1172bc5c170d9f888fe223a43246490d4) | fix(browser): complete task cancellation, screenshots, and result display (#3206) |
| 80 | 2026-09-12 | [eeaaa7f64c](https://github.com/Tencent/WeKnora/commit/eeaaa7f64cd27ef7b5ab64d44479b8ae71929819) | fix(preview): repair PptxGenJS slide masters so previews render |
| 81 | 2026-09-12 | [3db7fd5389](https://github.com/Tencent/WeKnora/commit/3db7fd53897be584c2c813ed4f1bdaf0fd119593) | fix(tracing): keep sandbox RPC spans under their owning operations |
| 82 | 2026-09-12 | [a799f23b26](https://github.com/Tencent/WeKnora/commit/a799f23b26511193f1870644ae12ac7362294b27) | feat(sandbox): stream live shell and install command output |
| 83 | 2026-09-12 | [815d32d3a9](https://github.com/Tencent/WeKnora/commit/815d32d3a9976b51f186dbfa97e76d4c8f437147) | fix(sandbox): use E2B BuildID as template version |
| 84 | 2026-09-12 | [462999ec3f](https://github.com/Tencent/WeKnora/commit/462999ec3f5c1467ef0ccf5cf8c422393f40a0e6) | feat(browser): add picture-in-picture task preview |
| 85 | 2026-09-13 | [fd7fa6a7d4](https://github.com/Tencent/WeKnora/commit/fd7fa6a7d407c3fbc6dfb31ea05d9a361373dcdf) | fix(memory): preserve extraction progress and atomic memory updates (#3224) |
| 86 | 2026-09-13 | [17f893865d](https://github.com/Tencent/WeKnora/commit/17f893865d4d8337f7e0bf942c6931a8ce7a08c1) | fix(memory): rank memories by the query, not by importance |
| 87 | 2026-09-14 | [e42f09d535](https://github.com/Tencent/WeKnora/commit/e42f09d5359e07212ed2d58c50ab8fdd689d1539) | fix(datasource): validate replacement credentials before saving |
| 88 | 2026-09-14 | [801ae18f91](https://github.com/Tencent/WeKnora/commit/801ae18f91b3d1cb370ccbe38ed987eaff9f9d5e) | fix(embed): enforce host origin allowlists on framed pages |
| 89 | 2026-09-14 | [935b456c3c](https://github.com/Tencent/WeKnora/commit/935b456c3c75429ad0804ad06bb9d97ce0341245) | fix(skills): isolate skill drops from global file uploads |
| 90 | 2026-09-14 | [fbad5b3730](https://github.com/Tencent/WeKnora/commit/fbad5b3730f382140a43f715fbf4cb235c7848b1) | fix: improve model editor connection tests and save flow |
| 91 | 2026-09-14 | [350b2ab364](https://github.com/Tencent/WeKnora/commit/350b2ab364ed43386e1cde88022159abba1ca7a1) | fix(milvus): support multilingual BM25 retrieval (#3235) |
| 92 | 2026-09-14 | [fc2c12e8ec](https://github.com/Tencent/WeKnora/commit/fc2c12e8ec08efc9b948438fb720db1ff88ca9ca) | fix(models): protect auto-tag model references |
| 93 | 2026-09-14 | [1f82fbc9b5](https://github.com/Tencent/WeKnora/commit/1f82fbc9b5e62bb0fd137968be5cd51481f7e196) | fix(datasource): preserve distinct GitLab files during deduplication |
| 94 | 2026-09-14 | [5d3114fc9e](https://github.com/Tencent/WeKnora/commit/5d3114fc9e571bd1540b8c7aa8e497db8cb65654) | feat(datasource): add DingTalk Docs integration with safe sync and localization (#2313) |
| 95 | 2026-09-14 | [98e216841c](https://github.com/Tencent/WeKnora/commit/98e216841cf7eb5952bb9eb2c96579febc727ca4) | fix(im): recover stale Yunzhijia WebSocket connections |
| 96 | 2026-09-14 | [6ccc5cedf1](https://github.com/Tencent/WeKnora/commit/6ccc5cedf11d43b05e0ff9a1e385e4aedf8d6818) | feat(sandbox): add a graphical desktop to session sandboxes (#3245) |
| 97 | 2026-09-14 | [b94612a7ec](https://github.com/Tencent/WeKnora/commit/b94612a7ecf286f1e3bcc15de638d9c7d1e66ba8) | fix(agent): PUT no longer erases the avatar, and driver errors stay server-side (#3230) |
| 98 | 2026-09-12 | [a84f866892](https://github.com/Tencent/WeKnora/commit/a84f86689265b3264dd208910bf553268f05dd51) | fix(i18n): add missing MCP OAuth locale keys |
| 99 | 2026-09-10 | [f8daa8d785](https://github.com/Tencent/WeKnora/commit/f8daa8d785cd847fe5aa514f3d83b8d49d70ac16) | fix(retrieval): floor ES script_score cosine at 0 to stop 400s on negative similarity |
| 100 | 2026-09-07 | [86f0903473](https://github.com/Tencent/WeKnora/commit/86f090347304027310d50dd2582a63883375ab7c) | fix: report qdrant update failures |
| 101 | 2026-09-08 | [780f2beb5a](https://github.com/Tencent/WeKnora/commit/780f2beb5af306fce79b0a7ab0b396a4c8aa28b1) | fix(agent): stop emitting terminal answer events for retryable empty stops |
| 102 | 2026-09-14 | [9d67a0eb4d](https://github.com/Tencent/WeKnora/commit/9d67a0eb4d8375e32cd7b936f479cfc09d2394db) | fix: clean up wiki data when deleting knowledge base (#3109) |
| 103 | 2026-09-14 | [e1bbc50b6e](https://github.com/Tencent/WeKnora/commit/e1bbc50b6ebc6d51c3546ba0f3e8b5d88fba3c69) | fix(feishu): preserve active streams during long-running QA |
| 104 | 2026-09-12 | [1ef38fdb8b](https://github.com/Tencent/WeKnora/commit/1ef38fdb8b19347b82d3a99f6f17d75ac09ad606) | feat(mcp): pass fileName and folder filters to knowledge APIs |
| 105 | 2026-09-09 | [ffe1d7302b](https://github.com/Tencent/WeKnora/commit/ffe1d7302badea9a8abfc2449ac5d74c1f656380) | fix(im): stop streaming when agent execution completes |
| 106 | 2026-09-15 | [86d245d892](https://github.com/Tencent/WeKnora/commit/86d245d8925299e761a7a9d2d76b3b855e3a2dd6) | feat(helm): expose the upload size limit as global.maxFileSizeMB |
| 107 | 2026-09-15 | [90e3977714](https://github.com/Tencent/WeKnora/commit/90e3977714a08e25ccd4506e317fa9c062c9c372) | test(datasource): keep yuque and ima config tests off live DNS |
| 108 | 2026-09-15 | [b23ae055fe](https://github.com/Tencent/WeKnora/commit/b23ae055feef678c72138298f4d97693c6e49920) | fix(im): finalize replies after agent QA returns (#3261) |
| 109 | 2026-09-15 | [d8126455ea](https://github.com/Tencent/WeKnora/commit/d8126455eaa50e60b1231102b1c251807053de2a) | fix(knowledge): allow XMind files in knowledge imports |
| 110 | 2026-09-15 | [e07e782b20](https://github.com/Tencent/WeKnora/commit/e07e782b20ddcc19afbfc3a259fe55525a188d39) | fix(im): preserve errors emitted after agent completion |
| 111 | 2026-09-15 | [bca3a9f448](https://github.com/Tencent/WeKnora/commit/bca3a9f448e9936bcf2160d5543a97b86b24ecc7) | fix(storage): use virtual-hosted addressing for OBS bucket requests (#3273) |
| 112 | 2026-09-15 | [42aceb17fb](https://github.com/Tencent/WeKnora/commit/42aceb17fba067aab7987a7a150f1da5024c4dfd) | fix(ui): distinguish file types and enlarge chat resource chips |
| 113 | 2026-09-15 | [6f98ff80c3](https://github.com/Tencent/WeKnora/commit/6f98ff80c3185679cc16f834faf42fc662770e68) | Revert "fix(ui): distinguish file types and enlarge chat resource chips" |
| 114 | 2026-09-16 | [06e8022476](https://github.com/Tencent/WeKnora/commit/06e802247661709eeaa2e0a9f09ca281b49cfb9b) | feat(website): add product homepage and unified docs deployment (#3297) |
| 115 | 2026-09-16 | [01f0700b35](https://github.com/Tencent/WeKnora/commit/01f0700b356b36c36be4c3110fc9932e9ce6a8df) | feat(datasource): add Confluence sync (#3287) |
| 116 | 2026-09-15 | [419ff2507c](https://github.com/Tencent/WeKnora/commit/419ff2507c255d5f39264a34f68581bc06f9e8e3) | feat(knowledge): add in-place file replacement primitive |
| 117 | 2026-09-16 | [079e96ae64](https://github.com/Tencent/WeKnora/commit/079e96ae64b7ba46f8cd9f2ecb8b33578951e0b3) | fix(knowledge): make in-place file replacement race-safe |
| 118 | 2026-09-16 | [a7d3545297](https://github.com/Tencent/WeKnora/commit/a7d35452972fd51ae6813852f4bfa73e2d41a110) | feat(knowledge): add batch file downloads (#3300) |
| 119 | 2026-09-16 | [2bc1261b87](https://github.com/Tencent/WeKnora/commit/2bc1261b87636c8ed4fddf9aabb4b998ba424cd9) | chore: remove vendored archives and unused Python helpers |
| 120 | 2026-09-16 | [d7ccd5ba51](https://github.com/Tencent/WeKnora/commit/d7ccd5ba517cc3968e6d5000a2ab0f4b51ef1dda) | fix: align client sessions and integration request handling (#3305) |
| 121 | 2026-09-16 | [d5efe5cedd](https://github.com/Tencent/WeKnora/commit/d5efe5cedde738fac2dc1aad65eb2ad530bedd1e) | feat(website): add WeChat and Tencent Cloud usage options (#3314) |
| 122 | 2026-09-16 | [bfa7f0ed69](https://github.com/Tencent/WeKnora/commit/bfa7f0ed696765791dacc24e8a6e9a21dac0c620) | refactor(frontend): remove unused components and share tag selection logic |
| 123 | 2026-09-16 | [5bcc43d728](https://github.com/Tencent/WeKnora/commit/5bcc43d728a237e3dac094904c604b0d63f571ad) | refactor(frontend): share skill installer model configuration |
| 124 | 2026-09-16 | [f2606629b3](https://github.com/Tencent/WeKnora/commit/f2606629b3204b6e7f07f024e1ee247bbf810f80) | refactor(frontend): share skill install progress subscriptions |
| 125 | 2026-09-17 | [84c9871d16](https://github.com/Tencent/WeKnora/commit/84c9871d16f9c0f58bbb757a977521e0f10ee406) | fix(browser): align BrowserSkill 0.3.0 and repair task lifecycle |
| 126 | 2026-09-15 | [38fee1e1f7](https://github.com/Tencent/WeKnora/commit/38fee1e1f7218a35bc5b682d27b8602c3a7f6dc6) | fix(milvus): preserve keyword retrieval scores |
| 127 | 2026-09-17 | [d8be87d940](https://github.com/Tencent/WeKnora/commit/d8be87d9408255bd12a3d860783be6d5d7e18025) | fix(retrieval): bound keyword-only BM25 scores before rerank (#3343) |
| 128 | 2026-09-17 | [55ec13ac41](https://github.com/Tencent/WeKnora/commit/55ec13ac41449a27ede696a3a2d74fd3ab762284) | fix(knowledge): stop deleting rows from blocking duplicate uploads |
| 129 | 2026-09-16 | [e80f5dfbc7](https://github.com/Tencent/WeKnora/commit/e80f5dfbc7996ea4e09f22255333bfff17941260) | feat(web-search): add Serply provider |
| 130 | 2026-09-16 | [0ad08d41be](https://github.com/Tencent/WeKnora/commit/0ad08d41be7c9279adf5a3f95cd1768686ffd7ea) | fix(repository): make message keyword search work on SQLite and MySQL |
| 131 | 2026-09-17 | [fc7f37da9b](https://github.com/Tencent/WeKnora/commit/fc7f37da9b6d625d505e69a5e7a87f9d79994e3d) | fix(skills): stop racing background install in retry tests |
| 132 | 2026-09-17 | [c9d5987a5d](https://github.com/Tencent/WeKnora/commit/c9d5987a5d7bb50b776eeff47da80199dbbd05fd) | fix(initialization): stamp the KB's tenant onto created models |
| 133 | 2026-09-17 | [42e6163981](https://github.com/Tencent/WeKnora/commit/42e61639819d934fea837d3849c7d0a8020c9cb9) | feat(session): fork conversations from a historical user message (#3304) |
| 134 | 2026-09-17 | [b3d3851ca1](https://github.com/Tencent/WeKnora/commit/b3d3851ca175d74c9adefd0719c44e92f6a0124a) | fix(chat): refine fork feedback and user message actions |
| 135 | 2026-09-17 | [8fbee673e6](https://github.com/Tencent/WeKnora/commit/8fbee673e69dc56b84de83df10965127e0e07ca5) | fix(browser): soften handoff UI and prevent preview overlay flicker |
| 136 | 2026-09-17 | [d645334d08](https://github.com/Tencent/WeKnora/commit/d645334d08e3fc6f99823347ad391927ccdc1bba) | refactor(frontend): share citation popover lifecycle |
| 137 | 2026-09-17 | [38be45c9d1](https://github.com/Tencent/WeKnora/commit/38be45c9d1c331ff27256fc078e75b3bacfe2410) | refactor: remove unused private Go helpers |
| 138 | 2026-09-17 | [10973ed321](https://github.com/Tencent/WeKnora/commit/10973ed321462c6f108bd95a5021bcc288f48a65) | test: isolate fork reset Git environment |
| 139 | 2026-09-17 | [5780cb0a71](https://github.com/Tencent/WeKnora/commit/5780cb0a71b114f57fc66f5683b9149cd672ab4c) | fix(sandbox): separate file access from artifact publishing (#3356) |
| 140 | 2026-09-17 | [820a14dafa](https://github.com/Tencent/WeKnora/commit/820a14dafa0ef9c6405baffebc70a3b631bfb95d) | refactor(datasource): share connector filename sanitization |
| 141 | 2026-09-17 | [559ad532ae](https://github.com/Tencent/WeKnora/commit/559ad532ae584fd4f2575f16f802deea04939630) | refactor(frontend): consolidate knowledge tag filters and remove stale styles |
| 142 | 2026-09-17 | [fa56b5b932](https://github.com/Tencent/WeKnora/commit/fa56b5b9322f039a3d4c2fb29f23a8119e63f705) | fix(docker): upgrade ParadeDB for CPUs without AVX2 |
| 143 | 2026-09-17 | [1c43934e53](https://github.com/Tencent/WeKnora/commit/1c43934e530d0d791a3104242dab4dc5f90025d9) | perf(sandbox): reduce redundant remote calls |
| 144 | 2026-09-17 | [ac4dd59712](https://github.com/Tencent/WeKnora/commit/ac4dd59712b652cbb5630979944808c13db767c3) | test(mcp): stabilize concurrent policy updates on SQLite |
| 145 | 2026-09-17 | [105c177b80](https://github.com/Tencent/WeKnora/commit/105c177b80070c95486bc6613b4a3bab0d91c74c) | feat(knowledge): make document summary generation optional |
| 146 | 2026-09-17 | [2514e42420](https://github.com/Tencent/WeKnora/commit/2514e42420a6f08c8eb9386ba4c65ee448603cd6) | docs: remove unsupported Homebrew installation guidance |
| 147 | 2026-09-17 | [68d99d7d21](https://github.com/Tencent/WeKnora/commit/68d99d7d2111303b70dce464c43d088aed7e873d) | perf(chunk): speed up chunk ingestion and batch updates (#3369) |
| 148 | 2026-09-17 | [182f67bd86](https://github.com/Tencent/WeKnora/commit/182f67bd86c08840c91670f38290903c42d8348f) | refactor(modelcontext): consolidate source_refs parsing, fixed data_analysis table, and identifier leak report (#3372) |
| 149 | 2026-09-17 | [5574ebc011](https://github.com/Tencent/WeKnora/commit/5574ebc0118e862ee1519d3700b8e73eb5b13c5b) | feat(knowledge): derive knowledge base descriptions from document profiles (#3370) |
| 150 | 2026-09-17 | [76a19f14c4](https://github.com/Tencent/WeKnora/commit/76a19f14c4aec6450421cc3f1256771c4214ff9f) | feat(mcp): add built-in MCP server with per-workspace endpoints (#3371) |
| 151 | 2026-09-17 | [8b1a64004d](https://github.com/Tencent/WeKnora/commit/8b1a64004d4a19fa288c60453ce40839905e0f10) | feat(mcp): show the endpoint token as the drawer's second step |
| 152 | 2026-09-17 | [b4aa5bbe8d](https://github.com/Tencent/WeKnora/commit/b4aa5bbe8db01fcff7d20620702d9c82e00ed28d) | refactor(frontend): style consistency pass — shared shells, tokens, unified confirmations (#3376) |
| 153 | 2026-09-18 | [c8ebb9ed96](https://github.com/Tencent/WeKnora/commit/c8ebb9ed96fafbb02b8e933910579404ffad1e73) | fix(knowledge): recover stranded deleting rows and align the KB count (#3367) |
| 154 | 2026-09-09 | [e3ac49cace](https://github.com/Tencent/WeKnora/commit/e3ac49cace44b19f1c729994adffdc9423219dcf) | fix(agent): report failed tool executions as tool_result, not error |
| 155 | 2026-09-18 | [6db8fdd601](https://github.com/Tencent/WeKnora/commit/6db8fdd60172449839f128928a2303f652a253b2) | feat(artifacts): artifacts library page backed by a message_artifacts table (#3384) |
| 156 | 2026-09-18 | [92e7c0ab37](https://github.com/Tencent/WeKnora/commit/92e7c0ab37bb553df950b91680f162127c8d1282) | refactor(agent): consolidate knowledge retrieval tools into search_knowledge, read_document, list_documents (#3381) |
| 157 | 2026-09-18 | [a8aa19f721](https://github.com/Tencent/WeKnora/commit/a8aa19f721f4121e0ff68fa4355e67ff3a567c3b) | fix(frontend): scale upload timeout with payload size |
| 158 | 2026-09-18 | [ad5d6a8fc2](https://github.com/Tencent/WeKnora/commit/ad5d6a8fc2e7b75e075e637d98a141f331b1e0b7) | fix(agent): make read_document queries keyword-based and bound its pages |
| 159 | 2026-09-18 | [62bb9da270](https://github.com/Tencent/WeKnora/commit/62bb9da270f78c3db7af4642394cc53a440a8696) | fix(agent): score search_knowledge reranks with the document title |
| 160 | 2026-09-18 | [2a6e5dcbb2](https://github.com/Tencent/WeKnora/commit/2a6e5dcbb2c3c03d1fff67ab360aea262d119703) | fix(agent): allow keyword retry after search_knowledge rerank rejection |
| 161 | 2026-09-18 | [a9c85fd17e](https://github.com/Tencent/WeKnora/commit/a9c85fd17ed1f227b2fef7454e63408f05111963) | fix(frontend): keep pipeline offsets within chart bounds |
| 162 | 2026-09-18 | [9d508c1c7a](https://github.com/Tencent/WeKnora/commit/9d508c1c7a6a1f155189bd47e1f23a616735e67f) | feat(frontend): show overall progress for knowledge file uploads (#3399) |
| 163 | 2026-09-18 | [2be7bd4063](https://github.com/Tencent/WeKnora/commit/2be7bd40631dda1dd485306038f07a62e9ee287e) | fix(agent): search bound knowledge bases before answering, and pass suggested-question origins (#3400) |
| 164 | 2026-09-18 | [aaec920a9c](https://github.com/Tencent/WeKnora/commit/aaec920a9cf9c633f2affa276ab735d01f6886b8) | feat(audit): 知识库活动记录对 API Key 操作额外显示 Key 名称 (#3405) |
| 165 | 2026-09-18 | [68e6c03dc4](https://github.com/Tencent/WeKnora/commit/68e6c03dc44e7faa2f807f70384ea5fc5534b0d7) | fix(skills): upgrade outdated sandbox skills without rolling back the catalog or taking them offline (#3406) |
| 166 | 2026-09-18 | [62f7da5eea](https://github.com/Tencent/WeKnora/commit/62f7da5eea3c52fd1887bdb0ef114562ce924a98) | feat(agent): persist compaction checkpoints across turns (#3407) |
| 167 | 2026-09-18 | [473c8726bb](https://github.com/Tencent/WeKnora/commit/473c8726bbdd743040e94f4a13b592091efd35c8) | fix(chunking): 修复内联 HTML 表格与保护模式导致的超长 chunk (#3404) |
| 168 | 2026-09-18 | [2a40cc37c0](https://github.com/Tencent/WeKnora/commit/2a40cc37c016776a9399301ab57a8fbc33264829) | fix(migrations): renumber the context checkpoint migration to 000105 |
| 169 | 2026-09-18 | [db8e7611be](https://github.com/Tencent/WeKnora/commit/db8e7611be7614e9cf625bea3146ab16a73d58bb) | fix(security): close cross-workspace access paths in shared spaces (#3408) |
| 170 | 2026-09-18 | [03d9d66a1f](https://github.com/Tencent/WeKnora/commit/03d9d66a1ff90022a5855f0e00df216a915bf9b6) | feat(agent): size agent history by the window and calibrate token estimates (#3413) |
| 171 | 2026-09-18 | [b82fd6c0f9](https://github.com/Tencent/WeKnora/commit/b82fd6c0f9939e8497fa7474999dd1bde5d0d88d) | fix(security): close the low-severity shared-space findings (#3416) |
| 172 | 2026-09-19 | [2a6a9c2517](https://github.com/Tencent/WeKnora/commit/2a6a9c251734d12bf65bf17d8f3412689ebed4e1) | fix(ui): 优化知识库文档交互与对话布局 (#3420) |
| 173 | 2026-09-19 | [a46a3c5996](https://github.com/Tencent/WeKnora/commit/a46a3c5996785fd7d3713a650e21d02ed7517710) | fix(ui): 优化对话布局、导航与产物和引用展示 (#3427) |
| 174 | 2026-09-19 | [aea51ecb62](https://github.com/Tencent/WeKnora/commit/aea51ecb6250eee5119ae4583673cbe22f5df568) | fix(wiki): batch taxonomy embeddings instead of one request per set |
| 175 | 2026-09-19 | [ca88fe74ae](https://github.com/Tencent/WeKnora/commit/ca88fe74aebe86d656116a9c42204f74b87cc2f0) | fix(embedding): preserve batching and concurrency limits for WeKnoraCloud |
| 176 | 2026-09-19 | [eec236b27f](https://github.com/Tencent/WeKnora/commit/eec236b27f9613d264be3a52b74c8f0322aca20c) | fix(ui): 优化资源卡片、标签分组与共享空间创建布局 |
| 177 | 2026-09-19 | [e986579731](https://github.com/Tencent/WeKnora/commit/e986579731c576f1c1894f2b330d0d6513fa7e7b) | refactor(ui): 清理资源列表废弃组件和无效样式 |
| 178 | 2026-09-19 | [014f95f139](https://github.com/Tencent/WeKnora/commit/014f95f1394a7bfd1c7bbbdcdc19057a80dcaca5) | fix(ui): 统一资源图标并调整会话分叉标识 (#3430) |
| 179 | 2026-09-19 | [a5bdf82a02](https://github.com/Tencent/WeKnora/commit/a5bdf82a02f831047b5945c26528ed1a85587cbb) | fix(storage): let object storage transfers outlive 30 seconds (#3368) |
| 180 | 2026-09-17 | [1b36030502](https://github.com/Tencent/WeKnora/commit/1b360305028fb90d328f689525a05adc8e94110b) | perf(chat): replay a stored stream a segment at a time |
| 181 | 2026-09-15 | [d3ed78313f](https://github.com/Tencent/WeKnora/commit/d3ed78313faf5d852f2a94615c78655736df9fdb) | fix(compose): target configured database in dev healthcheck |
| 182 | 2026-09-15 | [fca1cd06f8](https://github.com/Tencent/WeKnora/commit/fca1cd06f818f65d7776114ac2a697315af17efd) | fix(ima): traverse subfolders using list response media IDs |
| 183 | 2026-09-19 | [7ea5bf4349](https://github.com/Tencent/WeKnora/commit/7ea5bf4349509a352b2cf9584a4ada3316c074b3) | fix(image): stop appending KB custom instructions to the OCR prompt OCR提示词不再错误的附加图片解析用户自定义提示词 (#3291) |
| 184 | 2026-09-19 | [640e356e31](https://github.com/Tencent/WeKnora/commit/640e356e3117aa9f9fad982294fd8ee45c36edb8) | fix(browser): rebase BrowserSkill patches and use native session lifecycle |
| 185 | 2026-09-20 | [5c5d46d9fb](https://github.com/Tencent/WeKnora/commit/5c5d46d9fb556a7efa19d1a23331f4ce83ac2031) | fix(ui): add resource list title icons and highlight shared filter |
| 186 | 2026-09-20 | [562b131e6c](https://github.com/Tencent/WeKnora/commit/562b131e6c6c27f9dac2e9d1ee24a334ad1959a1) | fix(qdrant): treat a missing dimension collection as an empty delete target (#3419) |
| 187 | 2026-09-20 | [bcfb6dfc12](https://github.com/Tencent/WeKnora/commit/bcfb6dfc125374f455e6bea5024781339d5402db) | fix(storage): keep the SSRF redirect policy on the KS3 client (#3440) |
| 188 | 2026-09-20 | [68e21b6e74](https://github.com/Tencent/WeKnora/commit/68e21b6e74733b801fad6db1f3a2c4365a1d34bf) | docs(env): warn against SEARXNG_PORT colliding with APP_PORT (#3385) |
| 189 | 2026-09-20 | [110134f187](https://github.com/Tencent/WeKnora/commit/110134f1878c15e8f06889366769e94259fe26b4) | test(im): close the lifecycle test database and keep its name unique (#3433) |
| 190 | 2026-09-20 | [1e4389bb21](https://github.com/Tencent/WeKnora/commit/1e4389bb218639b4e5e0ac02fd60aa675024bbb4) | fix(wiki): clarify alias matching rules in deduplication prompt (#3418) |
| 191 | 2026-09-20 | [3017145102](https://github.com/Tencent/WeKnora/commit/3017145102a7a463b02c9eb4b2a5519ccebbb1ae) | fix(frontend): stop seeding new answer events with prior-round text (#3386) |
| 192 | 2026-09-20 | [a163bab32f](https://github.com/Tencent/WeKnora/commit/a163bab32f75b19d5b769a4d3795c539b032b089) | feat(knowledge): rework the parse timeline panel's layout and failure card |
| 193 | 2026-09-20 | [9e0c28fbc9](https://github.com/Tencent/WeKnora/commit/9e0c28fbc93f4b3abf982a0050a1cd5372cba72a) | feat(artifacts): let users delete generated files (#3447) |
| 194 | 2026-09-20 | [eaca7364cd](https://github.com/Tencent/WeKnora/commit/eaca7364cd776f81a39683b6f767428bb37146c4) | fix(knowledge): stop two parse paths from stranding a row in processing (#3445) |
| 195 | 2026-09-20 | [093c16cbb3](https://github.com/Tencent/WeKnora/commit/093c16cbb36725baf2388bcb4fd719c9f929a6c8) | fix(agent): end the turn when the completion cap truncates an answer (#3449) |
| 196 | 2026-09-20 | [a62a51e6b0](https://github.com/Tencent/WeKnora/commit/a62a51e6b0225ebd3edd1b6d56c1ea80efa54ba1) | docs(db): stop claiming MySQL is a supported DB_DRIVER |
| 197 | 2026-09-20 | [7fa6f028fc](https://github.com/Tencent/WeKnora/commit/7fa6f028fcea9e08720db1f565ec4911ea0e7aa6) | feat(chat): rework the saved-answer Markdown editor (#3467) |
| 198 | 2026-09-20 | [9b0b0910ba](https://github.com/Tencent/WeKnora/commit/9b0b0910ba5aa991d4d35797d2b6c20db0f08770) | ci(lint): filter new issues from git instead of the PR patch API |
| 199 | 2026-09-20 | [d4b9fd0c72](https://github.com/Tencent/WeKnora/commit/d4b9fd0c72e3ac0d0cb831300ac378d45d988284) | feat(session): in-place rewind with workspace checkpoint reset (#3469) |
| 200 | 2026-09-20 | [50bbf9ad5f](https://github.com/Tencent/WeKnora/commit/50bbf9ad5f7a504471978de841f7a8f1bd5e5d81) | fix(embedding): two wire-level bugs in the remote embedders (#3484) |
| 201 | 2026-09-20 | [c40a9dd194](https://github.com/Tencent/WeKnora/commit/c40a9dd1940f85d85a59787bd532eb0f290cb176) | refactor(models): rebuild vendor integration on protocol + catalog + vendor layers (#3470) |
| 202 | 2026-09-21 | [9aa63c866f](https://github.com/Tencent/WeKnora/commit/9aa63c866ff4570767110a6c7d3cf477a4ab955b) | fix(models): correct four vendor request shapes that contradict the docs (#3489) |
| 203 | 2026-09-21 | [d6c7665cb5](https://github.com/Tencent/WeKnora/commit/d6c7665cb581908ebdd236b3b53596b97a63464e) | refactor(rerank): move the wire formats into protocol packages, and correct the vendor facts (#3488) |
| 204 | 2026-09-21 | [9030f556f2](https://github.com/Tencent/WeKnora/commit/9030f556f2daaf67d9a1387e26aa01450674b75d) | refactor(embedding): move the wire formats into protocol packages, and correct the vendor facts (#3500) |
| 205 | 2026-09-21 | [13d0b9f94a](https://github.com/Tencent/WeKnora/commit/13d0b9f94a7e92971e22925d6d51a6ac7030ec81) | fix(queue): order ClaimBatch by fail_count so retries cannot starve new work (#3444) |
| 206 | 2026-09-21 | [a3c7be10ca](https://github.com/Tencent/WeKnora/commit/a3c7be10ca6f4133dc4a47396a4e1567466058f5) | fix(agent): scope the @Skill and @MCP pickers to the agent that will run (#3475) |
| 207 | 2026-09-21 | [2f2bcc7aa7](https://github.com/Tencent/WeKnora/commit/2f2bcc7aa7e82dab0b29501c39869cce55f0d07f) | fix(suggestion): avoid echoing the current query (#2582) |
| 208 | 2026-09-21 | [e3d52e8353](https://github.com/Tencent/WeKnora/commit/e3d52e83530126780696090d06fa0777160594cd) | fix(wiki): skip ambiguous single-Han auto-links (#2860) |
| 209 | 2026-09-21 | [07a5af673e](https://github.com/Tencent/WeKnora/commit/07a5af673ec551eec3fd8da0abc7b766d5d4d3d5) | fix(graph): increase entity extraction output budget (#2859) |
| 210 | 2026-09-21 | [07e6e2984a](https://github.com/Tencent/WeKnora/commit/07e6e2984a497d93d77f1dd6610e7c8ff46b1162) | fix(knowledge): hide mutation controls on read-only shared knowledge bases (#3103) |
| 211 | 2026-09-08 | [5036062355](https://github.com/Tencent/WeKnora/commit/5036062355da9099976c0b27a139d51f34a4c19b) | fix(wiki): degrade gracefully when revision snapshot is missing |
| 212 | 2026-09-21 | [911a1490ea](https://github.com/Tencent/WeKnora/commit/911a1490ead204ace609a98e4df247543177cdc4) | fix(wiki): enrich cited chunk content with image captions in Reduce phase 实体/摘要页面引用子块时同样使用富文本内容 (#3292) |
| 213 | 2026-09-20 | [2c5f90b78e](https://github.com/Tencent/WeKnora/commit/2c5f90b78e2c44cd16d1ae2d8f7cf7eb98b1ec40) | chore: remove empty root package-lock.json |
| 214 | 2026-09-21 | [7d688cb8d3](https://github.com/Tencent/WeKnora/commit/7d688cb8d360f51ca68e918f4769ff1a20e9202d) | feat(localsandbox): add host OS sandbox core and macOS Seatbelt backend (#3472) |
| 215 | 2026-09-21 | [4a9aa06274](https://github.com/Tencent/WeKnora/commit/4a9aa06274e69f8a12a809210a67f833323c90be) | refactor(asr): move the wire format into protocol packages, and cover every vendor whose docs offer ASR (#3503) |
| 216 | 2026-09-21 | [f307d3dc94](https://github.com/Tencent/WeKnora/commit/f307d3dc942cd636f5848804cf2b96e75e29ef73) | fix: remove temp file when TempFileContext write fails inside __enter__ (#3495) |
| 217 | 2026-09-21 | [cab16dcacb](https://github.com/Tencent/WeKnora/commit/cab16dcacb39f335f8396d29758cc01f70654bd5) | fix(storage): strip storage-backend scope in local provider path normalization (#2979) |
| 218 | 2026-09-21 | [2d37f8fd7d](https://github.com/Tencent/WeKnora/commit/2d37f8fd7d18c7a5871a9aa852c41bcb8eddcf2d) | fix(container): wire shared agent lookup for MCP handler |
| 219 | 2026-09-21 | [41dde3b05f](https://github.com/Tencent/WeKnora/commit/41dde3b05f2ebe68f30d4b0438bf38af354cf0f5) | refactor(models): delete the internal/models/provider façade (#3506) |
| 220 | 2026-09-21 | [56d075e213](https://github.com/Tencent/WeKnora/commit/56d075e213da57e6194d392907717e0eb7fe1f5b) | fix(sandbox): resolve a session's pinned sandbox in the workspace that owns it (#3471) |
| 221 | 2026-09-21 | [cc82e75cb5](https://github.com/Tencent/WeKnora/commit/cc82e75cb5ca5ac0fc66386917810519054870f2) | fix(parser): keep PDFs on builtin instead of defaulting to anydoc |
| 222 | 2026-09-21 | [4640a1ea0c](https://github.com/Tencent/WeKnora/commit/4640a1ea0cf73fe84c56da7b48ea6a44514a0fd7) | test(rerank): pin every vendor's request, and correct three facts the table surfaced (#3513) |
| 223 | 2026-09-21 | [da049f04e0](https://github.com/Tencent/WeKnora/commit/da049f04e0ac4b685109ad6a1428e0dd90cd51b8) | fix(browserskill): release connection capacity when members disconnect |
| 224 | 2026-09-21 | [58260d192c](https://github.com/Tencent/WeKnora/commit/58260d192c406374b565ebbd21fd4dfb49ade543) | fix(upload): accept XMind files before parser engine discovery |
| 225 | 2026-09-22 | [33c0333ec9](https://github.com/Tencent/WeKnora/commit/33c0333ec9474a6d090bab346e1d631a842bafaf) | fix(chat): restore shared workspace images after stream completion (#3523) |
| 226 | 2026-09-22 | [e55e342387](https://github.com/Tencent/WeKnora/commit/e55e34238799e5ee2b5ca7fa1f4b27512afb6d4e) | fix(skill): count create-time sandbox env vars in the required check |
| 227 | 2026-09-22 | [7feac9c570](https://github.com/Tencent/WeKnora/commit/7feac9c57081a6f60a698df369b745380f75066d) | fix(mcp): resolve MCP endpoint sessions by their storage ID |
| 228 | 2026-09-21 | [afeee71e48](https://github.com/Tencent/WeKnora/commit/afeee71e4807d9e32ab5c08c91c403d3cabeff32) | fix(wiki): release inflight slot on task cancellation |
| 229 | 2026-09-22 | [733fdd69a6](https://github.com/Tencent/WeKnora/commit/733fdd69a6515fb518919629eea5bda5a226c227) | fix(mcp): proxy server endpoints through frontend gateways |
| 230 | 2026-09-14 | [d1691cbc26](https://github.com/Tencent/WeKnora/commit/d1691cbc26235c3f16ac35b9f6711367eac2cf46) | fix(upload): clean multipart temporary files |
| 231 | 2026-09-20 | [6da4b3acc7](https://github.com/Tencent/WeKnora/commit/6da4b3acc73d95ac2dab44cf0ad91649e69c9fa0) | docs: link pgvector issues in retrieval engine guide |
| 232 | 2026-09-22 | [3a8ce97c29](https://github.com/Tencent/WeKnora/commit/3a8ce97c29e42387e64b67da77df5c1923eaa5c1) | fix(docker): pull MinIO image from quay.io |
| 233 | 2026-09-21 | [2dd36212d6](https://github.com/Tencent/WeKnora/commit/2dd36212d63503117d57567bd6e453f0b1cff1a1) | fix(knowledge): bind extracted images to their document and unscope shared-KB parent lookups |
| 234 | 2026-09-21 | [4b0dc3b687](https://github.com/Tencent/WeKnora/commit/4b0dc3b68792089ca4856b6cef33127466b844d8) | fix: satisfy revive unused-parameter and gofumpt in #3493 review findings |
| 235 | 2026-09-20 | [0b4cd706c2](https://github.com/Tencent/WeKnora/commit/0b4cd706c2405b193b8e27593f304edff7f4fd2c) | fix(knowledge): stop synthesizing never-run stages as failed |
| 236 | 2026-09-20 | [5f1b7a2b3b](https://github.com/Tencent/WeKnora/commit/5f1b7a2b3bdff2a68a8db99775a1cfec096f507d) | style: trim the comments added in this branch to the repo's cap |
| 237 | 2026-09-22 | [22cbcdfa57](https://github.com/Tencent/WeKnora/commit/22cbcdfa5789b33dbb8d47532c21b24da2fdf5aa) | feat(sandbox): run unpinned Lite sessions on the host OS sandbox (#3549) |
| 238 | 2026-09-21 | [ae185dd1a9](https://github.com/Tencent/WeKnora/commit/ae185dd1a94021e486ccbf965be980bf0e8437b5) | fix(docreader): stop a non-ASCII base64 payload from failing the whole parse |
| 239 | 2026-09-22 | [6c4f86becb](https://github.com/Tencent/WeKnora/commit/6c4f86becb6fe53ac454f343c7bcfb567657f306) | docs: point the RAG pipeline and agent docs at internal/modelcontext |
| 240 | 2026-09-21 | [78d7d862e0](https://github.com/Tencent/WeKnora/commit/78d7d862e07690e12ec439f39a45e66540958378) | fix(chat): surface truncated quick answers |
| 241 | 2026-09-22 | [8783a993ef](https://github.com/Tencent/WeKnora/commit/8783a993efd52479e266cbb32883d9a4dc1b530f) | fix(helm): point the MinIO image at quay.io |
| 242 | 2026-09-16 | [273db881c6](https://github.com/Tencent/WeKnora/commit/273db881c67176101a631e76507c7dc903f50bbc) | fix(mcp): assemble multiline SSE events before parsing |
| 243 | 2026-09-22 | [af3ac25bc1](https://github.com/Tencent/WeKnora/commit/af3ac25bc1370c6cbfdef92917cd6d713bdfa527) | fix(wiki): release documents a stranded wiki queue holds in finalizing |
| 244 | 2026-09-22 | [d951cb78d1](https://github.com/Tencent/WeKnora/commit/d951cb78d1329085f877dd256869b45a36aa6b7f) | fix(knowledge): stop pipeline errors from stranding rows in processing |
| 245 | 2026-09-22 | [0a99a2850b](https://github.com/Tencent/WeKnora/commit/0a99a2850baaef2b2f1cf3c114815202b6db1914) | fix(knowledge): address review of the stuck-processing fixes |
| 246 | 2026-09-22 | [6afb6ca9d5](https://github.com/Tencent/WeKnora/commit/6afb6ca9d541a5ad630895d8f20b40d3b3ca5518) | feat(knowledge): flag documents that stop making progress (#3557) |
| 247 | 2026-09-22 | [b57c362667](https://github.com/Tencent/WeKnora/commit/b57c362667e2436645e29066c9d91fd716a2e58b) | fix(chat): show full date for older message timestamps (#3576) |
| 248 | 2026-09-22 | [d69df50b7a](https://github.com/Tencent/WeKnora/commit/d69df50b7a879b067707cab50a504069192fef40) | fix(chat): align knowledge citation hover with outline cards |
| 249 | 2026-09-22 | [fbf9f54b6a](https://github.com/Tencent/WeKnora/commit/fbf9f54b6a1102c50b48b3437c93ffbe57f6ccf1) | fix(chat): prevent overscroll bounce from moving composer |
| 250 | 2026-09-22 | [63ef76dd80](https://github.com/Tencent/WeKnora/commit/63ef76dd8096109582855e9cbeed1763f961aaf2) | docs: 迁移有效文档至 website-docs 并清理旧文档 (#3580) |
| 251 | 2026-09-22 | [4afda18ec1](https://github.com/Tencent/WeKnora/commit/4afda18ec1ae3a166ab75cbb9b0c270f155929d0) | fix(chat): keep composer fixed while messages overscroll |
| 252 | 2026-09-23 | [a89a5a2c43](https://github.com/Tencent/WeKnora/commit/a89a5a2c43dec80df19ee79dc9412122048afe71) | fix(browser): use unmodified BrowserSkill upstream and remove downstream patches (#3563) |
| 253 | 2026-09-23 | [b8f081f03a](https://github.com/Tencent/WeKnora/commit/b8f081f03ae4a841777dbb71f47b199b3e3fcca7) | fix(browser): refine connection and preview UI with extension version |
| 254 | 2026-09-23 | [5279972eec](https://github.com/Tencent/WeKnora/commit/5279972eec845bc4d9e2460bdc53d5b3948d5a1d) | fix(agent): give local_browser one text cap and per-method signatures |
| 255 | 2026-09-22 | [407b184081](https://github.com/Tencent/WeKnora/commit/407b18408172f36412013fcb5e4d2b6e9df98172) | fix(chat): report truncated model fallback answers |
| 256 | 2026-09-23 | [8b059331b6](https://github.com/Tencent/WeKnora/commit/8b059331b6d5f2224ce659820469850b6472c36a) | fix(graph): 给 Neo4j SearchNode 加上种子上限与行数兜底，避免一次查询拉回整个邻域 (#3550) |
| 257 | 2026-09-23 | [ee3623ab76](https://github.com/Tencent/WeKnora/commit/ee3623ab76e1e9ca61b9d86117faf977869b37b6) | fix(wiki): reuse cross-type page for same-title items instead of creating twins (#3547) |
| 258 | 2026-09-23 | [402729341a](https://github.com/Tencent/WeKnora/commit/402729341a4456701277ebb40d307cee689f664d) | fix(docparser): make PaddleOCR-VL request timeout configurable (#3537) |
| 259 | 2026-09-23 | [ad20ea9c9b](https://github.com/Tencent/WeKnora/commit/ad20ea9c9bbcecba3ef8cb7d44fb64bcfa88c7f7) | feat: add configurable sorting to document and resource lists (#3318) |
| 260 | 2026-09-23 | [b86821f551](https://github.com/Tencent/WeKnora/commit/b86821f551704233640f067c46a1d5ce498bd643) | feat(im): add per-channel reply locale (#3455) |
| 261 | 2026-09-23 | [2c6ae0c262](https://github.com/Tencent/WeKnora/commit/2c6ae0c262fa68711431fe0ccebfeb9aa12792a6) | fix(session): reject invalid QA queries before session lookup (#3477) |
| 262 | 2026-09-23 | [171097f06e](https://github.com/Tencent/WeKnora/commit/171097f06e613acde5c01db527789278957795a2) | feat(security): add a whitelist-only outbound mode that refuses before DNS (#3393) |
| 263 | 2026-09-23 | [acd1a2a5a4](https://github.com/Tencent/WeKnora/commit/acd1a2a5a4919e2b58f039ed9861a3b5fe6518a7) | docs(ollama): document embedding model config and RAM guidance (#3531) |
| 264 | 2026-09-23 | [06ae6bcf04](https://github.com/Tencent/WeKnora/commit/06ae6bcf04e5ce3543c050383f7af8a033abf317) | fix(docreader): kill timed-out converter process groups (#2990) |
| 265 | 2026-09-23 | [23832fabad](https://github.com/Tencent/WeKnora/commit/23832fabad7b0a0fac98360a6d12825032f2fc76) | feat(lite): let new chats bind a host project directory (#3554) |
| 266 | 2026-09-23 | [c57766cc76](https://github.com/Tencent/WeKnora/commit/c57766cc76061acb4fdd2668ca0eb5e76aa4d015) | fix(tenant): stop wiki tasks of soft-deleted tenants from issuing model requests (#3602) |
| 267 | 2026-09-23 | [653307b5c7](https://github.com/Tencent/WeKnora/commit/653307b5c72d90aa6e43ac2f43bdcf859d14d0ad) | fix(wiki): 整页重写被补全预算截断时续写，续写用尽则拒绝写回（长名单 146→119） (#3582) |
| 268 | 2026-09-23 | [e179f44445](https://github.com/Tencent/WeKnora/commit/e179f4444592929f9e95be98713883d6c14a7677) | refactor(models): separate provider definitions, model catalogs, and runtime configuration (#3589) |
| 269 | 2026-09-23 | [02734d7499](https://github.com/Tencent/WeKnora/commit/02734d74991adb5d55f2eac2f4333e7158c7f621) | feat(chat): add session reasoning controls and polish model configuration |
| 270 | 2026-09-23 | [fd016df98c](https://github.com/Tencent/WeKnora/commit/fd016df98c23f220511bd195abc7f0f6485e8f57) | feat(toolbox): introduce toolbox feature with skills, MCP, and browser connection management |
| 271 | 2026-09-23 | [2e801677a0](https://github.com/Tencent/WeKnora/commit/2e801677a03449d112ec542529fc0356acd25016) | fix(frontend): use radius and motion tokens in toolbox styles |
| 272 | 2026-09-23 | [e6e41ed301](https://github.com/Tencent/WeKnora/commit/e6e41ed3011ca49a2a440a1759bbc9ae11e234d6) | feat(browser-connection): add sidebar connection status and localization support |
| 273 | 2026-09-23 | [1fa915d712](https://github.com/Tencent/WeKnora/commit/1fa915d712288902da7841dc8ff355a56e37a013) | chore: update BrowserSkill extension version to 0.3.1 and enhance localization support |
| 274 | 2026-09-23 | [11edd16819](https://github.com/Tencent/WeKnora/commit/11edd16819b2bc7ed90bb2b07fa7122c04159085) | fix(datasource): fail closed on GitLab compare errors |
| 275 | 2026-09-24 | [4a026fc896](https://github.com/Tencent/WeKnora/commit/4a026fc89668b50deed399dcafd4c94eb991e1a3) | fix(docs): update remaining references to migrated documentation |
| 276 | 2026-09-23 | [608027179f](https://github.com/Tencent/WeKnora/commit/608027179fb03b8cf31ee376ef1c624ff3dc6eec) | fix(gitlab): 补齐项目资源列表分页 |
| 277 | 2026-09-24 | [be46eabaf2](https://github.com/Tencent/WeKnora/commit/be46eabaf2eca2d5548d3afd5db562ef322d5575) | docs: sync website-docs, homepage and changelog for v0.8.2 (#3632) |
| 278 | 2026-09-24 | [e87607e852](https://github.com/Tencent/WeKnora/commit/e87607e8525a2f7dbe02e082c7f4a6992d1df3fd) | fix(mcp): expose wiki output in structured content (#3590) |
| 279 | 2026-09-24 | [2e712a97a1](https://github.com/Tencent/WeKnora/commit/2e712a97a10900c82bed74993362823a7c77254f) | fix(mcp): return accessible image URLs from existing tools (#3650) |
| 280 | 2026-09-24 | [a1b0d0cb6c](https://github.com/Tencent/WeKnora/commit/a1b0d0cb6c5fcc2c1f37eeab495908d8fae540c6) | fix(models): serve qwen3-rerank on the native DashScope rerank protocol |
| 281 | 2026-09-24 | [eb888d0d3d](https://github.com/Tencent/WeKnora/commit/eb888d0d3dfe44e2fd9547f6cf930e6a8bdb28fd) | fix(confluence): stop Server listings that link past the last page |
| 282 | 2026-09-24 | [5468e7c191](https://github.com/Tencent/WeKnora/commit/5468e7c191ac6a2eb2b587d7e8b95f59f70cb08a) | chore: bump version to 0.8.2 |
| 283 | 2026-09-24 | [9f57bbe4da](https://github.com/Tencent/WeKnora/commit/9f57bbe4da2ed0591115bad878b43387424ae5b0) | docs: date the v0.8.2 release and describe it in the READMEs |
| 284 | 2026-09-24 | [503f1abfb4](https://github.com/Tencent/WeKnora/commit/503f1abfb4e76df42a868655237c290a91195cd4) | fix(confluence): confirm unlisted pages before deleting after a cut-short listing |
| 285 | 2026-09-24 | [3e8b0bfc80](https://github.com/Tencent/WeKnora/commit/3e8b0bfc80b845b2d4b2ed683994748741450a97) | docs: describe the Lite host sandbox workspace choices accurately in the READMEs |
| 286 | 2026-09-24 | [cee697920e](https://github.com/Tencent/WeKnora/commit/cee697920e0aebf067b57cf9612aa330863b17c3) | fix(website-docs): serve everything but the homepage from /docs/ |
| 287 | 2026-09-24 | [b6b32c24e1](https://github.com/Tencent/WeKnora/commit/b6b32c24e152650cf94083c21195fe8bd24328f5) | test(website-docs): update integration test for the v0.8.2 header and release gallery |
| 288 | 2026-09-23 | [8d40150149](https://github.com/Tencent/WeKnora/commit/8d40150149ce00a952a98a10a80226ad6584abd5) | fix(session): keep auto-generated titles plain text |
| 289 | 2026-09-24 | [967ed09706](https://github.com/Tencent/WeKnora/commit/967ed097062c61723045ae103b665982728dc572) | feat(models): add system admin model catalog console (#3662) |
| 290 | 2026-09-24 | [eebc1d8643](https://github.com/Tencent/WeKnora/commit/eebc1d86436e9cd7b9f6d24c1f973527f318853f) | fix(ui): restore inner padding on form popups (#3673) |
| 291 | 2026-09-24 | [e6284f4eac](https://github.com/Tencent/WeKnora/commit/e6284f4eacc89a4b2f1ef2469e9f86241e2841cd) | fix(browser): let shared-space agents drive the caller's local browser |
| 292 | 2026-09-24 | [5d00985f72](https://github.com/Tencent/WeKnora/commit/5d00985f720440f88c88ccdf8f65232175d885d5) | docs: restructure the English and Chinese READMEs |
| 293 | 2026-09-24 | [44a1df6046](https://github.com/Tencent/WeKnora/commit/44a1df60460fc06fba9eed9022d57164605a487c) | docs(readme): website-matched design for all four READMEs (#3679) |
| 294 | 2026-09-24 | [65d8129cb6](https://github.com/Tencent/WeKnora/commit/65d8129cb64f4ccc23a7b73896ebe424d31ba228) | fix(retriever/es): keep vectors and enabled state on CopyIndices (#3678) |
| 295 | 2026-09-24 | [3cf0be2e97](https://github.com/Tencent/WeKnora/commit/3cf0be2e9745b97dd903feec1c66a936780237d8) | fix(embed): warn at startup when no signing key is configured |
| 296 | 2026-09-24 | [e22a8367ef](https://github.com/Tencent/WeKnora/commit/e22a8367efc9802353e6981e162226459a998404) | feat(docparser): support MinerU 4.0 V1 API for self-hosted MinerU |
| 297 | 2026-09-24 | [9253b2c1dc](https://github.com/Tencent/WeKnora/commit/9253b2c1dc8f4fe83fec8e66ba79dbf00144fd43) | fix(frontend): link docs to the public docs site instead of GitHub |
| 298 | 2026-09-24 | [9fc08c2079](https://github.com/Tencent/WeKnora/commit/9fc08c2079d89173c17c61356dee3838772931c5) | feat(retrieval): unify rerank across search APIs and expose diagnostics (#3675) |
| 299 | 2026-09-24 | [65d8a34ead](https://github.com/Tencent/WeKnora/commit/65d8a34ead4c99991d450ecc8c0acdbcf5b0e583) | fix(desktop): pair the Lite browser against the local API listener (#3677) |
| 300 | 2026-09-24 | [c029b50ef9](https://github.com/Tencent/WeKnora/commit/c029b50ef928d5c21ffb1922f6d541b0d27bb7a4) | fix(rss): 拒绝超限响应以避免正文静默截断 |
| 301 | 2026-09-24 | [7e82ee33dc](https://github.com/Tencent/WeKnora/commit/7e82ee33dcefc398ca455775700bdf8730054aa5) | docs(misc): add intro video source |
| 302 | 2026-09-24 | [e2433b1293](https://github.com/Tencent/WeKnora/commit/e2433b1293fa350d972902cd3676a1aa9a9c9aac) | feat(lite): install and run skills on the host (#3661) |
| 303 | 2026-09-25 | [e896270c41](https://github.com/Tencent/WeKnora/commit/e896270c418c01d230c9c6d11f05d0148677892a) | docs(readme): surface agent tools in the overview and condense feature tables (#3688) |
| 304 | 2026-09-25 | [04adc13272](https://github.com/Tencent/WeKnora/commit/04adc132726d93c9247898af79ad8f3b542df4f0) | fix(retrieval): fix recall correctness, score scales, cost and agent tools (#3687) |
| 305 | 2026-09-25 | [68fe9599cd](https://github.com/Tencent/WeKnora/commit/68fe9599cdb853eb1205499d44ed176613f875df) | feat(gallery): add an image gallery tab with schema-driven filters (#3694) |
| 306 | 2026-09-25 | [ebbcf96fab](https://github.com/Tencent/WeKnora/commit/ebbcf96fabb651c091c3ee2272d802b80c100c4e) | feat(gallery): rework the gallery toolbar, header tabs and image viewer |
| 307 | 2026-09-25 | [1bf717a54d](https://github.com/Tencent/WeKnora/commit/1bf717a54d4413e0e5cfe5e2b7ff008489b97c04) | feat(chat): open cited documents at the cited passage (#3700) |
| 308 | 2026-09-25 | [ab115c0944](https://github.com/Tencent/WeKnora/commit/ab115c0944b6e7f3c5d976970b67cb6070f6db1b) | fix(knowledge): stop parses from getting stuck in processing or finalizing (#3696) |
| 309 | 2026-09-24 | [8619bdf3c5](https://github.com/Tencent/WeKnora/commit/8619bdf3c5149acc8ca51263e4d485f97da39211) | docs(readme): swap in the new intro video |
| 310 | 2026-09-25 | [ef6f28c1ca](https://github.com/Tencent/WeKnora/commit/ef6f28c1cafec72bd05fcb4fbd3feb8144f93e3d) | fix(agent): recover tool panics instead of crashing the server |
| 311 | 2026-09-25 | [8356742dd5](https://github.com/Tencent/WeKnora/commit/8356742dd5a75cc84a5f18a527665eac4abf5bea) | fix(llm): treat cut-off streams as broken and tighten LLM retries |
| 312 | 2026-09-25 | [ef938f9114](https://github.com/Tencent/WeKnora/commit/ef938f9114cc82e799c56b5bde2b3aca9bc3832c) | fix(logger): keep LLM debug records inside their directory |
| 313 | 2026-09-25 | [8a75bece6d](https://github.com/Tencent/WeKnora/commit/8a75bece6d491a21e8b7b4c6171a8151ad77da9b) | fix(mcp): close MCP connections on shutdown without blocking exit |
| 314 | 2026-09-25 | [27b0c84bfd](https://github.com/Tencent/WeKnora/commit/27b0c84bfdb2641c66942aea4620f44b328096e1) | fix(docparser): make the MinerU parse timeouts configurable |
| 315 | 2026-09-25 | [929dad5747](https://github.com/Tencent/WeKnora/commit/929dad5747fb24e3cc85ec86eb1f4b4a3a497441) | fix(weaviate): accept a class another worker created first (#3689) |
| 316 | 2026-09-25 | [605d70fc52](https://github.com/Tencent/WeKnora/commit/605d70fc528a948591b647b0b805e7bfd617d4be) | fix(notion): 在资源发现失败时中止增量同步 (#3682) |
| 317 | 2026-09-25 | [5614c04d88](https://github.com/Tencent/WeKnora/commit/5614c04d88f2c28620a6bd3e16fc1093b9c54539) | fix(feishu): stop Wiki pagination on repeated page tokens (#3645) |
| 318 | 2026-09-25 | [6eb3cce164](https://github.com/Tencent/WeKnora/commit/6eb3cce164f6b1efe9c3a02973502c93f7d45126) | fix(agent): 合并相邻 assistant 消息时保留 tool_calls，避免孤儿 tool 消息导致整轮 400 (#3574) |
| 319 | 2026-09-25 | [e5a301ed3a](https://github.com/Tencent/WeKnora/commit/e5a301ed3ad156d7139231b43f7664e05eea6806) | fix(files): accept canonical storage:// paths in bare provider services (#3166) |
| 320 | 2026-09-25 | [5c717ca045](https://github.com/Tencent/WeKnora/commit/5c717ca045ab6742b5bb57dc1669a54330f6b65b) | fix(im): handle unresolved Feishu resource images (#2761) |
| 321 | 2026-09-25 | [591f1c0e4f](https://github.com/Tencent/WeKnora/commit/591f1c0e4fb70a9ece1d807b19c6fa0c24e6b102) | docs: drop references to source files that no longer exist (#3631) |
| 322 | 2026-09-25 | [d9444683da](https://github.com/Tencent/WeKnora/commit/d9444683dae2f382e29655f3efc6cc156d199329) | docs(dev): 补充 Swagger 容器访问与空白页排查步骤 (#3620) |
| 323 | 2026-09-25 | [0cb1e57fae](https://github.com/Tencent/WeKnora/commit/0cb1e57fae586908ca8d005459eb82acf54ca015) | fix(container): propagate host project lookup errors (#3611) |
| 324 | 2026-09-25 | [4668447aa6](https://github.com/Tencent/WeKnora/commit/4668447aa60b5190f90a8a53546151df4ec4ed11) | fix(helm): pin application images to chart version (#3669) |
| 325 | 2026-09-25 | [22404e92b7](https://github.com/Tencent/WeKnora/commit/22404e92b79dfc3ae7d60342ba16b0fd37364b30) | style(files): wrap long test case to satisfy lll |
| 326 | 2026-09-25 | [508048aee9](https://github.com/Tencent/WeKnora/commit/508048aee9739e4d034ab530d6a25603a6d9042d) | fix(frontend): mount the settings modal only once on /platform/settings |
| 327 | 2026-09-25 | [745a22d066](https://github.com/Tencent/WeKnora/commit/745a22d066cab0dc555372a7510c017c5ddb9c6a) | refactor(frontend): drop TTL resource caches and remove duplicate startup requests (#3562) |
| 328 | 2026-09-25 | [c7601b2d2a](https://github.com/Tencent/WeKnora/commit/c7601b2d2aaf298939b5ce775970c3005044d74a) | feat(wiki): 增加跨知识库 Wiki 搜索 POST /wiki-search (#3535) |
| 329 | 2026-09-25 | [a9d71038b6](https://github.com/Tencent/WeKnora/commit/a9d71038b6657fda5e01a85ad085b861c2021758) | fix(models): stop advertising graded thinking levels the wire format drops (#3552) |
| 330 | 2026-09-25 | [00e6366fd2](https://github.com/Tencent/WeKnora/commit/00e6366fd27a42bca7b08d6fef5b909339bf6be9) | fix(datasource): validate cron schedules and persist manual-only mode (#3653) |
| 331 | 2026-09-25 | [0bf105499b](https://github.com/Tencent/WeKnora/commit/0bf105499b093578ddbaabf24403ba0aa776657a) | fix(knowledge): reject invalid JSON uploads before async parse (#3360) |
| 332 | 2026-09-25 | [a7fc9e13da](https://github.com/Tencent/WeKnora/commit/a7fc9e13dab91cf5bfc3d33c0f09415879957868) | fix(client): accumulate multiline SSE data (#2122) |
| 333 | 2026-09-25 | [e7c27ccc9e](https://github.com/Tencent/WeKnora/commit/e7c27ccc9e7fc806c583f6bb5de012cba12e5c53) | feat(confluence): add directory sync and private image assets (#3319) |
| 334 | 2026-09-25 | [a7ac21ce76](https://github.com/Tencent/WeKnora/commit/a7ac21ce769fa75854c0b7c7cfb04d934464fa73) | fix(docreader): sanitize non-conforming styles.xml fills before openpyxl loads (#3637) (#3638) |
| 335 | 2026-09-25 | [1dd3fe867e](https://github.com/Tencent/WeKnora/commit/1dd3fe867e0d343df765961eb090ae0c72a3c031) | fix(docreader): drop data validation ranges openpyxl cannot parse instead of failing the sheet (#3599) (#3608) |
| 336 | 2026-09-25 | [4364e61afa](https://github.com/Tencent/WeKnora/commit/4364e61afa4bf1e086c34b4220bb21449aa15b25) | feat(datasource/yuque): preserve TOC folder hierarchy for synced documents (#3476) |
| 337 | 2026-09-26 | [47773b5bd1](https://github.com/Tencent/WeKnora/commit/47773b5bd1fc33ecbe96d11824d4aa4b97ae2438) | fix(frontend): initialize chat citation popover after botmsg props |
| 338 | 2026-09-26 | [c91a8d11f8](https://github.com/Tencent/WeKnora/commit/c91a8d11f8eaa1370970355019a001a36e993117) | feat(plugin): plugin system foundation — builtin plugins, config schemas and workspace switches (#3711) |
| 339 | 2026-09-26 | [ed63e35cc7](https://github.com/Tencent/WeKnora/commit/ed63e35cc7d49856c562b9f519534cba6d6b9ee8) | feat(plugin): install declarative plugins — MCP servers, skills and model vendors (#3716) |
| 340 | 2026-09-26 | [4328ab26a9](https://github.com/Tencent/WeKnora/commit/4328ab26a942a3fb9c397dbea57b41ac1074cc46) | feat(plugin): code plugins — extension protocol, Go SDK, plugin host, web search and data source adapters (#3720) |
| 341 | 2026-09-26 | [ff49f9491d](https://github.com/Tencent/WeKnora/commit/ff49f9491d6b73de5adea18f061a7d04073cbd44) | feat(plugin): Host API with a key-value store, and document parsers from plugins (#3721) |
| 342 | 2026-09-26 | [728007ada9](https://github.com/Tencent/WeKnora/commit/728007ada9df5461d1f70aca64ec358718b891bd) | feat(plugin): remote runtime, Python SDK and python host plugins (P2 batch 3) (#3724) |
| 343 | 2026-09-26 | [da4ef68d9d](https://github.com/Tencent/WeKnora/commit/da4ef68d9d5327dfd236299d882f092943a1a165) | feat(plugin): standalone plugin hosts with Redis discovery (P2 batch 4) (#3725) |
| 344 | 2026-09-26 | [996696057f](https://github.com/Tencent/WeKnora/commit/996696057f7ec229fd1930188955283042f0c852) | feat(plugin): sandboxed plugin pages and the @weknora/plugin-ui bridge (P2 batch 5) (#3727) |
| 345 | 2026-09-26 | [cf502ba46a](https://github.com/Tencent/WeKnora/commit/cf502ba46a629598db393d747be05da3df682c18) | feat(plugin): event delivery and webhooks for plugins (P2 batch 6) (#3728) |
| 346 | 2026-09-26 | [01222aac8e](https://github.com/Tencent/WeKnora/commit/01222aac8e884f9a14db531072c6074f059beb4a) | feat(plugin): dynamic form options, OAuth connections and resource limits (P2 batch 7) (#3729) |
| 347 | 2026-09-26 | [4f5d475c3e](https://github.com/Tencent/WeKnora/commit/4f5d475c3eebe47d7c0e921799bfdb8337f6a024) | feat(examples): Jira Cloud plugin with OAuth, dynamic options and a skill (#3730) |
| 348 | 2026-09-26 | [a829d7cc9c](https://github.com/Tencent/WeKnora/commit/a829d7cc9cc3ac96c3cd1bbd6210203966c97d56) | feat(plugin): agent tools served by plugins, with result views (P2 batch 8) (#3733) |
| 349 | 2026-09-26 | [6c83d7ab47](https://github.com/Tencent/WeKnora/commit/6c83d7ab479212c136094182a68ac6e238ea370c) | revert: take the plugin system back out of main |
| 350 | 2026-09-26 | [9fcc2c6743](https://github.com/Tencent/WeKnora/commit/9fcc2c674309b42a1d4077e6b4c2a2334d49e57c) | fix(dev): propagate dev.sh exit status and pull MinIO from pgsty/minio |
| 351 | 2026-09-26 | [4df7ccbb4f](https://github.com/Tencent/WeKnora/commit/4df7ccbb4f650082331452a92b52854f8922da2a) | fix(chat): drop stale minimap-target class binding |
| 352 | 2026-09-27 | [9114e4e4f9](https://github.com/Tencent/WeKnora/commit/9114e4e4f905be71976a77c684d3f92731e6f9a6) | fix(citations): 修复原文定位、高亮范围与 DocReader 启动失败 (#3794) |
| 353 | 2026-09-27 | [a570dffcfa](https://github.com/Tencent/WeKnora/commit/a570dffcfa51ad92a5c66a8260ceaaa831567dc0) | test(wiki): 表格夹具改用合成数据 |
| 354 | 2026-09-26 | [5bf2ae1cc2](https://github.com/Tencent/WeKnora/commit/5bf2ae1cc27aa555957dc1c258968c5b36246b26) | fix(faq): actually clean FAQ entry payloads before validation and storage |
| 355 | 2026-09-27 | [0b463e459a](https://github.com/Tencent/WeKnora/commit/0b463e459aa70ea42036a49664a75e4da95684bf) | test(skills): wait for the host install to release its lock before asserting |
| 356 | 2026-09-25 | [9cab226853](https://github.com/Tencent/WeKnora/commit/9cab2268531477f6cd879acfd975189973d099cd) | docs: extract active tenant ID in the API quickstart |
| 357 | 2026-09-26 | [157fb4d3c1](https://github.com/Tencent/WeKnora/commit/157fb4d3c1763c47fdda88aa4133b37439bbc54a) | fix(frontend): label host-sandbox skill secret cards like the sandbox cards |
| 358 | 2026-09-27 | [bdad5eea2b](https://github.com/Tencent/WeKnora/commit/bdad5eea2b46546db4c1481fc83d62bd69351fcc) | fix(chat): clear stale KB scope on session restore |
| 359 | 2026-09-26 | [d528d1fc3c](https://github.com/Tencent/WeKnora/commit/d528d1fc3ca60f331befe1ba8c0ce28962024bd3) | fix: skip table-summary tasks for missing documents |
| 360 | 2026-09-26 | [d270a844ac](https://github.com/Tencent/WeKnora/commit/d270a844ac7684ed6eeea7e0469f3f5e37aef8ca) | fix: reject empty table-summary model responses |
| 361 | 2026-09-26 | [675b06addc](https://github.com/Tencent/WeKnora/commit/675b06addcd76a9f9f0c515bcd6554008efc67d3) | test: consolidate table summary regressions |
| 362 | 2026-09-26 | [bad7afa6ae](https://github.com/Tencent/WeKnora/commit/bad7afa6ae3198b8e574c94614e41267dbe6e005) | fix(sandbox): pass desktop slot timing to the renew goroutine |
| 363 | 2026-09-26 | [4ba85e2edf](https://github.com/Tencent/WeKnora/commit/4ba85e2edf71a041fa6846c14eb74ccfa0cfb389) | fix(wiki): fan out follow-up triggers so backlogs drain at ingest_max_inflight |
| 364 | 2026-09-28 | [c819abd958](https://github.com/Tencent/WeKnora/commit/c819abd9589c23a3fa504f946fc477499cd0fe70) | fix(wiki): bound outstanding follow-up tasks |
| 365 | 2026-09-28 | [ec0ca46334](https://github.com/Tencent/WeKnora/commit/ec0ca46334556e5b49a981e9cf26189e162eca31) | fix(server): run ResourceCleaner on signal shutdown instead of racing Fatalf exit (#3804) |
| 366 | 2026-09-27 | [08f7e702cf](https://github.com/Tencent/WeKnora/commit/08f7e702cfb3a1344b537641739fa08c523df255) | fix(datasource/dingtalk): 跳过不再静默，展开叶子不再报错 |
| 367 | 2026-09-29 | [7d1d71e50d](https://github.com/Tencent/WeKnora/commit/7d1d71e50d7123dc14c749d5677b588f728dc7c9) | fix(datasource/dingtalk): 块级渲染不再静默丢内容（列表按线上真实形状渲染 + 未建模块类型可见） (#3822) |
| 368 | 2026-09-28 | [72e105f46a](https://github.com/Tencent/WeKnora/commit/72e105f46a7747d5acb1a887bd459419d4b9b24d) | fix(gitlab): 按提交快照比较增量同步变更 |
| 369 | 2026-09-29 | [a282d1091e](https://github.com/Tencent/WeKnora/commit/a282d1091ebf8c5d7d8d64796a25f1e1527bd296) | fix(feishu): 云盘分页补上 #3645 给 wiki 加的重复 page token 防护 (#3770) |
| 370 | 2026-09-26 | [b9be7d0d0d](https://github.com/Tencent/WeKnora/commit/b9be7d0d0d525435b67caeb49ef0502d6b6e7097) | fix(notion): 修复页面清空后旧正文残留 |
| 371 | 2026-09-25 | [15d4845521](https://github.com/Tencent/WeKnora/commit/15d484552140eafe491e01d7f7bbfb558515801b) | fix(rss): 修复全文回退后增量同步不重试 |
| 372 | 2026-09-26 | [cf079b8dcd](https://github.com/Tencent/WeKnora/commit/cf079b8dcd3315b46f1a2c848837f851ec61e04c) | fix(models): keep Retry-After on HTTPError and honour it in agent LLM retries |
| 373 | 2026-09-27 | [436d02fdf3](https://github.com/Tencent/WeKnora/commit/436d02fdf3865279df3ff3bc14a91057b56d0464) | fix(compaction): 流因取消而关闭时不得当成空摘要成功返回 |
| 374 | 2026-09-27 | [e39043f34b](https://github.com/Tencent/WeKnora/commit/e39043f34b22bec26ed6ddb4d08ff66e9555f3ea) | fix(repository): 关键词 LIKE/ILIKE 检索补显式 ESCAPE，修复 Lite 上静默漏结果 |
| 375 | 2026-09-27 | [1912ba26c5](https://github.com/Tencent/WeKnora/commit/1912ba26c544025adb869459a74c2cbac2a95cd9) | test(weaviate): 让并发建表的用例真的每个 worker 都去创建，消掉随机失败 |
| 376 | 2026-09-29 | [f46c990567](https://github.com/Tencent/WeKnora/commit/f46c9905677b36a454aa6669a2619f9eecf9fd23) | fix(embed): 开启 Host context 后点击追问建议不再返回 400 |
| 377 | 2026-09-29 | [75d2c1b583](https://github.com/Tencent/WeKnora/commit/75d2c1b58343b2872051a24b78185298fb8b88f3) | fix(frontend): normalize style guard paths |
| 378 | 2026-09-29 | [0c6b30cfa7](https://github.com/Tencent/WeKnora/commit/0c6b30cfa705c1d2f2ccd084cd76174738a7fb4b) | fix(docs): spawn nested npm through Node on Windows |
| 379 | 2026-09-29 | [9842e51d40](https://github.com/Tencent/WeKnora/commit/9842e51d40a54ca2dbb61153f3fb587da1265b7a) | fix(frontend): 修复复制模型时模型名称被改写的问题 (#3830) |
| 380 | 2026-09-30 | [5225edc0f7](https://github.com/Tencent/WeKnora/commit/5225edc0f746ae68e15dc4bd6fde9343efde1142) | feat(auth): 新增仅限邀请注册档位 (#3900) |
| 381 | 2026-09-30 | [176aa0d02b](https://github.com/Tencent/WeKnora/commit/176aa0d02b30b64634d55d000b6b44e00ec0310a) | fix(sandbox): build real arm64 sandbox images and ship arm64 Cube variants |
| 382 | 2026-09-30 | [dba9127558](https://github.com/Tencent/WeKnora/commit/dba9127558c3df160c8ff181e2a0a99172ae7c75) | docs(api): 补充模型复制接口说明 |
| 383 | 2026-09-29 | [2ec1774623](https://github.com/Tencent/WeKnora/commit/2ec1774623a77608902287b2ba42dcde3e4daaab) | fix(auth): 令牌写入失败不再当成登录成功，查询失败不再谎称已撤销 |
| 384 | 2026-09-29 | [b71b31c897](https://github.com/Tencent/WeKnora/commit/b71b31c897efae7b6b8320ab0e5bdf796eb1e006) | fix(ima): 笔记正文被清空后旧正文仍留在知识库 |
| 385 | 2026-09-30 | [1dbe9d0c1a](https://github.com/Tencent/WeKnora/commit/1dbe9d0c1a1b9dd8d7c345c5c1672f76269b86c5) | fix(gitlab): 仓库文件下载补体积上限，超限报错而不是无界读入 (#3887) |
| 386 | 2026-09-30 | [75fc0ad159](https://github.com/Tencent/WeKnora/commit/75fc0ad1590786202af57546a6879bb074671ab9) | feat(docparser): upgrade AnyDoc to 0.2.4 and parse assets once |
| 387 | 2026-09-29 | [bccb4b151b](https://github.com/Tencent/WeKnora/commit/bccb4b151bae403508da77fbb174efc79dc47c1a) | fix(embed): 内嵌网页只传图片不输入文字时不再返回 400 |
| 388 | 2026-10-07 | [402f69832e](https://github.com/Tencent/WeKnora/commit/402f69832e395d8d193d89fef7a248a1c48ad24c) | fix(notion): 查询撞厂商上限被截断时不再静默，截断轮次不下发删除 (#3845) |
| 389 | 2026-10-07 | [0ec945fe5d](https://github.com/Tencent/WeKnora/commit/0ec945fe5df0def761351aa640bdb588bd8e1bb2) | fix(dataset): use supported default QA model (#3965) |
| 390 | 2026-10-07 | [72b8a106c7](https://github.com/Tencent/WeKnora/commit/72b8a106c77a3337bd5d785c595c8575686b6025) | fix(sqlite): 检索后端批量更新索引副本失败不再吞掉错误 (#3768) |
| 391 | 2026-10-07 | [a8a2c17b89](https://github.com/Tencent/WeKnora/commit/a8a2c17b892c378686e4d0f3916b064d2b38da80) | fix(sqlite): 索引行写入失败不再被丢弃，BatchSave 回滚而不是留下「主行在、索引行不在」 (#3843) |
| 392 | 2026-10-07 | [f9eeb8dc85](https://github.com/Tencent/WeKnora/commit/f9eeb8dc8535d678e206b00c6e24540db542b868) | fix(retriever/es): _bulk 的 per-item 失败不再被吞，全部被拒时 BatchSave 返回错误 (#3842) |
| 393 | 2026-10-07 | [5f5093b9f7](https://github.com/Tencent/WeKnora/commit/5f5093b9f74e2c8621c11e1c53886758784d68b2) | fix(retriever): Qdrant/Milvus 关键词检索整批失败返回 error，不再当成「没有匹配」 (#3847) |
| 394 | 2026-10-07 | [5648d24a93](https://github.com/Tencent/WeKnora/commit/5648d24a93d353dca6d719178752e26d72a2478a) | fix(retriever/milvus): CopyIndices 按映射分块读源行，去掉会静默漏向量的 offset 窗口 (#3883) |
| 395 | 2026-10-07 | [9490035f15](https://github.com/Tencent/WeKnora/commit/9490035f1517bd68545df024e05b2244e4d1ac9a) | fix(agent): 图谱查询的三处截断必须可见，模型不再把子图当全集 (#3885) |
| 396 | 2026-10-07 | [9fa9e7c29d](https://github.com/Tencent/WeKnora/commit/9fa9e7c29d8eb727825b14f0b0ca23907969b91a) | fix: use consistent qdrant env settings (#3960) |
| 397 | 2026-10-08 | [afc6cbd502](https://github.com/Tencent/WeKnora/commit/afc6cbd50264a54018faa9552346d45f964834fe) | fix(milvus): validate connections and preserve gRPC resolver defaults (#3654) |
| 398 | 2026-10-08 | [e29301115a](https://github.com/Tencent/WeKnora/commit/e29301115a276afc4e7be7f0ffe568066c0c3386) | fix(graph): preserve relationship direction and document provenance (#3950) |
| 399 | 2026-10-08 | [7c284e7fc9](https://github.com/Tencent/WeKnora/commit/7c284e7fc9beb065c4da2587b82bed52b5073bef) | fix(faq): 保留克隆状态同步时的正文和索引状态 (#3993) |
| 400 | 2026-10-08 | [a8c291c6ff](https://github.com/Tencent/WeKnora/commit/a8c291c6ff52985e6729a8b3e0d404add08ecfd3) | fix: keep failed data-analysis table loads retryable (#3745) |
| 401 | 2026-10-07 | [27b9ee74a0](https://github.com/Tencent/WeKnora/commit/27b9ee74a02f2f33689d0adb2de091c5bd8d5342) | fix(storage): use virtual-hosted addressing for Tencent COS through the S3 driver (#3722) |
| 402 | 2026-10-08 | [a15dfdd941](https://github.com/Tencent/WeKnora/commit/a15dfdd941497b32198f8abaec71510b304ee070) | fix(summary): 坏 JSON 不再冒充文档摘要（截断/畸形 JSON 走失败路径，纯文本兜底保留） (#3769) |
| 403 | 2026-10-07 | [e915adb147](https://github.com/Tencent/WeKnora/commit/e915adb147d721fc265b22db24a56a535b5bcf6c) | fix(chunker): preserve source offsets for table child chunks (#3969) |
| 404 | 2026-10-07 | [98feef38de](https://github.com/Tencent/WeKnora/commit/98feef38de5c39e57ef1c817e43ef6a4879f4029) | fix(chunker): count an escaped pipe as cell text when matching table widths (#3979) |
| 405 | 2026-10-08 | [0d3a053eb6](https://github.com/Tencent/WeKnora/commit/0d3a053eb62c5150c8ad01442776de0225d06ea2) | fix(yuque): don't acknowledge failed detail fetches in the incremental cursor (#3702) |
| 406 | 2026-10-08 | [b5c74b2dd6](https://github.com/Tencent/WeKnora/commit/b5c74b2dd6f183f1e31ef98a166696664b716570) | fix(feishu): 多维表格字段分页补重复 page_token 防护与条数上限 (#3840) |
| 407 | 2026-10-08 | [a2aeb97a01](https://github.com/Tencent/WeKnora/commit/a2aeb97a01203b23b6093eba2369583b460aaa66) | fix(gitlab): 分页循环补重复页防护与跳数上限 (#3841) |
| 408 | 2026-10-08 | [a6c6c35ad2](https://github.com/Tencent/WeKnora/commit/a6c6c35ad292716570c44a42221f9e5ec68d2543) | fix(ima): 三条游标分页循环补重复游标防护与跳数上限 (#3889) |
| 409 | 2026-10-07 | [c9723b1198](https://github.com/Tencent/WeKnora/commit/c9723b1198bfa42d31278b4e44f1199593ee5a08) | fix(docreader): read EPUB chapters in spine order (#3977) |
| 410 | 2026-10-08 | [9de0946e53](https://github.com/Tencent/WeKnora/commit/9de0946e53348d063180b2af87008a51c4fd9dcc) | fix(docparser): anydoc 内嵌 EMF/WMF 图片的扩展名与 Rust 侧对齐 (#3938) |
| 411 | 2026-10-08 | [188d9026d8](https://github.com/Tencent/WeKnora/commit/188d9026d8d344464676c7919c09730a5181adb6) | fix(models): honor OpenRouter native thinking for legacy defaults (#3807) |
| 412 | 2026-10-08 | [f6119ae5fd](https://github.com/Tencent/WeKnora/commit/f6119ae5fd4032e00a621a2cc280972390d0cfa4) | fix(agent): 保留数据分析 SQL 字符串字面量 (#3812) |
| 413 | 2026-10-08 | [52893006ad](https://github.com/Tencent/WeKnora/commit/52893006ad26558164ddceabce89360a1b17b984) | fix(agent): validate tool string lengths by Unicode code points (#3937) |
| 414 | 2026-10-08 | [fa4f85d3dd](https://github.com/Tencent/WeKnora/commit/fa4f85d3dd1edba7689a3e235abd3f159b2af381) | fix(agent): 拒绝将断流摘要保存为记忆检查点 (#3994) |
| 415 | 2026-10-08 | [e394c7adec](https://github.com/Tencent/WeKnora/commit/e394c7adecbe38019f66b8490ee06d254efaa05a) | fix(agent): 工具自起 goroutine 的 panic 自行兜底，MCP 预热不再带走进程 (#3846) |
| 416 | 2026-10-08 | [412ce63f0b](https://github.com/Tencent/WeKnora/commit/412ce63f0b00ab45811f81b0ca254259e205acca) | fix(chat): clear stale file and tool selections on session restore (#3802) |
| 417 | 2026-10-08 | [f559473498](https://github.com/Tencent/WeKnora/commit/f559473498719ff0f5c94fc9c6918199986aaf3e) | fix(mcp): keep the client secret from dynamic client registration (#3763) |
| 418 | 2026-10-08 | [28734725a3](https://github.com/Tencent/WeKnora/commit/28734725a3567d5d12a0b190b0e5bccbb91b252c) | fix(mcp): reconnect after the server forgets the session (#3765) |
| 419 | 2026-10-08 | [172d0a85ab](https://github.com/Tencent/WeKnora/commit/172d0a85ab17e74abe8f987927ad1b9fd3eeb29a) | fix(mcp): pass embedded resources in tool results to the agent (#3766) |
| 420 | 2026-10-08 | [e8cc9c763e](https://github.com/Tencent/WeKnora/commit/e8cc9c763e639749bab486a0fc291d91c6b09597) | fix(mcp): include shared knowledge bases in endpoint scope (#3848) |
| 421 | 2026-10-08 | [69a7fd9931](https://github.com/Tencent/WeKnora/commit/69a7fd9931a0867dd7e1003262b315a330e5d606) | fix(mcp): close legacy chat responses on HTTP errors (#3930) |
| 422 | 2026-10-08 | [c0c194f625](https://github.com/Tencent/WeKnora/commit/c0c194f6253ba38ddfc868767bda21fd6f4ea900) | fix(im/wecom): preserve image in mixed messages with text captions (#3901) |
| 423 | 2026-10-08 | [95691daab2](https://github.com/Tencent/WeKnora/commit/95691daab2e292a8c45e0ae200737b1c9e5de222) | fix(knowledge): retry document processing after transient indexing failures (#3915) |
| 424 | 2026-10-08 | [fac9ebabf2](https://github.com/Tencent/WeKnora/commit/fac9ebabf2d4578f6192528b1e68f69af48d9a96) | fix(knowledge): close spans left open on terminal rows and superseded attempts (#3855) |
| 425 | 2026-10-08 | [4d7e663128](https://github.com/Tencent/WeKnora/commit/4d7e663128e70f4f4ba496f062fc5a7c23633358) | fix(wiki): preserve pending work when source lookups fail (#3715) |
| 426 | 2026-10-08 | [626185a120](https://github.com/Tencent/WeKnora/commit/626185a120ef55a741c5add46147c0af146c18aa) | fix(wiki): 重建链接逐页写失败不再静默，半成品不再回 200 (#3884) |
| 427 | 2026-10-08 | [11bfacad97](https://github.com/Tencent/WeKnora/commit/11bfacad97c7748542a988717bd3f4ffc53244b4) | chore(deps): bump the cli-deps group in /cli with 2 updates (#3920) |
| 428 | 2026-10-07 | [3206e15e83](https://github.com/Tencent/WeKnora/commit/3206e15e831672f0ae0d837d822150ca19415594) | fix(docs): make `make docs` reproduce the committed Swagger again (#3751) |
| 429 | 2026-10-08 | [9d4cc233c6](https://github.com/Tencent/WeKnora/commit/9d4cc233c68a3f3e7ee7fd81926708dd5c898b36) | feat(ui): add search to Skills and MCP toolbox lists (#3731) |
| 430 | 2026-10-08 | [4ea94afcb9](https://github.com/Tencent/WeKnora/commit/4ea94afcb91b7f2388d9440495f30cf555e88a5b) | fix(cli): propagate paginated response read errors (#3983) |
| 431 | 2026-10-08 | [4a0c7fa91f](https://github.com/Tencent/WeKnora/commit/4a0c7fa91f2f646e6f56181ba723a9eebf368ed9) | fix(cli): render paginated API results in text mode (#3984) |
| 432 | 2026-10-08 | [7d75ce3a6f](https://github.com/Tencent/WeKnora/commit/7d75ce3a6f871ae59668b440439ea27a9959c351) | fix(client): make multipart uploads replayable (#3989) |
| 433 | 2026-10-08 | [cefd5920f8](https://github.com/Tencent/WeKnora/commit/cefd5920f885b604cc760478e9dd87b5778a272e) | fix(cli): apply list limits after filtering and sorting (#3990) |
| 434 | 2026-10-08 | [ff33184433](https://github.com/Tencent/WeKnora/commit/ff33184433c438eea0ec40293289b819fa817c97) | fix(cli): preserve integer precision in jq output (#3991) |
| 435 | 2026-10-08 | [b413023d28](https://github.com/Tencent/WeKnora/commit/b413023d2885742048103d4d617922465788c0f2) | fix(cli): classify wrapped SDK HTTP errors correctly (#3992) |
| 436 | 2026-10-08 | [562127c654](https://github.com/Tencent/WeKnora/commit/562127c6542d9f3f073f390658cb67ae65b97484) | fix(client): accept optional spaces in SSE event fields (#3975) |
| 437 | 2026-10-08 | [b8d83eab1c](https://github.com/Tencent/WeKnora/commit/b8d83eab1cf426d19c8e0183b9d896b77a62872b) | fix(skills): skip vendored dep trees in install verification script enumeration (#3943) |
| 438 | 2026-10-08 | [dad81cb8c1](https://github.com/Tencent/WeKnora/commit/dad81cb8c10c52efd5d70dfe1abcc9f0c5672c99) | fix(sandbox): reclassify exec races with the idle sweep as NotFound so rebinding heals (#3980) |
| 439 | 2026-10-08 | [20038cbda6](https://github.com/Tencent/WeKnora/commit/20038cbda6e3c810b4ab3951d56bbcf8f81af19c) | fix(asynqdl): truncate dead-letter errors on a rune boundary (#3964) |
| 440 | 2026-10-08 | [6057a81017](https://github.com/Tencent/WeKnora/commit/6057a81017375a8d680cd32afca1c2ff03b0b17c) | fix(middleware): stop buffering whole request bodies in the access log (#3927) |
| 441 | 2026-10-08 | [5077f35800](https://github.com/Tencent/WeKnora/commit/5077f3580003bcf55eeb89694d43479ca216adae) | fix(notion): 超过 block 上限截断时补 Warn，避免内容与子页面静默缺失 (#3771) |
| 442 | 2026-10-08 | [dabbdaf156](https://github.com/Tencent/WeKnora/commit/dabbdaf156b8dfcb58dd68e9c0be3842d27d72b5) | fix(notion): 分页循环补重复游标防护与跳数上限 (#3890) |
| 443 | 2026-10-07 | [e7aaf8ae02](https://github.com/Tencent/WeKnora/commit/e7aaf8ae02608d1d9e0f84aa864106a5f1ae78a3) | fix(docreader): keep the space held by a whitespace-only inline element (#3978) |
| 444 | 2026-10-08 | [336f84c3be](https://github.com/Tencent/WeKnora/commit/336f84c3be4a4831c49b2708f98431dae2ab3312) | fix(data-analysis): skip style-only worksheets before DuckDB loading (#3660) |
| 445 | 2026-10-08 | [4dd19b5356](https://github.com/Tencent/WeKnora/commit/4dd19b535603d4a9dc8b64033192592572db5180) | fix(frontend): render single tilde literally instead of GFM strikethrough (#3963) |
| 446 | 2026-10-08 | [4eb23a8560](https://github.com/Tencent/WeKnora/commit/4eb23a85607c0cabfa050f5e87eb71fb0ce2554c) | fix(qdrant): CopyIndices 用服务端返回的 next_page_offset 翻页，不再按页尾 id 回填 (#3882) |
| 447 | 2026-10-08 | [e6ecb9adcc](https://github.com/Tencent/WeKnora/commit/e6ecb9adcc8ddbb500957f77490ca5cb4c2fa8d4) | fix(dataset): allow selecting the QA answer model (#3944) |
| 448 | 2026-10-08 | [49e262147c](https://github.com/Tencent/WeKnora/commit/49e262147ca504a28680930930ed0851a622d0c5) | fix(cli): preserve directory structure in doc upload --recursive (#3753) |
| 449 | 2026-10-08 | [de2dfd3034](https://github.com/Tencent/WeKnora/commit/de2dfd3034a7b506fa269910401bbdcbd4efc71d) | fix(multimodal): bound remote image downloads by the file size limit (#3906) |
| 450 | 2026-10-08 | [cbfd66fd92](https://github.com/Tencent/WeKnora/commit/cbfd66fd922df3632448e0bfd5925a6c10196fc8) | fix(knowledge): delete written chunks when the storage quota check fails (#3656) |
| 451 | 2026-10-08 | [11aa00f85f](https://github.com/Tencent/WeKnora/commit/11aa00f85f22ee54a6c37963010ac3f4d20d694b) | fix(agent): preserve untouched numeric tool arguments during casting (#3936) |
| 452 | 2026-10-08 | [9a2cd20063](https://github.com/Tencent/WeKnora/commit/9a2cd200637359081d87fb4c51b6bc911a0ff6a2) | fix(frontend): keep target="_blank" on sanitized markdown links (#3818) |
| 453 | 2026-10-08 | [a75e8e9ac0](https://github.com/Tencent/WeKnora/commit/a75e8e9ac0ac10ea23c3108e789a8d71f6f071e9) | fix(datasource): 语雀/Notion/IMA/飞书连接器 API 响应体补体积上限 (#3985) |
| 454 | 2026-10-08 | [b5fd79edf2](https://github.com/Tencent/WeKnora/commit/b5fd79edf283f4ac6a272403f231c3c07db11500) | fix(model): treat 400 as reachable in chat connection test again (#3929) |
| 455 | 2026-10-08 | [1826f46b12](https://github.com/Tencent/WeKnora/commit/1826f46b1217867fe15f188706a9334cacec4165) | fix(docparser): 云解析超时改为跟随既有配置，不再硬编码 (#3891) |
| 456 | 2026-10-08 | [99b4c3a9b0](https://github.com/Tencent/WeKnora/commit/99b4c3a9b0d3c2d3440036c6e3f4235c6f276390) | feat(rerank): support Hugging Face TEI protocol (#3956) |
| 457 | 2026-10-08 | [d84cdabac4](https://github.com/Tencent/WeKnora/commit/d84cdabac49046dc546ba04716afcc3785240a88) | fix(rss): refresh titles and source links when article text is unchanged (#3824) |
| 458 | 2026-10-08 | [bea0d7b689](https://github.com/Tencent/WeKnora/commit/bea0d7b6891ccbf94d3e1c6f649a3a6092978a12) | fix(docreader): preserve PPTX fallback image associations (#3973) |
| 459 | 2026-10-08 | [3fb6b9d631](https://github.com/Tencent/WeKnora/commit/3fb6b9d631d399104778cf8d4a79facd23ed8853) | fix(kb): let the update API set vlm_config (#3704) |
| 460 | 2026-10-08 | [b5c8d90cc2](https://github.com/Tencent/WeKnora/commit/b5c8d90cc20747be3f64356150ca69224d063be0) | fix(chat_pipeline): treat non-JSON model refusals as skipped graph extraction (#3606) |
| 461 | 2026-10-08 | [ae90546262](https://github.com/Tencent/WeKnora/commit/ae90546262a84dcfd90b5da370b56bd0774cc8f2) | fix(frontend): avoid false unsaved changes warnings |
| 462 | 2026-10-08 | [bcbd65a459](https://github.com/Tencent/WeKnora/commit/bcbd65a459e912f70ad4a9b0f39f8848db4a2d9f) | fix(datasource/dingtalk): Validate 不再因为某个工作区读不到就整库判失败 (#3782) |
| 463 | 2026-10-08 | [19542f2a1d](https://github.com/Tencent/WeKnora/commit/19542f2a1d958d26383bbb586adecae6be07a7d5) | feat(datasource): add Seafile data source connector (#3616) |
| 464 | 2026-10-08 | [d2bc0df6a2](https://github.com/Tencent/WeKnora/commit/d2bc0df6a2af4f953485e69c63972cf7d1cc2646) | feat(i18n): add Traditional Chinese (zh-TW) locale support (#3916) |
| 465 | 2026-10-08 | [1c528d83de](https://github.com/Tencent/WeKnora/commit/1c528d83de39d79e870fe8b46d21addf20136a60) | fix(i18n): add the zh-TW Seafile strings missing after #3616 and #3916 |
| 466 | 2026-10-08 | [0de62b2f80](https://github.com/Tencent/WeKnora/commit/0de62b2f801513982291311b59fad381f7959560) | fix(chat): restore citation source navigation and soften highlights |
| 467 | 2026-10-08 | [a729d1fc8c](https://github.com/Tencent/WeKnora/commit/a729d1fc8c1c7e05119b69696ef9f61bc2627cfb) | docs(models): state which rerank limits split requests and which fail the call |
| 468 | 2026-10-08 | [ea7cf8d5d7](https://github.com/Tencent/WeKnora/commit/ea7cf8d5d77cf0deb46215836262405ad91d4b44) | fix(models): emit explicit empty content for assistant tool-call turns (#3605) |
| 469 | 2026-10-08 | [5a25e26142](https://github.com/Tencent/WeKnora/commit/5a25e261421db4c76b720e9d7392bd8f88f36585) | feat(rerank): support Pinecone Inference protocol (#3957) |
| 470 | 2026-10-08 | [b7f3181066](https://github.com/Tencent/WeKnora/commit/b7f31810664e22e44cb6da2dae3b2115c2060f49) | docs: count 29 built-in model vendors and list Hugging Face TEI and Pinecone (#4048) |
| 471 | 2026-10-08 | [0c5d28abb7](https://github.com/Tencent/WeKnora/commit/0c5d28abb7917b49d36f79bafb78fb59796a068b) | feat: index and recall images with multimodal embedding models (#3683) |
| 472 | 2026-10-08 | [676a59307e](https://github.com/Tencent/WeKnora/commit/676a59307e5ca7efabfe8896354b9425703aef67) | fix(settings): improve modal sizing and integration page layout |
| 473 | 2026-10-03 | [2a02807a61](https://github.com/Tencent/WeKnora/commit/2a02807a618a18196ebc1ddbe9888fb742b24117) | fix(chunker): respect Markdown fences during heading detection |
| 474 | 2026-10-03 | [005627b180](https://github.com/Tencent/WeKnora/commit/005627b180c6c7509660b0a14401e463f48d53f8) | fix(agent): compare tool enum values without string coercion |
| 475 | 2026-10-09 | [b14bcbcd54](https://github.com/Tencent/WeKnora/commit/b14bcbcd541fc0fdd67eef518828a28302358e18) | fix(notion): 保留正文抓取失败时的同步游标 (#3693) |
| 476 | 2026-10-09 | [04da26024c](https://github.com/Tencent/WeKnora/commit/04da26024c2412b34b3b30e7b283c98197ef95f1) | fix(weaviate): CopyIndices 补游标防护，并修掉 GraphQL 报错时的 panic (#3881) |
| 477 | 2026-10-09 | [5d40f81ee4](https://github.com/Tencent/WeKnora/commit/5d40f81ee4020d8f3173ac64df4dc4ae140351c1) | fix(im): retry Yunzhijia attachment downloads (#3809) |
| 478 | 2026-10-09 | [419f17fc65](https://github.com/Tencent/WeKnora/commit/419f17fc651de93a34eb3586cb894a9e7f3db4a4) | fix(wiki): resolve titles for unloaded backlinks (#3931) |
| 479 | 2026-10-09 | [cbfc0fed4a](https://github.com/Tencent/WeKnora/commit/cbfc0fed4a1cb4a3ff0ffabe943d6f1edfa3b21e) | fix(chat): 在模型上下文中标注 FILTER_TOP_K 的子集截断，避免把前 N 条当作全集作答 (#3758) |
| 480 | 2026-10-09 | [abc365f919](https://github.com/Tencent/WeKnora/commit/abc365f919191317020d395e9af19e4129357ff3) | fix(wiki): 唯一写出口 UpdatePage 拦住丢表格行的机器重写（agent 整页写入不再静默改短页面） (#3793) |
| 481 | 2026-10-09 | [8319d59dd0](https://github.com/Tencent/WeKnora/commit/8319d59dd014fccb7daf04eed9d50d8a3ffdecf4) | feat(datasource/dingtalk): 摄取钉钉"上传文件"（docx/pptx/xlsx/pdf） (#3786) |
| 482 | 2026-10-09 | [09411106f0](https://github.com/Tencent/WeKnora/commit/09411106f0ec19e840e9c24cabf391b8c2852a8e) | fix(session): allow tenant admins to manage channel sessions (#3939) |
| 483 | 2026-10-09 | [3d01bebce3](https://github.com/Tencent/WeKnora/commit/3d01bebce3d7aa90aaa52914534f22d9cadf683b) | fix(docreader): resolve EPUB images relative to their chapters (#3924) |
| 484 | 2026-10-09 | [4243018ff1](https://github.com/Tencent/WeKnora/commit/4243018ff1f55ccd92baebf4da0a1057317cc644) | fix(mcp): report advertised directory failures in connection tests (#3905) |
| 485 | 2026-10-09 | [a1450d382e](https://github.com/Tencent/WeKnora/commit/a1450d382e39a06977cc395aaf822eebab87c69e) | fix(knowledge): claim cloned and released document images in the resource catalog (#3752) |
| 486 | 2026-10-09 | [65f88f4121](https://github.com/Tencent/WeKnora/commit/65f88f4121433702ea053ae9127647329bf98b05) | feat: judge image hits by their image in rerank and show them to vision models (#4013) |
| 487 | 2026-10-09 | [52617f6410](https://github.com/Tencent/WeKnora/commit/52617f64106e2ea9af770ff783a1818e1ce7b8fb) | fix(sandbox): omit env overrides when creating Aliyun E2B sandboxes from snapshots (#4054) |
| 488 | 2026-10-09 | [602d0c388b](https://github.com/Tencent/WeKnora/commit/602d0c388bf17d894f467fe0cd65d1d62c73d303) | feat: support app and docreader build download mirrors (#4073) |
| 489 | 2026-10-10 | [d074ec7b0b](https://github.com/Tencent/WeKnora/commit/d074ec7b0b63903f117543e09deb31f5f3d19ecb) | fix(websearch): return results for non-webpage Metaso scopes (#4010) |
| 490 | 2026-10-10 | [40eb34bc84](https://github.com/Tencent/WeKnora/commit/40eb34bc849cb29dc53ccddef971c9a7c4410f7b) | fix(agent): align web_search count limit with the configured maximum (#4011) |
| 491 | 2026-10-10 | [c3f077ed18](https://github.com/Tencent/WeKnora/commit/c3f077ed18f6e00aaf705d405023ed88c7c15b27) | fix(models): align configuration and multimodal embedding workflows (#4080) |
| 492 | 2026-10-10 | [0c71fff2a3](https://github.com/Tencent/WeKnora/commit/0c71fff2a387b539c449cd6b66a3922ed6839f76) | fix(sandbox): treat E2B in_progress template builds as building (#4108) |
| 493 | 2026-10-10 | [6e2e5d8bb0](https://github.com/Tencent/WeKnora/commit/6e2e5d8bb07cd9c8db360c03b80ca47ea2e81fd5) | fix(frontend): follow configured upload limit in hints (#4110) |
| 494 | 2026-10-10 | [097bd475f6](https://github.com/Tencent/WeKnora/commit/097bd475f601d48a82ed7aaa2f9676dd82f15d97) | fix(docker): install CA certificates before the browserskill mirror swap (#4104) |
| 495 | 2026-10-10 | [8125ed250f](https://github.com/Tencent/WeKnora/commit/8125ed250f5ea404445e7952359a6ac686d90afc) | fix(chat): keep custom metadata on knowledge-chat references (#4087) |
| 496 | 2026-10-10 | [d6269994b3](https://github.com/Tencent/WeKnora/commit/d6269994b33688c2b85488c24014d95db1948d52) | refactor: 出站拨号超时默认 5s 并改为 env 可配 (WEKNORA_OUTBOUND_DIAL_TIMEOUT) (#4122) |
| 497 | 2026-10-10 | [03e2c47aec](https://github.com/Tencent/WeKnora/commit/03e2c47aecbf60120820216baaa8abb2e98844f7) | feat(mcp): support 2026-07-28 alongside legacy clients (#4095) |
| 498 | 2026-10-10 | [e372b8e42e](https://github.com/Tencent/WeKnora/commit/e372b8e42ef71f8ba9e1c4ada6546d79ec056f91) | perf(embedding): batch Qwen3-VL message inputs (#4126) |
| 499 | 2026-10-10 | [8cb53fdbc5](https://github.com/Tencent/WeKnora/commit/8cb53fdbc58e4c15aafdba973d65742551e4c9cd) | fix(observability): distinguish retrieval diagnostic spans (#4128) |
| 500 | 2026-10-10 | [bb878af508](https://github.com/Tencent/WeKnora/commit/bb878af50873b919a68134dc007ba9ae859c6c39) | fix(gallery): improve image detail panel readability (#4130) |
| 501 | 2026-10-10 | [d27ecade65](https://github.com/Tencent/WeKnora/commit/d27ecade656a1f2335baa2520d2622a297036ea5) | fix(ocr): reject invalid output and record extraction failures (#4132) |
| 502 | 2026-10-10 | [5f18308b0a](https://github.com/Tencent/WeKnora/commit/5f18308b0a39bd44f3b44203c167663d0430a40b) | fix(ocr): reject split repetition and preserve no-text responses (#4135) |
| 503 | 2026-10-10 | [a58fc13206](https://github.com/Tencent/WeKnora/commit/a58fc13206c7d1e89c784d61189ab67295727744) | fix(chat): smooth streamed citation reveal cadence |
| 504 | 2026-10-10 | [aa535c22fe](https://github.com/Tencent/WeKnora/commit/aa535c22fed1f4ba5c965b7f208361c0b1c2d4e2) | fix(chat): keep queued follow-up answers attached to their own question |
| 505 | 2026-10-10 | [50a2963160](https://github.com/Tencent/WeKnora/commit/50a29631608b9b95c88b4ea98e09867bdbcad991) | fix(preview): refine source citation highlights (#4139) |
| 506 | 2026-10-10 | [d8fa5b49a7](https://github.com/Tencent/WeKnora/commit/d8fa5b49a7f4aa93ded9978c40d1b57346375211) | feat(embed): support pasting screenshots into chat |
| 507 | 2026-10-09 | [98e9dc0786](https://github.com/Tencent/WeKnora/commit/98e9dc078625265ba1ea85ee3538a547f85a5be0) | perf(reranking): accelerate MMR selection with sparse postings |
| 508 | 2026-10-09 | [54e8d7ffde](https://github.com/Tencent/WeKnora/commit/54e8d7ffde31ccc00ea59cec7b2eaec18b01f137) | fix(mcp): 隔离慢连接清理对其他服务的阻塞 |
| 509 | 2026-10-10 | [dccb1d6d2c](https://github.com/Tencent/WeKnora/commit/dccb1d6d2c974062976e8b8841c46bbddb06d226) | fix(mcp): stop advertising read-only endpoint tools as destructive |
| 510 | 2026-10-09 | [de616ed4b7](https://github.com/Tencent/WeKnora/commit/de616ed4b7d2a6e55c9d380bd6068f792fe9f3a1) | test(service): assert RRF ranks ignore list arrival order |
| 511 | 2026-10-08 | [1a21af1b82](https://github.com/Tencent/WeKnora/commit/1a21af1b827072589eddcc9def0b91433399480c) | fix(mcp): preserve structured tool results for agents |
| 512 | 2026-10-09 | [d9237166cc](https://github.com/Tencent/WeKnora/commit/d9237166cc4cb618818d2e65e62f820f1cc03f24) | fix(im/wecom): accept custom wss long-connection endpoints behind SSRF host checks |
| 513 | 2026-10-09 | [def582b98a](https://github.com/Tencent/WeKnora/commit/def582b98a2960ceb53d7051f27b90014c08858f) | test(im/wecom): satisfy errcheck and lll in endpoint tests |
| 514 | 2026-10-09 | [769a5aeca8](https://github.com/Tencent/WeKnora/commit/769a5aeca84331b1f1f5ad5f3b93934ec33d4fbf) | fix(frontend): preserve the caret when outdenting Markdown lists |
| 515 | 2026-10-10 | [57f8a65fec](https://github.com/Tencent/WeKnora/commit/57f8a65fec7b86fb20aa2c19a0e21b67a3864c73) | fix(datasource/dingtalk): tighten Validate sampling edges (#4079) |
| 516 | 2026-10-10 | [932b750b2f](https://github.com/Tencent/WeKnora/commit/932b750b2f5ef1ce2537ac57a117dec7423f46ab) | fix(agent): preserve steering order across persistence failures (#4078) |
| 517 | 2026-10-08 | [0466c294ef](https://github.com/Tencent/WeKnora/commit/0466c294ef1ad728e889058d331a5cfc37b2bd23) | fix(agent): 将空闲超时覆盖到模型流建立阶段 |
| 518 | 2026-10-08 | [3f9af2afa9](https://github.com/Tencent/WeKnora/commit/3f9af2afa912e16ebaec1ae172963a99dfc2e419) | fix(sqlite): add MCP metadata migration |
| 519 | 2026-10-07 | [45a022aac2](https://github.com/Tencent/WeKnora/commit/45a022aac23448e270f23046c9018cf8b2e37761) | fix: wait for qdrant writes to finish |
| 520 | 2026-10-08 | [89e3273daf](https://github.com/Tencent/WeKnora/commit/89e3273daf3c11f1a4498183b75c8257efd8db6c) | fix(mcp): preserve disabled state on service creation |
| 521 | 2026-10-08 | [0d39af1214](https://github.com/Tencent/WeKnora/commit/0d39af1214ff9d02382a4e851d320f5be8957985) | fix(retrieval): skip KB retrieval for summarize intent |
| 522 | 2026-10-08 | [a6327f8e39](https://github.com/Tencent/WeKnora/commit/a6327f8e39fe62b299348e16a9bc81763c0fcf00) | fix(i18n): correct Korean username rule message to match the pattern |
| 523 | 2026-10-08 | [b1d92048cf](https://github.com/Tencent/WeKnora/commit/b1d92048cf61d732a6cd9171237c8d05770183e1) | fix(agent): account for file paths in read pagination budgets |
| 524 | 2026-10-08 | [54cb6e9f9b](https://github.com/Tencent/WeKnora/commit/54cb6e9f9be8cf486bb88b62916f13e16351f9a5) | test(docreader): add unit tests for utils.split text-splitting helpers |
| 525 | 2026-10-10 | [150d8cc605](https://github.com/Tencent/WeKnora/commit/150d8cc605ed9305a5fc4ad1b8789dae238decce) | feat(sandbox): Docker 后端支持交互式终端 (#4023) |
| 526 | 2026-10-11 | [231a5c0b15](https://github.com/Tencent/WeKnora/commit/231a5c0b15b817bb13ee3a876c3fbaf337cb4d5e) | feat(wiki): improve graph layout, theme captions and group focus (#4147) |
| 527 | 2026-10-11 | [d2dbf72099](https://github.com/Tencent/WeKnora/commit/d2dbf7209963f0a9f6fafb70358ef9baec999aa3) | style: polish sidebar, list pages, toolbox and citation UI (#4154) |
| 528 | 2026-10-11 | [17dc7da514](https://github.com/Tencent/WeKnora/commit/17dc7da514e1fd86f6c2fe77ab05c799cf34c27a) | feat(gallery): replace pagination with infinite masonry scrolling (#4156) |
| 529 | 2026-10-11 | [643f64180e](https://github.com/Tencent/WeKnora/commit/643f64180eb1898a7a2e32d517ff73f978d5448f) | feat(system): add platform user management and streamline settings (#4153) |
| 530 | 2026-10-11 | [5343ac47b7](https://github.com/Tencent/WeKnora/commit/5343ac47b730d451ad267004530084e08a5f599b) | fix(models): stop sending temperature to Claude models that reject it (#4155) |
| 531 | 2026-10-11 | [00474f734d](https://github.com/Tencent/WeKnora/commit/00474f734d6c7c0911d7485465d7c7286d0047cb) | fix: align settings layouts and tolerate existing migration objects (#4164) |
| 532 | 2026-10-11 | [9efe94cc6f](https://github.com/Tencent/WeKnora/commit/9efe94cc6f6d623590098f1fb355dd1f2e63b8f5) | feat(models): add Claude Opus 5.5, Sonnet 5.5 and Haiku 5.5 to the catalog (#4158) |
| 533 | 2026-10-12 | [d272a83cfc](https://github.com/Tencent/WeKnora/commit/d272a83cfc5d4a92aa3655f6a6c806aa40a34c70) | fix(deploy): pass WEKNORA_OUTBOUND_DIAL_TIMEOUT through to the app container (#4157) |
| 534 | 2026-10-12 | [be581c41e0](https://github.com/Tencent/WeKnora/commit/be581c41e0f17db25fd50493c345a5c43a8db138) | fix(mcp): close only the authorizing user's connection after OAuth (#4138) |
| 535 | 2026-10-12 | [21ef506fc9](https://github.com/Tencent/WeKnora/commit/21ef506fc995259e48680477ce8eb06556bbd957) | fix(agent): scope database queries to authorized shared knowledge bases (#4136) |
| 536 | 2026-10-12 | [3c36cb19a2](https://github.com/Tencent/WeKnora/commit/3c36cb19a2bb4214835cf8e5d520d4aa64ac8f23) | fix(deploy): bump paradedb image to v0.25.11-pg17 (#4120) |
| 537 | 2026-10-12 | [c6bfe81894](https://github.com/Tencent/WeKnora/commit/c6bfe81894be411c910ad573bd1c80d88452414b) | feat(helm): allow Service name overrides with matching consumer references (#4109) |
| 538 | 2026-10-09 | [e5f9f9deab](https://github.com/Tencent/WeKnora/commit/e5f9f9deab36898676fe85a9fda6b3e6f00288fa) | feat(rerank): add Cohere provider for the v2 rerank API |
| 539 | 2026-10-12 | [fe444c851e](https://github.com/Tencent/WeKnora/commit/fe444c851ef4430bcdcbef4cb10342839c57807f) | fix(feishu): add shared request pacing and improve HTTP 429 / error 99991400 retries in Feishu Wiki synchronization (#4039) |
| 540 | 2026-10-12 | [24b5ca3f23](https://github.com/Tencent/WeKnora/commit/24b5ca3f2325e28acb4122384459ca66acc85ef4) | docs: clarify frontend development proxy configuration (#4171) |
| 541 | 2026-10-12 | [7925f774d1](https://github.com/Tencent/WeKnora/commit/7925f774d1baa72e8244b5fc65324542049677a8) | fix(datasource/gitlab): emptied files no longer leave their old body searchable (#4008) |
| 542 | 2026-10-12 | [9b031a6576](https://github.com/Tencent/WeKnora/commit/9b031a657656cddaddafe3cb2e124ccfc2f0693e) | fix: accept qdrant collection created concurrently (#3999) |
| 543 | 2026-10-12 | [7982c6a3bd](https://github.com/Tencent/WeKnora/commit/7982c6a3bda4b91da402faf39ca381cabd4832ca) | fix(agent): accept adjacent prefixed Python string literals (#4006) |
| 544 | 2026-10-12 | [6a7094ceac](https://github.com/Tencent/WeKnora/commit/6a7094ceacd3d9844ab9f2ad9083075e1f7cd088) | fix(mcp): bound resource listing pagination (#4007) |
| 545 | 2026-10-12 | [ecdcd14e1d](https://github.com/Tencent/WeKnora/commit/ecdcd14e1db5452e929ea846b0a1677502f81464) | docs: refresh LICENSE third-party notices (#4178) |
| 546 | 2026-10-12 | [c2edec192b](https://github.com/Tencent/WeKnora/commit/c2edec192b5b4abbe77c0e210b3e5abfb51792cc) | docs: explain knowledge upload folder paths (#4177) |
| 547 | 2026-10-12 | [089711f92c](https://github.com/Tencent/WeKnora/commit/089711f92c50ecad782ab95dbce8b04278a9cb35) | fix(mcp): ignore discover probe in resource list tests (#4179) |
| 548 | 2026-10-12 | [98362dd296](https://github.com/Tencent/WeKnora/commit/98362dd29698f4b78df5d8ecf3f5020517e50ddc) | feat(gallery): filter images by source document (#4159) |
| 549 | 2026-10-12 | [1c54beb551](https://github.com/Tencent/WeKnora/commit/1c54beb551673f15c6054ffb7f224a4bfdb495e3) | refactor(gallery): turn source document filter into a toolbar popover (#4190) |
| 550 | 2026-10-12 | [ef54e835c9](https://github.com/Tencent/WeKnora/commit/ef54e835c907718cedf7316bfa93daf69a28b2ba) | docs(api): clarify knowledge-search scope union of knowledge_base_ids and knowledge_ids (#4191) |
| 551 | 2026-10-12 | [8fae51c640](https://github.com/Tencent/WeKnora/commit/8fae51c64084a5cc938eadb32d27c61eea4f4d87) | style(frontend): polish knowledge base settings hints and sub-sections (#4193) |
| 552 | 2026-10-12 | [fd4cc45848](https://github.com/Tencent/WeKnora/commit/fd4cc4584884f7d21d8949d40de5404aa6b14a40) | fix(frontend): respect Markdown indentation selection boundaries (#4195) |
| 553 | 2026-10-12 | [97796c637a](https://github.com/Tencent/WeKnora/commit/97796c637aa53713edc846bd0a39470d687f0fdf) | fix(helm): configure app rollouts and document shared storage (#4180) |
| 554 | 2026-10-13 | [809bda7b2f](https://github.com/Tencent/WeKnora/commit/809bda7b2f2d44952eaa8765d91e3eb666a42c1c) | fix(agent): make agent tool selection match the tools the runtime registers (#4211) |
| 555 | 2026-10-13 | [194bffbb0a](https://github.com/Tencent/WeKnora/commit/194bffbb0a2b07a9a4bcf35355d18ef56dc6535b) | feat(agent): read chat attachments on demand (#4203) |
| 556 | 2026-10-13 | [e0712df3e0](https://github.com/Tencent/WeKnora/commit/e0712df3e020c3b9744a333cfa102b5a6e6371f7) | feat(mcp): let callers without a browser obtain MCP OAuth tokens (#4229) |
| 557 | 2026-10-13 | [ec382d7911](https://github.com/Tencent/WeKnora/commit/ec382d79114cef327da3f857004077f725da110c) | fix(frontend): preserve inline images in EPUB previews (#4226) |
| 558 | 2026-10-13 | [a7a5f47625](https://github.com/Tencent/WeKnora/commit/a7a5f47625f31698897a62c152b1e186c95b43c7) | fix(frontend): preserve nonempty lists when pressing Enter (#4224) |
| 559 | 2026-10-13 | [7972cdd251](https://github.com/Tencent/WeKnora/commit/7972cdd251acdb599eb41dcb66ad312982f82360) | fix(knowledge): make the knowledge base description token budget configurable (#4230) |
| 560 | 2026-10-13 | [621bc9b98a](https://github.com/Tencent/WeKnora/commit/621bc9b98ad19b32f553da83f6cbb14f7f74a4d6) | fix(website-docs): prebundle fastdom when mermaid imports it (#4208) |
| 561 | 2026-10-13 | [8af7938aed](https://github.com/Tencent/WeKnora/commit/8af7938aeddfaaf55fb0578817482a3468f2a045) | fix(mcp): keep OAuth token expiry independent of the process time zone (#4207) |
| 562 | 2026-10-13 | [94fae6de43](https://github.com/Tencent/WeKnora/commit/94fae6de43a36e44d64aecbbdf19c0b72b352aa8) | feat(models): add WeMM embedding presets and Transformers serving (#4238) |
| 563 | 2026-10-13 | [4356f8c16d](https://github.com/Tencent/WeKnora/commit/4356f8c16dcd5cae85652e9afebf63dc280334fc) | fix(skill): do not fail installs on non-blocking observations (#4246) |
| 564 | 2026-10-13 | [e51dee8297](https://github.com/Tencent/WeKnora/commit/e51dee8297e10db8b0037a8e0cf5d684010cea44) | feat(chat): polish rewind confirm and mark sessions with unread replies (#4247) |
| 565 | 2026-10-13 | [2511a1f932](https://github.com/Tencent/WeKnora/commit/2511a1f932a7a0dc0e9a04edb62986df0b147cca) | fix(mcp): offer OAuth prompt after tool approval and tidy tools panel (#4248) |
| 566 | 2026-10-13 | [9238220867](https://github.com/Tencent/WeKnora/commit/92382208674ae1b9a2dd62ac0fa98d67ce348bfe) | feat(chat): add global search to the @ mention menu and align its styling (#4249) |
| 567 | 2026-10-14 | [31c6055687](https://github.com/Tencent/WeKnora/commit/31c6055687d33f9ad7a92cf453a264ce13a1700d) | feat(agent): name the turn's chat model on the completion event (#4252) |
| 568 | 2026-10-14 | [76ff398e01](https://github.com/Tencent/WeKnora/commit/76ff398e015396b9a702aacfe9238caab465e6ed) | fix(chat): locate citations in manual Markdown documents (#4278) |
| 569 | 2026-10-14 | [3b3b519310](https://github.com/Tencent/WeKnora/commit/3b3b519310721295162ea94b9fd62a92c8cab341) | fix(agent): clarify attachment image paths and chart readings (#4283) |
| 570 | 2026-10-14 | [3e22822c1d](https://github.com/Tencent/WeKnora/commit/3e22822c1d2397be730c458c7539d0eb8c113a24) | chore(release): finalize v0.8.3 website and documentation (#4285) |
| 571 | 2026-10-14 | [6c3730a6ad](https://github.com/Tencent/WeKnora/commit/6c3730a6adfd9e77e3472c3f0e3d10fcd805bda6) | fix(docs): stabilize anchor navigation and link release topics (#4287) |
| 572 | 2026-10-14 | [383d26535a](https://github.com/Tencent/WeKnora/commit/383d26535adcd7c4c4d383d5c81889028bc46237) | feat(agent): persist human-in-the-loop prompts with a shared broker (#4250) |
| 573 | 2026-10-14 | [a77d83fe0e](https://github.com/Tencent/WeKnora/commit/a77d83fe0ee40518aa057fa3e69f50538c72873a) | feat(search): add RRF controls in a tabbed settings popover (#4300) |
| 574 | 2026-10-15 | [234aa948db](https://github.com/Tencent/WeKnora/commit/234aa948db29c10f5ed4b7939285c7b3711c5d4c) | docs(swagger): document the real request and response types (#4304) |
| 575 | 2026-10-15 | [687b178706](https://github.com/Tencent/WeKnora/commit/687b17870699e373c9be24e9d2adc1f220fb1f43) | feat(agent): add the built-in WeKnora assistant with an operation catalog (#4302) |
| 576 | 2026-10-15 | [7c9088d0fe](https://github.com/Tencent/WeKnora/commit/7c9088d0feeb252009ca7b3019a866c8febc7bfe) | feat(agent): let the WeKnora assistant read the bundled documentation (#4303) |
| 577 | 2026-10-15 | [33b77f8440](https://github.com/Tencent/WeKnora/commit/33b77f84401458a19a37a44bc47149b7817057f8) | fix(ui): compact resource list headers in Lite mode (#4316) |
| 578 | 2026-10-15 | [c40b33ecc3](https://github.com/Tencent/WeKnora/commit/c40b33ecc34f5a2c33bc47d33e33c577e2ae1c2d) | fix(agent): admit catalog operations only from knowledge API areas (#4317) |
| 579 | 2026-10-15 | [1b24132566](https://github.com/Tencent/WeKnora/commit/1b241325669d9e1ac06878858ce6ac24efa46a6b) | fix(agent): let the WeKnora assistant search without its own rerank model (#4315) |
| 580 | 2026-10-15 | [a131efa5e0](https://github.com/Tencent/WeKnora/commit/a131efa5e037040cc690f25a02e2f33dd0f6f1ce) | feat(frontend): open the WeKnora assistant from anywhere with page context (#4314) |
| 581 | 2026-10-15 | [a37a1bc51a](https://github.com/Tencent/WeKnora/commit/a37a1bc51ad85d096df341f6d3a098f60d586f52) | feat(agent): open more workspace operations to the WeKnora assistant (#4318) |
| 582 | 2026-10-15 | [aaa1accebd](https://github.com/Tencent/WeKnora/commit/aaa1accebdf673b9603aaa555c5610a8fd7727d9) | feat(agent): let the assistant delete and move documents in batches (#4319) |
| 583 | 2026-10-15 | [6c23dac3b0](https://github.com/Tencent/WeKnora/commit/6c23dac3b0105039bba7e362b502bbb2222914d3) | docs(site): add the model leaderboard page (#4333) |
| 584 | 2026-10-15 | [a730a411d4](https://github.com/Tencent/WeKnora/commit/a730a411d421a0117318ed560f049530e83d46aa) | feat(agent): let the assistant manage FAQ entries (#4327) |
| 585 | 2026-10-15 | [450b6ed1b7](https://github.com/Tencent/WeKnora/commit/450b6ed1b781c36983c53d9d564ffe65a863b3e8) | feat(assistant): side panel, access levels and role-aware operations (#4342) |
| 586 | 2026-10-15 | [4d4acec418](https://github.com/Tencent/WeKnora/commit/4d4acec418b5fa67d457d1bf85846476d1e92e73) | feat(assistant): manage agents and skills, separate history, refresh pages (#4344) |
| 587 | 2026-10-15 | [aeeb72a61a](https://github.com/Tencent/WeKnora/commit/aeeb72a61a196f06518ec908f21f8e8f363e666c) | fix(assistant): reject restored ordinary sessions (#4356) |
| 588 | 2026-10-15 | [2fbfb44715](https://github.com/Tencent/WeKnora/commit/2fbfb4471581cd5e7d731bc7a2d443fe7d8e95b5) | feat(assistant): let the assistant manage knowledge bases (#4353) |
| 589 | 2026-10-16 | [db7046b85e](https://github.com/Tencent/WeKnora/commit/db7046b85effa8104fed2f51a38cd65226c97b6d) | refactor(image): turn the hard-coded image flow into configurable pipelines (#3746) |
| 590 | 2026-10-16 | [973d9d051f](https://github.com/Tencent/WeKnora/commit/973d9d051f2d2a1704bf3a92dd6b1c0b3d1988d4) | feat(datasource): deep-adapt Feishu/Lark sync with folder mirror, block parsing and embedded objects (#3497) |
| 591 | 2026-10-16 | [03059997fe](https://github.com/Tencent/WeKnora/commit/03059997fe1972f3280c3bcd29e7ecc4afcca432) | fix(im): cap IM callback bodies before any adapter reads them (#4369) |
| 592 | 2026-10-16 | [c44f939211](https://github.com/Tencent/WeKnora/commit/c44f939211bbb1a504c4171e0f5eef9c583287db) | fix(web-search): bound provider response bodies instead of buffering them whole (#4370) |
| 593 | 2026-10-16 | [d3d4963989](https://github.com/Tencent/WeKnora/commit/d3d496398977d07891e6e049d08bc74e382c5f77) | fix(mcp): avoid replaying tools after ambiguous call failures (#4124) |
| 594 | 2026-10-16 | [6dd0a6a3d6](https://github.com/Tencent/WeKnora/commit/6dd0a6a3d62876f9f046db4b57e2b5cd87abe8f1) | fix(embedding): read the pool's first-error guard under its own mutex (#4014) |
| 595 | 2026-10-16 | [12e3e49d80](https://github.com/Tencent/WeKnora/commit/12e3e49d80699f8cdec4f42673940a3464b80971) | fix(chat): persist failed quick-answer turns so errors survive refresh (#4306) |
| 596 | 2026-10-16 | [1b2c7af01a](https://github.com/Tencent/WeKnora/commit/1b2c7af01aa497fef46f088198ad226c277a83d7) | fix(data-analysis): resolve workspace storage backends in quick answers (#4214) |
| 597 | 2026-10-16 | [cf7aabeaca](https://github.com/Tencent/WeKnora/commit/cf7aabeaca98f5b33d2f26efc1713d45c32cd401) | fix(mcp): 分离SSE连接生命周期与请求超时 (#4140) |
| 598 | 2026-10-16 | [1fe804bd5c](https://github.com/Tencent/WeKnora/commit/1fe804bd5cab126dd1187f3956ef3b56a4ec963c) | fix(docparser): bound MinerU result packages and their expansion (#4396) |
| 599 | 2026-10-16 | [7ba8e2eba9](https://github.com/Tencent/WeKnora/commit/7ba8e2eba90869aa685c8a28534b8e4d437a2180) | fix(embed): hide host context in user messages (#4388) |
| 600 | 2026-10-16 | [7705f77be9](https://github.com/Tencent/WeKnora/commit/7705f77be962e85739f81f594e56c8bbfc08bb26) | fix(embed): send images when an OpenAI-compatible row declares them (#4389) |
| 601 | 2026-10-16 | [f0adfac334](https://github.com/Tencent/WeKnora/commit/f0adfac3348e7784a3e1f5cca1236b8bc5710143) | ﻿fix(models/api): cap non-streaming provider response bodies (#4397) |
| 602 | 2026-10-16 | [e07a190219](https://github.com/Tencent/WeKnora/commit/e07a190219a98bbebd87f6ec60701682ef0df093) | fix(tools): reject legacy binary .xls with an actionable error (#4374) |
| 603 | 2026-10-16 | [9a7aa48e44](https://github.com/Tencent/WeKnora/commit/9a7aa48e449482f1d41caae6bae3b3862ea1bee7) | build(docker): mount the Go build cache when building the app image (#4320) |
| 604 | 2026-10-16 | [7115cb2a61](https://github.com/Tencent/WeKnora/commit/7115cb2a61e86a7c963ea3793937546286d7238c) | feat(assistant): trial runs, session grants, page links, and MCP/model/skill management (#4375) |
| 605 | 2026-10-16 | [86767d77e9](https://github.com/Tencent/WeKnora/commit/86767d77e91b55c6012a7f22d991ea8517e3e96a) | fix(frontend): cap concurrent status requests in upload polling (#4348) |
| 606 | 2026-10-16 | [ad63fc4b8c](https://github.com/Tencent/WeKnora/commit/ad63fc4b8caf9a2c33557a6aa85251b62d8339b4) | fix(frontend): label the parser engine the backend will actually pick (#4354) |
| 607 | 2026-10-17 | [f5cfa0ff5f](https://github.com/Tencent/WeKnora/commit/f5cfa0ff5fa87c45f59d4176f4ac8bb922bd5dd8) | fix(knowledge): leave the root span open while a retry is left (#4357) |
| 608 | 2026-10-17 | [0d7516ea0a](https://github.com/Tencent/WeKnora/commit/0d7516ea0a07ef92e245704b87457969c4322cfb) | feat(assistant): expand management operations and polish interaction UI (#4427) |
| 609 | 2026-10-17 | [8776a2d92d](https://github.com/Tencent/WeKnora/commit/8776a2d92d11ad3e76865d51cf3192e37e2a9976) | feat(assistant): let the assistant find skills and MCP services on the web (#4429) |
| 610 | 2026-10-17 | [345855b59c](https://github.com/Tencent/WeKnora/commit/345855b59cdb3181f6a7e584595bb395957d7e55) | fix(anydoc): upgrade patched zip dependency to 9 (#4437) |
| 611 | 2026-10-17 | [578ca1b519](https://github.com/Tencent/WeKnora/commit/578ca1b5193b66b7b08309dfefa7d89db20f8d2a) | fix(docreader): keep a pipe inside an HTML table cell in that cell (#4433) |
| 612 | 2026-10-17 | [1fc00cdc08](https://github.com/Tencent/WeKnora/commit/1fc00cdc08d8f08ebde2303d95a7b9115062e050) | fix(agent): share one configurable budget for web_fetch (#4432) |
| 613 | 2026-10-17 | [e2a58ea3cb](https://github.com/Tencent/WeKnora/commit/e2a58ea3cbfa28f24b828ebc1a721a12fbcf41a1) | fix(docparser): keep JSON numbers and <, >, & as written in JSON uploads (#4436) |
| 614 | 2026-10-17 | [e6dc60be32](https://github.com/Tencent/WeKnora/commit/e6dc60be3278a598064a667369eb698d6c819eda) | fix(knowledge): 文件夹树显示递归文档数，不再把有子目录的文件夹显示成 0 (#4428) |
| 615 | 2026-10-17 | [5044ef69f9](https://github.com/Tencent/WeKnora/commit/5044ef69f9db13bac00f920c70c362aa136cd836) | fix(chat): 消息历史与搜索的 limit 补上下界，limit=-1 不再取消 SQL LIMIT (#4409) |
| 616 | 2026-10-17 | [2b4c0e362f](https://github.com/Tencent/WeKnora/commit/2b4c0e362f11e3ef30c1125a42d18ddb304ac9f6) | fix(wiki): 列 wiki 页面的 page_size 补上界，与兄弟路径一致 (#4407) |
| 617 | 2026-10-17 | [340b72f141](https://github.com/Tencent/WeKnora/commit/340b72f1419e80aebd973591e3bde0e3da9785b4) | fix(models): 图片解析复用仓储层的 provider scheme 表，对象存储上的图片不再原样交给厂商 (#4405) |
| 618 | 2026-10-17 | [8ddb245d1b](https://github.com/Tencent/WeKnora/commit/8ddb245d1b08e603e099dffd8e2ee8cf68cbe1da) | fix(mcp): clarify read_document query and pagination semantics (#4403) |
| 619 | 2026-10-17 | [2fac6381f7](https://github.com/Tencent/WeKnora/commit/2fac6381f7e086a204e7cc66257336535049fa2f) | fix(search): preserve chunk types in tool results (#4402) |
| 620 | 2026-10-17 | [b373374bf3](https://github.com/Tencent/WeKnora/commit/b373374bf38a1437281d475109089962e1cb8389) | fix(feishu): keep nested text in native table cells (#4400) |
| 621 | 2026-10-17 | [164dafca85](https://github.com/Tencent/WeKnora/commit/164dafca85a477ced3c5f7f7be65433fa4435276) | feat(graph): add standalone JSON graph search API (#4395) |
| 622 | 2026-10-17 | [da3317039b](https://github.com/Tencent/WeKnora/commit/da3317039bdb3a3063d675b747245c8036d187cc) | fix(docparser): keep every row and cell of a CSV upload in its table (#4435) |
| 623 | 2026-10-17 | [42f91f7aa7](https://github.com/Tencent/WeKnora/commit/42f91f7aa7158e3a9debf6a6f5c292ddb21461ae) | fix(wiki): read scanned-page OCR when the text chunks are only images (#4430) |
| 624 | 2026-10-17 | [a7d915a70a](https://github.com/Tencent/WeKnora/commit/a7d915a70a96081dd697617ffb1f94af8e29fe29) | feat(resource): 相同图片字节复用一个 handle，不再重复写字节与重复注册 (#4419) |
| 625 | 2026-10-17 | [a9290dcb60](https://github.com/Tencent/WeKnora/commit/a9290dcb60e98eeb6b7b52dc252ee956338e0cd0) | fix(docreader): Excel 百分比被读小 100 倍 —— 加开关（默认关） (#4423) |
| 626 | 2026-10-17 | [0aa1a517c8](https://github.com/Tencent/WeKnora/commit/0aa1a517c86799ae16a23d9c5b4d8c6dab594f27) | fix(helm): make the data-files claim access modes configurable (#4425) |
| 627 | 2026-10-17 | [e849eaad5c](https://github.com/Tencent/WeKnora/commit/e849eaad5cc01d4b01e8eb449041b8673aa263a9) | feat(datasource/dingtalk): 保留来源目录结构（开关控制，默认关） (#4422) |
| 628 | 2026-10-17 | [d92728d88c](https://github.com/Tencent/WeKnora/commit/d92728d88c6db1d4c97267cb707388e8ca4862ee) | fix(datasource/dingtalk): Validate 只校验所选范围，不再扫全租户 (#4421) |
| 629 | 2026-10-17 | [073b5ea616](https://github.com/Tencent/WeKnora/commit/073b5ea6169254e29059e26fcb053e13730db33c) | fix(datasource): 数据源类型图标的 alt 用本地化显示名，不再用裸 type (#4420) |
| 630 | 2026-10-17 | [4d767d9dc4](https://github.com/Tencent/WeKnora/commit/4d767d9dc4eaacceca4d77851dd66fa074b6b6b7) | fix(yuque): 两处 offset 分页补跳数上限、重复页检测与 meta.total 提前结束 (#4415) |
| 631 | 2026-10-17 | [b2e81a9ac4](https://github.com/Tencent/WeKnora/commit/b2e81a9ac4e24b70d41c296c375af876e887c608) | fix(notion): 附件下载失败中止本轮不再推进游标，附件不再永久缺失 (#4413) |
| 632 | 2026-10-17 | [26ff988823](https://github.com/Tencent/WeKnora/commit/26ff9888238de05db115784724e9abaca05b48ac) | fix(notion): 递归子块分页补重复游标与跳数上限，兜底不再被空页绕过 (#4411) |
| 633 | 2026-10-17 | [75e7ee4d41](https://github.com/Tencent/WeKnora/commit/75e7ee4d41a8de5677734a576dbdfc548812809d) | test(mcp): cover read_document query and pagination contract (#4426) |
| 634 | 2026-10-17 | [f89a15e61c](https://github.com/Tencent/WeKnora/commit/f89a15e61c9ff587e5ae15b6cf8e6ae1069b30c8) | fix(frontend): bundle PDF.js wasm decoders (#4382) |
| 635 | 2026-10-17 | [44d90beea2](https://github.com/Tencent/WeKnora/commit/44d90beea22a08b701fe3e33b650af9816a4f9f4) | feat(docreader): 上传 HTML 只索引正文，与 URL 导入路径对齐（开关，默认关） (#4424) |
| 636 | 2026-10-17 | [2c0d838439](https://github.com/Tencent/WeKnora/commit/2c0d8384398ca756ee414eef01b71caf91e2caa6) | fix(feishu): stop logging tenant access token fragments (#4399) |
| 637 | 2026-10-19 | [966d841d19](https://github.com/Tencent/WeKnora/commit/966d841d193065603a3d72deaed2c1036d42f408) | feat(assistant): add durable multi-agent collaboration and live task views (#4438) |
| 638 | 2026-10-19 | [2e8b3aa729](https://github.com/Tencent/WeKnora/commit/2e8b3aa729cc05d5787a00803d834bfa078c33d3) | fix(agent): let open_page finish the turn with the answer above it (#4472) |
| 639 | 2026-10-19 | [737d237283](https://github.com/Tencent/WeKnora/commit/737d237283ee1a8abd19a4755ad3197f79ce1303) | style(assistant): tone down the running child agent highlight (#4476) |
| 640 | 2026-10-19 | [f07f7ff028](https://github.com/Tencent/WeKnora/commit/f07f7ff0282ae6b0a4be0e49e80ac08bdde03b7c) | fix: reject invalid image inputs before multimodal extraction (#4489) |
| 641 | 2026-10-19 | [ba253bcbac](https://github.com/Tencent/WeKnora/commit/ba253bcbacd28d2d465b7a6955a1cfd9f75fe0f8) | feat(agent): unified agent permission model with grants and run modes (#4490) |
| 642 | 2026-10-19 | [09b9c98d54](https://github.com/Tencent/WeKnora/commit/09b9c98d545cbdef978068c3ec298313243a07a7) | fix: preserve business session identity in model requests (#4491) |
| 643 | 2026-10-19 | [728966bc81](https://github.com/Tencent/WeKnora/commit/728966bc813735cc0276143ec212ff96616cb5c1) | feat(chat): fold long streaming timelines and steady their height (#4496) |
| 644 | 2026-10-20 | [eecfe4b3c8](https://github.com/Tencent/WeKnora/commit/eecfe4b3c8c34c94e9b559e5dd02187b4ae447cd) | feat(automation): scheduled agent runs with in-app notifications (#4492) |
