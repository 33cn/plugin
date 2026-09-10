changelog

# [1.72.0](https://github.com/33cn/plugin/compare/v1.71.3...v1.72.0) (2026-09-04)


### Bug Fixes

* **fork:** remove duplicated ForkAccountBlacklist/ForkParaFee configs ([9e5ba9e](https://github.com/33cn/plugin/commit/9e5ba9e74d15e1328cd73c58684e4f9d7c933ca2))


### Features

* add EVM account blacklist and bump chain33 to aa71469 ([80d3dfa](https://github.com/33cn/plugin/commit/80d3dfa19c724b9e536a5dbc671cec0b8b9ec2a2))

## [1.71.3](https://github.com/33cn/plugin/compare/v1.71.2...v1.71.3) (2026-09-03)


### Bug Fixes

* **relay:** decode verify order addr with its own btc network params ([3b2dd17](https://github.com/33cn/plugin/commit/3b2dd177d7d2d7c91db9ffac110b5dddf328b03f))

## [1.71.2](https://github.com/33cn/plugin/compare/v1.71.1...v1.71.2) (2026-09-02)


### Bug Fixes

* resolve shellcheck warnings and UNKOWN typo in RPC placeholders ([d48f7e3](https://github.com/33cn/plugin/commit/d48f7e319a9eb70911d7c541a823797da7ba46b4))

## [1.71.1](https://github.com/33cn/plugin/compare/v1.71.0...v1.71.1) (2026-09-02)


### Bug Fixes

* **relay:** bind verifyBtcTx to txid and header time ([949da7f](https://github.com/33cn/plugin/commit/949da7f19cbdc48e97c257cdbb6a25a2f35a07ac))
* **relay:** recompute btc tx hash from rawtx in verifyBtcTx with fork ForkRelayVerifyBtcTx ([5281dad](https://github.com/33cn/plugin/commit/5281dadf4a0cf1cfbe1a293a050c5876b6336517))

# [1.71.0](https://github.com/33cn/plugin/compare/v1.70.0...v1.71.0) (2026-09-01)


### Bug Fixes

* move native asset test before restart, remove redundant block_wait, add mint send error handling ([88b9b20](https://github.com/33cn/plugin/commit/88b9b20b7fa24294b828471d142f7fc456eb15a4))
* **rgbx:** add native asset mint CI scenario with btcMintSpend command ([fb08b30](https://github.com/33cn/plugin/commit/fb08b30b9f2fe1a25bb7908bfbd4d9caf6d24e82))
* **rgbx:** address PR [#1299](https://github.com/33cn/plugin/issues/1299) review - nil-client log, dead code, merkle proof tests ([801a7da](https://github.com/33cn/plugin/commit/801a7da338a22cd12db56fd5b8c760be7cfb7821))
* **rgbx:** derive spendHash from merkle-certified txid, reject trailing tx bytes ([b34b6b6](https://github.com/33cn/plugin/commit/b34b6b6463c9816cb5342df1f2c15581687007cf))
* **rgbx:** improve OP_RETURN matching in createConfirmPayload, prefer matching commitment ([600f885](https://github.com/33cn/plugin/commit/600f885a6d7603ce7b1f37c4bc38a7122fea0705))
* update exec_test.go to compute expected spendHash from BtcTxProof.TxData ([d3dcd6c](https://github.com/33cn/plugin/commit/d3dcd6cc88c330cec577d25583a897116d448760))


### Features

* **rgbx:** add BTC Merkle proof verification to native asset confirm ([92fe0d4](https://github.com/33cn/plugin/commit/92fe0d40145f7f002f7628408d53e3c02f1770fc))

# [1.70.0](https://github.com/33cn/plugin/compare/v1.69.0...v1.70.0) (2026-08-27)


### Bug Fixes

* adapt go-ethereum API changes for v1.14.8 upgrade ([e4a0274](https://github.com/33cn/plugin/commit/e4a0274a7b6e08bff4687de183b488ecde7c4d78))
* adapt mix/zksync to gnark v0.9.0 and gnark-crypto v0.12.1 ([d5bd728](https://github.com/33cn/plugin/commit/d5bd7285826fc23373d75ed46dbb6c2154e766f9))
* address PR [#1295](https://github.com/33cn/plugin/issues/1295) review — concurrency safety, nil dereference, gas estimation, and defensive checks ([27e66f6](https://github.com/33cn/plugin/commit/27e66f6f442241cb2e0296c6c3159d8f6eebe43f))
* **build:** move build_ci CGO comment out of shell recipe ([41fc318](https://github.com/33cn/plugin/commit/41fc318587c2bf7ae15c0f98b75f0a6de01939b3))
* **ci:** add btcd health dependency for para1 in rgbx docker-compose ([fc6559a](https://github.com/33cn/plugin/commit/fc6559a80a1f7e6a8288d45b1f44acbe44a172d8))
* **ci:** fix semantic-release preset and extra_plugins ([aea8692](https://github.com/33cn/plugin/commit/aea86924e59ca7001178d1842f27abf4214d4370))
* **ci:** raise rgbx deposit wait retries to absorb timing variance ([079ca9a](https://github.com/33cn/plugin/commit/079ca9a414409f6921591707f0f62c986d6785b3))
* **config:** remove redundant DisableForkCheck in chain33 toml templates ([27adee1](https://github.com/33cn/plugin/commit/27adee1553d8061e662747c6614fa4511db113b8))
* **cross-chain:** add overflow protection in evmxgo and cross2eth ([9436e04](https://github.com/33cn/plugin/commit/9436e048a27cf28ccf559144ca7bea10ba2b92a7))
* **cross2eth:** adapt tests to go-ethereum v1.14.8 simulated backend ([28f1d6e](https://github.com/33cn/plugin/commit/28f1d6e4e7b02e52f4d22d725b4417c25601a6de))
* **cross2eth:** replace secp256k1.Sign with crypto.Sign for CGO=0 ([b89969f](https://github.com/33cn/plugin/commit/b89969f4658b7946f7e66fbca7d72c1f0825eb79))
* **evm:** add ForkEVMFixOverflow gate for uint64→int64 overflow protection ([d4ef8c1](https://github.com/33cn/plugin/commit/d4ef8c1a3f0b62c0bffbeb49c1e551eb9b9974eb))
* **evm:** fix EVM log address issue ([d70a0cd](https://github.com/33cn/plugin/commit/d70a0cd4f1c60d56c20f11d724debb88a45fd909))
* **evm:** fix gasmultiple config not taking effect and extract quickEstimateGasValue ([cd1a9fe](https://github.com/33cn/plugin/commit/cd1a9fe268611a041452a3a3bd8be679563ad3d7))
* **lightclient:** enhance BTC header validation and error handling ([6c01acd](https://github.com/33cn/plugin/commit/6c01acdfbef45417f016fc8a72b6614d9bb2faf8))
* **lightclient:** fix BTC transaction scan gap and add state management to prevent duplicate processing ([ea2f19a](https://github.com/33cn/plugin/commit/ea2f19aeb17f0f5b531f68e0075cd5144dd7cd79))
* **lightclient:** prevent withdraw duplicates and optimize UTXO selection ([9e444cb](https://github.com/33cn/plugin/commit/9e444cb1032d78662ac577c0ae5a90e91175b2eb))
* **mix:** add defensive guards per review and remove dead code ([299a54f](https://github.com/33cn/plugin/commit/299a54f5fa6ec063d36dfeb2b42480f14fb60522))
* **mix:** add groth16 VK/proof compatibility for old v0.5.2 format ([a68e7e0](https://github.com/33cn/plugin/commit/a68e7e0d4d502858802048c7ec884b84762cc482))
* **mix:** add ProvingKey compatibility for old v0.5.2 format ([1914e6b](https://github.com/33cn/plugin/commit/1914e6b7f54e2b78c33fefdf39a11df64fc3421d))
* **mix:** harden error handling in pub input conversion and proof verify ([b2d25c7](https://github.com/33cn/plugin/commit/b2d25c7907f63104faf1b7f5ea962e076d135313))
* **mix:** preserve historical compatibility for CBC and witness formats ([001ac93](https://github.com/33cn/plugin/commit/001ac939cf5ffdf08219ac95c18bfd976cf6adcf))
* **para:** disable ForkParaFee in test para nodes via SetFork ([6ac019f](https://github.com/33cn/plugin/commit/6ac019f3d6c31cfba5cb417016ba75e77d33bf73))
* **relay:** use proto.Equal for BTC header comparison ([8387ec4](https://github.com/33cn/plugin/commit/8387ec4e560742e81c994744f58eb16a1d7f167d))
* **rgbx:** CI waits relied on chain33-cli exit codes which are always 0 ([56c5654](https://github.com/33cn/plugin/commit/56c5654ca235fc9579070acff32197e48fd2610a))
* **rgbx:** correct CI wait logic per final review ([c83d784](https://github.com/33cn/plugin/commit/c83d784d41673e7e2ea6b7a518e55ae248487e41))
* **zksync:** preserve v0.5.3 eddsa key derivation for historical addresses ([0041b7f](https://github.com/33cn/plugin/commit/0041b7fa78484601137fd423f65e6b5f77f5500b))
* **zksync:** use GenerateKeyCompat in l2txs SignTransaction for historical key compat ([1f48f83](https://github.com/33cn/plugin/commit/1f48f832ccc934a29c1bd30ff85310aeeebf2289))


### Features

* add legacymimc to preserve zksync/mix MiMC hash compatibility ([3d6c8db](https://github.com/33cn/plugin/commit/3d6c8dbd94840f8bf19a27cdbf4149a7e022453c))
* **btcwallet:** enhance transaction fee estimation and validation ([3c82862](https://github.com/33cn/plugin/commit/3c82862354b509950e4e8c6345d9404ee7857890))
* **btcwallet:** enhance transaction monitoring and state management ([420edb4](https://github.com/33cn/plugin/commit/420edb4d047cb28672eb5b23f66cbd28515f15e8))
* **btcwallet:** implement BTC cross-chain gateway with TSS integration ([a7dd728](https://github.com/33cn/plugin/commit/a7dd72886c338ada32e584f814a78a3cc8f31a40))
* **btcwallet:** implement SPV proof and enhance transaction monitoring ([44bcde6](https://github.com/33cn/plugin/commit/44bcde674e708e44a6536ab9406be298502f7b68))
* **config:** add new fork configurations for light client and RGBX in chain33.proxyminer.toml ([f241175](https://github.com/33cn/plugin/commit/f241175f589664cceecf26592e0e42e2e991e407))
* **lightclient/rgbx:** implement complete BTC withdrawal flow ([9538e52](https://github.com/33cn/plugin/commit/9538e5237e750c0cdbab99982015186a88bb4ec7))
* **lightclient:** add BTC deposit function with RPC support ([324fb2a](https://github.com/33cn/plugin/commit/324fb2a028e80860575bea41d53ccd7c8d7119a1))
* **lightclient:** add tx hash return and withdraw sign deduplication ([3022c53](https://github.com/33cn/plugin/commit/3022c5330dfa45fff195b42255fb418ce12c230c))
* **lightclient:** add withdraw state persistence ([5bd9384](https://github.com/33cn/plugin/commit/5bd938401dc7ae3f2b2c2de406c9e4625bac781d))
* **lightclient:** implement Bitcoin header validation and context management ([7680b6d](https://github.com/33cn/plugin/commit/7680b6d36395c838ba0e3bafb34b30d16c5c4c90))
* **mix:** add genzkkey tool, update CI to regenerate groth16 keys ([e3652e9](https://github.com/33cn/plugin/commit/e3652e9a4749fe7d69f97abfb0d31153575de91f))
* **neutrino:** add comprehensive configuration documentation for RGBX and Neutrino ([5f4a04b](https://github.com/33cn/plugin/commit/5f4a04b5683c050dee3737c0c0014324437b0bc5))
* **neutrino:** enhance Bitcoin header submission and synchronization ([1227846](https://github.com/33cn/plugin/commit/122784667b46d55018ae2c78ca36347479df06f7))
* **neutrino:** implement concurrent signing tasks in TSS service ([b701fca](https://github.com/33cn/plugin/commit/b701fca3ac89bb8595858fdeb8262b86450f52a8))
* **neutrino:** validate withdraw sign notify and receiver-pays-fee ([bf09167](https://github.com/33cn/plugin/commit/bf09167aa1c25f7fcacca8ecf6ff04ba13d50348))
* **rgbx/lightclient:** implement production-ready cross-chain transaction validation ([2567803](https://github.com/33cn/plugin/commit/2567803d8ce88865f770a8305a3679cf34622de5))
* **rgbx:** add BTC address script conversion and deposit transaction commands ([f4cc4fa](https://github.com/33cn/plugin/commit/f4cc4faf81e2f987f68a278ea669e516ccb62861))
* **rgbx:** add cross-chain account system with deposit/withdraw asset management ([a4e117e](https://github.com/33cn/plugin/commit/a4e117e29645fc54dabf16fb0e8fdf4532900bd1))
* **rgbx:** add cross-chain asset support in transfer transaction ([28bb459](https://github.com/33cn/plugin/commit/28bb4597b3439bee4b6e8ada64334f27ff1de555))
* **rgbx:** add new configuration options and enhance build scripts ([d87f31b](https://github.com/33cn/plugin/commit/d87f31bf1c7193624f0a775de0f2a6ce29858f4c))
* **rgbx:** enhance transaction handling and add end height parameter ([df2fa48](https://github.com/33cn/plugin/commit/df2fa487040e459e0d236b712dc6461163394955))
* **rgbx:** harden CommitDKG validation and guardian title handling ([605c1de](https://github.com/33cn/plugin/commit/605c1ded4f0f0da56d9064531393dd82a9f5b74e))


### Reverts

* Revert "fix(mix): add ProvingKey compatibility for old v0.5.2 format" ([c2a7111](https://github.com/33cn/plugin/commit/c2a7111a728135f6d310abbc8e26eedd063d68dc))

<a name="1.69.0"></a>
# [1.69.0](https://github.com/33cn/plugin/compare/v1.68.4...v1.69.0) (2024-06-17)


### Features

* update chain33 to v1.69.0 with block finalize consensus ([86ec62d](https://github.com/33cn/plugin/commit/86ec62d))

<a name="1.68.4"></a>
## [1.68.4](https://github.com/33cn/plugin/compare/v1.68.3...v1.68.4) (2023-09-13)


### Bug Fixes

* sync chain33 patch version 1.68.1 ([82f248a](https://github.com/33cn/plugin/commit/82f248a))

<a name="1.68.3"></a>
## [1.68.3](https://github.com/33cn/plugin/compare/v1.68.2...v1.68.3) (2023-05-24)


### Bug Fixes

* fix evm nonce rollback ([3609b74](https://github.com/33cn/plugin/commit/3609b74))

<a name="1.68.2"></a>
## [1.68.2](https://github.com/33cn/plugin/compare/v1.68.1...v1.68.2) (2023-03-16)


### Bug Fixes

* init evm exec address with format ([36e8875](https://github.com/33cn/plugin/commit/36e8875))

<a name="1.68.1"></a>
## [1.68.1](https://github.com/33cn/plugin/compare/v1.68.0...v1.68.1) (2023-03-14)


### Bug Fixes

* fix rpc config for parachain ([cd9ad35](https://github.com/33cn/plugin/commit/cd9ad35))

<a name="1.68.0"></a>
# [1.68.0](https://github.com/33cn/plugin/compare/v1.67.6...v1.68.0) (2023-02-21)


### Features

* add consensus commiter plugin rollup(33cn/chain33#1268) ([b0aca7d](https://github.com/33cn/plugin/commit/b0aca7d)), closes [33cn/chain33#1268](https://github.com/33cn/chain33/issues/1268)

<a name="1.67.6"></a>
## [1.67.6](https://github.com/33cn/plugin/compare/v1.67.5...v1.67.6) (2023-01-31)


### Bug Fixes

* add return for get evm tx recevier ([533f35a](https://github.com/33cn/plugin/commit/533f35a))

<a name="1.67.5"></a>
## [1.67.5](https://github.com/33cn/plugin/compare/v1.67.4...v1.67.5) (2023-01-16)


### Bug Fixes

* add evm mix address fork ([7fecaf4](https://github.com/33cn/plugin/commit/7fecaf4))

<a name="1.67.4"></a>
## [1.67.4](https://github.com/33cn/plugin/compare/v1.67.3...v1.67.4) (2022-10-11)


### Bug Fixes

* update chain33 patch version ([385028c](https://github.com/33cn/plugin/commit/385028c))

<a name="1.67.3"></a>
## [1.67.3](https://github.com/33cn/plugin/compare/v1.67.2...v1.67.3) (2022-05-27)


### Bug Fixes

* update chain33 patch version 1.67.3 ([303be37](https://github.com/33cn/plugin/commit/303be37))

<a name="1.67.2"></a>
## [1.67.2](https://github.com/33cn/plugin/compare/v1.67.1...v1.67.2) (2022-04-18)


### Bug Fixes

* sync chain33 patch version 1.67.2 ([0a0d7ec](https://github.com/33cn/plugin/commit/0a0d7ec))

<a name="1.67.1"></a>
## [1.67.1](https://github.com/33cn/plugin/compare/v1.67.0...v1.67.1) (2022-03-29)


### Bug Fixes

* update chain33 version to 1.67.1 ([f4394f8](https://github.com/33cn/plugin/commit/f4394f8))

<a name="1.67.0"></a>
# [1.67.0](https://github.com/33cn/plugin/compare/v1.66.3...v1.67.0) (2022-03-21)


### Features

* sync chain33 v1.67.0 ([c869934](https://github.com/33cn/plugin/commit/c869934))

<a name="1.66.3"></a>
## [1.66.3](https://github.com/33cn/plugin/compare/v1.66.2...v1.66.3) (2022-02-18)


### Bug Fixes

* update chain33 ([d0526e5](https://github.com/33cn/plugin/commit/d0526e5))

<a name="1.66.2"></a>
## [1.66.2](https://github.com/33cn/plugin/compare/v1.66.1...v1.66.2) (2022-01-20)


### Bug Fixes

* sync chain33 patch version v1.66.3 ([691168d](https://github.com/33cn/plugin/commit/691168d))

<a name="1.66.1"></a>
## [1.66.1](https://github.com/33cn/plugin/compare/v1.66.0...v1.66.1) (2022-01-10)


### Bug Fixes

* chain33 update to 1.66.2 ([3cdb80c](https://github.com/33cn/plugin/commit/3cdb80c))

<a name="1.66.0"></a>
# [1.66.0](https://github.com/33cn/plugin/compare/v1.65.4...v1.66.0) (2021-12-30)


### Features

* release 1.66 ([b94376e](https://github.com/33cn/plugin/commit/b94376e))

<a name="1.65.4"></a>
## [1.65.4](https://github.com/33cn/plugin/compare/v1.65.3...v1.65.4) (2021-10-27)


### Bug Fixes

* adjust semantic-release commit message conventions to jshint ([a0eade8](https://github.com/33cn/plugin/commit/a0eade8))
* Fixed semantic-release-replace-plugin config ([6c79cfd](https://github.com/33cn/plugin/commit/6c79cfd))

## [1.65.3](https://github.com/33cn/plugin/compare/v1.65.2...v1.65.3) (2021-10-19)


### Bug Fixes

* 🐛version control: Add github action for auto publish release and tag version ([72ab4fd](https://github.com/33cn/plugin/commit/72ab4fdf9625b348b06ae4b8ae90522a7aa3db6f))
* **doc:** release 1.65.3 ([7484035](https://github.com/33cn/plugin/commit/74840359adb86d9d920fe63b04fd790e8933fe53))
* ebrelayer log ([a8ee06d](https://github.com/33cn/plugin/commit/a8ee06da773bb015b6ec45762a87bbca54ea2268))
* fix ci and add manually auto publish release ([677029b](https://github.com/33cn/plugin/commit/677029bb4c2e6653626b0f0ef4a296f06102c604))
* lottery ci ([8941c81](https://github.com/33cn/plugin/commit/8941c81c70c6ab5a4e07b4d88cdf82b6e5a9f862))

## [6.1.1](https://github.com/33cn/plugin/compare/v6.1.0...v6.1.1) (2021-10-18)


### Bug Fixes

* 🐛version control: Add github action for auto publish release and tag version ([72ab4fd](https://github.com/33cn/plugin/commit/72ab4fdf9625b348b06ae4b8ae90522a7aa3db6f))
* **doc:** release 1.65.3 ([7484035](https://github.com/33cn/plugin/commit/74840359adb86d9d920fe63b04fd790e8933fe53))
* ebrelayer log ([a8ee06d](https://github.com/33cn/plugin/commit/a8ee06da773bb015b6ec45762a87bbca54ea2268))
* fix ci and add manually auto publish release ([677029b](https://github.com/33cn/plugin/commit/677029bb4c2e6653626b0f0ef4a296f06102c604))
* lottery ci ([8941c81](https://github.com/33cn/plugin/commit/8941c81c70c6ab5a4e07b4d88cdf82b6e5a9f862))
