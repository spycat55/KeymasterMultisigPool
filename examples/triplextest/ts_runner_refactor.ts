import { readFileSync } from 'fs';
import path from 'path';
import { PrivateKey } from '@bsv/sdk/primitives';
import Transaction from '@bsv/sdk/transaction/Transaction';
import Script from '@bsv/sdk/script/Script';

import MultiSig from '../../src/libs/MULTISIG';

// Re-export high-level helpers from the pool implementation
import {
  tripleBuildFeePoolBaseTx,
  tripleBuildFeePoolSpendTX,
  tripleSpendTXFeePoolBSign,
} from '../../src/triple_endpoint';

interface FixtureUTXO {
  txid: string;
  vout: number;
  satoshis: number;
}
interface Fixture {
  clientPrivHex: string;
  serverPrivHex: string;
  escrowPrivHex: string;
  clientUtxos: FixtureUTXO[];
  feePerByte: number;
}

/**
 * Read fixture.json that sits in the same directory as this runner.
 */
function loadFixture(): Fixture {
  const dir = path.resolve(__dirname);
  const data = readFileSync(path.join(dir, 'fixture.json'), 'utf8');
  return JSON.parse(data);
}

(async () => {
  const fixture = loadFixture();

  const clientPriv = PrivateKey.fromHex(fixture.clientPrivHex);
  const serverPriv = PrivateKey.fromHex(fixture.serverPrivHex);
  const escrowPriv = PrivateKey.fromHex(fixture.escrowPrivHex);

  const feeRate = fixture.feePerByte;

  /* ------------------------------------------------------------------
   * Step-1  Build pool funding (base) transaction
   * ------------------------------------------------------------------ */
  const { tx: baseTx } = await tripleBuildFeePoolBaseTx(
    fixture.clientUtxos,
    serverPriv.toPublicKey(),      // server pubkey comes first per API
    clientPriv,                    // client private key (A-party)
    escrowPriv.toPublicKey(),      // third-party pubkey
    feeRate,
  );
  console.log('Step1Hex:', baseTx.toHex());

  /* ------------------------------------------------------------------
   * Step-2  Client constructs spend transaction (client side partially signed)
   * ------------------------------------------------------------------ */
  const poolValue = baseTx.outputs[0].satoshis as number;

  const spendResp = await tripleBuildFeePoolSpendTX(
    baseTx.id('hex'),              // previous txid
    poolValue,    // value locked in pool output
    0,                             // lock-time / end height (0 for immediate)
    serverPriv.toPublicKey(),
    clientPriv,
    escrowPriv.toPublicKey(),
    feeRate,
  );

  const spendTx = spendResp.tx;
  const clientSig = spendResp.clientSignBytes;

  console.log('Step2Hex:', spendTx.toHex());

  /* ------------------------------------------------------------------
   * Step-3  Server adds its signature
   * ------------------------------------------------------------------ */


  const serverSig = await tripleSpendTXFeePoolBSign(
    spendTx,
    poolValue,                     // 使用原始池子金额（step1.Amount）而不是扣费后的金额
    serverPriv.toPublicKey(),
    clientPriv.toPublicKey(),
    escrowPriv.toPublicKey(),      // 传递 escrow 公钥而不是 server 私钥
    escrowPriv,                    // 使用 escrow 私钥进行签名，与 Go 保持一致
  );

  // Combine signatures into final unlocking script
  const unlockingScript = MultiSig.buildSignScript([clientSig, serverSig]);
  (spendTx.inputs[0] as any).unlockingScript = unlockingScript as unknown as Script;

  console.log('Step3Hex:', spendTx.toHex());
})();
