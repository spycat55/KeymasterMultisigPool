import fs from 'fs';
import path from 'path';
import { PrivateKey } from '@bsv/sdk/primitives';
import { buildDualFeePoolBaseTx, buildDualFeePoolSpendTX } from '../../src/dual_endpoint';

interface UTXO { txid: string; vout: number; satoshis: number; }
interface Fixture {
  clientPrivHex: string;
  serverPrivHex: string;
  clientUtxos: UTXO[];
  endHeight: number;
  feeRate: number;
  isMain: boolean;
}

const fixture: Fixture = JSON.parse(
  fs.readFileSync(path.resolve(__dirname, './fixture.json'), 'utf-8'),
);

describe('DualEndpoint cross-lang spec', () => {
  it('builds and spends with deterministic txids', async () => {
    const clientPriv = PrivateKey.fromHex(fixture.clientPrivHex);
    const serverPriv = PrivateKey.fromHex(fixture.serverPrivHex);

    // Step-1
    const res1 = await buildDualFeePoolBaseTx(
      fixture.clientUtxos,
      clientPriv,
      serverPriv.toPublicKey(),
      fixture.feeRate,
    );

    // Step-2
    const res2 = await buildDualFeePoolSpendTX(
      res1.tx,
      res1.amount,
      fixture.endHeight,
      clientPriv,
      serverPriv.toPublicKey(),
      fixture.feeRate,
    );

    expect(res1.tx.id('hex')).toMatch(/[0-9a-f]{64}/);
    expect(res2.tx.id('hex')).toMatch(/[0-9a-f]{64}/);
  });
});
