// Copyright (c) 2024 Fantom Foundation
//
// Use of this software is governed by the Business Source License included
// in the LICENSE file and at fantom.foundation/bsl11.
//
// Change Date: 2028-4-16
//
// On the date above, in accordance with the Business Source License, use of
// this software will be governed by the GNU Lesser General Public License v3.

package lfvm

import "github.com/Fantom-foundation/Tosca/go/tosca"

const (
	errOverflow               = tosca.ConstError("overflow")
	errInvalidOpCode          = tosca.ConstError("invalid op-code")
	errInvalidRevision        = tosca.ConstError("invalid revision")
	errInvalidJump            = tosca.ConstError("invalid jump destination")
	errOutOfGas               = tosca.ConstError("out of gas")
	errStaticContextViolation = tosca.ConstError("static context violation")
	errNotEnoughStaticGas     = tosca.ConstError("not enough static gas")
	errStackLimitsViolation   = tosca.ConstError("stack limits violation")
	errCodeLimitsViolation    = tosca.ConstError("code bounds violation")
	errInitCodeTooLarge       = tosca.ConstError("init code larger than allowed")
)
