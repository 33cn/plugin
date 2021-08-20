// SPDX-License-Identifier: GPL-3.0-or-later

pragma solidity >=0.6.0;

import '../utils/TransferHelper.sol';

// test helper for transfers
contract TransferHelperTest {
    function safeApprove(
        address token,
        address to,
        uint256 value
    ) external {
        TransferHelper.safeApprove(token, to, value);
    }

    function safeTransfer(
        address token,
        address to,
        uint256 value
    ) external {
        TransferHelper.safeTransfer(token, to, value);
    }

    function safeTransferFrom(
        address token,
        address from,
        address to,
        uint256 value
    ) external {
        TransferHelper.safeTransferFrom(token, from, to, value);
    }

    function safeTransferBNB(address to, uint256 value) external {
        TransferHelper.safeTransferBNB(to, value);
    }
}

// can revert on failure and returns true if successful
contract TransferHelperTestFakeBEP20Compliant {
    bool public success;
    bool public shouldRevert;

    function setup(bool success_, bool shouldRevert_) public {
        success = success_;
        shouldRevert = shouldRevert_;
    }

    function transfer(address, uint256) public view returns (bool) {
        require(!shouldRevert, 'REVERT');
        return success;
    }

    function transferFrom(
        address,
        address,
        uint256
    ) public view returns (bool) {
        require(!shouldRevert, 'REVERT');
        return success;
    }

    function approve(address, uint256) public view returns (bool) {
        require(!shouldRevert, 'REVERT');
        return success;
    }
}

// only reverts on failure, no return value
contract TransferHelperTestFakeBEP20Noncompliant {
    bool public shouldRevert;

    function setup(bool shouldRevert_) public {
        shouldRevert = shouldRevert_;
    }

    function transfer(address, uint256) public view {
        require(!shouldRevert);
    }

    function transferFrom(
        address,
        address,
        uint256
    ) public view {
        require(!shouldRevert);
    }

    function approve(address, uint256) public view {
        require(!shouldRevert);
    }
}

contract TransferHelperTestFakeFallback {
    bool public shouldRevert;

    function setup(bool shouldRevert_) public {
        shouldRevert = shouldRevert_;
    }

    receive() external payable {
        require(!shouldRevert);
    }
}
