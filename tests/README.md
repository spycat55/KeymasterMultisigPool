# 测试文档

本项目包含双签和三签多重签名池的完整测试套件，包括TypeScript和Go两个版本的单元测试。

## 测试结构

```
tests/
├── dual_endpoint/
│   └── dual_endpoint.test.ts     # 双签TypeScript单元测试
├── triple_endpoint/
│   └── triple_endpoint.test.ts   # 三签TypeScript单元测试
└── README.md                     # 本文档

pkg/
├── dual_endpoint/
│   └── dual_endpoint_test.go     # 双签Go单元测试
└── triple_endpoint/
    └── triple_endpoint_test.go   # 三签Go单元测试
```

## 测试数据

测试使用固定的测试数据，这些数据来自于成功通过的集成测试：

### 双签测试数据
- **客户端私钥**: `903b1b2c396f17203fa83444d72bf5c666119d9d681eb715520f99ae6f92322c`
- **服务器私钥**: `a2d2ca4c19e3c560792ca751842c29b9da94be09f712a7f9ba7c66e64a354829`
- **测试UTXO**: `0a1fd93f02e68d1a73fb499e948ee83a78aa9337e1476bd89f7092a7ef16a050:1` (99902 satoshis)
- **费率**: 0.5 sat/byte
- **结束高度**: 800000

### 三签测试数据
- **客户端私钥**: `a682814ac246ca65543197e593aa3b2633b891959c183416f54e2c63a8de1d8c`
- **服务器私钥**: `903b1b2c396f17203fa83444d72bf5c666119d9d681eb715520f99ae6f92322c`
- **托管私钥**: `a2d2ca4c19e3c560792ca751842c29b9da94be09f712a7f9ba7c66e64a354829`
- **测试UTXO**: `ffcfe296a596f01e5cef2d14f39bc61f55c8f0535a5f723c1b5b043b77053595:1` (19996 satoshis)
- **费率**: 1.2 sat/byte

## 预期输出

### 双签预期输出
- **Step1 (基础交易)**: `010000000150a016efa792709fd86b47e13793aa783ae88e949e49fb731a8de6023fd91f0a010000006b483045022100a5a54548db07e6c063ac05f646dbffbdb47a09398071ac89ce84a9f670ca3ef4022060a676631c36312a9b9c3f539e24b2bf24f732730949421cd4c27319e7ff91764121039e00beaeaab4162fa3d45326e3632303c394faf8f7a17bbcf27a01952a1e7646ffffffff013d8601000000000047522103f6552f24751f8618fe0b2a813c9c3e163fbeec92ab737af7990297568a63d62121039e00beaeaab4162fa3d45326e3632303c394faf8f7a17bbcf27a01952a1e764652ae00000000`
- **Step2 (花费交易)**: `0100000001201f8665d4d165761ab252e67405b38b0afdeeb8e3c62fd691e159367bee98d900000000920047304402205327dbfb02a36f64c6841c1b3f3559a56bdb5c282b6174cef1340903b1fe675b0220121ae327d8c124deba3eb2ec42dc6dbd7bcfb56e97cb9e45c60bd280217e3afc414830450221008186673e2a874c6c64ef008c4c67ae8235c4aa7fd34428b501842b8ea54dd8ec022016bb42d2a5919245a35514fa0ef26002c04081b52006aa3ff2c1046e9fd1ea0241010000000200000000000000001976a914789d07c284ff3f6c41633e2031b375e57434759688ac3c860100000000001976a9147e06a09c32ea06e80745cbfae60036968b64238888ac00350c00`

### 三签预期输出
- **Step1 (基础交易)**: `0100000001953505773b045b1b3c725f5a53f0c8551fc69bf3142def5c1ef096a596e2cfff010000006b483045022100f76e44c1685e658326c4bac76290df0cb4cc58738735dd42759028e723d4dabc0220170b6cfb649b0c4ec9934bb16e33430e1d89ec86d8a56104808ef7b5fda3762e4121032a33be07d7a12cbb2f178b8c6568223d1b8aa954cb929bebf7f3f855b2dae042ffffffff011b4e000000000000695221039e00beaeaab4162fa3d45326e3632303c394faf8f7a17bbcf27a01952a1e764621032a33be07d7a12cbb2f178b8c6568223d1b8aa954cb929bebf7f3f855b2dae0422103f6552f24751f8618fe0b2a813c9c3e163fbeec92ab737af7990297568a63d62153ae00000000`
- **Step3 (完整签名交易)**: `0100000001193bf65040f4c309fb4834a195eab9753fd3b5162551c10aade89d99f5afa6710000000091004730440220409f0a3c8b55e63f5c1b100fc81ff7bb3869e01e1267b9c790a2005798a887ad02204f75ae37d89b531b07c9740a5709aad6a2ab071cea5bbc124b16c8c2f0155d4d41473044022021a8ace2a74afed19531e202b544d6705d1c90ecdf0e41112d3cc8884e1cc0cb02200f68afe9712f9597a4c029984c3d0323d5499d8de6d8c9b6d4cdeb07e1f04de041010000000200000000000000001976a914789d07c284ff3f6c41633e2031b375e57434759688ac1a4e0000000000001976a914a8d0cb37061679d0523314d882d81b989254df7b88ac00000000`

## 运行测试

### 运行TypeScript测试
```bash
# 运行所有TypeScript测试
npm test tests/

# 运行双签测试
npm test tests/dual_endpoint/dual_endpoint.test.ts

# 运行三签测试
npm test tests/triple_endpoint/triple_endpoint.test.ts
```

### 运行Go测试
```bash
# 运行所有Go测试
go test ./pkg/... -v

# 运行双签测试
go test ./pkg/dual_endpoint -v

# 运行三签测试
go test ./pkg/triple_endpoint -v
```

### 运行集成测试
```bash
# 运行双签集成测试
go run examples/txtest/compare.go

# 运行三签集成测试
go run examples/triplextest/compare.go
```

### 运行所有测试
```bash
# 使用测试脚本运行所有测试
./scripts/run_all_tests.sh
```

## 测试覆盖范围

### TypeScript测试覆盖
- ✅ 基础交易构建
- ✅ 花费交易构建
- ✅ 完整签名流程
- ✅ 不同费率处理
- ✅ 输入参数验证
- ✅ 多重签名脚本结构验证

### Go测试覆盖
- ✅ 包编译验证
- 🚧 详细功能测试（待完善）

### 集成测试覆盖
- ✅ TypeScript与Go输出一致性验证
- ✅ 端到端交易流程验证

## 注意事项

1. **确定性输出**: 由于使用固定的私钥和UTXO，测试输出是确定性的
2. **费率计算**: 系统有最小费率限制，即使设置为0也会计算最小费用
3. **签名差异**: 由于签名算法的随机性，某些测试允许轻微的输出差异
4. **Go测试**: 当前Go测试主要验证包编译，详细功能测试需要根据实际函数签名进一步完善

## 测试数据来源

所有测试数据都来自于已经通过验证的集成测试：
- `examples/txtest/compare.go` - 双签集成测试
- `examples/triplextest/compare.go` - 三签集成测试

这确保了单元测试与实际功能的一致性。