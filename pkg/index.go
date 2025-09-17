// Package pkg provides multisig functionality for Bitcoin SV transactions
package pkg

// Re-export commonly used types and functions from subpackages
import (
	dual "github.com/spycat55/KeymasterMultisigPool/pkg/dual_endpoint"
	"github.com/spycat55/KeymasterMultisigPool/pkg/libs"
	triple "github.com/spycat55/KeymasterMultisigPool/pkg/triple_endpoint"
)

// Version of the KeymasterMultisigPool library
const Version = "1.4.0"

// Re-export multisig types and functions
type MultiSig = libs.MultiSig
type UTXO = libs.UTXO

var (
	// Multisig script creation
	Lock   = libs.Lock
	Unlock = libs.Unlock

	// Utility functions
	GetAddressFromPublicKey = libs.GetAddressFromPublicKey
	GetAddressFromPubKey    = libs.GetAddressFromPubKey

	// Dual endpoint functions
	DualPoolSpentScript        = dual.DualPoolSpentScript
	MergeDualPoolSigForSpendTx = dual.MergeDualPoolSigForSpendTx
	// Dual endpoint verify helpers
	ServerVerifyClientSpendSig  = dual.ServerVerifyClientSpendSig
	ClientVerifyServerSpendSig  = dual.ClientVerifyServerSpendSig
	ServerVerifyClientUpdateSig = dual.ServerVerifyClientUpdateSig
	ClientVerifyServerUpdateSig = dual.ClientVerifyServerUpdateSig

	// Triple endpoint functions
	TripleFeePoolSpentScript        = triple.TripleFeePoolSpentScript
	MergeTripleFeePoolSigForSpendTx = triple.MergeTripleFeePoolSigForSpendTx
	VerifySignature                 = triple.VerifySignature
	// Triple endpoint verify helpers
	ServerVerifyClientASig = triple.ServerVerifyClientASig
	ServerVerifyClientBSig = triple.ServerVerifyClientBSig
	ClientVerifyServerSig  = triple.ClientVerifyServerSig
)

// Common errors
var (
	ErrInvalidPublicKeys = libs.ErrInvalidPublicKeys
	ErrNoPrivateKeys     = libs.ErrNoPrivateKeys
	ErrInvalidM          = libs.ErrInvalidM
)
