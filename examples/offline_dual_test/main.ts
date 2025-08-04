import { PrivateKey } from '@bsv/sdk/primitives';
import { buildDualFeePoolBaseTx } from '../../src/dual_endpoint/1base_tx';
import { buildDualFeePoolSpendTX } from '../../src/dual_endpoint/2client_spend_tx';
import { spendTXServerSign } from '../../src/dual_endpoint/3server_sign';
import { loadTx, clientDualFeePoolSpendTXUpdateSign } from '../../src/dual_endpoint/4client_spend_tx_update';
import { serverDualFeePoolSpendTXUpdateSign } from '../../src/dual_endpoint/5server_sign_update';
import Script from '@bsv/sdk/script/Script';
import OP from '@bsv/sdk/script/OP';

const FEE_RATE = 0.5;
const END_HEIGHT = 1687365;

// 固定的测试数据
const CLIENT_PRIV_HEX = "2796e78fad7d383fa5236607eba52d9a1904325daf9b4da3d77be5ad15ab1dae";
const SERVER_PRIV_HEX = "e6d4d7685894d2644d1f4bf31c0b87f3f6aa8a3d7d4091eaa375e81d6c9f9091";

// 固定的 UTXO
const FIXED_UTXO = {
  txid: "3bc591b12d1d356c80eec9628a626c2676c27e21fe8e0ef34d6dab2e425d9629",
  vout: 1,
  satoshis: 206106,
};

// 期望的结果
const EXPECTED_RESULTS = {
  step1_tx: "010000000129965d422eab6d4df30e8efe217ec276266c628a62c9ee806c351d2db191c53b010000006a47304402204bbb904e6c0994bd3a7447b3cb3d5ae73525647e7d80973d1850a10dd754d00b022004028b0ac634ed5b9487caef95912300799f50018e025f07906196de28b673c94121028bd4b450d28a69ed1a5cc9f256d0f3f94c4dedb885aae7144868a511b03511b0ffffffff0119250300000000004752210257db5aff3592dcb574f54b0a448789d4049637acec8a4e66e192591ad56f2c2e21028bd4b450d28a69ed1a5cc9f256d0f3f94c4dedb885aae7144868a511b03511b052ae00000000",
  step1_amount: "206105",
  step2_client_sig: "304402206d7c4d4430f011e4a29fad9e959bf300789e2bee638f6ddb245316b99630fc7802205daee22202dbe58af3b60d63719d2233d9012335e0be29cbf12c08dec389d6cc41",
  step2_client_amount: "205104",
  step3_server_sig: "3045022100a17575b8be3889a4a388b53b8d71f1d547501feb48b0157bc7faeaeaf49f9e6d0220762d0f7698cece09463200a0546c76ec434bea74d3820cd60a9e60b86bd9dc7241",
  step3_complete_tx: "0100000001de468f3b75f979273edcadcd4a84d47ad46e48eaac6f2e759a35d4cd136e0c04000000009200483045022100a17575b8be3889a4a388b53b8d71f1d547501feb48b0157bc7faeaeaf49f9e6d0220762d0f7698cece09463200a0546c76ec434bea74d3820cd60a9e60b86bd9dc724147304402206d7c4d4430f011e4a29fad9e959bf300789e2bee638f6ddb245316b99630fc7802205daee22202dbe58af3b60d63719d2233d9012335e0be29cbf12c08dec389d6cc410100000002e8030000000000001976a91419c50751f7e477d98ec0189cdfa0967808a2a1e388ac30210300000000001976a914e803a69218895a1a8d3df0f33a5b3d95bbb5a9c688ac45bf1900",
  step4_client_update_sig: "3045022100e47d9d731eab999facc649397405194298f85756dbd877383e4d0d1594598b52022031555f9c5da7bfde25d101a9c6e17983dc54f98dc0f0785f545ed4f5cebaa95441",
  step5_server_update_sig: "3044022007bfd4595d03fa76716ab618dd11283a3a14fd1658d1c171c978011b76d5a00902207cbb2a1dfdfa9d7822a830acd1d6b709e2ef18b3bb78d8aa627cd866743a190741",
  step5_complete_tx: "0100000001de468f3b75f979273edcadcd4a84d47ad46e48eaac6f2e759a35d4cd136e0c04000000009200473044022007bfd4595d03fa76716ab618dd11283a3a14fd1658d1c171c978011b76d5a00902207cbb2a1dfdfa9d7822a830acd1d6b709e2ef18b3bb78d8aa627cd866743a190741483045022100e47d9d731eab999facc649397405194298f85756dbd877383e4d0d1594598b52022031555f9c5da7bfde25d101a9c6e17983dc54f98dc0f0785f545ed4f5cebaa954410200000002dc050000000000001976a91419c50751f7e477d98ec0189cdfa0967808a2a1e388ac3c1f0300000000001976a914e803a69218895a1a8d3df0f33a5b3d95bbb5a9c688ac45bf1900",
  final_client_sig: "3044022076e9e67c7f183a7bd0a413988dbd764214860ca265d326adef2db92f0de40b5202200c40bb973bfd0cb57dd2f5866195dd7996b582fd3385aa22d8d09c946457de7241",
  final_server_sig: "30450221009cafbdf540418b0ffc58f28c6ad511af23030084c1b86ee1ca780db92f714f3302206e10f4a1b93d787956bf88b7158c20250127e06f9d91fd556ca09fa6db9221ce41",
  final_tx: "0100000001de468f3b75f979273edcadcd4a84d47ad46e48eaac6f2e759a35d4cd136e0c040000000092004830450221009cafbdf540418b0ffc58f28c6ad511af23030084c1b86ee1ca780db92f714f3302206e10f4a1b93d787956bf88b7158c20250127e06f9d91fd556ca09fa6db9221ce41473044022076e9e67c7f183a7bd0a413988dbd764214860ca265d326adef2db92f0de40b5202200c40bb973bfd0cb57dd2f5866195dd7996b582fd3385aa22d8d09c946457de7241ffffffff02dc050000000000001976a91419c50751f7e477d98ec0189cdfa0967808a2a1e388ac3c1f0300000000001976a914e803a69218895a1a8d3df0f33a5b3d95bbb5a9c688acffffffff",
};

function createUnlockScript(sigServer: number[], sigClient: number[]): Script {
  const s = new Script([]);
  s.writeOpCode(OP.OP_0);
  s.writeBin(sigServer);
  s.writeBin(sigClient);
  return s;
}

function assertEqual(name: string, expected: string, actual: string): boolean {
  if (expected === actual) {
    console.log(`✅ ${name}: PASS`);
    return true;
  } else {
    console.log(`❌ ${name}: FAIL`);
    console.log(`   Expected: ${expected}`);
    console.log(`   Actual:   ${actual}`);
    return false;
  }
}

(async () => {
  console.log("=== Offline Dual Fee Pool Test (TypeScript) ===");
  console.log("Testing with fixed UTXO and expected results");
  console.log();

  // 初始化私钥
  const clientPriv = PrivateKey.fromHex(CLIENT_PRIV_HEX);
  const serverPriv = PrivateKey.fromHex(SERVER_PRIV_HEX);

  // 获取地址
  const clientAddress = clientPriv.toPublicKey().toAddress();
  const serverAddress = serverPriv.toPublicKey().toAddress();

  console.log(`Client Address: ${clientAddress}`);
  console.log(`Server Address: ${serverAddress}`);
  console.log(`Fixed UTXO: ${FIXED_UTXO.txid}:${FIXED_UTXO.vout} (${FIXED_UTXO.satoshis} satoshis)`);
  console.log();

  let testsPassed = 0;
  let totalTests = 0;

  // Step 1: 创建基础多签交易
  console.log("=== Step 1: Base Transaction ===");
  const clientUTXOs = [FIXED_UTXO];
  const res1 = await buildDualFeePoolBaseTx(
    clientUTXOs,
    clientPriv,
    serverPriv.toPublicKey(),
    FEE_RATE,
  );

  totalTests++;
  if (assertEqual("Step1 Transaction", EXPECTED_RESULTS.step1_tx, res1.tx.toHex())) {
    testsPassed++;
  }

  totalTests++;
  if (assertEqual("Step1 Amount", EXPECTED_RESULTS.step1_amount, res1.amount.toString())) {
    testsPassed++;
  }

  // Step 2: 客户端签名花费交易
  console.log("\n=== Step 2: Client Spend Transaction ===");
  const serverAmount = 1000;
  
  const res2 = await buildDualFeePoolSpendTX(
    res1.tx,
    res1.amount,
    serverAmount,
    END_HEIGHT,
    clientPriv,
    serverPriv.toPublicKey(),
    FEE_RATE,
  );

  totalTests++;
  if (assertEqual("Step2 Client Signature", EXPECTED_RESULTS.step2_client_sig, Buffer.from(res2.clientSignBytes).toString('hex'))) {
    testsPassed++;
  }

  totalTests++;
  if (assertEqual("Step2 Client Amount", EXPECTED_RESULTS.step2_client_amount, res2.amount.toString())) {
    testsPassed++;
  }

  // Step 3: 服务器签名
  console.log("\n=== Step 3: Server Sign ===");
  const serverSignBytes = spendTXServerSign(
    res2.tx,
    res1.amount,
    serverPriv,
    clientPriv.toPublicKey(),
  );

  totalTests++;
  if (assertEqual("Step3 Server Signature", EXPECTED_RESULTS.step3_server_sig, Buffer.from(serverSignBytes).toString('hex'))) {
    testsPassed++;
  }

  // 组合签名创建完整交易
  const unlockScript = createUnlockScript(serverSignBytes, res2.clientSignBytes);
  res2.tx.inputs[0].unlockingScript!.chunks = unlockScript.chunks;

  totalTests++;
  if (assertEqual("Step3 Complete Transaction", EXPECTED_RESULTS.step3_complete_tx, res2.tx.toHex())) {
    testsPassed++;
  }

  // Step 4: 客户端更新签名
  console.log("\n=== Step 4: Client Update Sign ===");
  const newServerAmount = 1500;
  const newSequenceNumber = 2;

  const updatedTx = loadTx(
    res2.tx.toHex(),
    undefined, // locktime
    newSequenceNumber,
    newServerAmount,
    serverPriv.toPublicKey(),
    clientPriv.toPublicKey(),
    res1.amount
  );

  const clientUpdateSignBytes = clientDualFeePoolSpendTXUpdateSign(
    updatedTx,
    clientPriv,
    serverPriv.toPublicKey()
  );

  totalTests++;
  if (assertEqual("Step4 Client Update Signature", EXPECTED_RESULTS.step4_client_update_sig, Buffer.from(clientUpdateSignBytes).toString('hex'))) {
    testsPassed++;
  }

  // Step 5: 服务器更新签名
  console.log("\n=== Step 5: Server Update Sign ===");
  const serverUpdateSignBytes = serverDualFeePoolSpendTXUpdateSign(
    updatedTx,
    serverPriv,
    clientPriv.toPublicKey()
  );

  totalTests++;
  if (assertEqual("Step5 Server Update Signature", EXPECTED_RESULTS.step5_server_update_sig, Buffer.from(serverUpdateSignBytes).toString('hex'))) {
    testsPassed++;
  }

  // 组合更新后的签名
  const updateUnlockScript = createUnlockScript(serverUpdateSignBytes, clientUpdateSignBytes);
  updatedTx.inputs[0].unlockingScript!.chunks = updateUnlockScript.chunks;

  totalTests++;
  if (assertEqual("Step5 Complete Transaction", EXPECTED_RESULTS.step5_complete_tx, updatedTx.toHex())) {
    testsPassed++;
  }

  // 最终步骤：关闭费用池
  console.log("\n=== Final Step: Close Fee Pool ===");
  const finalLocktime = 0xffffffff;
  const finalSequence = 0xffffffff;

  const finalTx = loadTx(
    res2.tx.toHex(),
    finalLocktime,
    finalSequence,
    newServerAmount,
    serverPriv.toPublicKey(),
    clientPriv.toPublicKey(),
    res1.amount
  );

  // 客户端最终签名
  const clientFinalSignBytes = clientDualFeePoolSpendTXUpdateSign(
    finalTx,
    clientPriv,
    serverPriv.toPublicKey()
  );

  totalTests++;
  if (assertEqual("Final Client Signature", EXPECTED_RESULTS.final_client_sig, Buffer.from(clientFinalSignBytes).toString('hex'))) {
    testsPassed++;
  }

  // 服务器最终签名
  const serverFinalSignBytes = serverDualFeePoolSpendTXUpdateSign(
    finalTx,
    serverPriv,
    clientPriv.toPublicKey()
  );

  totalTests++;
  if (assertEqual("Final Server Signature", EXPECTED_RESULTS.final_server_sig, Buffer.from(serverFinalSignBytes).toString('hex'))) {
    testsPassed++;
  }

  // 组合最终签名
  const finalUnlockScript = createUnlockScript(serverFinalSignBytes, clientFinalSignBytes);
  finalTx.inputs[0].unlockingScript!.chunks = finalUnlockScript.chunks;

  totalTests++;
  if (assertEqual("Final Transaction", EXPECTED_RESULTS.final_tx, finalTx.toHex())) {
    testsPassed++;
  }

  // 测试结果总结
  console.log("\n============================================================");
  console.log(`TEST RESULTS: ${testsPassed}/${totalTests} tests passed`);
  
  if (testsPassed === totalTests) {
    console.log("🎉 ALL TESTS PASSED!");
    console.log("\nTransaction Summary:");
    console.log(`Base TX ID:    ${res1.tx.id('hex')}`);
    console.log(`Spend TX ID:   ${res2.tx.id('hex')}`);
    console.log(`Updated TX ID: ${updatedTx.id('hex')}`);
    console.log(`Final TX ID:   ${finalTx.id('hex')}`);
    console.log(`\nFinal Distribution:`);
    console.log(`Server Amount: ${newServerAmount} satoshis`);
    console.log(`Client Amount: ${res1.amount - newServerAmount} satoshis`);
  } else {
    console.log(`❌ ${totalTests - testsPassed} tests failed`);
    process.exit(1);
  }
})();