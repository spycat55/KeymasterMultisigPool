# 在线区块链测试

使用真实的测试网络进行双端费用池完整流程测试。

## 运行方式

```bash
FEEPOOL_CLIENT_PRIV=2796e78fad7d383fa5236607eba52d9a1904325daf9b4da3d77be5ad15ab1dae \
FEEPOOL_SERVER_PRIV=e6d4d7685894d2644d1f4bf31c0b87f3f6aa8a3d7d4091eaa375e81d6c9f9091 \
go run examples/online_dual_test/main.go
```

## 功能

- 从 whatsonchain.com API 获取真实 UTXO
- 获取当前区块高度
- 广播交易到测试网络
- 完整的端到端测试

## 注意

需要客户端地址有足够的测试币才能运行。