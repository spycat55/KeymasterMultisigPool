import fs from 'fs';
import path from 'path';
import { PrivateKey, PublicKey } from '@bsv/sdk/primitives';
import { buildDualFeePoolBaseTx } from '../../src/dual_endpoint/1base_tx';
import { buildDualFeePoolSpendTX } from '../../src/dual_endpoint/2client_spend_tx';
import * as base from '../../src/dual_endpoint/1base_tx';
import Transaction from '@bsv/sdk/transaction/Transaction';
import Script from '@bsv/sdk/script/Script';
import OP from '@bsv/sdk/script/OP';
import { TransactionSignature } from '@bsv/sdk';
import { hash256 } from '@bsv/sdk/primitives/Hash';
import * as ECDSA from '@bsv/sdk/primitives/ECDSA';
import BigNumber from '@bsv/sdk/primitives/BigNumber';

interface FixtureUTXO {
  txid: string;
  vout: number;
  satoshis: number;
}
interface Fixture {
  clientPrivHex: string;
  serverPrivHex: string;
  clientUtxos: FixtureUTXO[];
  endHeight: number;
  feeRate: number;
  isMain: boolean;
}

function createUnlockScript(sigServer: number[], sigClient: number[]): Script {
  const s = new Script([]);
  s.writeOpCode(OP.OP_0);
  s.writeBin(sigServer);
  s.writeBin(sigClient);
  return s;
}

(async () => {
  const fixturePath = path.resolve(__dirname, 'fixture.json');
  const raw = fs.readFileSync(fixturePath, 'utf-8');
  const f: Fixture = JSON.parse(raw);

  const clientPriv = PrivateKey.fromHex(f.clientPrivHex);
  const serverPriv = PrivateKey.fromHex(f.serverPrivHex);

  // Step1
  const res1 = await buildDualFeePoolBaseTx(
    f.clientUtxos,
    clientPriv,
    serverPriv.toPublicKey(),
    f.feeRate,
  );

  // Step2 (without final unlock)
  const res2 = await buildDualFeePoolSpendTX(
    res1.tx,
    res1.amount,
    f.endHeight,
    clientPriv,
    serverPriv.toPublicKey(),
    f.feeRate,
  );

  // sign with both priv keys
  const multisigScript = base.createDualMultisigScript([
    serverPriv.toPublicKey(),
    clientPriv.toPublicKey(),
  ]);

  // prepare source output
  if (!res2.tx.inputs[0].sourceTransaction) {
    res2.tx.inputs[0].sourceTransaction = new Transaction();
    res2.tx.inputs[0].sourceTransaction!.outputs = [];
  }
  res2.tx.inputs[0].sourceTransaction!.outputs[0] = {
    satoshis: res1.amount,
    lockingScript: multisigScript as any,
  } as any;

  const scope = TransactionSignature.SIGHASH_ALL | TransactionSignature.SIGHASH_FORKID;
  const sighashData = TransactionSignature.format({
    sourceTXID: res2.tx.inputs[0].sourceTXID || '',
    sourceOutputIndex: res2.tx.inputs[0].sourceOutputIndex,
    sourceSatoshis: res1.amount,
    transactionVersion: res2.tx.version,
    otherInputs: [],
    outputs: res2.tx.outputs,
    inputIndex: 0,
    subscript: multisigScript,
    inputSequence: res2.tx.inputs[0].sequence || 1,
    lockTime: res2.tx.lockTime,
    scope,
  });

  console.log('TS sighash:', Buffer.from(sighashData).toString('hex'));

  const msgHash = hash256(sighashData);
  console.log('TS sighash32:', msgHash.toString('hex'));
  const bnHash = new BigNumber(Buffer.from(msgHash).toString('hex'), 16);
  const sigServerObj = ECDSA.sign(bnHash, serverPriv, true);
  const sigServer = sigServerObj.toDER() as number[];
  sigServer.push(scope);
  const sigClientObj = ECDSA.sign(bnHash, clientPriv, true);
  const sigClient = sigClientObj.toDER() as number[];
  sigClient.push(scope);

  const unlockScript = createUnlockScript(sigServer, sigClient);
  res2.tx.inputs[0].unlockingScript!.chunks = unlockScript.chunks;

  console.log('Step1Hex:', res1.tx.toHex());
  console.log('Step2Hex:', res2.tx.toHex());
})();
