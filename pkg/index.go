// Package pkg provides multisig functionality for Bitcoin SV transactions
package pkg

// Re-export commonly used types and functions from subpackages
import (
	"github.com/spycat55/KeymasterMultisigPool/pkg/libs"
	dual "github.com/spycat55/KeymasterMultisigPool/pkg/dual_endpoint"
	triple "github.com/spycat55/KeymasterMultisigPool/pkg/triple_endpoint"
)

// Version of the KeymasterMultisigPool library
const Version = "0.1.0"

// Re-export multisig types and functions
type MultiSig = libs.MultiSig

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
	
	// Triple endpoint functions
	TripleFeePoolSpentScript        = triple.TripleFeePoolSpentScript
	MergeTripleFeePoolSigForSpendTx = triple.MergeTripleFeePoolSigForSpendTx
	VerifySignature                 = triple.VerifySignature
)

// Common errors
var (
	ErrInvalidPublicKeys = libs.ErrInvalidPublicKeys
	ErrNoPrivateKeys     = libs.ErrNoPrivateKeys
	ErrInvalidM          = libs.ErrInvalidM
)
