// SPDX-License-Identifier: GPL-3.0-or-later

pragma solidity >=0.4.0;

import '../math/SafeMath.sol';

// used for testing the logic of token naming
contract SafeMathTest {
    function sqrt(uint256 num) public pure returns (uint256) {
        return SafeMath.sqrt(num);
    }
}
